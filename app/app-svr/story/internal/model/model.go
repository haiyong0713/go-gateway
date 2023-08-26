package model

import (
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	storymdl "go-gateway/app/app-svr/app-card/interface/model/card/story"
)

type StoryParam struct {
	AID           int64  `form:"aid"`
	MobiApp       string `form:"mobi_app"`
	Platform      string `form:"platform"`
	Device        string `form:"device"`
	Build         int    `form:"build"`
	Qn            int    `form:"qn" default:"0"`
	Fnver         int    `form:"fnver" default:"0"`
	Fnval         int    `form:"fnval" default:"0"`
	ForceHost     int    `form:"force_host"`
	Fourk         int    `form:"fourk"`
	DeviceName    string `form:"device_name"`
	TrackID       string `form:"trackid"`
	DisplayID     int    `form:"display_id"`
	Pull          int    `form:"pull"`
	Network       string `form:"network"`
	StoryParam    string `form:"story_param"`
	From          int    `form:"from"`
	FromSpmid     string `form:"from_spmid"`
	Spmid         string `form:"spmid"`
	AutoPlay      int    `form:"auto_play"`
	TfType        int32
	NetType       int32
	AdExtra       string `form:"ad_extra"`
	FeedStatus    int64  `form:"feed_status"`
	Bvid          string `form:"bvid"`
	DisableRcmd   int    `form:"disable_rcmd"`
	TeenagersMode int    `form:"teenagers_mode"`
	RequestFrom   int    `form:"request_from"`
	VideoMode     int8   `form:"video_mode"`
}

type SpaceStoryParam struct {
	VMid int64 `form:"vmid" validate:"required"`
	PS   int64 `form:"ps" validate:"required"`
	PN   int64 `form:"pn" validate:"required"`

	MobiApp    string `form:"mobi_app"`
	Platform   string `form:"platform"`
	Device     string `form:"device"`
	Build      int    `form:"build"`
	Qn         int    `form:"qn" default:"0"`
	Fnver      int    `form:"fnver" default:"0"`
	Fnval      int    `form:"fnval" default:"0"`
	ForceHost  int    `form:"force_host"`
	Fourk      int    `form:"fourk"`
	DeviceName string `form:"device_name"`
	TrackID    string `form:"trackid"`
	DisplayID  int    `form:"display_id"`
	Pull       int    `form:"pull"`
	Network    string `form:"network"`
	StoryParam string `form:"story_param"`
	TfType     int32
	NetType    int32

	Mid   int64
	Buvid string
	Plat  int8
}

type SpaceStoryCursorParam struct {
	VMid       int64  `form:"vmid" validate:"required"`
	Aid        int64  `form:"aid"`
	Contain    bool   `form:"contain"`
	BeforeSize int64  `form:"before_size"`
	AfterSize  int64  `form:"after_size"`
	Position   string `form:"position"`
	Index      int64  `form:"index"`

	MobiApp    string `form:"mobi_app"`
	Platform   string `form:"platform"`
	Device     string `form:"device"`
	Build      int    `form:"build"`
	Qn         int    `form:"qn" default:"0"`
	Fnver      int    `form:"fnver" default:"0"`
	Fnval      int    `form:"fnval" default:"0"`
	ForceHost  int    `form:"force_host"`
	Fourk      int    `form:"fourk"`
	DeviceName string `form:"device_name"`
	TrackID    string `form:"trackid"`
	DisplayID  int    `form:"display_id"`
	Pull       int    `form:"pull"`
	Network    string `form:"network"`
	StoryParam string `form:"story_param"`
	TfType     int32
	NetType    int32

	Mid   int64
	Buvid string
	Plat  int8
}

// SpaceStoryReply is
type SpaceStoryReply struct {
	Meta struct {
		TitleTail string `json:"title_tail"`
	} `json:"meta"`
	Items []*storymdl.SpaceItem `json:"items"`
	Page  struct {
		PN      int64 `json:"pn"`
		PS      int64 `json:"ps"`
		Total   int64 `json:"total"`
		HasNext bool  `json:"has_next"`
	} `json:"page"`
}

