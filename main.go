package main

import (
    "encoding/hex"
    "encoding/json"
    "fmt"
    "github.com/jinzhu/configor"
    "io/ioutil"
    "math"
    "net/http"
    "stc-proofreading-go/configs"
    "strconv"
    "strings"
)

//
// Body
//  @Description: rpc交互的json body
//
type Body struct {
    Id      int           `json:"id"`
    Jsonrpc string        `json:"jsonrpc"`
    Method  string        `json:"method"`
    Params  []interface{} `json:"params"`
}

type BlockInfo struct {
    BlockHash          string `json:"block_hash"`
    TotalDifficulty    string `json:"total_difficulty"`
    TxnAccumulatorInfo struct {
        AccumulatorRoot    string   `json:"accumulator_root"`
        FrozenSubtreeRoots []string `json:"frozen_subtree_roots"`
        NumLeaves          string   `json:"num_leaves"`
        NumNodes           string   `json:"num_nodes"`
    } `json:"txn_accumulator_info"`
    BlockAccumulatorInfo struct {
        AccumulatorRoot    string   `json:"accumulator_root"`
        FrozenSubtreeRoots []string `json:"frozen_subtree_roots"`
        NumLeaves          string   `json:"num_leaves"`
        NumNodes           string   `json:"num_nodes"`
    } `json:"block_accumulator_info"`
}

type NodeInfo struct {
    Jsonrpc string `json:"jsonrpc"`
    Result  struct {
        PeerInfo struct {
            PeerId    string `json:"peer_id"`
            ChainInfo struct {
                ChainId     int    `json:"chain_id"`
                GenesisHash string `json:"genesis_hash"`
                Head        struct {
                    BlockHash            string      `json:"block_hash"`
                    ParentHash           string      `json:"parent_hash"`
                    Timestamp            string      `json:"timestamp"`
                    Number               string      `json:"number"`
                    Author               string      `json:"author"`
                    AuthorAuthKey        interface{} `json:"author_auth_key"`
                    TxnAccumulatorRoot   string      `json:"txn_accumulator_root"`
                    BlockAccumulatorRoot string      `json:"block_accumulator_root"`
                    StateRoot            string      `json:"state_root"`
                    GasUsed              string      `json:"gas_used"`
                    Difficulty           string      `json:"difficulty"`
                    BodyHash             string      `json:"body_hash"`
                    ChainId              int         `json:"chain_id"`
                    Nonce                int         `json:"nonce"`
                    Extra                string      `json:"extra"`
                } `json:"head"`
                BlockInfo BlockInfo `json:"block_info"`
            } `json:"chain_info"`
            NotifProtocols string `json:"notif_protocols"`
            RpcProtocols   string `json:"rpc_protocols"`
        } `json:"peer_info"`
        SelfAddress string `json:"self_address"`
        Net         string `json:"net"`
        Consensus   struct {
            Type string `json:"type"`
        } `json:"consensus"`
        NowSeconds int `json:"now_seconds"`
    } `json:"result"`
    Id int `json:"id"`
}

type Block struct {
    Jsonrpc string `json:"jsonrpc"`
    Result  struct {
        Header struct {
            BlockHash            string      `json:"block_hash"`
            ParentHash           string      `json:"parent_hash"`
            Timestamp            string      `json:"timestamp"`
            Number               string      `json:"number"`
            Author               string      `json:"author"`
            AuthorAuthKey        interface{} `json:"author_auth_key"`
            TxnAccumulatorRoot   string      `json:"txn_accumulator_root"`
            BlockAccumulatorRoot string      `json:"block_accumulator_root"`
            StateRoot            string      `json:"state_root"`
            GasUsed              string      `json:"gas_used"`
            Difficulty           string      `json:"difficulty"`
            BodyHash             string      `json:"body_hash"`
            ChainId              int         `json:"chain_id"`
            Nonce                int         `json:"nonce"`
            Extra                string      `json:"extra"`
        } `json:"header"`
        Body struct {
            Full []interface{} `json:"Full"`
        } `json:"body"`
        Uncles []interface{} `json:"uncles"`
    } `json:"result"`
    Id int `json:"id"`
}

