package bangumi

import (
	"go-gateway/app/app-svr/app-car/interface/model"

	"encoding/json"
)

type MyAnimeParam struct {
	model.DeviceInfo
	Pn         int    `form:"pn" default:"1" validate:"min=1"`
	Ps         int    `form:"ps" default:"20" validate:"min=1,max=20"`
	FollowType string `form:"follow_type"`
	FromType   string `form:"from_type"`
	ParamStr   string `form:"param"`
}

type ListParam struct {
	model.DeviceInfo
	Pn         int    `form:"pn" default:"1" validate:"min=1"`
	Ps         int    `form:"ps" default:"20" validate:"min=1,max=20"`
	FollowType string `form:"follow_type"`
	FromType   string `form:"from_type"`
	ParamStr   string `form:"param"`
}

type Module struct {
	Badge      string `json:"badge"`
	BadgeType  int    `json:"badge_type"`
	Cover      string `json:"cover"`
	Desc       string `json:"desc"`
	IsAuto     int    `json:"is_auto"`
	Link       string `json:"link"`
	SeasonID   int32  `json:"season_id"`
	SeasonType int    `json:"season_type"`
	Stat       struct {
		Danmaku    int    `json:"danmaku"`
		Follow     int    `json:"follow"`
		FollowView string `json:"follow_view"`
		View       int    `json:"view"`
	} `json:"stat"`
	Status struct {
		Follow       int `json:"follow"`
		FollowStatus int `json:"follow_status"`
	} `json:"status"`
	Title            string `json:"title"`
	Logo             string `json:"logo"`
	IsNew            int    `json:"is_new"`
	CanWatch         int    `json:"can_watch"`
	Icon             string `json:"icon"`
	SeasonStyles     string `json:"season_styles"`
	EvaluateOmission string `json:"evaluate_omission"`
}

type View struct {
	Title       string       `json:"title"`
	SeasonTitle string       `json:"season_title"`
	Cover       string       `json:"cover"`
	Detail      string       `json:"detail"`
	Alias       string       `json:"alias"`
	OriginName  string       `json:"origin_name"`
	TypeName    string       `json:"type_name"`
	BadgeInfo   *BadgeInfo   `json:"badge_info"`
	BadgeType   int32        `json:"badge_type"` // ogv已经废弃该字段
	Badge       string       `json:"badge"`
	SquareCover string       `json:"square_cover"`
	RefineCover string       `json:"refine_cover"`
	TypeDesc    string       `json:"type_desc"`
	Mode        int          `json:"mode"`
	ShareURL    string       `json:"share_url"`
	ShortLink   string       `json:"short_link"`
	Link        string       `json:"link"`
	Evaluate    string       `json:"evaluate"`
	SeasonType  int          `json:"season_type"`
	SeasonID    int64        `json:"season_id"`
	MediaID     int64        `json:"media_id"`
	SeriesID    int64        `json:"series_id"`
	FollowTip   *FollowTip   `json:"follow_tip"`
	NewEP       *NewEP       `json:"new_ep"`
	UserStatus  *UserStatus  `json:"user_status"`
	Stat        *Stat        `json:"stat"`
	Publish     *Publish     `json:"publish"`
	Rights      *Rights      `json:"rights"`
	FollowLayer *FollowLayer `json:"follow_layer"`
	Modules     []*Modules   `json:"modules"`
	// pgc番剧类型 1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧
	Type   int     `json:"type"`
	Areas  []*Area `json:"areas"`
	Total  int     `json:"total"`
	Rating *Rating `json:"rating"`
	Status int     `json:"status"`
}

type BadgeInfo struct {
	Text         string `json:"text"`
	BgColor      string `json:"bg_color"`
	BgColorNight string `json:"bg_color_night"`
	Type         int    `json:"type"`
}

type FollowTip struct {
	Followers int64 `json:"followers"`
}

type NewEP struct {
	ID    int64  `json:"id"`
	Index string `json:"index"`
	Desc  string `json:"desc"`
	More  string `json:"more"`
	Title string `json:"title"`
}

