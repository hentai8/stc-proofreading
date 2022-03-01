package stc

import (
    "encoding/hex"
    "encoding/json"
    "fmt"
    //"github.com/jmoiron/sqlx"
    "io/ioutil"
    "math"
    "net/http"
    mcsql "stc-proofreading-go/tools/mysql"
    "strconv"
    "strings"
)

type Node struct {

}

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

type Cache struct {
    Id      int64  `db:"id"`
    Txid    int64  `db:"txid"`
    Txhash  string `db:"txhash"`
    Event   string `db:"event"`
    Address string `db:"address"`
    Amount  int64  `db:"amount"`
}

//
// GetTimestampBoundary
//  @Description: 获取timestamp的边界值
//  @return newestTime 最新block中的timestamp
//  @return oldestTime 0号block中的timestamp
//  @return newestTransactionGlobalIndexInt 最新block中的txid
//
func (n *Node)GetTimestampBoundary() (newestTime int64, oldestTime int64, newestTransactionGlobalIndexInt int64) {
    // 获取最新的block和最旧的block的timestamp
    nodeInfo := n.GetNodeInfo()
    //newestNumber := nodeInfo.Result.PeerInfo.ChainInfo.Head.Number
    newestBlockHash := nodeInfo.Result.PeerInfo.ChainInfo.Head.BlockHash
    blockTxnInfos := n.GetBlockTxnInfos(newestBlockHash)

    newestTXHash := blockTxnInfos.Result[0].TransactionHash

    newestTransactionInfo := n.GetTransactionInfo(newestTXHash)
    //fmt.Println(newestTransactionInfo)
    newestTransactionGlobalIndex := newestTransactionInfo.Result.TransactionGlobalIndex

    //fmt.Println(newestTransactionGlobalIndex)
    newestTransactionGlobalIndexInt, err := strconv.ParseInt(newestTransactionGlobalIndex, 10, 64)
    fmt.Println("newestTransactionGlobalIndexInt", newestTransactionGlobalIndexInt)
    if err != nil {
        fmt.Println("strconv.ParseInt err")
    }

    newestTimeString := n.GetTimeByTransactionGlobalIndex(newestTransactionGlobalIndexInt)
    oldestTimeString := n.GetTimeByTransactionGlobalIndex(0)

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
// GetOutputSumByIndexPeriod
//  @Description: 通过起始的txid和结束的txid获取该address的所有支出的统计
//  @param startIndex 起始的txid
//  @param endIndex 结束的txid
//  @param address 带salt的address地址
//  @return OutputSum 统计出的出帐总值
//
func (n *Node)GetOutputSumByIndexPeriod(startIndex int64, endIndex int64, address string) (OutputSum int64) {
    if startIndex > endIndex {
        x := startIndex
        startIndex = endIndex
        endIndex = x
    }
    leftIndex := startIndex
    rightIndex := endIndex
    for rightIndex > leftIndex {
        // 判断数据库中是否有该address在该txid的交易缓存
        outputExist, err := n.IsExistInMysql(leftIndex, "WithdrawEvent", address)
        if err == true {
            //fmt.Println("is exist")
            OutputSum = OutputSum + outputExist
            leftIndex = leftIndex + 1
        } else {
            if rightIndex-leftIndex > 32 {
                // 查看是否存在相同的txid,event,address存储在mysql数据库中
                transactionInfos := n.GetTransactionInfos(leftIndex, false, 32)
                var offset int64
                offset = 0
                for _, transactionInfoResult := range transactionInfos.Result {
                    signal := false
                    events := n.GetEventsByTxnHash(transactionInfoResult.TransactionHash)
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
                            //fmt.Println("txid:", leftIndex+offset)
                            n.SaveIntoMysql(leftIndex+offset, event.TransactionHash, "WithdrawEvent", address, output)
                            signal = true
                        }
                    }
                    if signal == false {
                        //fmt.Println("txid:", leftIndex+offset)
                        n.SaveIntoMysql(leftIndex+offset, transactionInfoResult.TransactionHash, "WithdrawEvent", address, 0)
                    }
                    offset++
                }
                leftIndex = leftIndex + 32
            } else if rightIndex-leftIndex > 0 {
                transactionInfos := n.GetTransactionInfos(leftIndex, false, rightIndex-leftIndex)
                var offset int64
                offset = 0
                for _, transactionInfoResult := range transactionInfos.Result {
                    signal := false
                    events := n.GetEventsByTxnHash(transactionInfoResult.TransactionHash)
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
                            n.SaveIntoMysql(leftIndex+offset, event.TransactionHash, "WithdrawEvent", address, output)
                            signal = true
                        }
                    }
                    if signal == false {
                        n.SaveIntoMysql(leftIndex+offset, transactionInfoResult.TransactionHash, "WithdrawEvent", address, 0)
                    }
                    offset++
                }
                leftIndex = rightIndex
            }
        }
    }
    return OutputSum
}

