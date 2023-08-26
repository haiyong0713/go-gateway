package splash

import (
	"encoding/json"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	garbmdl "git.bilibili.co/bapis/bapis-go/garb/model"
)

type EventSplashRequest struct {
	Mid          int64  `form:"-"`
	MobiApp      string `form:"mobi_app"`
	Platform     string `form:"platform"`
	Device       string `form:"device"`
	Build        int    `form:"build"`
	Network      string `form:"network"`
	ScreenWidth  int64  `form:"screen_width"`
	ScreenHeight int64  `form:"screen_height"`
}

type EventSplash interface {
	GetParam() string
	EventType() string
	AsJSON() ([]byte, error)
}

type EventSplashListReply struct {
	EventSplash []EventSplash `json:"event_splash"`
}

type registrationDateEvent struct {
	Event        string      `json:"event"`
	Param        string      `json:"param"`
	BeginTime    int64       `json:"begin_time"`
	EndTime      int64       `json:"end_time"`
	ResourceType string      `json:"resource_type"`
	Image        string      `json:"image,omitempty"`
	VideoURI     string      `json:"video_uri,omitempty"`
	VideoHash    string      `json:"video_hash,omitempty"`
	Screen       string      `json:"screen"`
	Element      []*Element  `json:"element"`
	Logo         string      `json:"logo"`
	ShowTimes    int64       `json:"show_times"`
	AccountCard  AccountCard `json:"account_card"`
	URI          string      `json:"uri,omitempty"`
	SkipButton   bool        `json:"skip_button"`
	Duration     int64       `json:"duration"`
}

type AccountCard struct {
	Mid            int64  `json:"mid"`
	Uname          string `json:"uname"`
	Face           string `json:"face"`
	Sign           string `json:"sign"`
	Level          int64  `json:"level"`
	OfficialVerify struct {
		Desc string `json:"desc"`
		Type int32  `json:"type"`
	} `json:"official_verify"`
	Vip     accountgrpc.VipInfo     `json:"vip"`
	Pendant accountgrpc.PendantInfo `json:"pendant"`
}

type Text struct {
	Text string `json:"text"`
}

type Element struct {
	Type              string  `json:"type"`
	MaxWidthPX        int64   `json:"max_width_px"`
	PaddingTopPercent float64 `json:"padding_top_percent"`
	FontSize          int64   `json:"font_size,omitempty"`
	Text              string  `json:"text,omitempty"`
}

func NewRegistrationDateEvent() *registrationDateEvent {
	return &registrationDateEvent{
		Event: "registration_date",
	}
}
func (r *registrationDateEvent) EventType() string {
	return r.Event
}

func (r *registrationDateEvent) AsJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *registrationDateEvent) GetParam() string {
	return r.Param
}

type EventSplashList2Reply struct {
	EventList []EventSplashV2 `json:"event_list"`
	Account   Account         `json:"account"`
}

type EventSplashV2 struct {
	ID            int64                     `json:"id"`
	EventType     int64                     `json:"event_type"`
	BeginTime     int64                     `json:"begin_time"`
	EndTime       int64                     `json:"end_time"`
	Resources     []*garbmdl.SplashResource `json:"resources"`
	Elements      []*garbmdl.SplashElement  `json:"elements"`
	ShowTimes     int64                     `json:"show_times"`
	ShowSkip      int64                     `json:"show_skip"`
	Duration      int64                     `json:"duration"`
	ShowCountDown int64                     `json:"show_countdown"`
	WifiDownload  int64                     `json:"wifi_download"`
}

type Account struct {
	Mid      int64  `json:"mid,omitempty"`
	Uname    string `json:"uname,omitempty"`
	Level    int32  `json:"level,omitempty"`
	Uimage   string `json:"uimage,omitempty"`
	Birthday int64  `json:"birthday,omitempty"`
	JoinTime int64  `json:"join_time,omitempty"`
}
