package menu

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/time"
)

const (
	TabDel    = -1
	TabOnline = 1
	SideType  = 1 //固定导航
	MenuType  = 0 //运营导航模块
	// PlatAndroid is int8 for android.
	PlatAndroid = int8(0)
	// PlatIPhone is int8 for iphone.
	PlatIPhone  = int8(1)
	PlatIPhoneI = int8(5)
	// PlatAndroidI is int8 for Android Global.
	PlatAndroidI = int8(8)
	// condition type.
	ConditionsLt = "lt"
	ConditionsGt = "gt"
	ConditionsEq = "eq"
	ConditionsNe = "ne"
	//attribute bit
	AttrYes = int64(1)
	// attribute bit
	AttrBitImage          = uint(0)
	AttrBitColor          = uint(1)
	AttrBitBgImage        = uint(2)
	AttrBitFollowBusiness = uint(3)
)

// TabSaveReply .
type TabSaveReply struct {
	ID int64 `json:"id,omitempty"`
}

// TabSaveParam .
type TabSaveParam struct {
	ID             int64     `form:"id" default:"0" validate:"min=0"`
	Type           int       `form:"type" validate:"min=0"`
	TabID          int64     `form:"tab_id" validate:"min=1"`
	Attribute      int64     `form:"attribute" validate:"min=1"`
	InactiveIcon   string    `form:"inactive_icon" validate:"max=255"`
	Inactive       int       `form:"inactive" default:"0" validate:"min=0,max=3"`
	InactiveType   int       `form:"inactive_type" default:"0" validate:"min=0,max=1"`
	ActiveIcon     string    `form:"active_icon" validate:"max=255"`
	Active         int       `form:"active" default:"0" validate:"min=0,max=3"`
	ActiveType     int       `form:"active_type" default:"0" validate:"min=0,max=1"`
	TabTopColor    string    `form:"tab_top_color" validate:"max=100"`
	TabMiddleColor string    `form:"tab_middle_color" validate:"max=100"`
	TabBottomColor string    `form:"tab_bottom_color" validate:"max=100"`
	BgImage1       string    `form:"bg_image1" validate:"max=255"`
	BgImage2       string    `form:"bg_image2" validate:"max=255"`
	FontColor      string    `form:"font_color" validate:"max=100"`
	BarColor       int       `form:"bar_color" default:"0" validate:"min=0,max=1"`
	Stime          time.Time `form:"stime" validate:"required"`
	Etime          time.Time `form:"etime" validate:"required"`
	Limit          string    `form:"limit" validate:"required"`
}

// BuildLimit .
type BuildLimit struct {
	Type       int    `json:"type"`
	Plat       int8   `json:"plat"`
	Build      int64  `json:"build"`
	Conditions string `json:"conditions"`
}

func (lt BuildLimit) ValidateParam() (err error) {
	if lt.Type < 0 || lt.Build < 0 {
		err = ecode.RequestErr
		return
	}
	if lt.Plat != PlatAndroid && lt.Plat != PlatIPhone && lt.Plat != PlatIPhoneI && lt.Plat != PlatAndroidI {
		err = ecode.RequestErr
		return
	}
	if lt.Conditions != ConditionsLt && lt.Conditions != ConditionsGt && lt.Conditions != ConditionsEq && lt.Conditions != ConditionsNe {
		err = ecode.RequestErr
	}
	return
}

type ListParam struct {
	TabID int64 `form:"tab_id" validate:"min=0"`
	Pn    int   `form:"pn" default:"1" validate:"min=1"`
	Ps    int   `form:"ps" default:"15" validate:"min=1"`
}

// ListReply .
type ListReply struct {
	List []*TabList `json:"list"`
	Page *Page      `json:"page"`
}

