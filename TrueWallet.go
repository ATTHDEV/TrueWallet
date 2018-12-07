// author ATTH
package TrueWallet

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/resty.v1"
)

type Data struct {
	Data string `json:"data"`
}

const (
	host             = `https://mobile-api-gateway.truemoney.com/mobile-api-gateway/api/v1/`
	enpointSignin    = `signin/`
	enpointProfile   = `profile/`
	enpointTopup     = `topup/mobile/`
	enpointGettran   = `profile/transactions/history/`
	enpointChecktran = `profile/activities/`
)

var headers = map[string]string{
	"Host":         "mobile-api-gateway.truemoney.com",
	"Content-Type": "application/json",
	"User-Agent":   "okhttp/3.8.0",
}

type TrueWallet struct {
	username     string
	password     string
	passwordHash string
	loginType    string
	Token        string
}

type Activity struct {
	ReportID string `json:"reportID"`
	Date     string `json:"text2En"`
	Money    string `json:"text4En"`
	Phone    string `json:"text5En"`
	Action string `json:"originalAction"`
}

type Transaction struct {
	Code string `json:"code"`
	Data struct {
		Total      int        `json:"total"`
		TotalPage  int        `json:"totalPage"`
		Activities []Activity `json:"activities"`
	} `json:"data"`
}

type Profile struct {
	Data struct {
		Occupation         string `json:"occupation"`
		Birthdate          string `json:"birthdate"`
		TmnID              string `json:"tmnId"`
		MobileNumber       string `json:"mobileNumber"`
		CurrentBalance     string `json:"currentBalance"`
		ProfileImageStatus string `json:"profileImageStatus"`
		LastnameEn         string `json:"lastnameEn"`
		HasPassword        bool   `json:"hasPassword"`
		Title              string `json:"title"`
		ThaiID             string `json:"thaiId"`
		ProfileType        string `json:"profileType"`
		FirstnameEn        string `json:"firstnameEn"`
		AddressList        []struct {
			AddressID  int    `json:"addressID"`
			Address    string `json:"address"`
			Province   string `json:"province"`
			PostalCode string `json:"postalCode"`
		} `json:"addressList"`
		ImageURL string `json:"imageURL"`
		Fullname string `json:"fullname"`
		Email    string `json:"email"`
	} `json:"data"`
}

func New(id string, password string, options ...interface{}) *TrueWallet {
	var loginType string
	if len(options) >= 1 {
		loginType = options[0].(string)
	} else {
		loginType = "mobile"
	}

	h := sha1.New()
	h.Write([]byte(id + password))
	wallet := TrueWallet{
		username:     id,
		password:     password,
		passwordHash: hex.EncodeToString(h.Sum(nil)),
		loginType:    loginType,
	}
	in := wallet.GetToken()
	var raw map[string]interface{}
	json.Unmarshal(in, &raw)
	data := raw["data"].(map[string]interface{})
	if len(data) == 0 {
		log.Fatal("username or password incorret")
		return nil
	}
	wallet.Token = data["accessToken"].(string)
	return &wallet
}

func (w *TrueWallet) GetToken() []byte {
	url := host + enpointSignin
	resp, err := resty.R().
		SetBody(map[string]interface{}{
			"username": w.username,
			"password": w.passwordHash,
			"type":     w.loginType,
		}).
		SetHeaders(headers).
		Post(url)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return resp.Body()
}

func (w *TrueWallet) GetProfile() *Profile {
	url := host + enpointProfile + w.Token
	resp, err := resty.R().
		SetHeaders(headers).
		Get(url)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	var profile Profile
	json.Unmarshal(resp.Body(), &profile)
	return &profile
}

func (w *TrueWallet) GetRawTransaction(options ...interface{}) []byte {
	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, +1).Format("2006-01-02")
	size := len(options)
	var url bytes.Buffer
	url.WriteString(host + enpointGettran + w.Token)
	if size > 0 {
		if reflect.TypeOf(options[0]).String() == "int" {
			url.WriteString("/?startDate=" + today + "&endDate=" + tomorrow)
			url.WriteString("&limit=" + strconv.Itoa(options[0].(int)))
		} else {
			url.WriteString("/?startDate=" + options[0].(string))
			if size > 1 {
				url.WriteString("&endDate=" + options[1].(string))
				if size > 2 {
					url.WriteString("&limit=" + strconv.Itoa(options[2].(int)))
					if size > 3 {
						url.WriteString("&page=" + strconv.Itoa(options[3].(int)))
						if size > 4 {
							url.WriteString("&type=" + options[4].(string))
							if size > 5 {
								url.WriteString("&action=" + options[5].(string))
							}
						}
					}
				}
			} else {
				t, err := time.Parse("2006-01-02", options[0].(string))
				if err != nil {
					url.WriteString("&endDate=" + tomorrow)
				} else {
					url.WriteString("&endDate=" + t.AddDate(0, 0, +1).Format("2006-01-02"))
				}
			}
		}
	} else {
		url.WriteString("/?startDate=" + today + "&endDate=" + tomorrow)
	}

	resp, err := resty.R().
		SetHeaders(headers).
		Get(url.String())
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return resp.Body()
}

