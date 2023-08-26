package resource

import (
	"time"

	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
)

// const cc TP
const (
	CustonConfigTPArchive = int64(1)
)

// const cc state
const (
	CustomConfigStateEnable  = int64(1)
	CustomConfigStateDisable = int64(0)
)

// const cc status
const (
	CustomConfigStatusExpired = "cc_expired"
	CustomConfigStatusStaging = "cc_staging"
	CustomConfigStatusOnline  = "cc_online"
	CustomConfigStatusOffline = "cc_offline"
	CustomConfigStatusUnknow  = "cc_unknow"
)

// CCListReq is
type CCListReq struct {
	TP         int64  `form:"tp"`
	Oid        string `form:"oid"` //兼容bvid
	PN         int64  `form:"pn" default:"1"`
	PS         int64  `form:"ps" default:"10"`
	OidNum     int64  `form:"-"`
	OriginType int32  `form:"origin_type"`
	AuditCode  int32  `form:"audit_code"`
}

// CCListReply is
type CCListReply struct {
	Data []*CustomConfigReply `json:"data"`
	Page common.Page          `json:"page"`
}

// CCListArchiveReply is
type CCListArchiveReply struct {
	Data []*CustomConfigArchiveReply `json:"data"`
	Page common.Page                 `json:"page"`
}

// CustomConfigReply is
type CustomConfigReply struct {
	CustomConfig
	Status string `json:"status"`
}

// CCArchive is
type CCArchive struct {
	Aid   int64  `json:"aid"`
	Title string `json:"title"`
	Mid   int64  `json:"mid"`
	State int64  `json:"state"`
	Attr  int64  `json:"attr"`
}

// CCAuthor is
type CCAuthor struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Sex  string `json:"sex"`
	Face string `json:"face"`
	Sign string `json:"sign"`
	Rank int32  `json:"rank"`
}

// CustomConfigArchiveReply is
type CustomConfigArchiveReply struct {
	CustomConfigReply
	CCArchive CCArchive `json:"archive"`
	CCAuthor  CCAuthor  `json:"author"`
}

// CustomConfig is
type CustomConfig struct {
	ID               int64      `json:"id" gorm:"column:id"`
	TP               int64      `json:"tp" gorm:"column:tp"`
	Oid              int64      `json:"oid" gorm:"column:oid"`
	Content          string     `json:"content" gorm:"column:content"`
	URL              string     `json:"url" gorm:"column:url"`
	HighlightContent string     `json:"highlight_content" gorm:"column:highlight_content"`
	Image            string     `json:"image" gorm:"column:image"`
	ImageBig         string     `json:"image_big" gorm:"column:image_big"`
	STime            xtime.Time `json:"stime" gorm:"column:stime"`
	ETime            xtime.Time `json:"etime" gorm:"column:etime"`
	State            int64      `json:"state" gorm:"column:state"`
	CTime            xtime.Time `json:"ctime" gorm:"column:ctime"`
	MTime            xtime.Time `json:"mtime" gorm:"column:mtime"`
	OriginType       int32      `json:"origin_type" gorm:"column:origin_type"`
	AuditCode        int32      `json:"audit_code"`
}

// ResolveCCStatus is
func ResolveCCStatus(ccState int64, stime time.Time, etime time.Time, now time.Time) string {
	if now.After(etime) {
		return CustomConfigStatusExpired
	}
	if now.Before(stime) {
		return CustomConfigStatusStaging
	}

	switch ccState {
	case CustomConfigStateDisable:
		return CustomConfigStatusOffline
	case CustomConfigStateEnable:
		return CustomConfigStatusOnline
	}
	return CustomConfigStatusUnknow
}

// ResolveStatusAt is
func (cc *CustomConfig) ResolveStatusAt(now time.Time) string {
	stime := cc.STime.Time()
	etime := cc.ETime.Time()
	return ResolveCCStatus(cc.State, stime, etime, now)
}

// CCAddReq is
type CCAddReq struct {
	TP               int64      `form:"tp"`
	Oid              string     `form:"oid" validate:"required"` //兼容bvid
	Content          string     `form:"content"`
	URL              string     `form:"url"`
	HighlightContent string     `form:"highlight_content"`
	Image            string     `form:"image"`
	ImageBig         string     `form:"image_big"`
	STime            xtime.Time `form:"stime" validate:"required"`
	ETime            xtime.Time `form:"etime" validate:"required"`
	Operator         string     `form:"-"`
	OperatorID       int64      `form:"-"`
	OidNum           int64      `form:"-"`
	OriginType       int32      `form:"-"`
	AuditCode        int32      `form:"-"`
}

// CCUpdateReq is
type CCUpdateReq struct {
	ID               int64      `form:"id" validate:"required"`
	TP               int64      `form:"tp"`
	Oid              string     `form:"oid" validate:"required"`
	Content          string     `form:"content"`
	URL              string     `form:"url"`
	HighlightContent string     `form:"highlight_content"`
	Image            string     `form:"image"`
	ImageBig         string     `form:"image_big"`
	STime            xtime.Time `form:"stime" validate:"required"`
	ETime            xtime.Time `form:"etime" validate:"required"`
	Operator         string     `form:"-"`
	OperatorID       int64      `form:"-"`
	OidNum           int64      `form:"-"`
	OriginType       int32      `form:"-"`
	AuditCode        int32      `form:"_"`
}

// CCLogReq is
type CCLogReq struct {
	ID int64 `form:"id" validate:"required"`
}

// CCLog is
type CCLog struct {
	Operator   string `json:"operator"`
	OperatorID int64  `json:"operator_id"`
	OperateAt  string `json:"operate_at"`
	Operation  string `json:"operation"`
}

// CCLogReply is
type CCLogReply struct {
	ID      int64    `json:"id"`
	TP      int64    `json:"tp"`
	Oid     int64    `json:"oid"`
	Logging []*CCLog `json:"logging"`
}

// GetConfigReq is
type GetConfigReq struct {
	ID int64 `form:"id" validate:"required"`
}

// CCOptReq is
type CCOptReq struct {
	ID         int64  `form:"id" validate:"required"`
	State      int64  `form:"state"`
	Operator   string `form:"-"`
	OperatorID int64  `form:"-"`
}

// ConfigListReply is
type ConfigListReply struct {
	Data []*ConfigListItem `json:"data"`
	Page common.Page       `json:"page"`
}

// ConfigListItem is
type ConfigListItem struct {
	ID               int64  `json:"id"`
	Oid              int64  `json:"oid"`
	BVID             string `json:"bvid"`
	Title            string `json:"title"`
	Mid              int64  `json:"mid"`
	Name             string `json:"name"`
	OState           int64  `json:"ostate"`
	Content          string `json:"content"`
	URL              string `json:"url"`
	Image            string `json:"image"`
	ImageBig         string `json:"image_big"`
	HighlightContent string `json:"highlight_content"`
	STime            string `json:"stime"`
	ETime            string `json:"etime"`
	State            int64  `json:"state"`
	StateDesc        string `json:"state_desc"`
	OriginType       int32  `json:"origin_type"`
	AuditCode        int32  `json:"audit_code"`
}