type TransactionInfos struct {
    Jsonrpc string                  `json:"jsonrpc"`
    Result  []TransactionInfoResult `json:"result"`
    Id      int                     `json:"id"`
}

type TransactionInfoResult struct {
    BlockHash              string `json:"block_hash"`
    BlockNumber            string `json:"block_number"`
    TransactionHash        string `json:"transaction_hash"`
    TransactionIndex       int    `json:"transaction_index"`
    TransactionGlobalIndex string `json:"transaction_global_index"`
    StateRootHash          string `json:"state_root_hash"`
    EventRootHash          string `json:"event_root_hash"`
    GasUsed                string `json:"gas_used"`
    //Status                 string `json:"status"`
}

type TransactionInfo struct {
    Jsonrpc string                `json:"jsonrpc"`
    Result  TransactionInfoResult `json:"result"`
    Id      int                   `json:"id"`
}

type BlockTxnInfos struct {
    Jsonrpc string `json:"jsonrpc"`
    Result  []struct {
        BlockHash              string `json:"block_hash"`
        BlockNumber            string `json:"block_number"`
        TransactionHash        string `json:"transaction_hash"`
        TransactionIndex       int    `json:"transaction_index"`
        TransactionGlobalIndex string `json:"transaction_global_index"`
        StateRootHash          string `json:"state_root_hash"`
        EventRootHash          string `json:"event_root_hash"`
        GasUsed                string `json:"gas_used"`
        Status                 string `json:"status"`
    } `json:"result"`
    Id int `json:"id"`
}

type EventsByTxnHash struct {
    Jsonrpc string `json:"jsonrpc"`
    Result  []struct {
        BlockHash              string `json:"block_hash"`
        BlockNumber            string `json:"block_number"`
        TransactionHash        string `json:"transaction_hash"`
        TransactionIndex       int    `json:"transaction_index"`
        TransactionGlobalIndex string `json:"transaction_global_index"`
        Data                   string `json:"data"`
        TypeTag                string `json:"type_tag"`
        EventIndex             int    `json:"event_index"`
        EventKey               string `json:"event_key"`
        EventSeqNumber         string `json:"event_seq_number"`
    } `json:"result"`
    Id int `json:"id"`
}

var cfg configs.Config

//
// main
//  @Description: 主函数入口
//
func main() {
    // 读取配置文件
    err := configor.Load(&cfg, "configs/config.json")
    if err != nil {
        fmt.Println("read config err")
    }

    startTime := cfg.StartTime
    endTime := cfg.EndTime

    // 获取最新的block和最旧的block的timestamp
    newestTimeInt, oldestTimeInt, newestTransactionGlobalIndexInt := getTimestampBoundary()

    // 判断输入的两个时间是否都在这个区间内,如果不是,则抛出异常
    if startTime < oldestTimeInt || startTime > newestTimeInt || endTime < oldestTimeInt || endTime > newestTimeInt {
        fmt.Println("Input Time Error")
        return
    }

    // 通过二分法,分别找到输入的两个时间对应的交易id
    startLeftIndex, _ := findTransactionByTime(startTime, 0, newestTransactionGlobalIndexInt)
    _, endRightIndex := findTransactionByTime(endTime, 0, newestTransactionGlobalIndexInt)

    // 根据两个时间对应的交易id,遍历其中所有的交易
    IndexSum := endRightIndex - startLeftIndex
    fmt.Println("IndexSum:", IndexSum)

    // 获取所有的coinbase收入
    coinbaseSum := getCoinbaseSumByIndexPeriod(startLeftIndex, endRightIndex, cfg.CoinbaseAddress)
    fmt.Println("coinbaseSum:", coinbaseSum)

    // 获取所有的输出
    outputSum := getOutputSumByIndexPeriod(startLeftIndex, endRightIndex, cfg.WithdrawAddress)
    fmt.Println("outputSum:", outputSum)
}