type Page struct {
	Num   int   `json:"num"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

// TabList .
type TabList struct {
	*TabExt
	MenuName string      `json:"menu_name"`
	Plat     int         `json:"plat"`
	Limit    []*TabLimit `json:"limit"`
}

// TabExt .
type TabExt struct {
	ID             int64     `gorm:"column:id" json:"id"`
	Type           int       `gorm:"column:type" json:"type"`
	TabID          int64     `gorm:"column:tab_id" json:"tab_id"`
	Attribute      int64     `gorm:"column:attribute" json:"attribute"`
	InactiveIcon   string    `gorm:"column:inactive_icon" json:"inactive_icon"`
	Inactive       int       `gorm:"column:inactive" json:"inactive"`
	InactiveType   int       `gorm:"column:inactive_type" json:"inactive_type"`
	ActiveIcon     string    `gorm:"column:active_icon" json:"active_icon"`
	Active         int       `gorm:"column:active" json:"active"`
	ActiveType     int       `gorm:"column:active_type" json:"active_type"`
	TabTopColor    string    `gorm:"column:tab_top_color" json:"tab_top_color"`
	TabMiddleColor string    `gorm:"column:tab_middle_color" json:"tab_middle_color"`
	TabBottomColor string    `gorm:"column:tab_bottom_color" json:"tab_bottom_color"`
	BgImage1       string    `gorm:"column:bg_image1" json:"bg_image1"`
	BgImage2       string    `gorm:"column:bg_image2" json:"bg_image2"`
	FontColor      string    `gorm:"column:font_color" json:"font_color"`
	BarColor       int       `gorm:"column:bar_color" json:"bar_color"`
	State          int       `gorm:"column:state" json:"state"`
	Ctime          time.Time `gorm:"column:ctime" json:"ctime"`
	Mtime          time.Time `gorm:"column:mtime" json:"mtime"`
	Stime          time.Time `gorm:"column:stime" json:"stime"`
	Etime          time.Time `gorm:"column:etime" json:"etime"`
	Operator       string    `gorm:"column:operator" json:"operator"`
	Ver            string    `gorm:"column:ver" json:"ver"`
}

// AttrVal get attr val by bit.
func (a *TabExt) AttrVal(bit uint) int64 {
	return (a.Attribute >> bit) & int64(1)
}

// TableName .
func (a TabExt) TableName() string {
	return "tab_ext"
}

func (a *TabExt) ExtMD5() string {
	str := fmt.Sprintf("md5:%d:%s:%d:%d:%s:%d", a.Active, a.ActiveIcon, a.ActiveType, a.Inactive, a.InactiveIcon, a.InactiveType)
	randOne := md5.Sum([]byte(str))
	return hex.EncodeToString(randOne[:])
}

type TabLimit struct {
	ID         int64     `gorm:"column:id" json:"id"`
	Type       int       `gorm:"column:title" json:"type"`
	TID        int64     `gorm:"column:t_id" json:"t_id"`
	Plat       int8      `gorm:"column:plat" json:"plat"`
	Build      int64     `gorm:"column:build" json:"build"`
	Conditions string    `gorm:"column:conditions" json:"conditions"`
	State      int       `gorm:"column:state" json:"state"`
	Ctime      time.Time `gorm:"column:ctime" json:"ctime"`
	Mtime      time.Time `gorm:"column:mtime" json:"mtime"`
}

// TableName .
func (a TabLimit) TableName() string {
	return "tab_limit"
}

// AppMenus .
type AppMenus struct {
	ID   int64  `gorm:"column:id" json:"id"`
	Name string `gorm:"column:name" json:"name"`
}

// TableName .
func (a AppMenus) TableName() string {
	return "app_menus"
}

// Sidebar.
type Sidebar struct {
	ID   int64  `gorm:"column:id" json:"id"`
	Name string `gorm:"column:name" json:"name"`
	Plat int    `gorm:"column:plat" json:"plat"`
}

// TableName .
func (a Sidebar) TableName() string {
	return "sidebar"
}

// SearchReply .
type SearchReply struct {
	Total int64
	List  []*TabExt
}

// SideReply .
type SideReply struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Plat int    `json:"plat"`
}
