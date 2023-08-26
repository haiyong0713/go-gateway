package hidden

import (
	"go-common/library/ecode"
	"go-common/library/time"
)

const (
	StateOnline  = 1
	StateOffline = 0
	StateDel     = -1
	// PlatAndroid is int8 for android.
	PlatAndroid = int8(0)
	// PlatAndroidI is int8 for Android Global.
	PlatAndroidI = int8(8)
	// condition type.
	ConditionsLt = "lt"
	ConditionsGt = "gt"
	ConditionsEq = "eq"
	ConditionsNe = "ne"

	EntranceHome    = 0
	EntranceChannel = 1
	EntranceSidebar = 2
	EntranceModule  = 3
)

// HiddenSaveReply .
type HiddenSaveReply struct {
	ID int64 `json:"id,omitempty"`
}

// HiddenSaveParam .
type HiddenSaveParam struct {
	ID      int64     `form:"id" default:"0" validate:"min=0"`
	Channel string    `form:"channel" validate:"required"`
	SID     int64     `form:"sid"`
	RID     int64     `form:"rid"`
	CID     int64     `form:"cid"`
	AreaIDs []int64   `form:"area_ids,split" validate:"required"`
	Limit   string    `form:"limit" validate:"required"`
	Stime   time.Time `form:"stime" validate:"required"`
	Etime   time.Time `form:"etime" validate:"required"`
	//渠道包判断条件 include exclude
	HiddenCondition string `form:"hidden_condition" validate:"required"`
	//一级模块id
	ModuleID    int64 `form:"module_id"`
	HideDynamic int64 `form:"hide_dynamic"`
}

func (lt HiddenLimit) ValidateParam() (err error) {
	if lt.Build < 0 {
		err = ecode.RequestErr
		return
	}
	if lt.Plat != PlatAndroid && lt.Plat != PlatAndroidI {
		err = ecode.RequestErr
		return
	}
	if lt.Conditions != ConditionsLt && lt.Conditions != ConditionsGt && lt.Conditions != ConditionsEq && lt.Conditions != ConditionsNe {
		err = ecode.RequestErr
	}
	return
}

// ListParam
type ListParam struct {
	Pn int `form:"pn" default:"1" validate:"min=1"`
	Ps int `form:"ps" default:"15" validate:"min=1"`
}

// ListReply .
type ListReply struct {
	List []*HiddenInfo `json:"list"`
	Page *Page         `json:"page"`
}

type Page struct {
	Num   int   `json:"num"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

// HiddenInfo .
type HiddenInfo struct {
	*Hidden
	SName   string         `json:"s_name,omitempty"`
	RName   string         `json:"r_name,omitempty"`
	CName   string         `json:"c_name,omitempty"`
	MName   string         `json:"m_name,omitempty"`
	AreaIDs []int64        `json:"area_ids"`
	Limit   []*HiddenLimit `json:"limit"`
}

// Hidden .
type Hidden struct {
	ID      int64     `gorm:"column:id" json:"id"`
	SID     int64     `gorm:"column:sid" json:"sid"`
	RID     int64     `gorm:"column:rid" json:"rid"`
	CID     int64     `gorm:"column:cid" json:"cid"`
	Channel string    `gorm:"column:channel" json:"channel"`
	PID     int64     `gorm:"column:pid" json:"pid"`
	State   int       `gorm:"column:state" json:"state"`
	Stime   time.Time `gorm:"column:stime" json:"stime"`
	Etime   time.Time `gorm:"column:etime" json:"etime"`
	//渠道包判断条件
	HiddenCondition string `gorm:"column:hidden_condition" json:"hidden_condition"`
	//一级模块id
	ModuleID int64 `gorm:"column:module_id" json:"module_id"`
	//动态游戏卡是否展示  1表示不展示  0表示展示
	HideDynamic int64 `gorm:"column:hide_dynamic" json:"hide_dynamic"`
}

// TableName Hidden
func (a Hidden) TableName() string {
	return "entrance_hidden"
}

// TableName HiddenLimit
func (a HiddenLimit) TableName() string {
	return "entrance_hidden_limit"
}

type HiddenLimit struct {
	ID         int64  `gorm:"column:id" json:"id"`
	OID        int64  `gorm:"column:oid" json:"oid"`
	Plat       int8   `gorm:"column:plat" json:"plat"`
	Build      int64  `gorm:"column:build" json:"build"`
	Conditions string `gorm:"column:conditions" json:"conditions"`
	State      int    `gorm:"column:state" json:"state"`
}

// Region .
type Region struct {
	ID   int64  `gorm:"column:rid" json:"rid"`
	Name string `gorm:"column:name" json:"name"`
	Plat int    `gorm:"column:plat" json:"plat"`
}

// TableName .
func (a Region) TableName() string {
	return "region_copy"
}

// Entrance .
type Entrance struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Plat  int    `json:"plat"`
}