func (w *TrueWallet) GetTransaction(options ...interface{}) *Transaction {
	var transaction Transaction
	json.Unmarshal(w.GetRawTransaction(options...), &transaction)
	return &transaction
}

func (w *TrueWallet) GetTransactionByPhone(options ...interface{}) []Activity {
	act := []Activity{}
	var start string
	var end string
	var phone string
	currentTime := time.Now()
	size := len(options)
	if size == 1 {
		phone = options[0].(string)
		start = currentTime.Format("2006-01-02")
		end = currentTime.AddDate(0, 0, +1).Format("2006-01-02")
	} else if size == 2 {
		phone = options[0].(string)
		start = options[1].(string)
		t, err := time.Parse("2006-01-02", options[1].(string))
		if err != nil {
			end = t.AddDate(0, 0, +1).Format("2006-01-02")
		}
	} else if size == 3 {
		phone = options[0].(string)
		start = options[1].(string)
		end = options[2].(string)
	} else {
		log.Fatalln("Argument is invalid")
	}

	transaction := w.GetTransaction(start, end, 1, 1, "transfer", "creditor")
	n := 10
	limit := transaction.Data.Total/n + 1
	var wg sync.WaitGroup
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go w.getTransactionWorker(&wg, &act, start, end, limit, i, phone)
	}
	wg.Wait()
	return act
}

func (w *TrueWallet) getTransactionWorker(wg *sync.WaitGroup, act *[]Activity, start string, end string, limit int, page int, phone string) {
	transaction := w.GetTransaction(start, end, limit, page, "transfer", "creditor")
	for _, activity := range transaction.Data.Activities {
		if strings.Replace(activity.Phone, "-", "", -1) == phone {
			*act = append(*act, activity)
		}
	}
	wg.Done()
}

func (w *TrueWallet) CheckTransaction(phone string, money string) *Activity {
	currentTime := time.Now()
	today := currentTime.Format("2006-01-02")
	tomorrow := currentTime.AddDate(0, 0, +1).Format("2006-01-02")
	act := []Activity{}
	w.createCheckTransactionWorker(&act, today, tomorrow, phone, money) // check today
	if len(act) == 0 {
		yesterday := currentTime.AddDate(0, 0, -1).Format("2006-01-02")
		w.createCheckTransactionWorker(&act, yesterday, today, phone, money) //check yesterday
	}
	size := len(act)
	if size > 1 {
		t, _ := time.Parse("02/01/06 15:04", act[0].Date)
		a := act[0]
		for i := 1; i < size; i++ {
			t2, _ := time.Parse("02/01/06 15:04", act[i].Date)
			if t.Unix() < t2.Unix() {
				a = act[i]
			}
		}
		return &a
	} else if size == 1 {
		return &act[0]
	}
	return nil
}

func (w *TrueWallet) createCheckTransactionWorker(act *[]Activity, start string, end string, phone string, money string) {
	limit := 27
	transaction := w.GetTransaction(start, end, limit, 1, "transfer", "creditor")
	for _, activity := range transaction.Data.Activities {
		if strings.Replace(activity.Phone, "-", "", -1) == phone && activity.Money == money {
			*act = append(*act, activity)
			return
		}
	}
	n := 10
	limit = transaction.Data.Total/n + 1
	var wg sync.WaitGroup
	for i := 0; i <= n; i++ {
		wg.Add(1)
		go w.checkTransactionWorker(&wg, act, start, end, limit, i, phone, money)
	}
	wg.Wait()
}

func (w *TrueWallet) checkTransactionWorker(wg *sync.WaitGroup, act *[]Activity, start string, end string, limit int, page int, phone string, money string) {
	transaction := w.GetTransaction(start, end, limit, page, "transfer", "creditor")
	for _, activity := range transaction.Data.Activities {
		if strings.Replace(activity.Phone, "-", "", -1) == phone && activity.Money == money {
			*act = append(*act, activity)
			wg.Done()
			return
		}
	}
	wg.Done()
}

func (w *TrueWallet) GetReport(id string) []byte {
	url := host + enpointChecktran + id + "/detail/" + w.Token
	resp, err := resty.R().
		SetHeaders(headers).
		Get(url)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return resp.Body()
}

func (w *TrueWallet) GetBalance() string {
	url := host + enpointProfile + "balance/" + w.Token
	resp, err := resty.R().
		SetHeaders(headers).
		Get(url)
	if err != nil {
		log.Fatalln(err)
		return ""
	}
	var raw map[string]interface{}
	json.Unmarshal(resp.Body(), &raw)
	data := raw["data"].(map[string]interface{})
	return data["currentBalance"].(string)
}

func (w *TrueWallet) TopupMoney(cashcard string) []byte {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	url := host + enpointTopup + timeStamp + "/" + w.Token + "/cashcard/" + cashcard
	resp, err := resty.R().
		SetHeaders(headers).
		Post(url)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return resp.Body()
}