//
// getTimestampBoundary
//  @Description: 获取timestamp的边界值
//  @return newestTime 最新block中的timestamp
//  @return oldestTime 0号block中的timestamp
//  @return newestTransactionGlobalIndexInt 最新block中的txid
//
func getTimestampBoundary() (newestTime int64, oldestTime int64, newestTransactionGlobalIndexInt int64) {
    // 获取最新的block和最旧的block的timestamp
    nodeInfo := getNodeInfo()
    //newestNumber := nodeInfo.Result.PeerInfo.ChainInfo.Head.Number
    newestBlockHash := nodeInfo.Result.PeerInfo.ChainInfo.Head.BlockHash
    blockTxnInfos := getBlockTxnInfos(newestBlockHash)

    newestTXHash := blockTxnInfos.Result[0].TransactionHash

    newestTransactionInfo := getTransactionInfo(newestTXHash)
    //fmt.Println(newestTransactionInfo)
    newestTransactionGlobalIndex := newestTransactionInfo.Result.TransactionGlobalIndex

    //fmt.Println(newestTransactionGlobalIndex)
    newestTransactionGlobalIndexInt, err := strconv.ParseInt(newestTransactionGlobalIndex, 10, 64)
    fmt.Println("newestTransactionGlobalIndexInt", newestTransactionGlobalIndexInt)
    if err != nil {
        fmt.Println("strconv.ParseInt err")
    }

    newestTimeString := getTimeByTransactionGlobalIndex(newestTransactionGlobalIndexInt)
    oldestTimeString := getTimeByTransactionGlobalIndex(0)

    newestTime, err = strconv.ParseInt(newestTimeString, 10, 64)
    if err != nil {
        fmt.Println("strconv.ParseInt err")
    }
    oldestTime, err = strconv.ParseInt(oldestTimeString, 10, 64)
    if err != nil {
        fmt.Println("strconv.ParseInt err")
    }
    return newestTime, oldestTime, newestTransactionGlobalIndexInt
}

//
// getOutputSumByIndexPeriod
//  @Description: 通过起始的txid和结束的txid获取该address的所有支出的统计
//  @param startIndex 起始的txid
//  @param endIndex 结束的txid
//  @param address 带salt的address地址
//  @return OutputSum 统计出的出帐总值
//
func getOutputSumByIndexPeriod(startIndex int64, endIndex int64, address string) (OutputSum int64) {
    if startIndex > endIndex {
        x := startIndex
        startIndex = endIndex
        endIndex = x
    }
    leftIndex := startIndex
    rightIndex := endIndex
    for rightIndex > leftIndex {
        if rightIndex-leftIndex > 32 {
            transactionInfos := getTransactionInfos(leftIndex, false, 32)
            for _, transactionInfoResult := range transactionInfos.Result {
                events := getEventsByTxnHash(transactionInfoResult.TransactionHash)
                for _, event := range events.Result {
                    if event.TypeTag == "0x00000000000000000000000000000001::Account::WithdrawEvent" && event.EventKey == address {
                        //fmt.Println("event.TransactionHash:", event.TransactionHash)
                        var output int64
                        outputAmount := event.Data[2:34]
                        //fmt.Println("outputAmount:", outputAmount)
                        //fmt.Println("event.Data:", event.Data)
                        for i := 0; i < (len(outputAmount) / 2); i++ {
                            num1String := outputAmount[0+2*i : 2+2*i]
                            num1, err := hex.DecodeString(num1String)
                            if err != nil {
                                fmt.Println("strconv.ParseInt err")
                            }
                            output = int64(num1[0])*int64(math.Pow(256, float64(i))) + output
                        }
                        //fmt.Println("output:", output)
                        OutputSum = OutputSum + output
                    }
                }
            }
            leftIndex = leftIndex + 32
        } else if rightIndex-leftIndex > 0 {
            transactionInfos := getTransactionInfos(leftIndex, false, rightIndex-leftIndex)
            for _, transactionInfoResult := range transactionInfos.Result {
                events := getEventsByTxnHash(transactionInfoResult.TransactionHash)
                for _, event := range events.Result {
                    if event.TypeTag == "0x00000000000000000000000000000001::Account::WithdrawEvent" && event.EventKey == address {
                        //fmt.Println("event.TransactionHash:", event.TransactionHash)
                        var output int64
                        outputAmount := event.Data[2:34]
                        //fmt.Println("outputAmount:", outputAmount)
                        //fmt.Println("event.Data:", event.Data)
                        for i := 0; i < (len(outputAmount) / 2); i++ {
                            num1String := outputAmount[0+2*i : 2+2*i]
                            num1, err := hex.DecodeString(num1String)
                            if err != nil {
                                fmt.Println("strconv.ParseInt err")
                            }
                            output = int64(num1[0])*int64(math.Pow(256, float64(i))) + output
                        }
                        OutputSum = OutputSum + output
                    }
                }
            }
            leftIndex = rightIndex
        }
    }
    return OutputSum
}

