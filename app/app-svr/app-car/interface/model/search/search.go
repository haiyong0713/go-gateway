package search

import (
	"encoding/json"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-car/interface/model"
)

type SearchParam struct {
	model.DeviceInfo
	Keyword  string `form:"keyword"`
	Pn       int    `form:"pn" default:"1" validate:"min=1"`
	Ps       int    `form:"ps" default:"20" validate:"min=1,max=20"`
	FromType string `form:"from_type"`
	ParamStr string `form:"param"`
}

type SearchSuggestParam struct {
	model.DeviceInfo
	Keyword   string `form:"keyword"`
	Highlight int    `form:"highlight"`
}

// Search all
type Search struct {
	Code           int     `json:"code,omitempty"`
	Trackid        string  `json:"seid,omitempty"`
	Page           int     `json:"page,omitempty"`
	PageSize       int     `json:"pagesize,omitempty"`
	Total          int     `json:"total,omitempty"`
	NumResults     int     `json:"numResults,omitempty"`
	NumPages       int     `json:"numPages,omitempty"`
	SuggestKeyword string  `json:"suggest_keyword,omitempty"`
	CrrQuery       string  `json:"crr_query,omitempty"`
	Result         *Result `json:"result,omitempty"`
}

// Flow struct
type Result struct {
	Video        []*Video `json:"video"`
	MediaBangumi []*Media `json:"media_bangumi"`
	MediaFt      []*Media `json:"media_ft"`
	BiliUser     []*User  `json:"bili_user,omitempty"`
}

// Video struct
type Video struct {
	ID         int64       `json:"id"`
	Author     string      `json:"author"`
	Title      string      `json:"title"`
	Pic        string      `json:"pic"`
	Desc       string      `json:"description"`
	Play       interface{} `json:"play"`
	Danmaku    int         `json:"video_review"`
	Duration   string      `json:"duration"`
	Pages      int         `json:"numPages"`
	ViewType   string      `json:"view_type"`
	RecTags    []string    `json:"rec_tags"`
	IsPay      int         `json:"is_pay"`
	NewRecTags []*RecTag   `json:"new_rec_tags"`
}

// RecTag from video
type RecTag struct {
	Name  string `json:"tag_name"`
	Style int8   `json:"tag_style"`
}

// Media struct
type Media struct {
	MediaID    int64  `json:"media_id,omitempty"`
	SeasonID   int32  `json:"season_id,omitempty"`
	Title      string `json:"title,omitempty"`
	OrgTitle   string `json:"org_title,omitempty"`
	Styles     string `json:"styles,omitempty"`
	Cover      string `json:"cover,omitempty"`
	PlayState  int    `json:"play_state,omitempty"`
	MediaScore *struct {
		Score     float64 `json:"score,omitempty"`
		UserCount int     `json:"user_count,omitempty"`
	} `json:"media_score,omitempty"`
	MediaType   int             `json:"media_type,omitempty"`
	CV          string          `json:"cv,omitempty"`
	Staff       string          `json:"staff,omitempty"`
	Areas       string          `json:"areas,omitempty"`
	GotoURL     string          `json:"goto_url,omitempty"`
	Pubtime     xtime.Time      `json:"pubtime,omitempty"`
	HitColumns  []string        `json:"hit_columns,omitempty"`
	AllNetName  string          `json:"all_net_name,omitempty"`
	AllNetIcon  string          `json:"all_net_icon,omitempty"`
	AllNetURL   string          `json:"all_net_url,omitempty"`
	DisplayInfo json.RawMessage `json:"display_info,omitempty"`
	HitEpids    string          `json:"hit_epids,omitempty"`
	Position    int             `json:"position,omitempty"`
}

