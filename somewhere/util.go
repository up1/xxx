package somewhere

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	// "strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

func JSONResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Errorln("json response error:", err)
	}
}

func RandomEAIUserID(userType string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch userType {
	case "settlement":
		user := conf.EAI.PrepaidToEAIUserID["online_settlement_max"].(string)
		prefix := "%0" + fmt.Sprint(len(user[4:])) + "d"
		maxLen, err := strconv.Atoi(user[4:])
		log.Println("settlement MaxLen : ", maxLen)
		if err != nil {
			maxLen = 99
		}
		randLen := fmt.Sprintf(prefix, r.Intn(maxLen))
		return user[0:4] + randLen
	default:
		user := conf.EAI.PrepaidToEAIUserID["qr_payment_reverse_online_max"].(string)
		prefix := "%0" + fmt.Sprint(len(user[4:])) + "d"
		maxLen, err := strconv.Atoi(user[4:])
		log.Println("default MaxLen : ", maxLen)
		if err != nil {
			maxLen = 99
		}
		randLen := fmt.Sprintf(prefix, r.Intn(maxLen))
		return user[0:4] + randLen
	}
}

func consumeMessageFromElk(consumeUrl string, msg []byte) ([]byte, interface{}, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 5 * time.Second,
		}).Dial,
		MaxIdleConns:          100,
		IdleConnTimeout:       5 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(5 * time.Second),
	}

	reqStat, err := http.NewRequest("POST", consumeUrl, bytes.NewBuffer(msg))
	if err != nil {
		log.Errorln("http new request:", consumeUrl, err)
		return nil, nil, err
	}
	reqStat.Header.Set("Content-Type", "application/x-ndjson; charset=UTF-8")
	respStat, err := client.Do(reqStat)
	if err != nil {
		log.Errorln("Cannot Consume Message From ELK ", err)
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(respStat.Body)
	if err != nil {
		log.Errorln("ioutil read all error:", err)
		return nil, nil, err
	}
	defer respStat.Body.Close()

	var respInf interface{}
	if err := json.Unmarshal(body, &respInf); err != nil {
		log.Errorln("json unmarshal error:", err)
		return nil, nil, err
	}

	return body, respInf, nil
}

func RemoveDataFromElastic(url string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 5 * time.Second,
		}).Dial,
		MaxIdleConns:          100,
		IdleConnTimeout:       5 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(5 * time.Second),
	}

	reqStat, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Errorln("http new request:", url, err)
		return err
	}
	reqStat.Header.Set("Content-Type", "application/x-ndjson; charset=UTF-8")
	respStat, err := client.Do(reqStat)
	if err != nil {
		log.Errorln("Cannot Consume Message From ELK ", err)
		return err
	}

	if _, err = ioutil.ReadAll(respStat.Body); err != nil {
		log.Errorln("ioutil read all error:", err)
		return err
	}
	defer respStat.Body.Close()
	return nil
}

func RandomNumbertoString() string {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	date := time.Now().Format("20060102")
	time := time.Now().Format("030405")
	return date + time + fmt.Sprint(r1.Int31())
}

func GenerateAppIdWithSeqNumberForEAI() string {

	var resultGenerate bytes.Buffer

	appNumber := conf.PrepaidAppNumberString
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	date := time.Now().Format("20060102")
	time := time.Now().Format("030405")
	resultGenerate.WriteString(appNumber)
	resultGenerate.WriteString("_")
	resultGenerate.WriteString(date)
	resultGenerate.WriteString("_")
	resultGenerate.WriteString(time)
	resultGenerate.WriteString("_")
	resultGenerate.WriteString(fmt.Sprint(r1.Int31()))

	return resultGenerate.String()
}

// func GenerateRetailNoWithDate() string {
// 	date := time.Now().Format("20060102")
// 	prepaidTransactionRunning := GetPaymentRunningSeqNew(date)
// 	running_string := fmt.Sprint(prepaidTransactionRunning.Running_no)
// 	return date + strings.Repeat("0", 7-len(running_string)) + running_string
// }

func GenerateRequestNoToKplus() string {
	tn := time.Now()
	tn.Nanosecond()
	rand.Seed(tn.UnixNano())
	sInv := fmt.Sprintf("733_LF%s%06d", tn.Format("20060102"), tn.Nanosecond())
	return sInv
}

func GenerateLifeStyleInvoice(invoiceType string) string {
	prefixInv := "LIF"
	if invoiceType == "paynow" {
		prefixInv = "PAY"
	}

	tn := time.Now()
	tn.Nanosecond()
	rand.Seed(tn.UnixNano())
	sInv := fmt.Sprintf("%s-%s%06d", prefixInv, tn.Format("20060102"), tn.Nanosecond())
	return sInv
}
