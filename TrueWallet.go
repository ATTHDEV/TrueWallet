// author ATTH
package TrueWallet

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

type Wallet struct {
	username     string
	password     string
	passwordHash string
	loginType    string
	Token        string
}

type Error struct {
	Code    int
	Message string
}

const (
	RequestError = 1
	LoginError   = 2
	TokenError   = 3
	UnknownError = 4
)

func NewError(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

const (
	StatusSuccess = 1
	StatusFail    = 0
)

type Activity struct {
	ReportID string `json:"reportID"`
	Date     string `json:"text2En"`
	Money    string `json:"text4En"`
	Phone    string `json:"text5En"`
	Action   string `json:"originalAction"`
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
	Code string `json:"code"`
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

func New(id string, password string, options ...interface{}) (*Wallet, *Error) {
	var loginType string
	if len(options) >= 1 {
		loginType = options[0].(string)
	} else {
		loginType = "mobile"
	}
	h := sha1.New()
	h.Write([]byte(id + password))
	wallet := Wallet{
		username:     id,
		password:     password,
		passwordHash: hex.EncodeToString(h.Sum(nil)),
		loginType:    loginType,
	}
	t, err := wallet.GetToken()
	if err != nil {
		return &wallet, err
	}
	wallet.Token = t
	return &wallet, nil
}

func (w *Wallet) GetToken() (string, *Error) {
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
		return "", NewError(RequestError, err.Error())
	}
	in := resp.Body()
	var raw map[string]interface{}
	json.Unmarshal(in, &raw)
	data := raw["data"].(map[string]interface{})
	if len(data) == 0 {
		return "", NewError(LoginError, "User not found.")
	}
	return data["accessToken"].(string), nil
}

func (w *Wallet) GetProfile() (*Profile, *Error) {
	url := host + enpointProfile + w.Token
	resp, err := resty.R().
		SetHeaders(headers).
		Get(url)
	if err != nil {
		return nil, NewError(RequestError, err.Error())
	}
	var profile Profile
	json.Unmarshal(resp.Body(), &profile)
	return &profile, nil
}

func (w *Wallet) GetRawTransaction(options ...interface{}) []byte {
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
		return nil
	}
	return resp.Body()
}

func (w *Wallet) GetTransaction(options ...interface{}) (*Transaction, *Error) {
	if w.Token == "" {
		return nil, NewError(TokenError, "Token not found.")
	}
	var transaction Transaction
	json.Unmarshal(w.GetRawTransaction(options...), &transaction)
	var err *Error
	code := transaction.Code
	switch transaction.Code {
	case "20000":
		err = nil
	case "40000":
		err = NewError(TokenError, "Token not found.")
	default:
		err = NewError(UnknownError, "Unknown Error Code "+code)
	}
	return &transaction, err
}

func (w *Wallet) GetActivities(options ...interface{}) ([]Activity, *Error) {
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
	transaction, err := w.GetTransaction(start, end, 1, 1, "transfer")
	if err != nil {
		return act, err
	}
	for _, activity := range transaction.Data.Activities {
		if strings.Replace(activity.Phone, "-", "", -1) == phone {
			act = append(act, activity)
		}
	}
	n := 8
	limit := transaction.Data.Total/n + 1
	var wg sync.WaitGroup
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, act *[]Activity, start, end string, page int) {
			t, _ := w.GetTransaction(start, end, limit, page, "transfer")
			for _, activity := range t.Data.Activities {
				if strings.Replace(activity.Phone, "-", "", -1) == phone {
					*act = append(*act, activity)
				}
			}
			wg.Done()
		}(&wg, &act, start, end, i)
	}
	wg.Wait()
	return act, nil
}

func (w *Wallet) GetLastTransfer(phone string, money float32, options ...interface{}) (*Activity, *Error) {
	if money < 0 {
		log.Fatalln("Errer money must be positive.")
	}
	var start string
	var end string
	if len(options) > 0 {
		start = options[0].(string)
		t, err := time.Parse("2006-01-02", options[1].(string))
		if err != nil {
			end = t.AddDate(0, 0, +1).Format("2006-01-02")
		}
	} else {
		currentTime := time.Now()
		start = currentTime.Format("2006-01-02")
		end = currentTime.AddDate(0, 0, +1).Format("2006-01-02")
	}
	m := fmt.Sprintf("+%.2f", money)
	limit := 30
	transaction, err := w.GetTransaction(start, end, limit, 1, "transfer", "creditor")
	if err != nil {
		return nil, err
	}
	for _, activity := range transaction.Data.Activities {
		if strings.Replace(activity.Phone, "-", "", -1) == phone && activity.Money == m {
			return &activity, nil
		}
	}
	for i := 2; i <= transaction.Data.TotalPage; i++ {
		transaction, err = w.GetTransaction(start, end, limit, i, "transfer", "creditor")
		if err != nil {
			return nil, err
		}
		for _, activity := range transaction.Data.Activities {
			if strings.Replace(activity.Phone, "-", "", -1) == phone && activity.Money == m {
				return &activity, nil
			}
		}
	}
	return nil, nil
}

func (w *Wallet) GetReport(id string) ([]byte, *Error) {
	url := host + enpointChecktran + id + "/detail/" + w.Token
	resp, err := resty.R().
		SetHeaders(headers).
		Get(url)
	if err != nil {
		return nil, NewError(RequestError, err.Error())
	}
	return resp.Body(), nil
}

func (w *Wallet) GetBalance() (string, *Error) {
	url := host + enpointProfile + "balance/" + w.Token
	resp, err := resty.R().
		SetHeaders(headers).
		Get(url)
	if err != nil {
		return "", NewError(RequestError, err.Error())
	}
	var raw map[string]interface{}
	json.Unmarshal(resp.Body(), &raw)
	data := raw["data"].(map[string]interface{})
	return data["currentBalance"].(string), nil
}

func (w *Wallet) TopupMoney(cashcard string) ([]byte, *Error) {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	url := host + enpointTopup + timeStamp + "/" + w.Token + "/cashcard/" + cashcard
	resp, err := resty.R().
		SetHeaders(headers).
		Post(url)
	if err != nil {
		return nil, NewError(RequestError, err.Error())
	}
	return resp.Body(), nil
}