// User struct
type User struct {
	Mid            int64           `json:"mid,omitempty"`
	Name           string          `json:"uname,omitempty"`
	SName          string          `json:"name,omitempty"`
	OfficialVerify *OfficialVerify `json:"official_verify,omitempty"`
	Usign          string          `json:"usign,omitempty"`
	Fans           int             `json:"fans,omitempty"`
	Videos         int             `json:"videos,omitempty"`
	Level          int             `json:"level,omitempty"`
	Pic            string          `json:"upic,omitempty"`
	Pages          int             `json:"numPages,omitempty"`
	Res            []*struct {
		Play     interface{} `json:"play,omitempty"`
		Danmaku  int         `json:"dm,omitempty"`
		Pubdate  int64       `json:"pubdate,omitempty"`
		Title    string      `json:"title,omitempty"`
		Aid      int64       `json:"aid,omitempty"`
		Pic      string      `json:"pic,omitempty"`
		ArcURL   string      `json:"arcurl,omitempty"`
		Duration string      `json:"duration,omitempty"`
		IsPay    int         `json:"is_pay,omitempty"`
	} `json:"res,omitempty"`
	IsLive         int             `json:"is_live,omitempty"`
	RoomID         int64           `json:"room_id,omitempty"`
	IsUpuser       int             `json:"is_upuser,omitempty"`
	Position       int             `json:"position,omitempty"`
	BackgroundInfo *BackgroundInfo `json:"background_info,omitempty"`
	Version        int             `json:"version,omitempty"`
	IsInlineLive   int64           `json:"is_inline_live,omitempty"`
}

type BackgroundInfo struct {
	BgPic string `json:"bg_pic,omitempty"`
	FgPic string `json:"fg_pic,omitempty"`
}

// OfficialVerify struct
type OfficialVerify struct {
	Type int    `json:"type"`
	Desc string `json:"desc,omitempty"`
}

type SearchArgs struct {
	Trackid string `json:"trackid,omitempty"`
	Page    int    `json:"page,omitempty"`
}

func (a *SearchArgs) SearchArgsFrom(s *Search) {
	a.Page = s.Page
	a.Trackid = s.Trackid
}

// Suggest struct
type Suggest struct {
	Code    int    `json:"code"`
	TrackID string `json:"trackid"`
	ExpStr  string `json:"exp_str"`
	Result  []*Sug `json:"result"`
}

// Sug struct
type Sug struct {
	ShowName string `json:"show_name,omitempty"`
	Term     string `json:"term,omitempty"`
	Ref      int64  `json:"ref,omitempty"`
	Pos      int    `json:"pos,omitempty"`
	TermType int    `json:"term_type,omitempty"`
}

type SuggestItem struct {
	Position int    `json:"position,omitempty"`
	Title    string `json:"title,omitempty"`
	From     string `json:"from,omitempty"`
	KeyWord  string `json:"keyword,omitempty"`
	TermType int    `json:"term_type,omitempty"`
	ModuleID int64  `json:"module_id,omitempty"`
}

type UpItem struct {
	Mid   int64  `json:"mid,omitempty"`
	Name  string `json:"name,omitempty"`
	Desc1 string `json:"desc_1,omitempty"`
	Desc2 string `json:"desc_2,omitempty"`
	Face  string `json:"face,omitempty"`
	URI   string `json:"uri,omitempty"`
}

func (i *SuggestItem) FromSuggest(st *Sug) {
	i.Position = st.Pos
	i.Title = st.ShowName
	i.From = "search"
	i.KeyWord = st.Term
	i.TermType = st.TermType
	i.ModuleID = st.Ref
}

func (i *SuggestItem) FromSuggestWeb(st *Sug) {
	i.Title = st.ShowName
	i.KeyWord = st.Term
}

func (a *SearchArgs) FromSuggestArgs(s *Suggest) {
	a.Trackid = s.TrackID
}

type MediaSearchParam struct {
	Pn      int    `form:"pn" validate:"min=1"`
	Ps      int    `form:"ps" validate:"min=1,max=50"`
	Keyword string `form:"keyword"`
}