// SpaceStoryCursorReply is
type SpaceStoryCursorReply struct {
	Meta struct {
		TitleTail string `json:"title_tail"`
	} `json:"meta"`
	Items []*storymdl.SpaceCursorItem `json:"items"`
	Page  struct {
		Total   int64 `json:"total"`
		HasPrev bool  `json:"has_prev"`
		HasNext bool  `json:"has_next"`
	} `json:"page"`
	Config struct {
		ShowButton      []string `json:"show_button"`
		ReplyZoomExp    int8     `json:"reply_zoom_exp"`
		ReplyNoDanmu    bool     `json:"reply_no_danmu"`
		ReplyHighRaised bool     `json:"reply_high_raised"`
		SpeedPlayExp    bool     `json:"speed_play_exp"`
	} `json:"config"`
}

type StoryCartParam struct {
	MobiApp   string `form:"mobi_app"`
	Build     int    `form:"build"`
	Platform  string `form:"platform"`
	Device    string `form:"device"`
	Network   string `form:"network"`
	AccessKey string `form:"access_key"`
	AdExtra   string `form:"ad_extra"`
	Aid       int64  `form:"aid"`
	Cid       int64  `form:"cid"`
	AvRid     int64  `form:"av_rid"`
	AvUpId    int64  `form:"av_up_id"`
	Ua        string `form:"ua"`
	Resource  string `form:"resource"`
	Country   string `form:"country"`
	Province  string `form:"province"`
	City      string `form:"city"`
	IP        string `form:"ip"`

	Buvid string
	Mid   int64
}

type StoryCartReply struct {
	Ads []*StoryCartAds `json:"ads"`
}

type StoryCartAds struct {
	RequestID  string     `json:"request_id"`
	Index      int64      `json:"index"`
	CmMark     int64      `json:"cm_mark"`
	AdInfo     *cm.AdInfo `json:"ad_info"`
	ResourceID int64      `json:"resource_id"`
	SourceID   int64      `json:"source_id"`
	ClientIP   string     `json:"client_ip"`
	IsAdLoc    bool       `json:"is_ad_loc"`
}

type StoryGameParam struct {
	MobiApp    string `form:"mobi_app"`
	Build      int    `form:"build"`
	Platform   string `form:"platform"`
	GameBaseId int64  `form:"game_base_id"`

	Mid int64
}

type StoryGameReply struct {
	GiftNum     int64       `json:"gift_num"`
	GiftName    string      `json:"gift_name"`
	GiftIconNum int64       `json:"gift_icon_num"`
	IconURLs    interface{} `json:"icon_urls"`
	GiftInfoIds interface{} `json:"gift_info_ids"`
}

// ArcSearchParam is
type ArcSearchParam struct {
	Mid       int64
	Tid       int64
	Order     string
	Keyword   string
	Pn        int64
	Ps        int64
	CheckType string
	CheckID   int64
	AttrNot   uint64
}

// ArcSearchReply is
type ArcSearchReply struct {
	TList map[string]*ArcSearchTList `json:"tlist"`
	VList []*ArcSearchVList          `json:"vlist"`
}

// ArcSearchTList is
type ArcSearchTList struct {
	Tid   int64  `json:"tid"`
	Count int64  `json:"count"`
	Name  string `json:"name"`
}

// ArcSearchVList is
type ArcSearchVList struct {
	Comment      int64       `json:"comment"`
	TypeID       int64       `json:"typeid"`
	Play         interface{} `json:"play"`
	Pic          string      `json:"pic"`
	SubTitle     string      `json:"subtitle"`
	Description  string      `json:"description"`
	Copyright    string      `json:"copyright"`
	Title        string      `json:"title"`
	Review       int64       `json:"review"`
	Author       string      `json:"author"`
	Mid          int64       `json:"mid"`
	Created      interface{} `json:"created"`
	Length       string      `json:"length"`
	VideoReview  int64       `json:"video_review"`
	Aid          int64       `json:"aid"`
	Bvid         string      `json:"bvid"`
	HideClick    bool        `json:"hide_click"`
	IsPay        int         `json:"is_pay"`
	IsUnionVideo int         `json:"is_union_video"`
}
