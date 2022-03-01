package main

import (
    "fmt"
    "github.com/jinzhu/configor"
    "stc-proofreading-go/cmd/stc"
    "stc-proofreading-go/configs"
    "sync"
    "time"
)

var cfg configs.Config

//
// main
//  @Description: 主函数入口
//
func main() {
    var wg sync.WaitGroup
    duration := time.Hour * 24
    timer := time.NewTimer(duration)
    wg.Add(1)
    go func() {
        for {
            select {
            case <-timer.C:
                //获取前一天的0点和24点的时间戳
                currentTime := time.Now()
                yesterdayTime := time.Unix(currentTime.Unix()-86400, 0)
                startTime := time.Date(yesterdayTime.Year(), yesterdayTime.Month(), yesterdayTime.Day(), 0, 0, 0, 0, yesterdayTime.Location()).UnixNano() / 1000000
                endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location()).UnixNano() / 1000000

                //运行统计,计入mysql数据库
                periodicTask(startTime, endTime)
                //periodicTask(1645574400000, 1645575400000)
                timer.Reset(duration)
            }
        }
        wg.Done()
    }()
    wg.Wait()
}

//
// periodicTask
//  @Description: 单次统计
//  @param startTime
//  @param endTime
//
func periodicTask(startTime int64, endTime int64) {
    var stc stc.Node

    // 读取配置文件
    err := configor.Load(&cfg, "../configs/config.json")
    if err != nil {
        fmt.Println("read config err")
    }

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
