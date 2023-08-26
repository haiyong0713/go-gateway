package reward_conf

import (
	xtime "go-common/library/time"
)

type AddOneRewardReq struct {
	ActivityId string     `form:"activity_id" json:"activity_id" validate:"required"`
	AwardId    string     `form:"award_id" json:"award_id" validate:"required"`
	CostType   int8       `form:"cost_type" json:"cost_type" validate:"required"`
	CostValue  int32      `form:"cost_value" json:"cost_value" validate:"required"`
	ShowTime   xtime.Time `form:"show_time"  json:"show_time" validate:"required"`
	EndTime    xtime.Time `form:"end_time"  json:"end_time" validate:"required"`
	Creator    string     `form:"creator" json:"creator"`
	Order      int32      `form:"order" json:"order"`
	StoreNum   int32      `form:"store_num" json:"store_num"`
}

type StockCycleLimit struct {
	CycleType int   `json:"cycle_type"`
	LimitType int   `json:"limit_type"`
	Store     int32 `json:"store"`
}

type UpdateOneRewardReq struct {
	Id         int64      `json:"id" validate:"required"`
	ActivityId string     `json:"activity_id" validate:"required"`
	StockId    int64      `json:"stock_id"`
	AwardId    string     `form:"award_id" json:"award_id"`
	CostType   int8       `form:"cost_type" json:"cost_type"`
	CostValue  int32      `form:"cost_value" json:"cost_value"`
	ShowTime   xtime.Time `form:"show_time"  json:"show_time"`
	EndTime    xtime.Time `form:"end_time"  json:"end_time"`
	Creator    string     `form:"creator" json:"creator"`
	Order      int32      `form:"order" json:"order"`
	StoreNum   int32      `form:"store_num" json:"store_num"`
	Status     int        `json:"status" default:"1"`
}

type SearchReq struct {
	ActivityId string     `form:"activity_id" json:"activity_id" validate:"required"`
	STime      xtime.Time `form:"s_time"  json:"s_time"`
	ETime      xtime.Time `form:"e_time" json:"e_time"`
	CostType   int        `form:"cost_type"`
	Pn         int        `form:"pn" json:"pn" default:"1"`
	Ps         int        `form:"ps" json:"ps" default:"50"`
}

type SearchRes struct {
	List  []*OneAwardRes `json:"list"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
	Total int            `json:"total"`
}

type OneAwardRes struct {
	ID         uint64     `json:"id"`       // 主键
	AwardID    string     `json:"award_id"` // 奖品id/奖池id
	AwardName  string     `json:"award_name"`
	AwardIcon  string     `json:"award_icon"`
	StockID    int32      `json:"stock_id"` // 库存id
	StockNum   int32      `json:"stock_num"`
	CostType   int8       `json:"cost_type"`   // 是否有效 1=抽奖 2=积分兑换
	CostValue  int32      `json:"cost_value"`  // 消耗数量
	ShowTime   xtime.Time `json:"show_time"`   // 展示时间
	Order      int32      `json:"order"`       // 排序
	Creator    string     `json:"creator"`     // 创建人
	Status     int8       `json:"status"`      // 是否有效 1=有效 0=无效
	Ctime      xtime.Time `json:"ctime"`       // 创建时间
	Mtime      xtime.Time `json:"mtime"`       // 修改时间
	ActivityID string     `json:"activity_id"` // 活动唯一标识
	EndTime    xtime.Time `json:"end_time"`    // 展示时间
}
