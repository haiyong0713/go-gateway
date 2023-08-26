package model

import "go-common/library/time"

// VideoUpView .
type VideoUpView struct {
	Archive *Archive `json:"archive"`
	Videos  []*Video `json:"videos"`
}

// Archive is archive model.
type Archive struct {
	Aid          int64     `json:"aid"`
	Mid          int64     `json:"mid"`
	TypeID       int16     `json:"tid"`
	HumanRank    int       `json:"-"`
	Title        string    `json:"title"`
	Author       string    `json:"-"`
	Cover        string    `json:"cover"`
	RejectReason string    `json:"reject_reason"`
	Tag          string    `json:"tag"`
	Duration     int64     `json:"duration"`
	Copyright    int8      `json:"copyright"`
	Desc         string    `json:"desc"`
	MissionID    int64     `json:"mission_id"`
	Round        int8      `json:"-"`
	Forward      int64     `json:"-"`
	Attribute    int32     `json:"attribute"`
	Access       int16     `json:"-"`
	State        int8      `json:"state"`
	Source       string    `json:"source"`
	NoReprint    int32     `json:"no_reprint"`
	UGCPay       int32     `json:"ugcpay"`
	OrderID      int64     `json:"order_id"`
	ADOrderID    int64     `json:"adorder_id"`
	UpFrom       int8      `json:"up_from"`
	Dynamic      string    `json:"dynamic"`
	DescFormatID int64     `json:"desc_format_id"`
	DTime        time.Time `json:"dtime"`
	PTime        time.Time `json:"ptime"`
	CTime        time.Time `json:"ctime"`
	MTime        time.Time `json:"mtime"`
}

// Video is archive_video model.
type Video struct {
	ID           int64      `json:"id"`
	Aid          int64      `json:"aid"`
	Title        string     `json:"title"`
	Desc         string     `json:"desc"`
	Filename     string     `json:"filename"`
	SrcType      string     `json:"src_type"`
	Cid          int64      `json:"cid"`
	Sid          int64      `json:"-"`
	Duration     int64      `json:"duration"`
	Filesize     int64      `json:"-"`
	Resolutions  string     `json:"-"`
	Index        int        `json:"index"`
	Playurl      string     `json:"-"`
	Status       int16      `json:"status"`
	FailCode     int8       `json:"fail_code"`
	XcodeState   int8       `json:"xcode_state"`
	Attribute    int32      `json:"-"`
	RejectReason string     `json:"reject_reason"`
	CTime        time.Time  `json:"ctime"`
	MTime        time.Time  `json:"-"`
	Dimension    *Dimension `json:"dimension"`
}

// Dimension Archive video dimension
type Dimension struct {
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
	Rotate int64 `json:"rotate"`
}

// DimensionReply is the reply of BVC's api for dimension
type DimensionReply struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Info    []*Dimension `json:"info"`
}
