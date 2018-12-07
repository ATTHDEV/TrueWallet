# TrueWallet

---

**Installation**

go get github.com/ATTHDEV/TrueWallet/

---

Example
```
package main

import (
	"fmt"

	"github.com/ATTHDEV/TruwWallet"
)

func main() {
	wallet := TrueWallet.New("xxxxxxxxxx", "xxxx") //put you mobile number and pin

	fmt.Println("this is last 50 transaction for today")

	transaction := wallet.GetTransaction()
	fmt.Println(transaction)
}

```
