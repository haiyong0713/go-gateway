package vogue

import (
	"go-common/library/time"
)

const (
	// 商品状态 1进行中
	UserTaskStatusInProgress = 1
	// 商品状态 2填写地址
	UserTaskStatusFillingAddress = 2
	// 商品状态 3已兑换
	UserTaskStatusExchangeDone = 3
)

type PrizeSearch struct {
	Uid int64 `form:"uid" default:"-1"`
	Pn  int64 `form:"pn" default:"1" validate:"min=1"`
	Ps  int64 `form:"ps" default:"15" validate:"min=1,max=50000"`
}

type PrizeExportSearch struct {
	Uid int64 `form:"uid" default:"-1"`
	Pn  int64 `form:"pn" default:"1"`
	Ps  int64 `form:"ps" default:"50000"`
}

// ListRsp
type PrizesListRsp struct {
	List []*PrizeData `json:"list"`
	Page *Page        `json:"page"`
}

// PrizeData 兑换信息
type PrizeData struct {
	ID  int64 `json:"id" gorm:"column:id"`
	Uid int64 `json:"uid" gorm:"column:uid"`
	// 用户昵称，后期拼进来
	NickName       string `json:"nickname" gorm:"-"`
	Goods          int64  `json:"goods" gorm:"column:goods"`
	GoodsState     int64  `json:"goods_state" gorm:"column:goods_state"`
	GoodsAddressId int64  `json:"goods_address_id" gorm:"column:goods_address"`
	// 收货信息
	GoodsAddress string `json:"goods_address" gorm:"-"`
	GoodsAttr    int    `json:"goods_attr" gorm:"column:goods_attr"`
	GoodsName    string `json:"goods_name" gorm:"column:goods_name"`
	// 商品当前设置的积分
	GoodsScoreSetting int `json:"goods_score_setting" gorm:"column:goods_score_setting"`
	// 用户兑换时消耗的积分
	GoodsScore int `json:"goods_score" gorm:"column:goods_score"`
	// 商品类型，从GoodsAttr中提取
	GoodsAttrReal int `json:"goods_type" gorm:"-"`
	// 是否存在异常
	Risk         bool      `json:"risk" gorm:"-"`
	RiskMsg      string    `json:"risk_msg"`
	Ctime        time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime        time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
	ExchangeTime string    `json:"exchange_time" gorm:"-"`
	TimeCost     string    `json:"time_cost" gorm:"-"`
}

// TableName act_vogue_user_task def
func (PrizeData) TableName() string {
	return "act_vogue_user_task"
}
