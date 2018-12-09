package main

import (
	"fmt"

	"github.com/ATTHDEV/TruwWallet"
)

func main() {
	wallet, err := TrueWallet.New("xxxxxxxxxx", "xxxx") //put you mobile number and pin

	fmt.Println("this is activities for 0812345678 from 2018-12-01 to 2018-12-02")

	transaction := wallet.GetActivities("0812345678", "2018-12-01", "2018-12-02")
	fmt.Println(transaction)
}
