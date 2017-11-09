package main

import (
	"time"

	"encoding/json"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/boj/redistore"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

var (
	rdPool  *redis.Pool
	rdStore *redistore.RediStore
)

func createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     100,
		MaxActive:   3000, // max number of connections
		Wait:        true,
		IdleTimeout: 5 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", conf.Redis.AddrWithPort)
			if err != nil {
				panic(err)
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < 10*time.Second {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func CreateRedisStore(pool *redis.Pool) *redistore.RediStore {
	store, err := redistore.NewRediStoreWithPool(pool, conf.Redis.KeyPair)
	if err != nil {
		panic(err)
	}
	return store
}

var (
	prepaidProd     *mgo.Session
	prepaidMerchant *mgo.Session

	mgoPool mongoPool
)

var RP mongoPool

func createMongoPool() {
	var err error
	prepaidMerchant, err = mgo.DialWithInfo(conf.MongoDB.DialInfo.Merchant)
	if err != nil {
		panic(err)
	}
	prepaidMerchant.SetMode(mgo.Strong, true)

	mi := []dbIndex{
		//{
		//	Collection: "notifications_map",
		//	Index: mgo.Index{
		//		Key:         []string{"created_at"},
		//		Unique:      false,
		//		DropDups:    false,
		//		Background:  true,
		//		ExpireAfter: time.Hour * time.Duration(24 * 90),
		//	},
		//},
		{
			Collection: "deals", // all deals VIP
			Index: mgo.Index{
				Key:        []string{"deal_class", "status", "deal_type", "expire_date", "deal_available"},
				Unique:     false,
				DropDups:   false,
				Background: true,
			},
		},
		{
			Collection: "deals", // all deal normals
			Index: mgo.Index{
				Key:        []string{"status", "expire_date", "deal_type", "deal_available", "rank_normal"},
				Unique:     false,
				DropDups:   false,
				Background: true,
			},
		},
		{
			Collection: "pruanfun_mobile", // all deal normals, pruanfun mobile query
			Index: mgo.Index{
				Key:        []string{"mobile_no"},
				Unique:     true,
				DropDups:   false,
				Background: true,
			},
		},
		{
			Collection: "deal_payment",
			Index: mgo.Index{
				Key:        []string{"request_header.request_id", "request_header.mobile_number"},
				Unique:     false,
				DropDups:   false,
				Background: true,
			},
		},
		{
			Collection: "deal_payment",
			Index: mgo.Index{
				Key:        []string{"request_body.ref_no", "request_header.mobile_number"},
				Unique:     false,
				DropDups:   false,
				Background: true,
			},
		},
		{
			Collection: "lifestyle_payment",
			Index: mgo.Index{
				Key:        []string{"invoice_no", "mobile_no"},
				Unique:     false,
				DropDups:   false,
				Background: true,
			},
		},
		{
			Collection: "my_deals",
			Index: mgo.Index{
				Key:        []string{"mobile_no"},
				Unique:     false,
				DropDups:   false,
				Sparse:     true,
				Background: true,
			},
		},
		{
			Collection: "deals",
			Index: mgo.Index{
				Key:        []string{"merchant.mid"},
				Unique:     false,
				DropDups:   false,
				Background: true,
			},
		},
		{
			Collection: "retail_transactions",
			Index: mgo.Index{
				Key:        []string{"request_header.mobile_number"},
				Unique:     false,
				DropDups:   false,
				Background: true,
			},
		},
		{
			Collection: "favorite_merchant",
			Index: mgo.Index{
				Key:        []string{"tel_no"},
				Unique:     true,
				DropDups:   false,
				Background: true,
			},
		},
	}

	for _, i := range mi {
		if err := prepaidMerchant.DB("PrepaidMerchant").C(i.Collection).EnsureIndex(i.Index); err != nil {
			panic(err)
		}
	}

	prepaidProd, err = mgo.DialWithInfo(conf.MongoDB.DialInfo.Customer)
	if err != nil {
		panic(err)
	}
	prepaidProd.SetMode(mgo.Monotonic, true)

	ci := []dbIndex{
	//{
	//	Collection: "notifications",
	//	Index: mgo.Index{
	//		Key:         []string{"created_at"},
	//		Unique:      false,
	//		DropDups:    false,
	//		Background:  true,
	//		ExpireAfter: (time.Hour * time.Duration(24 * 90)) + time.Duration(10),
	//	},
	//},
	}

	for _, i := range ci {
		if err := prepaidProd.DB("PrepaidProd").C(i.Collection).EnsureIndex(i.Index); err != nil {
			panic(err)
		}
	}
}

type mongoPool struct {
	MS *mgo.Database
	CS *mgo.Database

	ms *mgo.Session
	cs *mgo.Session
}

func (p *mongoPool) RemoveFavoriteMerchant(tel_no, mid string) error {
	mgoPool.get()
	defer mgoPool.close()
	c := mgoPool.MS.C("favorite_merchant")
	err := c.Update(bson.M{"tel_no": tel_no}, bson.M{"$pull": bson.M{"merchants": bson.M{"mid": mid}}})
	if err != nil {
		return err
	}
	return nil
}

func (p *mongoPool) DoFavorite(favRequest FavRequest) bool {
	mgoPool.get()
	defer mgoPool.close()

	m := mgoPool.MS.C("merchants")
	c := mgoPool.MS.C("favorite_merchant")

	merChants := Merchants{}

	//Get name image from merchants collections
	m.Find(bson.M{"mid": favRequest.MID}).One(&merChants)

	//Check null
	if merChants.MID == "" {
		log.Errorln("Merchant not found ", favRequest.MID)
		return false
	}

	merChants.DateCreate = time.Now()
	merChants.Weight = CalculateSortWeight(merChants.Name)

	//Check user is exist
	count, err := c.Find(bson.M{"tel_no": favRequest.MobileNo}).Count()
	if err != nil {
		log.Errorln("Search err : ", err.Error())
		return false
	}

	//Insert
	if count == 0 {
		favShop := FavoriteShops{favRequest.MobileNo, []Merchants{merChants}}
		err = c.Insert(&favShop)
		if err != nil {
			log.Errorln(err.Error())
			return false
		}
		return true
	}

	//Update
	err = c.Update(bson.M{"tel_no": favRequest.MobileNo}, bson.M{"$pull": bson.M{"merchants": bson.M{"mid": merChants.MID}}})
	err = c.Update(bson.M{"tel_no": favRequest.MobileNo}, bson.M{"$push": bson.M{"merchants": merChants}})
	if err != nil {
		log.Errorln("Add favorite merchant error : ", err.Error())
		return false
	}
	return true
}

func (p *mongoPool) get() {
	p.ms = prepaidMerchant.Clone()
	p.MS = p.ms.DB(conf.MongoDB.Schema.Merchant)

	p.cs = prepaidProd.Clone()
	p.CS = p.cs.DB(conf.MongoDB.Schema.Customer)
}

func (p mongoPool) close() {
	p.ms.Close()
	p.cs.Close()
}

func testMongo() map[string]interface{} {
	mgoPool.get()
	defer mgoPool.close()

	c := mgoPool.CS.C("grand_line")

	t := map[string]interface{}{}
	c.Find(bson.M{"hello": "world"}).One(&t)

	c.Upsert(bson.M{"hello": "world"}, bson.M{"$set": bson.M{"updated_at": time.Now()}})
	return t
}

func testRedis() {
	r1, err := rdPool.Get().Do("SET", "bui", "hello world")
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}
	fmt.Println("redis", r1)

	r2, err := redis.String(rdPool.Get().Do("GET", "bui"))
	if err != nil {
		fmt.Errorf("%s", err)
		return
	}
	fmt.Println("redis", r2)
}

func SetSessionDataRedis(w http.ResponseWriter, r *http.Request, name string, expireTime int, data map[string]interface{}) error {
	session, err := rdStore.Get(r, name)
	if err != nil || session.ID == "" {
		log.Warnf("redis session '%s' not found: create new", name)
		session, err = rdStore.New(r, name)
		if err != nil {
			log.Errorln("unable to create new redis session:", err)
		}
	}
	if err := session.Save(r, w); err != nil {
		log.Errorln("unable to save new redis session:", err)
		return err
	}

	for k, v := range data {
		session.Values[k] = v
	}
	session.Options.MaxAge = 60 * expireTime
	if err := session.Save(r, w); err != nil {
		log.Errorln("unable to save new redis session:", err)
		return err
	}

	log.Infoln("successfully create redis session:", name, session.ID)
	return nil
}

func setEmailPersistentSession(email, sessionID string) error {
	conn := rdPool.Get()
	defer conn.Close()

	key := fmt.Sprintf("%s_pmsession", email)
	sess, _ := redis.String(conn.Do("GET", key))
	if sess != "" {
		log.Infoln("remove old psession:", email, sess)
		if _, err := conn.Do("DEL", fmt.Sprintf("session_%s", sess)); err != nil {
			log.Errorf("unable to remove session %s from redis: %s", sess, err)
		}
	}

	conn.Do("SET", key, sessionID)
	return nil
}

func uniqSlices(sl []string) []string {
	rtn := []string{}
	seen := map[string]string{}
	for _, v := range sl {
		if _, ok := seen[v]; !ok {
			rtn = append(rtn, v)
			seen[v] = v
		}
	}
	return rtn
}

func DelSessionDataRedis(w http.ResponseWriter, r *http.Request, name string) error {
	session, err := rdStore.Get(r, name)
	if err != nil || session.ID == "" {
		log.Infoln("no data on redis to remove skipped:", name)
		return nil
	}

	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		log.Errorf("unable to save %s session", name)
		return err
	}
	return nil
}

func GetSessionDataRedis(r *http.Request, name string) map[interface{}]interface{} {
	session, err := rdStore.Get(r, name)
	if err != nil {
		log.Errorln("unable to get data from redis:", err)
		return nil
	}
	return session.Values
}

type CryptoMiddleware struct {
	handler http.Handler
}

func MyCryptoMiddleware(inner http.Handler) *CryptoMiddleware {
	return &CryptoMiddleware{inner}
}

func (c *CryptoMiddleware) DecryptServeHTTP(w http.ResponseWriter, r *http.Request) {

}

func GetDataOnRedis(key string) (map[string]interface{}, error) {
	conn := rdPool.Get()
	defer conn.Close()

	log.Errorln("key in redis :", key)
	val, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}

	rtn := make(map[string]interface{})
	if err := json.Unmarshal(val, &rtn); err != nil {
		return nil, err
	}
	return rtn, nil
}

