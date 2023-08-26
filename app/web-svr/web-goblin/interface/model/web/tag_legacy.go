package web

import (
	xtime "go-common/library/time"
)

// ArgID .
type ArgID struct {
	ID     int64
	Mid    int64
	RealIP string
}

// ArgIDs .
type ArgIDs struct {
	IDs    []int64
	Mid    int64
	RealIP string
}

// ArgResTags .
type ArgResTags struct {
	Oids   []int64
	Type   int8
	Mid    int64
	RealIP string
}

// ArgChannelResource ArgChannelResource.
type ArgChannelResource struct {
	Tid           int64  `form:"tid"`
	Mid           int64  `form:"mid"`
	Plat          int32  `form:"plat"`
	LoginEvent    int32  `form:"login_event"`
	TeenagersMode int32  `form:"teenagers_mode"`
	RequestCNT    int32  `form:"request_cnt"`
	DisplayID     int32  `form:"display_id"`
	From          int32  `form:"from"`
	Type          int32  `form:"type"`
	Build         int32  `form:"build"`
	Name          string `form:"tname"`
	Buvid         string `form:"buvid"`
	Channel       int32
	RealIP        string
}

// Tag .
type Tag struct {
	ID           int64      `json:"tag_id"`
	Name         string     `json:"tag_name"`
	Cover        string     `json:"cover"`
	HeadCover    string     `json:"head_cover"`
	Content      string     `json:"content"`
	ShortContent string     `json:"short_content"`
	Type         int8       `json:"type"`
	State        int8       `json:"state"`
	CTime        xtime.Time `json:"ctime"`
	MTime        xtime.Time `json:"-"`
	// tag count
	Count struct {
		View  int `json:"view"`
		Use   int `json:"use"`
		Atten int `json:"atten"`
	} `json:"count"`
	// subscriber
	IsAtten int8 `json:"is_atten"`
	// archive_tag
	Role      int8  `json:"-"`
	Likes     int64 `json:"likes"`
	Hates     int64 `json:"hates"`
	Attribute int8  `json:"attribute"`
	Liked     int8  `json:"liked"`
	Hated     int8  `json:"hated"`
	ExtraAttr int32 `json:"extra_attr"`
}

// ChannelResource ChannelResource.
type ChannelResource struct {
	Oids      []int64              `json:"resource"`
	Failover  bool                 `json:"failover"`
	IsChannel bool                 `json:"is_channel"`
	Pages     *ChannelResourcePage `json:"page"`
}

// Page page.
type ChannelResourcePage struct {
	Page     int32 `json:"page"`
	PageSize int32 `json:"pagesize"`
	Total    int64 `json:"count"`
}
