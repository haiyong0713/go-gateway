package model

import (
	v1 "go-gateway/app/app-svr/resource/service/api/v1"
	"time"

	xtime "go-common/library/time"
)

// resource const
const (
	IconTypeFix     = 1
	IconTypeRandom  = 2
	IconTypeBangumi = 3

	NoCategory = 0
	IsCategory = 1

	AsgTypePic   = int8(0)
	AsgTypeVideo = int8(1)
	// pgc mobile
	AsgTypeURL     = int8(2)
	AsgTypeBangumi = int8(3)
	AsgTypeLive    = int8(4)
	AsgTypeGame    = int8(5)
	AsgTypeAv      = int8(6)

	// play icon type
	PlayIconOverall = 1
	PlayIconType    = 2
	PlayIconTag     = 3
	PlayIconArchive = 4
	PlayIconPgc     = 5
	// custom config type
	CustomConfigTPArchive = 1
	// custom config display status
	CustomConfigStatusStaging = 1
	CustomConfigStatusOnline  = 2
	CustomConfigStatusOffline = 3
	CustomConfigStatusExpired = 4
	CustomConfigStatusUnknow  = 5
	// custom config state
	CustomConfigStateEnable  = int64(1)
	CustomConfigStateDisable = int64(0)

	//数据来源方:banner配置后台标识
	Operater = "manager_banner_10948"
)

// IconTypes icon_type
var IconTypes = map[int]string{
	IconTypeFix:     "fix",
	IconTypeRandom:  "random",
	IconTypeBangumi: "bangumi",
}

// Rule resource_assignmen rule
type Rule struct {
	Cover int32  `json:"is_cover"`
	Style int32  `json:"style"`
	Label string `json:"label"`
	Intro string `json:"intro"`
}

// Resource struct
type Resource struct {
	ID          int           `json:"id"`
	Platform    int           `json:"platform"`
	Name        string        `json:"name"`
	Parent      int           `json:"parent"`
	State       int           `json:"-"`
	Counter     int           `json:"counter"`
	Position    int           `json:"position"`
	Rule        string        `json:"rule"`
	Size        string        `json:"size"`
	Previce     string        `json:"preview"`
	Desc        string        `json:"description"`
	Mark        string        `json:"mark"`
	Assignments []*Assignment `json:"assignments"`
	CTime       xtime.Time    `json:"ctime"`
	MTime       xtime.Time    `json:"mtime"`
	Level       int64         `json:"level"`
	Type        int           `json:"type"`
	IsAd        int           `json:"is_ad"`
}

// Assignment struct
type Assignment struct {
	ID             int        `json:"id"`
	AsgID          int        `json:"-"`
	Name           string     `json:"name"`
	ContractID     string     `json:"contract_id"`
	ResID          int        `json:"resource_id"`
	Pic            string     `json:"pic"`
	LitPic         string     `json:"litpic"`
	URL            string     `json:"url"`
	Rule           string     `json:"rule"`
	Weight         int        `json:"weight"`
	Agency         string     `json:"agency"`
	Price          float32    `json:"price"`
	State          int        `json:"state"`
	Atype          int8       `json:"atype"`
	Username       string     `json:"username"`
	PlayerCategory int8       `json:"player_category"`
	ApplyGroupID   int        `json:"-"`
	STime          xtime.Time `json:"stime"`
	ETime          xtime.Time `json:"etime"`
	CTime          xtime.Time `json:"ctime"`
	MTime          xtime.Time `json:"mtime"`
	ActivityID     int64      `json:"activity_id"`
	ActivitySTime  xtime.Time `json:"activity_stime"`
	ActivityETime  xtime.Time `json:"activity_etime"`
	// 投放类型 0固定投放 1推荐池 2强运营帧
	Category       int8   `json:"category"`
	SubTitle       string `json:"sub_title"`
	PositionWeight int    `json:"-"`
	//主色调,封面pic对应的主色调
	PicMainColor string `json:"pic_main_color"`
	//inline播放相关配置
	Inline Inline `json:"inline"`
	//消息推送方,数据来源方:英文名加固定数字
	Operater string `json:"operater"`
}

// inline播放配置
type Inline struct {
	//inline播放和跳转是否相同 1:不相同 2:相同
	InlineUseSame int8 `json:"inline_use_same"`
	//inline播放类型 0:默认值 1:web稿件
	InlineType int8 `json:"inline_type"`
	//inline_type 对应的value(ID)
	InlineUrl string `json:"inline_url"`
	//inline弹幕开关 1:关闭 2:开启
	InlineBarrageSwitch int8 `json:"inline_barrage_switch"`
}