func GetKeysOnRedis(key string) ([]string, error) {
	conn := rdPool.Get()
	defer conn.Close()

	log.Debugln("[GetKeysOnRedis] query :", key)

	//If keys not found will return nil in error
	val, err := redis.Strings(conn.Do("KEYS", key))
	if err != nil {
		return []string{}, err
	}

	if len(val) == 0 {
		return []string{}, errors.New("Keys not found")
	}

	return val, nil
}

func SaveDataOnRedis(key string, data map[string]interface{}, expire time.Duration) error {
	conn := rdPool.Get()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return err
	}

	ek := make(map[string]interface{})
	if exists {
		val, err := redis.Bytes(conn.Do("GET", key))
		if err != nil {
			return err
		}

		if err := json.Unmarshal(val, &ek); err != nil {
			return err
		}
	}

	for k, v := range data {
		ek[k] = v
	}

	jd, err := json.Marshal(ek)
	if err != nil {
		return err
	}

	t := expire * time.Minute
	if _, err := conn.Do("SET", key, jd, "EX", t.Seconds()); err != nil {
		return err
	}

	return nil
}

func RemoveDataOnRedis(key string) error {
	conn := rdPool.Get()
	defer conn.Close()

	if _, err := conn.Do("DEL", key); err != nil {
		return err
	}
	return nil
}

type dbIndex struct {
	Collection string
	Index      mgo.Index
}

func myDealRedisKey(mobileNo string) string {
	return fmt.Sprintf("mydeal:%s", mobileNo)
}
