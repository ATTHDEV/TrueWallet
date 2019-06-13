package main

import (
	"fmt"

	TrueWallet "github.com/ATTHDEV/TrueWallet-API"
)

func main() {

	// in example1 you have mobileTracking and ReferencToken
	// you will use it aways
	mobileTracking := "xxxxx" // frome example01
	refToken := "xxxxx"       // frome example01 , after confirm OTP

	wallet, err := TrueWallet.New("xxxxxxxxxx", "xxxx", "email", mobileTracking) //put you email and password
	wallet.SetReferenceToken(refToken)
	if err != nil {
		panic(err)
	}

	// Login for receive access token
	wallet.Login()

	fmt.Println("this is activities for 0812345678 from 2018-12-01 to 2018-12-02")
	fmt.Println(wallet.GetActivities("0812345678", "2018-12-01", "2018-12-02"))
}
