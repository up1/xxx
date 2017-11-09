package main

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/mgo.v2"
	"kbtg.tech/prepaid/go-eai"
	"kbtg.tech/prepaid/go-gtm"
	"strings"
)

type Stage string

const (
	StageLocal  Stage = "localhost"
	StageDEV    Stage = "development"
	StageSIT    Stage = "sit"
	StageOLDSIT Stage = "oldsit"
	StageProd   Stage = "production"
)

func ParseStage(s string) Stage {
	switch s {
	case "local", "localhost", "l":
		return StageLocal
	case "dev", "develop", "development", "d":
		return StageDEV
	case "sit", "staging", "s":
		return StageSIT
	case "oldsit":
		return StageOLDSIT
	case "prod", "production", "p":
		return StageProd
	}
	return StageLocal
}

var conf Configs

type Configs struct {
	Stage                  Stage
	ProjectCode            string
	PrepaidSavingAccount   string
	PrepaidAppNumberString string
	DonateMerchant       []string
	MongoDB                struct {
		Addr         string
		Port         int
		AddrWithPort string
		Timeout      time.Duration
		Username     string
		Password     string
		Schema       struct {
			Customer string
			Merchant string
		}
		DialInfo struct {
			Customer *mgo.DialInfo
			Merchant *mgo.DialInfo
		}
	}
	Redis struct {
		Addr         string
		Port         int
		AddrWithPort string
		Password     string
		KeyPair      []byte
	}
	EAI struct {
		EAIConfig        *eai.Config
		Username         string
		Password         string
		Timeout          time.Duration
		PrepaidAppID     string
		PrepaidAppUserID string
		TerminalID       string
		UserLangPref     string
		AuthUserID       string
		AuthLevel        string
		AutoAdjustUser   string
		SOAPEnvAttribute string
		CBS1189I01       string
		CBS1186I01       string
		CBS1182F01       string
		CBS1601F01       string
		CBS1602F01       string
		CIS0367I01       string
		SVCBranchID      string
		PrepaidAuthLevel string
		DoService        struct {
			URL        string
			CBS1189I01 string
			CBS1186I01 string
			CBS1182F01 string
			CIS0367I01 string
			CBS1601F01 string
			CBS1602F01 string
		}
		JWS struct {
			CBS1189I01 string
			CBS1186I01 string
			CBS1182F01 string
			CIS0367I01 string
			CBS1601F01 string
			CBS1602F01 string
		}
		Descriptions struct {
			Transfer       string
			Payment        string
			ReversePayment string
			TopUp          string
		}
		PrepaidToEAIUserID map[string]interface{}
	}
	Toggle struct {
		Inquiry bool
	}
	GTM struct {
		Config               *gtm.Config
		ProductID            string
		AllowTransferoutCASA string
	}
	Kafka struct {
		Addr         string
		Port         int
		AddrWithPort string
		Timeout      time.Duration
	}
	ELK struct {
		URL        string
		ShopIndex  string
		ShopType   string
		DealIndex  string
		DealType   string
		ScrollTime int
	}
	RecommendZone map[string]interface{}
	Retail        struct {
		URLPFPath string
	}
	KPLUS   map[string]interface{}
	Session struct {
		LifeStyleTimeout int
	}
	Push struct {
		NotiURL string
		FeedURL string
		Timeout time.Duration
	}
}

func (c *Configs) InitViper() {
	path := "."
	name := "dev"
	switch c.Stage {
	case StageProd:
		name = "prod"
	case StageSIT:
		name = "sit"
	case StageOLDSIT:
		name = "oldsit"
	case StageDEV:
	case StageLocal:
	}

	viper.SetConfigName("config." + name)
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		log.Errorln("unable to read config file:", err)
	}

	c.binding()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infoln("config file changed:", e.Name)
		c.binding()
	})
}