//
// GetCoinbaseSumByIndexPeriod
//  @Description: 通过起始的txid和结束的txid获取该address的所有coinbase收入的统计
//  @param startIndex 起始的txid
//  @param endIndex 结束的txid
//  @param address 带salt的address地址
//  @return OutputSum 统计出的coinbase收入总值
//
func (n *Node)GetCoinbaseSumByIndexPeriod(startIndex int64, endIndex int64, address string) (CoinbaseSum int64) {
    if startIndex > endIndex {
        x := startIndex
        startIndex = endIndex
        endIndex = x
    }
    leftIndex := startIndex
    rightIndex := endIndex
    for rightIndex > leftIndex {
        // 判断数据库中是否有该address在该txid的交易缓存
        coinbaseExist, err := n.IsExistInMysql(leftIndex, "DepositEvent", address)
        if err == true {
            //fmt.Println("txid:", leftIndex, "is exist")
            CoinbaseSum = CoinbaseSum + coinbaseExist
            leftIndex = leftIndex + 1
        } else {
            if rightIndex-leftIndex > 32 {
                transactionInfos := n.GetTransactionInfos(leftIndex, false, 32)
                var offset int64
                offset = 0
                for _, transactionInfoResult := range transactionInfos.Result {
                    signal := false
                    events := n.GetEventsByTxnHash(transactionInfoResult.TransactionHash)
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
                            //fmt.Println("txid:", leftIndex+offset)
                            n.SaveIntoMysql(leftIndex+offset, event.TransactionHash, "DepositEvent", address, coinbase)
                            signal = true
                        }
                    }
                    if signal == false {
                        //fmt.Println("txid:", leftIndex+offset)
                        n.SaveIntoMysql(leftIndex+offset, transactionInfoResult.TransactionHash, "DepositEvent", address, 0)
                    }
                    offset++
                }
                leftIndex = leftIndex + 32
            } else if rightIndex-leftIndex > 0 {
                transactionInfos := n.GetTransactionInfos(leftIndex, false, rightIndex-leftIndex)
                var offset int64
                offset = 0
                for _, transactionInfoResult := range transactionInfos.Result {
                    signal := false
                    events := n.GetEventsByTxnHash(transactionInfoResult.TransactionHash)
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
                            n.SaveIntoMysql(leftIndex+offset, event.TransactionHash, "DepositEvent", address, coinbase)
                            signal = true
                        }
                    }
                    if signal == false {
                        n.SaveIntoMysql(leftIndex+offset, transactionInfoResult.TransactionHash, "DepositEvent", address, 0)
                    }
                    offset++
                }
                leftIndex = rightIndex
            }
        }
    }
    return CoinbaseSum
}

