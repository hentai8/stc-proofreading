package main

import (
    "fmt"
    "github.com/jinzhu/configor"
    "stc-proofreading-go/cmd/stc"
    "stc-proofreading-go/configs"
)

var cfg configs.Config

//
// main
//  @Description: 主函数入口
//
func main() {
    var stc stc.Node

    // 读取配置文件
    err := configor.Load(&cfg, "configs/config.json")
    if err != nil {
        fmt.Println("read config err")
    }

    startTime := cfg.StartTime
    endTime := cfg.EndTime

    // 获取最新的block和最旧的block的timestamp
    newestTimeInt, oldestTimeInt, newestTransactionGlobalIndexInt := stc.GetTimestampBoundary()

    // 判断输入的两个时间是否都在这个区间内,如果不是,则抛出异常
    if startTime < oldestTimeInt || startTime > newestTimeInt || endTime < oldestTimeInt || endTime > newestTimeInt {
        fmt.Println("Input Time Error")
        return
    }

    // 通过二分法,分别找到输入的两个时间对应的交易id
    startLeftIndex, _ := stc.FindTransactionByTime(startTime, 0, newestTransactionGlobalIndexInt)
    _, endRightIndex := stc.FindTransactionByTime(endTime, 0, newestTransactionGlobalIndexInt)

    // 根据两个时间对应的交易id,遍历其中所有的交易
    IndexSum := endRightIndex - startLeftIndex
    fmt.Println("startLeftIndex:", startLeftIndex)
    fmt.Println("endRightIndex:", endRightIndex)
    fmt.Println("IndexSum:", IndexSum)

    // 获取所有的coinbase收入
    coinbaseSum := stc.GetCoinbaseSumByIndexPeriodNew(startLeftIndex, endRightIndex, cfg.CoinbaseAddress)
    fmt.Println("coinbaseSum:", coinbaseSum)

    // 获取所有的输出
    outputSum := stc.GetOutputSumByIndexPeriodNew(startLeftIndex, endRightIndex, cfg.WithdrawAddress)
    fmt.Println("outputSum:", outputSum)
}
