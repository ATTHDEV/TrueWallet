package main

import (
	"fmt"

	"github.com/ATTHDEV/TruwWallet"
)

func main() {
	wallet, err := TrueWallet.New("xxxxxxxxxx", "xxxx") //put you mobile number and pin

	if err != nil {

		fmt.Println("this is last 100 transaction for today")

		transaction := wallet.GetTransaction(100)
		fmt.Println(transaction)

	}
}