//
// FindTransactionByTime
//  @Description: 通过timestamp寻找到Transaction的txid,采用二分法快速查询
//  @param time 目标时间戳
//  @param oldestTransactionGlobalIndex 最旧的txid
//  @param newestTransactionGlobalIndex 最新的txid
//  @return leftIndex 该时间戳对应的前一个txid
//  @return rightIndex 该时间戳对应的后一个txid
//
func (n *Node)FindTransactionByTime(time int64, oldestTransactionGlobalIndex int64, newestTransactionGlobalIndex int64) (leftIndex int64, rightIndex int64) {
    leftIndex = oldestTransactionGlobalIndex
    rightIndex = newestTransactionGlobalIndex

    for leftIndex <= rightIndex {

        middleIndex := (leftIndex + rightIndex) / 2
        middleTransactionInfos := n.GetTransactionInfos(middleIndex, false, 1)
        middleBlockHash := middleTransactionInfos.Result[0].BlockHash
        middleBlock := n.GetBlockByHash(middleBlockHash)
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
    //fmt.Println("leftIndex:", leftIndex)
    //fmt.Println("rightIndex:", rightIndex)
    return leftIndex, rightIndex
}

//
// GetTimeByTransactionGlobalIndex
//  @Description: 通过txid获取到timestamp
//  @param transactionGlobalIndex txid
//  @return time 该txid对应的timestamp
//
func (n *Node)GetTimeByTransactionGlobalIndex(transactionGlobalIndex int64) (time string) {
    transactionInfos := n.GetTransactionInfos(transactionGlobalIndex, false, 1)
    blockHash := transactionInfos.Result[0].BlockHash
    block := n.GetBlockByHash(blockHash)
    time = block.Result.Header.Timestamp
    return time
}

func (n *Node)GetTransactionInfo(transactionHash string) (transactionInfo TransactionInfo) {
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
        fmt.Println("err2 GetTransactionInfo")
        fmt.Println(string(x))
    }
    return transactionInfo
}

//
// GetBlockTxnInfos
//  @Description: 通过blockHash获取该block中所有的tx信息
//  @param blockHash
//  @return blockTxnInfos
//
func (n *Node)GetBlockTxnInfos(blockHash string) (blockTxnInfos BlockTxnInfos) {
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
        fmt.Println("err2 GetBlockTxnInfos")
        fmt.Println(string(x))
    }
    return blockTxnInfos
}

//
// GetEventsByTxnHash
//  @Description: 通过txHash获取该tx中所有的events
//  @param txnHash
//  @return eventsByTxnHash
//
func (n *Node)GetEventsByTxnHash(txnHash string) (eventsByTxnHash EventsByTxnHash) {
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
        fmt.Println("err2 GetEventsByTxnHash")
    }
    return eventsByTxnHash
}

//
// GetTransactionInfos
//  @Description: 获取多个tx的详细信息
//  @param startGlobalIndex txid的起始值
//  @param reverse 是否降序
//  @param maxSize 展示tx个数 最大为32
//  @return transactionInfos 多个tx的详细信息
//
func (n *Node)GetTransactionInfos(startGlobalIndex int64, reverse bool, maxSize int64) (transactionInfos TransactionInfos) {
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
        fmt.Println("err2 GetTransactionInfos:", err.Error())
        fmt.Println("params:", body.Params)
        fmt.Println(string(x))
        fmt.Println(transactionInfos)
    }
    return transactionInfos
}

//
// GetBlockInfoByNumber
//  @Description: 通过number获取block的详细信息
//  @param number
//  @return blockInfo
//
func (n *Node)GetBlockInfoByNumber(number int64) (blockInfo BlockInfo) {
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
// GetBlockByNumber
//  @Description: 通过number获取block
//  @param number
//  @return block
//
func (n *Node)GetBlockByNumber(number int64) (block Block) {
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
// GetBlockByHash
//  @Description: 通过blockHash获取block
//  @param blockHash
//  @return block
//
func (n *Node)GetBlockByHash(blockHash string) (block Block) {
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
        fmt.Println("err2 GetBlockByHash")
    }
    return block
}

//
// GetNodeInfo
//  @Description: 获取当前节点信息
//  @return nodeInfo
//
func (n *Node)GetNodeInfo() (nodeInfo NodeInfo) {
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
        fmt.Println("err2 GetNodeInfo")
    }
    return nodeInfo
}

