package main

import (
	"fmt"

	TrueWallet "github.com/ATTHDEV/TrueWallet-API"
)

func main() {

	mobileTracking, _ := TrueWallet.GenerateRandomString(40)
	wallet, err := TrueWallet.New("xxxxxxxxxx", "xxxx", "email", mobileTracking) //put you email and password
	if err != nil {
		panic(err)
	}
	ref, err := wallet.GetOtp()
	if err != nil {
		panic(err)
	}
	wallet.ConfirmOtp("You mobile number", "You OTP", ref)

	// if you confirm otp , you will have reference token
	fmt.Println(wallet.ReferenceToken)

	// now you can fetch data from you wallet..

	fmt.Println("this is last 100 transaction for today")
	fmt.Println(wallet.GetTransaction(100))
}
