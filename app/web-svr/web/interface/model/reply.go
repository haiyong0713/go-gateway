package model

import xtime "go-common/library/time"

// ReplyHot reply hot
type ReplyHot struct {
	Page    *ReplyPage   `json:"page"`
	Replies []*ReplyItem `json:"replies"`
}

type ReplyItem struct {
	RpID      int64      `json:"rpid"`
	Oid       int64      `json:"oid"`
	Type      int8       `json:"type"`
	Mid       int64      `json:"mid"`
	Root      int64      `json:"root"`
	Parent    int64      `json:"parent"`
	Dialog    int64      `json:"dialog"`
	Count     int        `json:"count"`
	RCount    int        `json:"rcount"`
	Floor     int        `json:"floor,omitempty"`
	State     int8       `json:"state"`
	FansGrade int32      `json:"fansgrade"`
	Attr      uint32     `json:"attr"`
	CTime     xtime.Time `json:"ctime"`
	MTime     xtime.Time `json:"-"`
	// string
	RpIDStr   string `json:"rpid_str,omitempty"`
	RootStr   string `json:"root_str,omitempty"`
	ParentStr string `json:"parent_str,omitempty"`
	DialogStr string `json:"dialog_str,omitempty"`
	// action count, from ReplyAction count
	Like    int           `json:"like"`
	Hate    int           `json:"-"`
	Action  int8          `json:"action"`
	Content *ReplyContent `json:"content"`
	Replies []*ReplyItem  `json:"replies"`
	Assist  int           `json:"assist"`
	// 是否有折叠评论
	RawInput   string `json:"-"`
	ShowFollow bool   `json:"show_follow"`
}

type ReplyContent struct {
	RpID    int64      `json:"-"`
	Message string     `json:"message"`
	Ats     []int64    `json:"ats,omitempty"`
	Topics  []string   `json:"topics,omitempty"`
	IP      uint32     `json:"ipi,omitempty"`
	Plat    int8       `json:"plat"`
	Device  string     `json:"device"`
	Version string     `json:"version,omitempty"`
	CTime   xtime.Time `json:"-"`
	MTime   xtime.Time `json:"-"`
}

// ReplyPage .
type ReplyPage struct {
	Acount int64 `json:"acount"`
	Count  int64 `json:"count"`
	Num    int64 `json:"num"`
	Size   int64 `json:"size"`
}