type UserStatus struct {
	Follow       int       `json:"follow"`
	FollowStatus int       `json:"follow_status"`
	FollowBubble int       `json:"follow_bubble"`
	Pay          int       `json:"pay"`
	Sponsor      int       `json:"sponsor"`
	Vip          int       `json:"vip"`
	VipFrozen    int       `json:"vip_frozen"`
	Progress     *Progress `json:"progress"`
}

type Progress struct {
	LastEpID    int64  `json:"last_ep_id"`
	LastEpIndex string `json:"last_ep_index"`
	LastTime    int64  `json:"last_time"`
}

type Stat struct {
	Favorites  int64  `json:"favorites"`
	Views      int64  `json:"views"`
	Danmakus   int64  `json:"danmakus"`
	Coins      int64  `json:"coins"`
	Reply      int64  `json:"reply"`
	Share      int64  `json:"share"`
	Hot        int64  `json:"hot"`
	Play       string `json:"play"`
	Followers  string `json:"followers"`
	SeriesPlay int64  `json:"series_play"`
	Likes      int64  `json:"likes"`
}

type Publish struct {
	PubTime   string `json:"pub_time"`
	IsStarted int    `json:"is_started"`
	IsFinish  int    `json:"is_finish"`
}

type Rights struct {
	IsCoverShow     int    `json:"is_cover_show"`
	CanWatch        int    `json:"can_watch"`
	SeriesNew       int    `json:"series_new"`
	Copyright       string `json:"copyright"`
	AllowBp         int    `json:"allow_bp"`
	AllowDownload   int    `json:"allow_download"`
	AreaLimit       int    `json:"area_limit"`
	IsPreview       int    `json:"is_preview"`
	AllowReview     int    `json:"allow_review"`
	Resource        string `json:"resource"`
	ForbidPre       int    `json:"forbid_pre"`
	OnlyVipDownload int    `json:"only_vip_download"`
	IsHasFormal     int    `json:"is_has_formal"`
}

type FollowLayer struct {
	Info  string `json:"info"`
	Title string `json:"title"`
}

type Modules struct {
	ID          int64  `json:"id"`
	Style       string `json:"style"`
	Title       string `json:"title"`
	SectionID   int64  `json:"section_id"`
	SectionType int    `json:"section_type"`
	ModuleStyle struct {
		Line      int      `json:"line"`
		Hidden    int      `json:"hidden"`
		ShowPages []string `json:"show_pages"`
	} `json:"module_style"`
	More       string `json:"more"`
	CanOrdDesc int    `json:"can_ord_desc"`
	Data       struct {
		Episodes []*Episodes `json:"episodes"`
	} `json:"data"`
}

type Episodes struct {
	ID          int64      `json:"id"`
	BadgeInfo   *BadgeInfo `json:"badge_info"`
	BadgeType   int32      `json:"badge_type"`
	Badge       string     `json:"badge"`
	Title       string     `json:"title"`
	LongTitle   string     `json:"long_title"`
	Cover       string     `json:"cover"`
	Aid         int64      `json:"aid"`
	Cid         int64      `json:"cid"`
	ReleaseDate string     `json:"release_date"`
	Dimension   struct {
		Width  int64 `json:"width"`
		Height int64 `json:"height"`
		Rotate int64 `json:"rotate"`
	} `json:"dimension"`
	Stat struct {
		Play     int64 `json:"play"`
		Danmakus int64 `json:"danmakus"`
		Reply    int64 `json:"reply"`
		Coin     int64 `json:"coin"`
		Likes    int64 `json:"likes"`
	} `json:"stat"`
	Rights struct {
		DM            int `json:"dm"`
		AllowDownload int `json:"allow_download"`
	}
	Interaction json.RawMessage `json:"interaction"`
	Duration    int64           `json:"duration"`
}

type Area struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Rating struct {
	Score float64 `json:"score"`
	Count int     `json:"count"`
}

type MediaPGCParam struct {
	Pn         int    `form:"pn" validate:"min=1"`
	Ps         int    `form:"ps" validate:"min=1,max=50"`
	FollowType string `form:"follow_type" validate:"required"`
}
