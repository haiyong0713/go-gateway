package lottery

import (
	"go-common/library/time"
)

// VIPInfo VIP information
type VIPInfo struct {
	ID            int    `json:"id"`
	Platform      int    `json:"platform"`
	ProductName   string `json:"product_name"`
	ProductID     string `json:"product_id"`
	SuitType      int    `json:"suit_type"`
	Month         int    `json:"month"`
	SubType       int    `json:"sub_type"`
	OriginalPrice int    `json:"original_price"`
	Selected      int    `json:"selected"`
	Remark        string `json:"remark"`
	Status        int    `json:"status"`
	Operator      string `json:"operator"`
	OperID        int    `json:"oper_id"`
	Ctime         int    `json:"ctime"`
	Mtime         int    `json:"mtime"`
	ActToken      string `json:"act_token"`
	OnSale        int    `json:"on_sale"`
}

// BatchInfo
type BatchInfo struct {
	ID             int       `json:"id"`
	PoolID         int       `json:"pool_id"`
	Unit           int       `json:"unit"`
	Count          int       `json:"count"`
	Ver            int       `json:"ver"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	SurplusCount   int       `json:"surplus_count"`
	CodeUseCount   int       `json:"code_use_count"`
	DirectUseCount int       `json:"direct_use_count"`
	Token          string    `json:"token"`
}

// GouponInfo
type CouponInfo struct {
	ID                  int       `json:"id"`
	AppID               int       `json:"app_id"`
	AppName             string    `json:"app_name"`
	BatchName           string    `json:"batch_name"`
	BatchToken          string    `json:"batch_token"`
	MaxCount            int       `json:"max_count"`
	CurrentCount        int       `json:"current_count"`
	StartTime           time.Time `json:"start_time"`
	ExpireTime          time.Time `json:"expire_time"`
	Operator            string    `json:"operator"`
	LimitCount          int       `json:"limit_count"`
	ProductLimitExplain string    `json:"product_limit_explain"`
	PlatfromLimit       []int     `json:"platfrom_limit"`
	UseLimitExplain     string    `json:"use_limit_explain"`
	State               int       `json:"state"`
	ProductLimitMonth   int       `json:"product_limit_month"`
	ProductLimitRenewal int       `json:"product_limit_Renewal"`
	ActivateStart       time.Time `json:"activate_start"`
	ActivateEnd         time.Time `json:"activate_end"`
}

// Check vip check response
type CheckRsp struct {
	Check int `json:"check"`
}