func (n *Node)IsExistInMysql(txid int64, event string, address string) (amount int64, exist bool) {
    //startTime := time.Now().UnixNano()
    db := mcsql.GetInstance()
    sql := `select amount from stc_cache where txid = ? and event = ? and address = ?`
    rows, err := db.Query(sql, txid, event, address)
    if err != nil {
        fmt.Println("record to mysql err=", err)
    }
    amount = 1
    for rows.Next() {
        err = rows.Scan(&amount)
    }
    //endTime := time.Now().UnixNano()
    //fmt.Println("time:", (endTime-startTime)/1000000)
    if amount != 1 {
        return amount, true
    } else {
        return 0, false
    }
}

func (n *Node)IsExistInMemoryCache(caches []Cache, txid int64, event string, address string) (amount int64, exist bool) {
    for _, cache := range caches {
        if cache.Txid == txid && cache.Event == event && cache.Address == address {
            return cache.Amount, true
        }
    }
    return 0, false
}

func (n *Node)SaveIntoCacheFromMysql(startTxid int64, endTxid int64) (caches []Cache, err error) {
    //startTime := time.Now().UnixNano()
    db := mcsql.GetInstance()
    //db.Select(&caches, `select * from stc_cache where txid >= $1 and txid < $2`, startTxid, endTxid)
    err = db.Select(&caches, `select * from stc_cache where txid>=? and txid<?`, startTxid, endTxid)
    if err != nil {
        fmt.Println(err)
    }
    return caches, err

    //sql := `select * from stc_cache where txid>4699825`
    //rows, err := db.Queryx(sql)
    //if err != nil {
    //    fmt.Println("record to mysql err=", err)
    //}
    //results := make(map[string]interface{})
    //for rows.Next() {
    //    err = rows.MapScan(results)
    //}
    //
    //fmt.Println(results)

    //endTime := time.Now().UnixNano()
    //fmt.Println("time:", (endTime-startTime)/1000000)
    //if amount != 1 {
    //    return amount, true
    //} else {
    //    return 0, false
    //}
}

func (n *Node)SaveIntoMysql(txid int64, txhash string, event string, address string, amount int64) {
    db := mcsql.GetInstance()
    //fmt.Println("SaveIntoMysql:", amount)
    sql := `insert into stc_cache (txid, txhash, event, address, amount) values (?,?,?,?,?)`
    _, err := db.Exec(sql, txid, txhash, event, address, amount)
    if err != nil {
        fmt.Println("save into cache err: ", err.Error())
    }
    return
}

//
// GetCoinbaseSumByIndexPeriod
//  @Description: 通过起始的txid和结束的txid获取该address的所有coinbase收入的统计
//  @param startIndex 起始的txid
//  @param endIndex 结束的txid
//  @param address 带salt的address地址
//  @return OutputSum 统计出的coinbase收入总值
//
func (n *Node)GetCoinbaseSumByIndexPeriodDeparted(startIndex int64, endIndex int64, address string) (CoinbaseSum int64) {
    if startIndex > endIndex {
        x := startIndex
        startIndex = endIndex
        endIndex = x
    }
    leftIndex := startIndex
    rightIndex := endIndex
    count := 100
    sum := rightIndex - leftIndex
    c := make(chan struct{}, count)
    sc := make(chan struct{}, sum)
    defer close(c)
    defer close(sc)
    for rightIndex > leftIndex {
        c <- struct{}{}
        go func(leftIndex int64, address string) {
            // 判断数据库中是否有该address在该txid的交易缓存
            coinbaseExist, err := n.IsExistInMysql(leftIndex, "DepositEvent", address)
            if err == true {
                //fmt.Println("txid:", leftIndex, "is exist")
                CoinbaseSum = CoinbaseSum + coinbaseExist
                leftIndex = leftIndex + 1
            } else {
                fmt.Println("not in cache")
            }
            <-c
            sc <- struct{}{}
        }(leftIndex, address)
        leftIndex++
    }
    for i := sum; i > 0; i-- {
        <-sc
    }
    return CoinbaseSum
}