//
// getCoinbaseSumByIndexPeriod
//  @Description: 通过起始的txid和结束的txid获取该address的所有coinbase收入的统计
//  @param startIndex 起始的txid
//  @param endIndex 结束的txid
//  @param address 带salt的address地址
//  @return OutputSum 统计出的coinbase收入总值
//
func getCoinbaseSumByIndexPeriod(startIndex int64, endIndex int64, address string) (CoinbaseSum int64) {
    if startIndex > endIndex {
        x := startIndex
        startIndex = endIndex
        endIndex = x
    }
    leftIndex := startIndex
    rightIndex := endIndex
    for rightIndex > leftIndex {
        if rightIndex-leftIndex > 32 {
            transactionInfos := getTransactionInfos(leftIndex, false, 32)
            for _, transactionInfoResult := range transactionInfos.Result {
                events := getEventsByTxnHash(transactionInfoResult.TransactionHash)
                for _, event := range events.Result {
                    if event.TypeTag == "0x00000000000000000000000000000001::Account::DepositEvent" && event.EventKey == address {
                        var coinbase int64
                        coinbaseAmount := event.Data[2:34]
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
                        CoinbaseSum = CoinbaseSum + coinbase
                    }
                }
            }
            leftIndex = leftIndex + 32
        } else if rightIndex-leftIndex > 0 {
            transactionInfos := getTransactionInfos(leftIndex, false, rightIndex-leftIndex)
            for _, transactionInfoResult := range transactionInfos.Result {
                events := getEventsByTxnHash(transactionInfoResult.TransactionHash)
                for _, event := range events.Result {
                    if event.TypeTag == "0x00000000000000000000000000000001::Account::DepositEvent" && event.EventKey == address {
                        var coinbase int64
                        coinbaseAmount := event.Data[2:34]
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
                        CoinbaseSum = CoinbaseSum + coinbase
                    }
                }
            }
            leftIndex = rightIndex
        }
    }
    return CoinbaseSum
}

//
// findTransactionByTime
//  @Description: 通过timestamp寻找到Transaction的txid,采用二分法快速查询
//  @param time 目标时间戳
//  @param oldestTransactionGlobalIndex 最旧的txid
//  @param newestTransactionGlobalIndex 最新的txid
//  @return leftIndex 该时间戳对应的前一个txid
//  @return rightIndex 该时间戳对应的后一个txid
//
func findTransactionByTime(time int64, oldestTransactionGlobalIndex int64, newestTransactionGlobalIndex int64) (leftIndex int64, rightIndex int64) {
    leftIndex = oldestTransactionGlobalIndex
    rightIndex = newestTransactionGlobalIndex

    for leftIndex <= rightIndex {

        middleIndex := (leftIndex + rightIndex) / 2
        middleTransactionInfos := getTransactionInfos(middleIndex, false, 1)
        middleBlockHash := middleTransactionInfos.Result[0].BlockHash
        middleBlock := getBlockByHash(middleBlockHash)
        middleTimeString := middleBlock.Result.Header.Timestamp
        middleTime, err := strconv.ParseInt(middleTimeString, 10, 64)
        if err != nil {
            fmt.Println("strconv.ParseInt err")
        }
        if middleTime > time {
            rightIndex = middleIndex - 1
        } else if middleTime < time {
            leftIndex = middleIndex + 1
        } else {
            return leftIndex, rightIndex
        }
    }
    fmt.Println("leftIndex:", leftIndex)
    fmt.Println("rightIndex:", rightIndex)
    //fmt.Println("leftIndex:", leftIndex)
    //fmt.Println("leftIndex:", leftIndex)
    return leftIndex, rightIndex
}

