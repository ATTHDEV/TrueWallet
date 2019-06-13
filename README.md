# TrueWalletAPI

---

**Installation**

go get github.com/ATTHDEV/TrueWallet/

---

Example01
```
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

```

---
Example02
```
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

	fmt.Println("this is activities for 0812345678 from 2018-12-01 to 2018-12-02")
	fmt.Println(wallet.GetActivities("0812345678", "2018-12-01", "2018-12-02"))
}

```

---


**License**

MIT license.

---
