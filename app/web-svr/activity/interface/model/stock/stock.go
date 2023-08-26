package stock

import xtime "go-common/library/time"

type ConfItemDB struct {
	ID             int64      `json:"id"`
	ResourceId     string     `json:"resource_id"`
	ResourceVer    int64      `json:"resource_ver"`
	ForeignActId   string     `json:"foreign_act_id"`
	DescribeInfo   string     `json:"describe_info"`
	RulesInfo      string     `json:"rules_info"`
	State          int        `json:"state"`
	StockStartTime xtime.Time `json:"stock_start_time"`
	StockEndTime   xtime.Time `json:"stock_end_time"`
	Ctime          xtime.Time `json:"ctime"`
	Mtime          xtime.Time `json:"mtime"`
}

type StockBaseReq struct {
	StockId  int64
	LimitKey string
}

type StockReq struct {
	StockBaseReq
	GiftVer int64
}

type ConsumerStockReq struct {
	StockId  int64
	StoreVer int64

	TotalStore       int32
	CycleStore       int32
	UserStore        int32
	CycleStoreKeyPre string
	ConsumerStock    int
}

type UserStockCache struct {
	StockId  int64
	LimitKey string
	Mid      int64
}
