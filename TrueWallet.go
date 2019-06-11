// author ATTH
package TrueWallet

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	resty "gopkg.in/resty.v1"
)

type Data struct {
	Data string `json:"data"`
}

const (
	host = `https://mobile-api-gateway.truemoney.com/mobile-api-gateway/api/v1/`

	enpointOtp        = "login/otp/"
	enpointConfirmOtp = "login/otp/verification/"
	enpointLogin      = `login/`
	enpointProfile    = `profile/`
	enpointTopup      = `topup/mobile/`
	enpointGettran    = `profile/transactions/history/`
	enpointChecktran  = `profile/activities/`

	secretKey = "9LXAVCxcITaABNK48pAVgc4muuTNJ4enIKS5YzKyGZ"
)

var headers = map[string]string{
	"Host":         "mobile-api-gateway.truemoney.com",
	"Content-Type": "application/json",
	"User-Agent":   "okhttp/3.8.0",
	"username":     "",
	"password":     "",
	// "Authorization": "",
}

type Wallet struct {
	username       string
	password       string
	passwordHash   string
	loginType      string
	MobileTracking string
	DeviceID       string
	AccessToken    string
	ReferenceToken string
	Headers        map[string]string
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
	OtpError     = 5
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
	ReportID string `json:"report_id"`
	Date     string `json:"date_time"`
	Money    string `json:"amount"`
	Phone    string `json:"sub_title"`
	Action   string `json:"original_action"`
}