//
// GetCoinbaseSumByIndexPeriodNew
//  @Description: 通过起始的txid和结束的txid获取该address的所有coinbase收入的统计
//  @param startIndex 起始的txid
//  @param endIndex 结束的txid
//  @param address 带salt的address地址
//  @return OutputSum 统计出的coinbase收入总值
//
func (n *Node)GetCoinbaseSumByIndexPeriodNew(startIndex int64, endIndex int64, address string) (CoinbaseSum int64) {
    if startIndex > endIndex {
        x := startIndex
        startIndex = endIndex
        endIndex = x
    }
    leftIndex := startIndex
    rightIndex := startIndex + 1000
    //cache := 1000

    // 在这个位置重新定义leftIndex和rightIndex,对caches进行切片

    for rightIndex <= endIndex {
        //fmt.Println("leftIndex:", leftIndex)
        //fmt.Println("rightIndex:", rightIndex)
        caches, err := n.SaveIntoCacheFromMysql(leftIndex, rightIndex)
        if err != nil {
            fmt.Println("SaveIntoCacheFromMysql err")
        }
        // 判断内存中是否有该address在该txid的交易缓存
        for rightIndex > leftIndex {
            coinbaseExist, exist := n.IsExistInMemoryCache(caches, leftIndex, "DepositEvent", address)
            if exist == true {
                //fmt.Println("txid:", leftIndex, "is exist")
                CoinbaseSum = CoinbaseSum + coinbaseExist
                leftIndex = leftIndex + 1
            } else {
                if rightIndex-leftIndex > 32 {
                    transactionInfos := n.GetTransactionInfos(leftIndex, false, 32)
                    var offset int64
                    offset = 0
                    for _, transactionInfoResult := range transactionInfos.Result {
                        signal := false
                        events := n.GetEventsByTxnHash(transactionInfoResult.TransactionHash)
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
                                //fmt.Println("txid:", leftIndex+offset)
                                n.SaveIntoMysql(leftIndex+offset, event.TransactionHash, "DepositEvent", address, coinbase)
                                signal = true
                            }
                        }
                        if signal == false {
                            //fmt.Println("txid:", leftIndex+offset)
                            n.SaveIntoMysql(leftIndex+offset, transactionInfoResult.TransactionHash, "DepositEvent", address, 0)
                        }
                        offset++
                    }
                    leftIndex = leftIndex + 32
                } else if rightIndex-leftIndex > 0 {
                    transactionInfos := n.GetTransactionInfos(leftIndex, false, rightIndex-leftIndex)
                    var offset int64
                    offset = 0
                    for _, transactionInfoResult := range transactionInfos.Result {
                        signal := false
                        events := n.GetEventsByTxnHash(transactionInfoResult.TransactionHash)
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
                                n.SaveIntoMysql(leftIndex+offset, event.TransactionHash, "DepositEvent", address, coinbase)
                                signal = true
                            }
                        }
                        if signal == false {
                            n.SaveIntoMysql(leftIndex+offset, transactionInfoResult.TransactionHash, "DepositEvent", address, 0)
                        }
                        offset++
                    }
                    leftIndex = rightIndex
                }
            }
        }
        if rightIndex == endIndex {
            break
        } else if rightIndex+1000 >= endIndex {
            rightIndex = endIndex
        } else {
            rightIndex = rightIndex + 1000
        }
    }
    return CoinbaseSum
}