//
// getTimeByTransactionGlobalIndex
//  @Description: 通过txid获取到timestamp
//  @param transactionGlobalIndex txid
//  @return time 该txid对应的timestamp
//
func getTimeByTransactionGlobalIndex(transactionGlobalIndex int64) (time string) {
    transactionInfos := getTransactionInfos(transactionGlobalIndex, false, 1)
    blockHash := transactionInfos.Result[0].BlockHash
    block := getBlockByHash(blockHash)
    time = block.Result.Header.Timestamp
    return time
}

func getTransactionInfo(transactionHash string) (transactionInfo TransactionInfo) {
    var body Body
    body.Id = 101
    body.Jsonrpc = "2.0"
    body.Method = "chain.get_transaction_info"
    body.Params = append(body.Params, transactionHash)

    bodyString, err := json.Marshal(body)
    reader := strings.NewReader(string(bodyString))
    request, err := http.NewRequest("POST", "http://52.199.102.11:9850", reader)
    request.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    response, err := client.Do(request)
    defer response.Body.Close()
    x, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("err1")
    }
    err = json.Unmarshal(x, &transactionInfo)
    if err != nil {
        fmt.Println("err2 getTransactionInfo")
        fmt.Println(string(x))
    }
    return transactionInfo
}

//
// getBlockTxnInfos
//  @Description: 通过blockHash获取该block中所有的tx信息
//  @param blockHash
//  @return blockTxnInfos
//
func getBlockTxnInfos(blockHash string) (blockTxnInfos BlockTxnInfos) {
    var body Body
    body.Id = 101
    body.Jsonrpc = "2.0"
    body.Method = "chain.get_block_txn_infos"
    body.Params = append(body.Params, blockHash)

    bodyString, err := json.Marshal(body)
    reader := strings.NewReader(string(bodyString))
    request, err := http.NewRequest("POST", "http://52.199.102.11:9850", reader)
    request.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    response, err := client.Do(request)
    defer response.Body.Close()
    x, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("err1")
    }
    err = json.Unmarshal(x, &blockTxnInfos)
    if err != nil {
        fmt.Println("err2 getBlockTxnInfos")
        fmt.Println(string(x))
    }
    return blockTxnInfos
}

//
// getEventsByTxnHash
//  @Description: 通过txHash获取该tx中所有的events
//  @param txnHash
//  @return eventsByTxnHash
//
func getEventsByTxnHash(txnHash string) (eventsByTxnHash EventsByTxnHash) {
    var body Body
    body.Id = 101
    body.Jsonrpc = "2.0"
    body.Method = "chain.get_events_by_txn_hash"
    body.Params = append(body.Params, txnHash)

    bodyString, err := json.Marshal(body)
    reader := strings.NewReader(string(bodyString))

    request, err := http.NewRequest("POST", "http://52.199.102.11:9850", reader)
    request.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    response, err := client.Do(request)
    defer response.Body.Close()
    x, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("err1")
    }

    err = json.Unmarshal(x, &eventsByTxnHash)
    if err != nil {
        fmt.Println("err2 getEventsByTxnHash")
    }
    return eventsByTxnHash
}