type Transaction struct {
	Code string `json:"code"`
	Data struct {
		Total      int        `json:"total"`
		TotalPage  int        `json:"total_page"`
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

type Report struct {
	Code string `json:"code"`
	Data struct {
		ServiceType string  `json:"service_type"`
		Amount      float64 `json:"amount"`
		Section2    struct {
			Column1 struct {
				Cell1 struct {
					Title string `json:"title"`
					Value string `json:"value"`
				} `json:"cell1"`
				Cell2 struct {
					Title string `json:"title"`
					Value string `json:"value"`
				} `json:"cell2"`
			} `json:"column1"`
		} `json:"section2"`
		Section4 struct {
			Column1 struct {
				Cell1 struct {
					Title string `json:"title"`
					Value string `json:"value"`
				} `json:"cell1"`
			} `json:"column1"`
			Column2 struct {
				Cell1 struct {
					Title string `json:"title"`
					Value string `json:"value"`
				} `json:"cell1"`
			} `json:"column2"`
		} `json:"section4"`
	} `json:"data"`
}

func TimestampString() string {
	msec := time.Now().UnixNano() / 1000000
	return strconv.Itoa(int(msec))
}

func CreateSignature(msg string) string {
	h := hmac.New(sha1.New, []byte(secretKey))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func New(id, password, loginType string, options ...interface{}) (*Wallet, *Error) {
	var mobileTracking string
	if len(options) == 1 {
		mobileTracking = options[0].(string)
	}

	h := sha1.New()
	h.Write([]byte(id + password))
	wallet := Wallet{
		username:     id,
		password:     password,
		passwordHash: hex.EncodeToString(h.Sum(nil)),
		loginType:    loginType,
	}

	if mobileTracking != "" {
		wallet.MobileTracking = mobileTracking
	} else {
		mobileTracking, err := GenerateRandomString(40)
		if err == nil {
			wallet.MobileTracking = mobileTracking
		}
	}

	deviceID := md5.Sum([]byte(wallet.MobileTracking))
	wallet.DeviceID = hex.EncodeToString(deviceID[:])[0:16]

	wallet.Headers = headers
	wallet.Headers["username"] = wallet.username
	wallet.Headers["password"] = wallet.passwordHash
	return &wallet, nil
}

func (w *Wallet) GetOtp() (string, *Error) {
	timestamp := TimestampString()
	resp, err := resty.R().
		SetBody(map[string]interface{}{
			"device_id": w.DeviceID,
			"signature": CreateSignature(w.loginType + "|" + w.DeviceID + "|" + timestamp),
			"timestamp": timestamp,
			"type":      w.loginType,
		}).
		SetHeaders(w.Headers).
		Post(host + enpointOtp)

	if err != nil {
		return "", NewError(UnknownError, "unknow error.")
	}

	in := resp.Body()
	var raw map[string]interface{}
	json.Unmarshal(in, &raw)

	code := raw["code"].(string)
	if code == "1014" {
		return "", NewError(LoginError, "รหัสผ่านหรืออีเมลล์ไม่ถูกต้อง")
	}

	if code == "3" {
		return "", NewError(LoginError, "ขออภัย ไม่พบบัญชีผู้ใช้")
	}

	data := raw["data"].(map[string]interface{})
	if len(data) == 0 {
		return "", NewError(UnknownError, "unknow error.")
	}

	if val, ok := data["otp_reference"]; ok {
		return val.(string), nil
	}

	return "", NewError(UnknownError, "unknow error.")
}

func (w *Wallet) ConfirmOtp(phone, otp, ref string) *Error {
	timestamp := TimestampString()
	resp, err := resty.R().
		SetBody(map[string]interface{}{
			"device_id":       w.DeviceID,
			"mobile_number":   phone,
			"mobile_tracking": w.MobileTracking,
			"otp_code":        otp,
			"otp_reference":   ref,
			"signature":       CreateSignature(w.loginType + "|" + otp + "|" + phone + "|" + ref + "|" + w.DeviceID + "|" + w.MobileTracking + "|" + timestamp),
			"timestamp":       timestamp,
			"type":            w.loginType,
		}).
		SetHeaders(w.Headers).
		Post(host + enpointConfirmOtp)

	if err != nil {
		return NewError(UnknownError, "unknow error.")
	}

	in := resp.Body()
	var raw map[string]interface{}
	json.Unmarshal(in, &raw)

	code := raw["code"].(string)
	if code == "1001" {
		return NewError(OtpError, "คุณกรอก otp ไม่ถูกต้อง")
	}

	data := raw["data"].(map[string]interface{})
	if len(data) == 0 {
		return NewError(UnknownError, "unknow error.")
	}

	if val, ok := data["access_token"]; ok {
		w.AccessToken = val.(string)
		w.Headers["Authorization"] = w.AccessToken
	}
	if val, ok := data["reference_token"]; ok {
		w.ReferenceToken = val.(string)
	}
	return nil
}

func (w *Wallet) SetReferenceToken(token string) {
	w.ReferenceToken = token
}

func (w *Wallet) Login() *Error {
	timestamp := TimestampString()
	resp, err := resty.R().
		SetBody(map[string]interface{}{
			"device_id":       w.DeviceID,
			"mobile_tracking": w.MobileTracking,
			"reference_token": w.ReferenceToken,
			"signature":       CreateSignature(w.loginType + "|" + w.ReferenceToken + "|" + w.DeviceID + "|" + w.MobileTracking + "|" + timestamp),
			"timestamp":       timestamp,
			"type":            w.loginType,
		}).
		SetHeaders(headers).
		Post(host + enpointLogin)

	if err != nil {
		return NewError(UnknownError, "unknow error.")
	}

	in := resp.Body()
	var raw map[string]interface{}
	json.Unmarshal(in, &raw)
	data := raw["data"].(map[string]interface{})
	if len(data) == 0 {
		return NewError(LoginError, "User not found.")
	}

	if val, ok := data["access_token"]; ok {
		w.AccessToken = val.(string)
		w.Headers["Authorization"] = w.AccessToken
	}

	return nil
}

func (w *Wallet) GetProfile() (*Profile, *Error) {
	url := host + enpointProfile + w.AccessToken
	resp, err := resty.R().
		SetHeaders(w.Headers).
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
	// url.WriteString(host + enpointGettran + w.AccessToken)
	url.WriteString("https://mobile-api-gateway.truemoney.com/mobile-api-gateway/user-profile-composite/v1/users/transactions/history")
	if size > 0 {
		if reflect.TypeOf(options[0]).String() == "int" {
			url.WriteString("?start_date=" + today + "&end_date=" + tomorrow)
			url.WriteString("&limit=" + strconv.Itoa(options[0].(int)))
		} else {
			url.WriteString("?start_date=" + options[0].(string))
			if size > 1 {
				url.WriteString("&end_date=" + options[1].(string))
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
				if err == nil {
					url.WriteString("&end_date=" + tomorrow)
				} else {
					url.WriteString("&end_date=" + t.AddDate(0, 0, +1).Format("2006-01-02"))
				}
			}
		}
	} else {
		url.WriteString("?start_date=" + today + "&end_date=" + tomorrow)
	}

	resp, err := resty.R().
		SetHeaders(w.Headers).
		Get(url.String())
	if err != nil {
		return nil
	}
	return resp.Body()
}

func (w *Wallet) GetTransaction(options ...interface{}) (*Transaction, *Error) {
	if w.AccessToken == "" {
		return nil, NewError(TokenError, "Token not found.")
	}
	var transaction Transaction
	json.Unmarshal(w.GetRawTransaction(options...), &transaction)
	var err *Error
	code := transaction.Code
	switch transaction.Code {
	case "UPC-200":
		err = nil
	case "UPC-400", "MAS-401":
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
		t, err := time.Parse("2006-01-02", options[0].(string))
		if err == nil {
			end = t.AddDate(0, 0, +1).Format("2006-01-02")
		}
	} else {
		currentTime := time.Now()
		start = currentTime.Format("2006-01-02")
		end = currentTime.AddDate(0, 0, +1).Format("2006-01-02")
	}
	m := fmt.Sprintf("%.2f", money)
	limit := 30
	transaction, err := w.GetTransaction(start, end, limit, 1, "transfer", "creditor")
	if err != nil {
		return nil, err
	}
	for _, activity := range transaction.Data.Activities {
		acm := strings.Replace(activity.Money, ",", "", -1)
		acm = acm[1:len(acm)]
		if strings.Replace(activity.Phone, "-", "", -1) == phone && acm == m {
			return &activity, nil
		}
	}

	if transaction.Data.Total == 0 {
		return nil, nil
	}
	transaction, err = w.GetTransaction(start, end, transaction.Data.Total, 1, "transfer", "creditor")
	if err != nil {
		return nil, err
	}
	for _, activity := range transaction.Data.Activities {
		acm := strings.Replace(activity.Money, ",", "", -1)
		acm = acm[1:len(acm)]
		if strings.Replace(activity.Phone, "-", "", -1) == phone && acm == m {
			return &activity, nil
		}
	}
	return nil, nil
}

func (w *Wallet) GetReport(id string) (*Report, *Error) {
	url := "https://mobile-api-gateway.truemoney.com/mobile-api-gateway/user-profile-composite/v1/users/transactions/history/detail/" + id
	resp, err := resty.R().
		SetHeaders(w.Headers).
		Get(url)
	if err != nil {
		return nil, NewError(RequestError, err.Error())
	}
	var data Report
	json.Unmarshal(resp.Body(), &data)
	return &data, nil
}

func (w *Wallet) GetBalance() (string, *Error) {
	url := host + enpointProfile + "balance/" + w.AccessToken
	resp, err := resty.R().
		SetHeaders(headers).
		Get(url)
	if err != nil {
		return "", NewError(RequestError, err.Error())
	}
	var raw map[string]interface{}
	json.Unmarshal(resp.Body(), &raw)
	data := raw["data"].(map[string]interface{})
	if data != nil {
		balance := data["currentBalance"]
		if balance != nil {
			if str, ok := balance.(string); ok {
				return str, nil
			}
		}
	}
	return "0", nil
}

func (w *Wallet) TopupMoney(cashcard string) ([]byte, *Error) {
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	url := host + enpointTopup + timeStamp + "/" + w.AccessToken + "/cashcard/" + cashcard
	resp, err := resty.R().
		SetHeaders(headers).
		Post(url)
	if err != nil {
		return nil, NewError(RequestError, err.Error())
	}
	return resp.Body(), nil
}

func (w *Wallet) ClearToken() {
	w.AccessToken = w.AccessToken + "A"
	w.Headers["Authorization"] = w.AccessToken
}
