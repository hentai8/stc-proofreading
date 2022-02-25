package configs

type Config struct {
    StartTime       int64  `json:"start_time"`
    EndTime         int64  `json:"end_time"`
    CoinbaseAddress string `json:"coinbase_address"`
    WithdrawAddress string `json:"withdraw_address"`
    Mysql           struct {
        Username string `json:"username"`
        Password string `json:"password"`
        Network  string `json:"network"`
        Server   string `json:"server"`
        Port     int    `json:"port"`
        Database string `json:"database"`
    } `json:"mysql"`
}
