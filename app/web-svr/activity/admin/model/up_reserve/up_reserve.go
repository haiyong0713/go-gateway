package up_reserve

import (
	xtime "go-common/library/time"
)

type ParamList struct {
	Sid int64 `form:"sid"`
	Mid int64 `form:"mid"`
	Pn  int   `form:"pn" default:"1" validate:"min=1"`
	Ps  int   `form:"ps" default:"15" validate:"min=1"`
}

type UpReserveList struct {
	ID                int64      `json:"id" gorm:"column:id"`
	Sid               int64      `json:"sid" gorm:"column:sid"`
	Mid               int64      `json:"mid" gorm:"column:mid"`
	Oid               string     `json:"oid" gorm:"column:oid"`
	Type              int64      `json:"type" gorm:"column:type"`
	State             int64      `json:"state" gorm:"column:state"`
	Ctime             xtime.Time `json:"ctime" time_format:"2006-01-02 15:04:05" gorm:"column:ctime"`
	Mtime             xtime.Time `json:"mtime" time_format:"2006-01-02 15:04:05" gorm:"column:mtime"`
	LivePlanStartTime xtime.Time `json:"live_plan_start_time" time_format:"2006-01-02 15:04:05" gorm:"column:live_plan_start_time"`
	Audit             int64      `json:"audit" gorm:"column:audit"`
	AuditChannel      int64      `json:"audit_channel" gorm:"column:audit_channel"`
	DynamicID         string     `json:"dynamic_id" gorm:"column:dynamic_id"`
	DynamicAudit      int64      `json:"dynamic_audit" gorm:"column:dynamic_audit"`
	LotteryType       int64      `json:"lottery_type" gorm:"column:lottery_type"`
	LotteryID         string     `json:"lottery_id" gorm:"column:lottery_id"`
	LotteryAudit      int64      `json:"lottery_audit" gorm:"column:lottery_audit"`
}

type UpReserveListReply struct {
	List  []*UpReserveList `json:"list"`
	Num   int              `json:"num"`
	Size  int              `json:"size"`
	Total int64            `json:"total"`
}

type ParamHang struct {
	Mid      string `form:"mid" validate:"required"`
	Sid      int64  `form:"sid" validate:"min=1"`
	Operator string `form:"operator" validate:"required"`
}

type CreateHangLog struct {
	Operator string `json:"operator" gorm:"column:operator"`
	Type     int64  `json:"type" gorm:"column:type"`
	Detail   string `json:"detail" gorm:"column:detail"`
	Result   string `json:"result" gorm:"column:result"`
	Remark   string `json:"remark" gorm:"column:remark"`
	Sid      int64  `json:"sid" gorm:"column:sid"`
}

func (CreateHangLog) TableName() string {
	return "up_act_reserve_hang_log"
}

type HangLogListParams struct {
	Sid int64 `form:"sid" validate:"min=1"`
	Pn  int64 `form:"pn" default:"1" validate:"min=1"`
	Ps  int64 `form:"ps" default:"15" validate:"min=1"`
}

type HangLogListReply struct {
	List  []*HangLogItem `json:"list"`
	Pager *Pager         `json:"pager"`
}

type HangLogItem struct {
	ID int64 `json:"id" gorm:"column:id"`
	CreateHangLog
	Ctime xtime.Time `json:"ctime" gorm:"column:ctime"`
}

type Pager struct {
	Num   int64 `json:"num"`
	Size  int64 `json:"size"`
	Total int64 `json:"total"`
}