func (c *Configs) binding() {
	c.ProjectCode = viper.GetString("project_code")
	c.PrepaidSavingAccount = viper.GetString("prepaid_saving_account")
	c.PrepaidAppNumberString = viper.GetString("prepaid_app_number_string")
	c.MongoDB.Addr = viper.GetString("mongodb.addr")
	c.MongoDB.Port = viper.GetInt("mongodb.port")
	c.MongoDB.AddrWithPort = fmt.Sprintf("%s:%d", c.MongoDB.Addr, c.MongoDB.Port)
	c.MongoDB.Timeout = viper.GetDuration("mongodb.timeout") * time.Second
	c.MongoDB.Username = viper.GetString("mongodb.username")
	c.MongoDB.Password = viper.GetString("mongodb.password")
	c.MongoDB.Schema.Customer = viper.GetString("mongodb.schema.customer")
	c.MongoDB.Schema.Merchant = viper.GetString("mongodb.schema.merchant")
	c.MongoDB.DialInfo.Customer = &mgo.DialInfo{
		Addrs:    []string{c.MongoDB.AddrWithPort},
		Timeout:  c.MongoDB.Timeout,
		Database: c.MongoDB.Schema.Customer,
		Username: c.MongoDB.Username,
		Password: c.MongoDB.Password,
	}
	c.MongoDB.DialInfo.Merchant = &mgo.DialInfo{
		Addrs:    []string{c.MongoDB.AddrWithPort},
		Timeout:  c.MongoDB.Timeout,
		Database: c.MongoDB.Schema.Merchant,
		Username: c.MongoDB.Username,
		Password: c.MongoDB.Password,
	}

	c.Redis.Addr = viper.GetString("redis.addr")
	c.Redis.Port = viper.GetInt("redis.port")
	c.Redis.AddrWithPort = fmt.Sprintf("%s:%d", c.Redis.Addr, c.Redis.Port)
	c.Redis.Password = viper.GetString("redis.password")
	c.Redis.KeyPair = []byte(viper.GetString("redis.key_pair"))

	c.EAI.Username = viper.GetString("eai.username")
	c.EAI.Password = viper.GetString("eai.password")

	c.EAI.PrepaidAppID = viper.GetString("eai.prepaid_app_id")
	c.EAI.PrepaidAppUserID = viper.GetString("eai.prepaid_app_user_id")
	c.EAI.Timeout = viper.GetDuration("eai.timeout") * time.Second
	c.EAI.TerminalID = viper.GetString("eai.terminal_id")
	c.EAI.UserLangPref = viper.GetString("eai.user_lang_pref")
	c.EAI.AuthUserID = viper.GetString("eai.auth_user_id")
	c.EAI.AuthLevel = viper.GetString("eai.auth_level")
	c.EAI.AutoAdjustUser = viper.GetString("eai.auto_adjust_user")
	c.EAI.SOAPEnvAttribute = viper.GetString("eai.soap_env_attribute")
	c.EAI.CBS1186I01 = "CBS1186I01"
	c.EAI.CBS1601F01 = "CBS1601F01"
	c.EAI.SVCBranchID = viper.GetString("eai.svc_branch_id")
	c.EAI.DoService.URL = viper.GetString("eai.do_service_url")
	c.EAI.DoService.CBS1182F01 = fmt.Sprintf("%s%s/v2", c.EAI.DoService.URL, c.EAI.CBS1182F01)
	c.EAI.DoService.CBS1186I01 = fmt.Sprintf("%s%s/v2", c.EAI.DoService.URL, c.EAI.CBS1186I01)
	c.EAI.DoService.CBS1189I01 = fmt.Sprintf("%s%s/v2", c.EAI.DoService.URL, c.EAI.CBS1189I01)
	c.EAI.DoService.CIS0367I01 = fmt.Sprintf("%s%s/v2", c.EAI.DoService.URL, c.EAI.CIS0367I01)
	c.EAI.DoService.CBS1601F01 = fmt.Sprintf("%s%s/v2", c.EAI.DoService.URL, c.EAI.CBS1601F01)
	c.EAI.DoService.CBS1602F01 = fmt.Sprintf("%s%s/v2", c.EAI.DoService.URL, c.EAI.CBS1602F01)
	c.EAI.JWS.CBS1182F01 = viper.GetString("eai.jws_1182F01")
	c.EAI.JWS.CBS1186I01 = viper.GetString("eai.jws_1186I01")
	c.EAI.JWS.CBS1189I01 = viper.GetString("eai.jws_1189I01")
	c.EAI.JWS.CIS0367I01 = viper.GetString("eai.jws_0367I01")
	c.EAI.JWS.CBS1601F01 = viper.GetString("eai.jws_1601F01")
	c.EAI.JWS.CBS1602F01 = viper.GetString("eai.jws_1602F01")
	c.EAI.PrepaidAuthLevel = viper.GetString("prepaid_auth_level")
	c.EAI.Descriptions.Transfer = viper.GetString("eai.description.transfer")
	c.EAI.Descriptions.Payment = viper.GetString("eai.description.payment")
	c.EAI.Descriptions.ReversePayment = viper.GetString("eai.description.reverse_payment")
	c.EAI.Descriptions.TopUp = viper.GetString("eai.description.topup")
	c.EAI.PrepaidToEAIUserID = viper.GetStringMap("eai.prepaid_to_eai_user_id")
	c.ELK.URL = viper.GetString("elk.url")
	c.ELK.ShopIndex = viper.GetString("elk.shop_index")
	c.ELK.ShopType = viper.GetString("elk.shop_type")
	c.ELK.DealIndex = viper.GetString("elk.deal_index")
	c.ELK.DealType = viper.GetString("elk.deal_type")
	c.ELK.ScrollTime = viper.GetInt("elk.scroll_time")
	c.EAI.EAIConfig = &eai.Config{
		Username: "pprdlrkppapp01",
		Password: "password",
		URL: eai.URL{
			SOAP1189I01: c.EAI.JWS.CBS1189I01,
			SOAP1182F01: c.EAI.JWS.CBS1182F01,
			SOAP1186I01: c.EAI.JWS.CBS1186I01,
			SOAP0367I01: c.EAI.JWS.CIS0367I01,
			SOAP1601F01: c.EAI.JWS.CBS1601F01,
			SOAP1602F01: c.EAI.JWS.CBS1602F01,
		},
	}

	c.Toggle.Inquiry = viper.GetBool("toggle.inquiry")

	c.GTM.Config = &gtm.Config{
		ProjectCode:    c.ProjectCode,
		MPayCustomerID: "N/A",
		TellerID:       "go-lifestyle",
		KafkaAddress:   c.Kafka.Addr,
		KafkaPort:      c.Kafka.Port,
		KafkaTimeout:   c.Kafka.Timeout,
	}

	c.Kafka.Addr = viper.GetString("kafka.addr")
	c.Kafka.Port = viper.GetInt("kafka.port")
	c.Kafka.AddrWithPort = fmt.Sprintf("%s:%d", c.Kafka.Addr, c.Kafka.Port)
	c.Kafka.Timeout = viper.GetDuration("kafka.timeout") * time.Second

	c.RecommendZone = viper.GetStringMap("recommend_zone")

	c.Retail.URLPFPath = viper.GetString("retail.url_pf_image_path")
	c.KPLUS = viper.GetStringMap("kplus")
	c.Session.LifeStyleTimeout = viper.GetInt("session.lifestyle_timeout")

	c.Push.NotiURL = viper.GetString("push.notiurl")
	c.Push.FeedURL = viper.GetString("push.feedurl")
	c.Push.Timeout = viper.GetDuration("push.timeout") * time.Millisecond

	c.DonateMerchant = strings.Split(viper.GetString("donate_merchant"), ",")
}
