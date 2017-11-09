package somewhere

import (

	. "lifestyle/fav"

	"math"
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

func CreateRedisPool(conf *Configs) *redis.Pool {
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

func CreateRedisStore(pool *redis.Pool, conf *Configs) *redistore.RediStore {
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
var conf Configs

func CreateMongoPool(conf *Configs) {
	conf = conf
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

func CalculateSortWeight(text string) string {
	textTest := ""
	var vowelValue float64
	characterCounter := 0.0
	vowelCounter := 0.0
	for _, v := range text {
		if Alphabet[string(v)] == 0 {
			continue
		}

		if Alphabet[string(v)] >= 100 {
			alphabetValue := Alphabet[string(v)]
			vowelCounter = 0 //If found character reset vowelCounter
			if characterCounter == 0.0 {
				textTest = fmt.Sprint(alphabetValue+vowelValue, ".")
			} else {
				textTest += fmt.Sprint(alphabetValue + vowelValue)
			}
			vowelValue = 0
			characterCounter++
		} else {
			temp := int(vowelValue + (Alphabet[string(v)] / (math.Pow(10, vowelCounter))))
			vowelValue = float64(temp)
			vowelCounter++
		}
	}
	return textTest
}

func InitData() map[string]float64 {
	Alphabet["เ"] = 5
	Alphabet["แ"] = 6
	Alphabet["โ"] = 7
	Alphabet["ใ"] = 8
	Alphabet["ไ"] = 9
	Alphabet["ะ"] = 10
	Alphabet["ั"] = 11
	Alphabet["า"] = 12
	Alphabet["ำ"] = 13
	Alphabet["ิ"] = 14
	Alphabet["ี"] = 15
	Alphabet["ึ"] = 16
	Alphabet["ื"] = 17
	Alphabet["ุ"] = 18
	Alphabet["ู"] = 19

	Alphabet["็"] = 20
	Alphabet["์"] = 21
	Alphabet["่"] = 22
	Alphabet["้"] = 23
	Alphabet["๊"] = 24
	Alphabet["๋"] = 25

	Alphabet["ก"] = 1000
	Alphabet["ข"] = 1030
	Alphabet["ฃ"] = 1060
	Alphabet["ค"] = 1090
	Alphabet["ฅ"] = 1120
	Alphabet["ฆ"] = 1150
	Alphabet["ง"] = 1180
	Alphabet["จ"] = 1210
	Alphabet["ฉ"] = 1240
	Alphabet["ช"] = 1270
	Alphabet["ซ"] = 1300
	Alphabet["ฌ"] = 1330
	Alphabet["ญ"] = 1360
	Alphabet["ฎ"] = 1390
	Alphabet["ฏ"] = 1420
	Alphabet["ฐ"] = 1450
	Alphabet["ฑ"] = 1480
	Alphabet["ฒ"] = 1510
	Alphabet["ณ"] = 1540
	Alphabet["ด"] = 1570
	Alphabet["ต"] = 1600
	Alphabet["ถ"] = 1630
	Alphabet["ท"] = 1660
	Alphabet["ธ"] = 1690
	Alphabet["น"] = 1720
	Alphabet["บ"] = 1750
	Alphabet["ป"] = 1780
	Alphabet["ผ"] = 1810
	Alphabet["ฝ"] = 1840
	Alphabet["พ"] = 1870
	Alphabet["ฟ"] = 1900
	Alphabet["ภ"] = 1930
	Alphabet["ม"] = 1960
	Alphabet["ย"] = 1990
	Alphabet["ร"] = 2020
	Alphabet["ฤ"] = 2050
	Alphabet["ล"] = 2080
	Alphabet["ฦ"] = 2110
	Alphabet["ว"] = 2140
	Alphabet["ศ"] = 2170
	Alphabet["ษ"] = 2200
	Alphabet["ส"] = 2230
	Alphabet["ห"] = 2260
	Alphabet["ฬ"] = 2290
	Alphabet["อ"] = 2320
	Alphabet["ฮ"] = 2350

	Alphabet["a"] = 2380
	Alphabet["A"] = 2410
	Alphabet["b"] = 2440
	Alphabet["B"] = 2470
	Alphabet["c"] = 2500
	Alphabet["C"] = 2530
	Alphabet["d"] = 2560
	Alphabet["D"] = 2590
	Alphabet["e"] = 2620
	Alphabet["E"] = 2650
	Alphabet["f"] = 2680
	Alphabet["F"] = 2710
	Alphabet["g"] = 2740
	Alphabet["G"] = 2770
	Alphabet["h"] = 2800
	Alphabet["H"] = 2830
	Alphabet["i"] = 2860
	Alphabet["I"] = 2890
	Alphabet["j"] = 2920
	Alphabet["J"] = 2950
	Alphabet["k"] = 2980
	Alphabet["K"] = 3010
	Alphabet["l"] = 3040
	Alphabet["L"] = 3070
	Alphabet["m"] = 3100
	Alphabet["M"] = 3130
	Alphabet["n"] = 3160
	Alphabet["N"] = 3190
	Alphabet["o"] = 3220
	Alphabet["O"] = 3250
	Alphabet["p"] = 3280
	Alphabet["P"] = 3310
	Alphabet["q"] = 3340
	Alphabet["Q"] = 3370
	Alphabet["r"] = 3400
	Alphabet["R"] = 3430
	Alphabet["s"] = 3460
	Alphabet["S"] = 3490
	Alphabet["t"] = 3520
	Alphabet["T"] = 3550
	Alphabet["u"] = 3580
	Alphabet["U"] = 3610
	Alphabet["v"] = 3640
	Alphabet["V"] = 3670
	Alphabet["w"] = 3700
	Alphabet["W"] = 3730
	Alphabet["x"] = 3760
	Alphabet["X"] = 3790
	Alphabet["y"] = 3820
	Alphabet["Y"] = 3850
	Alphabet["z"] = 3880
	Alphabet["Z"] = 3910

	Alphabet["๐"] = 3940
	Alphabet["๑"] = 3970
	Alphabet["๒"] = 4000
	Alphabet["๓"] = 4030
	Alphabet["๔"] = 4060
	Alphabet["๕"] = 4090
	Alphabet["๖"] = 4120
	Alphabet["๗"] = 4150
	Alphabet["๘"] = 4180
	Alphabet["๙"] = 4210

	Alphabet["0"] = 4240
	Alphabet["1"] = 4270
	Alphabet["2"] = 4300
	Alphabet["3"] = 4330
	Alphabet["4"] = 4360
	Alphabet["5"] = 4390
	Alphabet["6"] = 4420
	Alphabet["7"] = 4450
	Alphabet["8"] = 4480
	Alphabet["9"] = 4510

	Alphabet[""] = 4540
	Alphabet["!"] = 4570
	Alphabet["\""] = 4600
	Alphabet["#"] = 4630
	Alphabet["$"] = 4660
	Alphabet["%"] = 4690
	Alphabet["&"] = 4720
	Alphabet["'"] = 4750
	Alphabet["("] = 4780
	Alphabet[")"] = 4810
	Alphabet["*"] = 4840
	Alphabet["+"] = 4870
	Alphabet[","] = 4900
	Alphabet["-"] = 4930
	Alphabet["."] = 4960
	Alphabet["/"] = 4990

	Alphabet[":"] = 5020
	Alphabet[";"] = 5050
	Alphabet["<"] = 5080
	Alphabet["="] = 5110
	Alphabet[">"] = 5140
	Alphabet["?"] = 5170
	Alphabet["@"] = 5200

	Alphabet["["] = 5230
	Alphabet["\\"] = 5260
	Alphabet["]"] = 5290
	Alphabet["^"] = 5320
	Alphabet["_"] = 5350
	Alphabet["`"] = 5380

	Alphabet["{"] = 5410
	Alphabet["|"] = 5440
	Alphabet["}"] = 5470
	Alphabet["~"] = 5500

	return Alphabet
}
