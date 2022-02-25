package main

import (
    "encoding/hex"
    "fmt"
    "math"
)

func main() {
    var coinbase int64
    data := "0xc8f0814f0c000000000000000000000000000000000000000000000000000001035354430353544300"
    coinbaseAmount := data[2:34]
    //fmt.Println("coinbaseAmount:", coinbaseAmount)
    //fmt.Println("event.Data:", event.Data)
    for i := 0; i < (len(coinbaseAmount) / 2); i++ {
        num1String := coinbaseAmount[0+2*i : 2+2*i]
        num1, err := hex.DecodeString(num1String)
        if err != nil {
            fmt.Println("strconv.ParseInt err")
        }
        coinbase = int64(num1[0])*int64(math.Pow(256, float64(i))) + coinbase
    }
    fmt.Println(coinbase)
}