//
// GetOutputSumByIndexPeriodNew
//  @Description: 通过起始的txid和结束的txid获取该address的所有支出的统计
//  @param startIndex 起始的txid
//  @param endIndex 结束的txid
//  @param address 带salt的address地址
//  @return OutputSum 统计出的出帐总值
//
func (n *Node)GetOutputSumByIndexPeriodNew(startIndex int64, endIndex int64, address string) (OutputSum int64) {
    if startIndex > endIndex {
        x := startIndex
        startIndex = endIndex
        endIndex = x
    }
    leftIndex := startIndex
    rightIndex := startIndex + 1000
    //cache := 1000

    // 在这个位置重新定义leftIndex和rightIndex,对caches进行切片

    for rightIndex <= endIndex {
        //fmt.Println("leftIndex:", leftIndex)
        //fmt.Println("rightIndex:", rightIndex)
        caches, err := n.SaveIntoCacheFromMysql(leftIndex, rightIndex)
        if err != nil {
            fmt.Println("SaveIntoCacheFromMysql err")
        }
        // 判断内存中是否有该address在该txid的交易缓存
        for rightIndex > leftIndex {
            ouputExist, exist := n.IsExistInMemoryCache(caches, leftIndex, "WithdrawEvent", address)
            if exist == true {
                //fmt.Println("txid:", leftIndex, "is exist")
                OutputSum = OutputSum + ouputExist
                leftIndex = leftIndex + 1
            } else {
                if rightIndex-leftIndex > 32 {
                    transactionInfos := n.GetTransactionInfos(leftIndex, false, 32)
                    var offset int64
                    offset = 0
                    for _, transactionInfoResult := range transactionInfos.Result {
                        signal := false
                        events := n.GetEventsByTxnHash(transactionInfoResult.TransactionHash)
                        for _, event := range events.Result {
                            if event.TypeTag == "0x00000000000000000000000000000001::Account::WithdrawEvent" && event.EventKey == address {
                                var ouput int64
                                ouputAmount := event.Data[2:34]
                                //fmt.Println("ouputAmount:", ouputAmount)
                                //fmt.Println("event.Data:", event.Data)
                                for i := 0; i < (len(ouputAmount) / 2); i++ {
                                    num1String := ouputAmount[0+2*i : 2+2*i]
                                    num1, err := hex.DecodeString(num1String)
                                    if err != nil {
                                        fmt.Println("strconv.ParseInt err")
                                    }
                                    ouput = int64(num1[0])*int64(math.Pow(256, float64(i))) + ouput
                                }
                                OutputSum = OutputSum + ouput
                                //fmt.Println("txid:", leftIndex+offset)
                                n.SaveIntoMysql(leftIndex+offset, event.TransactionHash, "WithdrawEvent", address, ouput)
                                signal = true
                            }
                        }
                        if signal == false {
                            //fmt.Println("txid:", leftIndex+offset)
                            n.SaveIntoMysql(leftIndex+offset, transactionInfoResult.TransactionHash, "WithdrawEvent", address, 0)
                        }
                        offset++
                    }
                    leftIndex = leftIndex + 32
                } else if rightIndex-leftIndex > 0 {
                    transactionInfos := n.GetTransactionInfos(leftIndex, false, rightIndex-leftIndex)
                    var offset int64
                    offset = 0
                    for _, transactionInfoResult := range transactionInfos.Result {
                        signal := false
                        events := n.GetEventsByTxnHash(transactionInfoResult.TransactionHash)
                        for _, event := range events.Result {
                            if event.TypeTag == "0x00000000000000000000000000000001::Account::WithdrawEvent" && event.EventKey == address {
                                var ouput int64
                                ouputAmount := event.Data[2:34]
                                //fmt.Println("ouputAmount:", ouputAmount)
                                //fmt.Println("event.Data:", event.Data)
                                for i := 0; i < (len(ouputAmount) / 2); i++ {
                                    num1String := ouputAmount[0+2*i : 2+2*i]
                                    num1, err := hex.DecodeString(num1String)
                                    if err != nil {
                                        fmt.Println("strconv.ParseInt err")
                                    }
                                    ouput = int64(num1[0])*int64(math.Pow(256, float64(i))) + ouput
                                }
                                OutputSum = OutputSum + ouput
                                n.SaveIntoMysql(leftIndex+offset, event.TransactionHash, "WithdrawEvent", address, ouput)
                                signal = true
                            }
                        }
                        if signal == false {
                            n.SaveIntoMysql(leftIndex+offset, transactionInfoResult.TransactionHash, "WithdrawEvent", address, 0)
                        }
                        offset++
                    }
                    leftIndex = rightIndex
                }
            }
        }
        if rightIndex == endIndex {
            break
        } else if rightIndex+1000 >= endIndex {
            rightIndex = endIndex
        } else {
            rightIndex = rightIndex + 1000
        }
    }
    return OutputSum
}