//
// getTransactionInfos
//  @Description: 获取多个tx的详细信息
//  @param startGlobalIndex txid的起始值
//  @param reverse 是否降序
//  @param maxSize 展示tx个数 最大为32
//  @return transactionInfos 多个tx的详细信息
//
func getTransactionInfos(startGlobalIndex int64, reverse bool, maxSize int64) (transactionInfos TransactionInfos) {
    var body Body
    body.Id = 101
    body.Jsonrpc = "2.0"
    body.Method = "chain.get_transaction_infos"
    body.Params = append(body.Params, startGlobalIndex)
    body.Params = append(body.Params, reverse)
    body.Params = append(body.Params, maxSize)

    bodyString, err := json.Marshal(body)
    reader := strings.NewReader(string(bodyString))

    request, err := http.NewRequest("POST", "http://52.199.102.11:9850", reader)
    request.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    response, err := client.Do(request)
    defer response.Body.Close()
    x, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("err1")
    }
    err = json.Unmarshal(x, &transactionInfos)
    if err != nil {
        fmt.Println("err2 getTransactionInfos:", err.Error())
        fmt.Println("params:", body.Params)
        fmt.Println(string(x))
        fmt.Println(transactionInfos)
    }
    return transactionInfos
}

//
// getBlockInfoByNumber
//  @Description: 通过number获取block的详细信息
//  @param number
//  @return blockInfo
//
func getBlockInfoByNumber(number int64) (blockInfo BlockInfo) {
    var body Body
    body.Id = 101
    body.Jsonrpc = "2.0"
    body.Method = "chain.get_block_info_by_number"
    body.Params = append(body.Params, number)

    bodyString, err := json.Marshal(body)
    reader := strings.NewReader(string(bodyString))

    request, err := http.NewRequest("POST", "http://52.199.102.11:9850", reader)
    request.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    response, err := client.Do(request)
    defer response.Body.Close()
    x, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("err1")
    }

    err = json.Unmarshal(x, &blockInfo)
    if err != nil {
        fmt.Println("err2")
    }
    return blockInfo
}

//
// getBlockByNumber
//  @Description: 通过number获取block
//  @param number
//  @return block
//
func getBlockByNumber(number int64) (block Block) {
    var body Body
    body.Id = 101
    body.Jsonrpc = "2.0"
    body.Method = "chain.get_block_by_number"
    body.Params = append(body.Params, number)

    bodyString, err := json.Marshal(body)
    reader := strings.NewReader(string(bodyString))

    request, err := http.NewRequest("POST", "http://52.199.102.11:9850", reader)
    request.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    response, err := client.Do(request)
    defer response.Body.Close()
    x, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("err1")
    }
    err = json.Unmarshal(x, &block)
    if err != nil {
        fmt.Println("err2")
    }
    return block
}

//
// getBlockByHash
//  @Description: 通过blockHash获取block
//  @param blockHash
//  @return block
//
func getBlockByHash(blockHash string) (block Block) {
    var body Body
    body.Id = 101
    body.Jsonrpc = "2.0"
    body.Method = "chain.get_block_by_hash"
    body.Params = append(body.Params, blockHash)
    bodyString, err := json.Marshal(body)
    reader := strings.NewReader(string(bodyString))

    request, err := http.NewRequest("POST", "http://52.199.102.11:9850", reader)
    request.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    response, err := client.Do(request)
    defer response.Body.Close()
    x, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("err1")
    }
    err = json.Unmarshal(x, &block)
    if err != nil {
        fmt.Println("err2 getBlockByHash")
    }
    return block
}

//
// getNodeInfo
//  @Description: 获取当前节点信息
//  @return nodeInfo
//
func getNodeInfo() (nodeInfo NodeInfo) {
    var body Body
    body.Id = 101
    body.Jsonrpc = "2.0"
    body.Method = "node.info"

    bodyString, err := json.Marshal(body)
    reader := strings.NewReader(string(bodyString))

    request, err := http.NewRequest("POST", "http://52.199.102.11:9850", reader)
    request.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    response, err := client.Do(request)
    defer response.Body.Close()
    x, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("err1")
    }

    err = json.Unmarshal(x, &nodeInfo)
    if err != nil {
        fmt.Println("err2 getNodeInfo")
    }
    return nodeInfo
}
