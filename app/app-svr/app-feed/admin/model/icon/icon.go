package icon

import (
	"go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/show"
)

const (
	StateDel       = 0
	StateNormal    = 1
	EffectGroupMid = 3
)

// IconSaveReply .
type IconSaveReply struct {
	ID int64 `json:"id,omitempty"`
}

// HiddenSaveParam .
type IconSaveParam struct {
	ID          int64     `form:"id" default:"0" validate:"min=0"`
	Module      string    `form:"module" validate:"required"`
	Icon        string    `form:"icon" validate:"required"`
	GlobalRed   int32     `form:"global_red_dot"`
	EffectGroup int32     `form:"effect_group" validate:"required"`
	EffectURL   string    `form:"effect_url"`
	Stime       time.Time `form:"stime" validate:"required"`
	Etime       time.Time `form:"etime" validate:"required"`
}

// ListParam
type ListParam struct {
	Pn int `form:"pn" default:"1" validate:"min=1"`
	Ps int `form:"ps" default:"15" validate:"min=1"`
}

// ListReply .
type ListReply struct {
	List []*IconInfo `json:"list"`
	Page *Page       `json:"page"`
}

type Page struct {
	Num   int   `json:"num"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

// IconInfo .
type IconInfo struct {
	*Icon
	ModuleInfo []*ModuleInfo `json:"module_info,omitempty"`
}

// Icon .
type Icon struct {
	ID          int64     `gorm:"column:id" json:"id"`
	Module      string    `gorm:"column:module" json:"module"`
	Icon        string    `gorm:"column:icon" json:"icon"`
	GlobalRed   int32     `gorm:"column:global_red_dot" json:"global_red_dot"`
	EffectGroup int32     `gorm:"column:effect_group" json:"effect_group"`
	EffectURL   string    `gorm:"column:effect_url" json:"effect_url"`
	Operator    string    `gorm:"column:operator" json:"operator"`
	State       int32     `gorm:"column:state" json:"state"`
	Stime       time.Time `gorm:"column:stime" json:"stime"`
	Etime       time.Time `gorm:"column:etime" json:"etime"`
}

// Module is original module
type Module struct {
	Oid  int64 `json:"oid"`
	Plat int32 `json:"plat"`
}

// ModuleInfo is combine sidebar info
type ModuleInfo struct {
	Oid   int64                `json:"oid"`
	Plat  int32                `json:"plat"`
	Name  string               `json:"name"`
	Limit []*show.SidebarLimit `json:"limit"`
}

// Limit
type Limit struct {
	Conditions string `json:"conditions"`
	Build      int32  `json:"build"`
}

// TableName Icon
func (a Icon) TableName() string {
	return "mng_icon"
}