// IndexIcon struct
type IndexIcon struct {
	ID       int64      `json:"id"`
	Type     int        `json:"type"`
	Title    string     `json:"title"`
	State    int        `json:"state"`
	Links    []string   `json:"links"`
	Icon     string     `json:"icon"`
	Weight   int        `json:"weight"`
	UserName string     `json:"-"`
	StTime   xtime.Time `json:"sttime"`
	EndTime  xtime.Time `json:"endtime"`
	DelTime  xtime.Time `json:"deltime"`
	CTime    xtime.Time `json:"ctime"`
	MTime    xtime.Time `json:"mtime"`
}

type PlayerIconRly struct {
	Item *PlayerIcon `json:"item"`
}

// PlayerIcon struct
type PlayerIcon struct {
	URL1         string       `json:"url1,omitempty"`
	Hash1        string       `json:"hash1,omitempty"`
	URL2         string       `json:"url2,omitempty"`
	Hash2        string       `json:"hash2,omitempty"`
	CTime        xtime.Time   `json:"ctime,omitempty"`
	Type         int8         `json:"-"`
	TypeValue    string       `json:"-"`
	MTime        xtime.Time   `json:"-"`
	DragLeftPng  string       `json:"drag_left_png,omitempty"`
	MiddlePng    string       `json:"middle_png,omitempty"`
	DragRightPng string       `json:"drag_right_png,omitempty"`
	DragData     *v1.IconData `json:"drag_data,omitempty"`
	NoDragData   *v1.IconData `json:"no_drag_data,omitempty"`
}

// ResWarnInfo for email
type ResWarnInfo struct {
	AID            int64
	URL            string
	AssignmentID   int
	AssignmentName string
	ResourceName   string
	ResourceID     int
	MaterialID     int
	UserName       string
	STime          xtime.Time `json:"stime"`
	ETime          xtime.Time `json:"etime"`
	ApplyGroupID   int
}

// Cmtbox live danmaku box
type Cmtbox struct {
	ID            int64      `json:"id"`
	LoadCID       int64      `json:"load_cid"`
	Server        string     `json:"server"`
	Port          string     `json:"port"`
	SizeFactor    string     `json:"size_factor"`
	SpeedFactor   string     `json:"speed_factor"`
	MaxOnscreen   string     `json:"max_onscreen"`
	Style         string     `json:"style"`
	StyleParam    string     `json:"style_param"`
	TopMargin     string     `json:"top_margin"`
	State         string     `json:"state"`
	RenqiVisible  string     `json:"renqi_visible"`
	RenqiFontsize string     `json:"renqi_fontsize"`
	RenqiFmt      string     `json:"renqi_fmt"`
	RenqiOffset   string     `json:"renqi_offset"`
	RenqiColor    string     `json:"renqi_color"`
	CTime         xtime.Time `json:"ctime"`
	MTime         xtime.Time `json:"mtime"`
}

// CustomConfig is
type CustomConfig struct {
	TP               int32     `json:"tp"`
	Oid              int64     `json:"oid"`
	Content          string    `json:"content"`
	URL              string    `json:"url"`
	HighlightContent string    `json:"highlight_content"`
	Image            string    `json:"image"`
	ImageBig         string    `json:"image_big"`
	STime            time.Time `json:"stime"`
	ETime            time.Time `json:"etime"`
	State            int64     `json:"state"`
	ID               int64     `json:"id"`
	AuditCode        int32     `json:"audit_code"`
	CTime            time.Time `json:"ctime"`
	MTime            time.Time `json:"mtime"`
}

type WhiteCheckForm struct {
	Uid int64 `form:"uid" validate:"required"`
}

type WhiteCheckStatus struct {
	Status int `json:"status"`
}

// ResolveCCStatus is
//func ResolveCCStatus(ccState int64, stime time.Time, etime time.Time, now time.Time) int {
//	if now.After(etime) {
//		return CustomConfigStatusExpired
//	}
//	if now.Before(stime) {
//		return CustomConfigStatusStaging
//	}
//
//	switch ccState {
//	case CustomConfigStateDisable:
//		return CustomConfigStatusOffline
//	case CustomConfigStateEnable:
//		return CustomConfigStatusOnline
//	}
//	return CustomConfigStatusUnknow
//}

// ResolveStatusAt is
//func (cc *CustomConfig) ResolveStatusAt(now time.Time) int {
//	stime := cc.STime.Time()
//	etime := cc.ETime.Time()
//	return ResolveCCStatus(cc.State, stime, etime, now)
//}

func IsUnderAndroid604(mobiApp string, build int32) bool {
	if mobiApp == "android" && build <= 6040000 {
		return true
	}
	return false

}
