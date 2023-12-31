package feed

import (
	"encoding/json"
	"strconv"
	"strings"

	xtime "go-common/library/time"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/card/rank"
	"go-gateway/app/app-svr/app-card/interface/model/card/show"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/app-feed/interface/model/dislike"
	livemdl "go-gateway/app/app-svr/app-feed/interface/model/live"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	feed "git.bilibili.co/bapis/bapis-go/community/service/feed"
	feedArcApi "git.bilibili.co/bapis/bapis-go/community/service/feed/archive"
	feedArtApi "git.bilibili.co/bapis/bapis-go/community/service/feed/article"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

const (
	_rankTitle          = "全站排行榜"
	_rankURI            = "http://www.bilibili.com/index/rank/all-03.json"
	_rankCount          = 3
	_convergeMinCount   = 2
	_bangumiRcmdUpdates = 99

	FlagYes                     = 1
	FlagBitCanSupportThreePoint = uint(0)
	FlagBitCanSupportGuide      = uint(1)
	FlagBitCanResetColumn       = uint(2)

	DiscardReasonCannotFindCardHandler  = 1
	DiscardReasonCannotBuildCard        = 2
	DiscardReasonNilItem                = 3
	DiscardReasonOther                  = 4
	DiscardReasonAd                     = 5
	DiscardReasonRepeatedId             = 7
	DiscardReasonUnexpectedCardTemplate = 8
	DiscardReasonUnexpectedGoto         = 9
	DiscardReasonCardIsNotNormal        = 10
	DiscardReasonEmptyWindow            = 11
)

// Item is feed item, contains av, bangumi, live, banner, feed...
type Item struct {
	Title      string      `json:"title,omitempty"`
	Subtitle   string      `json:"subtitle,omitempty"`
	Cover      string      `json:"cover,omitempty"`
	URI        string      `json:"uri,omitempty"`
	Redirect   string      `json:"redirect,omitempty"`
	Param      string      `json:"param,omitempty"`
	Goto       string      `json:"goto,omitempty"`
	Desc       string      `json:"desc,omitempty"`
	Play       int         `json:"play,omitempty"`
	Danmaku    int         `json:"danmaku,omitempty"`
	Reply      int         `json:"reply,omitempty"`
	Fav        int         `json:"favorite,omitempty"`
	Coin       int         `json:"coin,omitempty"`
	Share      int         `json:"share,omitempty"`
	Like       int         `json:"like,omitempty"`
	Dislike    int         `json:"dislike,omitempty"`
	Duration   int64       `json:"duration,omitempty"`
	Count      int         `json:"count,omitempty"`
	Status     int8        `json:"status,omitempty"`
	Type       int8        `json:"type,omitempty"`
	Badge      string      `json:"badge,omitempty"`
	StatType   int8        `json:"stat_type,omitempty"`
	RcmdReason *RcmdReason `json:"rcmd_reason,omitempty"`
	Item       []*Item     `json:"item,omitempty"`
	// sortedset index
	Idx int64 `json:"idx,omitempty"`
	// av
	Cid             int64                     `json:"cid,omitempty"`
	Rid             int32                     `json:"tid,omitempty"`
	TName           string                    `json:"tname,omitempty"`
	Tag             *Tag                      `json:"tag,omitempty"`
	Button          *Button                   `json:"button,omitempty"`
	DisklikeReasons []*dislike.DisklikeReason `json:"dislike_reasons,omitempty"`
	CTime           xtime.Time                `json:"ctime,omitempty"`
	Autoplay        int32                     `json:"autoplay,omitempty"`
	// upper
	Mid      int64         `json:"mid,omitempty"`
	Name     string        `json:"name,omitempty"`
	Face     string        `json:"face,omitempty"`
	IsAtten  int8          `json:"is_atten,omitempty"`
	Fans     int64         `json:"fans,omitempty"`
	RecCnt   int           `json:"recent_count,omitempty"`
	Recent   []*Item       `json:"recent,omitempty"`
	Official *OfficialInfo `json:"official,omitempty"`
	// live
	Online int    `json:"online,omitempty"`
	Area   string `json:"area,omitempty"`
	AreaID int    `json:"area_id,omitempty"`
	Area2  *Area2 `json:"area2,omitempty"`
	// bangumi
	Index       string `json:"index,omitempty"`
	IndexTitle  string `json:"index_title,omitempty"`
	CoverMark   string `json:"cover_mark,omitempty"`
	Finish      bool   `json:"finish,omitempty"`
	LatestIndex string `json:"last_index,omitempty"`
	// bangumi ai
	Updates int `json:"updates,omitempty"`
	// live or bangumi
	From int8 `json:"from,omitempty"`
	// adviertisement
	RequestID  string          `json:"request_id,omitempty"`
	CreativeID int64           `json:"creative_id,omitempty"`
	SrcID      int32           `json:"src_id,omitempty"`
	IsAd       bool            `json:"is_ad,omitempty"`
	IsAdLoc    bool            `json:"is_ad_loc,omitempty"`
	AdCb       string          `json:"ad_cb,omitempty"`
	ShowURL    string          `json:"show_url,omitempty"`
	ClickURL   string          `json:"click_url,omitempty"`
	ClientIP   string          `json:"client_ip,omitempty"`
	CmMark     int64           `json:"cm_mark,omitempty"`
	AdIndex    int32           `json:"ad_index,omitempty"`
	Extra      json.RawMessage `json:"extra,omitempty"`
	CardIndex  int32           `json:"card_index,omitempty"`
	// tag
	Tags []*taggrpc.Tag `json:"tags,omitempty"`
	// rank
	Cover1 string `json:"cover1,omitempty"`
	Cover2 string `json:"cover2,omitempty"`
	Cover3 string `json:"cover3,omitempty"`
	// banner
	BannerItem []*banner.Banner `json:"banner_item,omitempty"`
	Hash       string           `json:"hash,omitempty"`
	// article
	Covers    []string  `json:"covers,omitempty"`
	Template  int       `json:"template,omitempty"`
	Temple    int       `json:"temple,omitempty"`
	Category  *Category `json:"category,omitempty"`
	BannerURL string    `json:"banner_url,omitempty"`
	// game download
	Download int32  `json:"download,omitempty"`
	BigCover string `json:"big_cover,omitempty"`
	// special
	HideBadge bool    `json:"hide_badge,omitempty"`
	Ratio     float64 `json:"ratio,omitempty"`
	// shopping
	City   string `json:"city,omitempty"`
	PType  string `json:"ptype,omitempty"`
	Price  string `json:"price,omitempty"`
	Square string `json:"square,omitempty"`
	STime  string `json:"stime,omitempty"`
	ETime  string `json:"etime,omitempty"`
	// news
	Content string `json:"content,omitempty"`
	// subscribe
	Kind string `json:"kind,omitempty"`
	// audio
	SongTitle string `json:"song_title,omitempty"`
	// bigdata source
	Source    string          `json:"-"`
	AvFeature json.RawMessage `json:"-"`
	// common
	GotoOrg string `json:"-"`
	// rank score
	Score string `json:"score,omitempty"`
	// ai recommend
	AI *ai.Item `json:"-"`
	// abtest
	AutoplayCard int `json:"autoplay_card,omitempty"`
}

type Dimension struct {
	Width  int64 `json:"width,omitempty"`
	Height int64 `json:"height,omitempty"`
	Rotate int64 `json:"rotate,omitempty"`
}

type Button struct {
	Name        string `json:"name,omitempty"`
	URI         string `json:"uri,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

type RcmdReason struct {
	ID           int    `json:"id,omitempty"`
	Content      string `json:"content,omitempty"`
	BgColor      string `json:"bg_color,omitempty"`
	IconLocation string `json:"icon_location,omitempty"`
	Message      string `json:"message,omitempty"`
}

type Category struct {
	ID       int64     `json:"id,omitempty"`
	Name     string    `json:"name,omitempty"`
	Children *Category `json:"children,omitempty"`
}

type Area2 struct {
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Children *Area2 `json:"children,omitempty"`
}

type Tag struct {
	// new
	ID      int64  `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Face    string `json:"face,omitempty"`
	Fans    int    `json:"fans,omitempty"`
	IsAtten int8   `json:"is_atten,omitempty"`
	URI     string `json:"uri,omitempty"`

	// old
	TagID   int64     `json:"tag_id,omitempty"`
	TagName string    `json:"tag_name,omitempty"`
	Count   *TagCount `json:"count,omitempty"`
}

type TagCount struct {
	Atten int `json:"atten,omitempty"`
}

type OfficialInfo struct {
	Role  int32  `json:"role,omitempty"`
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
}

type IndexParam struct {
	Build    int    `form:"build"`
	Platform string `form:"platform"`
	MobiApp  string `form:"mobi_app"`
	Device   string `form:"device"`
	Network  string `form:"network"`
	// idx, err := strconv.ParseInt(idxStr, 10, 64)
	// if err != nil || idx < 0 {
	// 	idx = 0
	// }
	Idx int64 `form:"idx" default:"0"`
	// pull, err := strconv.ParseBool(pullStr)
	// if err != nil {
	// 	pull = true
	// }
	Pull   bool             `form:"pull" default:"true"`
	Column cdm.ColumnStatus `form:"column"`
	// loginEvent, err := strconv.Atoi(loginEventStr)
	// if err != nil {
	// 	loginEvent = 0
	// }
	LoginEvent    int    `form:"login_event" default:"0"`
	OpenEvent     string `form:"open_event"`
	BannerHash    string `form:"banner_hash"`
	AdExtra       string `form:"ad_extra"`
	Interest      string `form:"interest"`
	Flush         int    `form:"flush"`
	AutoPlayCard  int    `form:"autoplay_card"`
	DeviceType    int    `form:"device_type"`
	ParentMode    int    `form:"parent_mode"`
	RecsysMode    int    `form:"recsys_mode"`
	TeenagersMode int    `form:"teenagers_mode"`
	LessonsMode   int    `form:"lessons_mode"`
	DeviceName    string `form:"device_name"`
	AccessKey     string `form:"access_key"`
	ActionKey     string `form:"actionKey"`
	Statistics    string `form:"statistics"`
	Appver        int    `form:"appver"`
	Filtered      int    `form:"filtered"`
	AppKey        string `form:"appkey"`
	HttpsUrlReq   int    `form:"https_url_req"`
	InterestV2    string `form:"interest_v2"`
	SplashID      int64  `form:"splash_id"`
	// 表示当前用户设备是否是新设备	0:不是新设备、1:新设备
	Guidance            int    `form:"guidance"`
	RedPoint            int64  `form:"red_point"`
	IsMelloi            string `form:"is_melloi"` // 来自melloi
	InlineSound         int64  `form:"inline_sound"`
	InlineDanmu         int64  `form:"inline_danmu"`
	SLocale             string `form:"s_locale"`
	CLocale             string `form:"c_locale"`
	ColumnTimestamp     int64  `form:"column_timestamp"`
	AutoplayTimestamp   int64  `form:"autoplay_timestamp"`
	ScreenWindowType    int64  `form:"screen_window_type"`
	DisableRcmd         int    `form:"disable_rcmd"`
	LocalBuvid          string `form:"local_buvid"`
	OpenAppURL          string `form:"open_app_url"`
	VoiceBalance        int64  `form:"voice_balance"`
	DituiLanding        int64  `form:"ditui_landing"`
	PrivacyDisagreeMode int    `form:"privacy_disagree_mode"`
	InterestId          int64  `form:"interest_id"`
	InterestResult      string `form:"interest_result"`
	Ua                  string
	VideoMode           int8 `form:"video_mode"`
}

type ConvergeParam struct {
	ID            int64  `form:"id" validate:"required,min=1"`
	MobiApp       string `form:"mobi_app"`
	Platform      string `form:"platform"`
	Device        string `form:"device"`
	Build         int    `form:"build"`
	Network       string `form:"network"`
	ConvergeParam string `form:"converge_param"`
	DisplayID     int    `form:"display_id"`
	ConvergeType  int    `form:"converge_type"`
}

type NewUser struct {
	HideCase int
	Buvid    string
}

func (i *Item) FromRcmd(r *ai.Item) {
	i.Title = r.Name
	i.Param = strconv.FormatInt(r.ID, 10)
	if r.Goto == "" {
		r.Goto = model.GotoAv
	}
	i.From = r.From
	i.Source = r.Source
	i.AvFeature = r.AvFeature
	if r.Config != nil {
		i.Title = r.Config.Title
		i.Cover = r.Config.Cover
		i.URI = r.Config.URI
	}
	i.StatType = r.StatType
	i.GotoOrg = r.Goto
}

type Infoc struct {
	UserFeature             json.RawMessage
	IsRcmd                  bool
	NewUser                 bool
	Code                    int
	AutoPlayInfoc           string
	IsGifCover              map[int64]int
	BannerHash              string
	AdCode                  int
	AdError                 error
	AdPos                   []string
	AdPkCode                []string
	SubGotoMap              map[int][]string
	PendantMap              map[int64]string
	AiTunnelOidMap          map[int][]string
	AiBangumiRcmdOgvInfoMap map[int]*BangumiRcmdInfoc
	DiscardReason           map[int64]*Discard
	GameBadge               map[int64]string
	BadgeMap                map[int64]string
}

type BangumiRcmdInfoc struct {
	SeasonId []int64 `json:"season_id"`
	Epid     []int64 `json:"epid"`
}

type Config struct {
	Column           cdm.ColumnStatus `json:"column"`
	AutoplayCard     int8             `json:"autoplay_card"`
	FeedCleanAbtest  int8             `json:"feed_clean_abtest"`
	FollowMode       *FollowMode      `json:"follow_mode,omitempty"`
	HomeTransferTest int8             `json:"home_transfer_test"`
	NewDouble        int8             `json:"new_double,omitempty"`
	AutoRefreshTime  int64            `json:"auto_refresh_time,omitempty"`
	Interest         *Interest        `json:"interest,omitempty"`
	// 最终用户是否能展示新用户引导，0或者字段不存在：不展示、1:展示新用户引导
	InterGuidance     int8   `json:"inter_guidance,omitempty"`
	SceneURI          string `json:"scene_uri,omitempty"`
	ShowInlineDanmaku int    `json:"show_inline_danmaku,omitempty"`
	FeedTopClean      int    `json:"feed_top_clean,omitempty"`
	Toast             Toast  `json:"toast,omitempty"`
	// no_preload： 0：需要预加载；1：不需要预加载，默认 0
	NoPreload                  int8           `json:"no_preload,omitempty"`
	SingleAutoplayFlag         int8           `json:"single_autoplay_flag,omitempty"`
	IpadHDAbtest               int8           `json:"ipad_hd_abtest,omitempty"`
	NeedResetAutoplay          bool           `json:"need_reset_autoplay,omitempty"`
	NeedResetColumn            bool           `json:"need_reset_column,omitempty"`
	SwitchColumnGuidance       *PopupGuidance `json:"switch_column_guidance,omitempty"`
	RecoverColumnGuidance      *PopupGuidance `json:"recover_column_guidance,omitempty"`
	IsBackToHomepage           bool           `json:"is_back_to_homepage,omitempty"`
	EnableRcmdGuide            bool           `json:"enable_rcmd_guide,omitempty"`
	InlineSound                int8           `json:"inline_sound,omitempty"`
	AutoRefreshTimeByAppear    int64          `json:"auto_refresh_time_by_appear,omitempty"`
	AutoRefreshTimeByActive    int64          `json:"auto_refresh_time_by_active,omitempty"`
	TriggerLoadmoreLeftLineNum int64          `json:"trigger_loadmore_left_line_num,omitempty"`
	IsNaviExp                  int8           `json:"is_navi_exp,omitempty"`
	RefreshTopFirstToast       string         `json:"refresh_top_first_toast,omitempty"`
	RefreshTopSecondToast      string         `json:"refresh_top_second_toast,omitempty"`
	HistoryCacheSize           int64          `json:"history_cache_size,omitempty"`
	RefreshBarType             int8           `json:"refresh_bar_type,omitempty"`
	RefreshOnBack              int8           `json:"refresh_on_back,omitempty"`
	VisibleArea                int64          `json:"visible_area,omitempty"`
	CardDensityExp             int8           `json:"card_density_exp,omitempty"`
	SmallCoverWhRatio          float32        `json:"small_cover_wh_ratio,omitempty"`
	VideoMode                  int8           `json:"video_mode,omitempty"`
	TopRefreshLatestExp        int8           `json:"top_refresh_latest_exp,omitempty"`
	PegasusRefreshGuidanceExp  int8           `json:"pegasus_refresh_guidance_exp,omitempty"`
	SpaceEnlargeExp            int8           `json:"space_enlarge_exp,omitempty"`
	IconGuidanceExp            int8           `json:"icon_guidance_exp,omitempty"`
}

type PopupGuidance struct {
	Title          string         `json:"title,omitempty"`
	SubTitle       string         `json:"sub_title,omitempty"`
	SourceURL      string         `json:"source_url,omitempty"`
	SourceNightURL string         `json:"source_night_url,omitempty"`
	Option         []*GuideOption `json:"option,omitempty"`
}

type GuideOption struct {
	Desc  string           `json:"desc,omitempty"`
	Toast string           `json:"toast,omitempty"`
	Value int8             `json:"value,omitempty"`
	Type  cdm.ColumnStatus `json:"type,omitempty"`
}

type InterestChoose struct {
	Style          int64                 `json:"style,omitempty"`
	Items          []*InterestChooseItem `json:"items,omitempty"`
	Title          string                `json:"title,omitempty"`
	SubTitle       string                `json:"sub_title,omitempty"`
	ConfirmText    string                `json:"confirm_text,omitempty"`
	ConfirmOutText string                `json:"confirm_out_text,omitempty"`
	CancelText     string                `json:"cancel_text,omitempty"`
	UniqueId       int64                 `json:"unique_id,omitempty"`
}

type InterestChooseItem struct {
	Name string `json:"name,omitempty"`
	Icon string `json:"icon,omitempty"`
	Id   int8   `json:"id,omitempty"`
	Desc string `json:"desc,omitempty"`
}

type Toast struct {
	HasToast     bool   `json:"has_toast,omitempty"`
	ToastMessage string `json:"toast_message,omitempty"`
}

type Interest struct {
	TitleHide string          `json:"title_hide,omitempty"`
	DescHide  string          `json:"desc_hide,omitempty"`
	TitleShow string          `json:"title_show,omitempty"`
	DescShow  string          `json:"desc_show,omitempty"`
	Message   string          `json:"message,omitempty"`
	Items     []*InterestItem `json:"items,omitempty"`
}

type InterestItem struct {
	ID     int64           `json:"id,omitempty"`
	Title  string          `json:"title,omitempty"`
	Option []*InterestItem `json:"option,omitempty"`
}

type FollowMode struct {
	Title        string    `json:"title,omitempty"`
	Option       []*Option `json:"option,omitempty"`
	Card         *Card     `json:"-"`
	ToastMessage string    `json:"toast_message,omitempty"`
}

type Option struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
	Value int8   `json:"value"`
}

type Card struct {
	Title  string   `json:"-"`
	Desc   string   `json:"-"`
	Button []string `json:"-"`
}

type Abtest struct {
	Banner         int8                `json:"-"`
	ThreePoint     int8                `json:"-"`
	DislikeExp     int                 `json:"-"`
	RcmdReason     int8                `json:"-"`
	IsNewAd        int8                `json:"-"`
	GifType        int                 `json:"-"`
	AllGifState    map[string]struct{} `json:"-"`
	NewUser        int                 `json:"-"`
	FeedLive       int8                `json:"-"`
	IsNewAdBigCard int8                `json:"-"`
	BannerExp      int                 `json:"-"`
	// 不展示用户引导
	HideGuidance int8 `json:"-"`
	// ai广告
	AdExp             int   `json:"-"`
	StoryThreePoint   bool  `json:"-"`
	ManualInline      int8  `json:"-"`
	ResourceID        int64 `json:"-"`
	IpadHDThreeColumn int8  `json:"-"`
	SingleGuide       int64 `json:"-"`
	RsNewUser         bool  `json:"-"`
	DislikeText       int8  `json:"-"`
	SingleRcmdReason  int8  `json:"-"`
	LiveContentMode   int64 `json:"-"`
}

func (t *Abtest) singleGuideFlag(bit uint) int64 {
	return (t.SingleGuide >> bit) & 1
}

func (t *Abtest) CanSupportThreePoint() bool {
	return t.singleGuideFlag(FlagBitCanSupportThreePoint) == FlagYes
}

func (t *Abtest) CanSupportGuide() bool {
	return t.singleGuideFlag(FlagBitCanSupportGuide) == FlagYes
}

func (t *Abtest) CanResetColumn() bool {
	return t.singleGuideFlag(FlagBitCanResetColumn) == FlagYes
}

type AIResponse struct {
	Items           []*ai.Item
	InterestList    []*ai.Interest
	Ad              *cm.Ad
	UserFeature     json.RawMessage
	RespCode        int
	DislikeExp      int
	AutoRefreshTime int64
	NewUser         bool
	IsRcmd          bool
	// ai ad
	BizData                    *ai.BizData
	AdCode                     int
	SceneURI                   string
	FeedTopClean               int
	NoPreload                  int8
	ManualInline               int8
	SingleGuide                int64
	DislikeText                int8
	OpenSound                  int8
	AutoRefreshTimeByAppear    int64
	AutoRefreshTimeByActive    int64
	TriggerLoadmoreLeftLineNum int64
	RefreshToast               string
	IsNaviExp                  int8
	RefreshTopFirstToast       string
	RefreshTopSecondToast      string
	HistoryCacheSize           int64
	RefreshBarType             int8
	RefreshOnBack              int8
	SingleRcmdReason           int8
	SmallCoverWhRatio          float32
	VideoMode                  int8
	TopRefreshLatestExp        int8
	ValidShowThres             int64
	PegasusRefreshGuidanceExp  int8
	SpaceEnlargeExp            int8
	IconGuidanceExp            int8
}

type SceneData struct {
	Key  string `json:"key,omitempty"`
	Goto string `json:"goto,omitempty"`
}

func (s *SceneData) URI() string {
	if s == nil || s.Key == "" {
		return ""
	}
	// https: //info.bilibili.co/pages/viewpage.action?pageId=4551903
	switch s.Goto {
	case "av":
		uri := "bilibili://video/" + s.Key + "?from_spmid=tm.recommend.from-advertise-link.0"
		return uri
	case "pgc":
		uri := "https://www.bilibili.com/bangumi/play/ss" + s.Key + "?from_spmid=tm.recommend.from-advertise-link.0"
		return uri
	case "url":
		return s.Key
	default:
		return ""
	}
}

func (i *Item) FromAv(a *arcgrpc.Arc) {
	if i.Title == "" {
		i.Title = a.Title
	}
	if i.Cover == "" {
		i.Cover = model.CoverURLHTTPS(a.Pic)
	} else {
		i.Cover = model.CoverURLHTTPS(i.Cover)
	}
	i.Param = strconv.FormatInt(a.Aid, 10)
	i.Goto = model.GotoAv
	i.URI = model.FillURI(i.Goto, i.Param, 0, 0, model.AvPlayHandler(a))
	i.Cid = a.FirstCid
	i.Rid = a.TypeID
	i.TName = a.TypeName
	i.Desc = strconv.Itoa(int(a.Stat.Danmaku)) + "弹幕"
	i.fillArcStat(a)
	i.Duration = a.Duration
	i.Mid = a.Author.Mid
	i.Name = a.Author.Name
	i.Face = a.Author.Face
	i.CTime = a.PubDate
	i.Autoplay = a.Rights.Autoplay
}

func (i *Item) FromFeedAv(a *feedArcApi.Arc) {
	if i.Title == "" {
		i.Title = a.Title
	}
	if i.Cover == "" {
		i.Cover = model.CoverURLHTTPS(a.Pic)
	} else {
		i.Cover = model.CoverURLHTTPS(i.Cover)
	}
	i.Param = strconv.FormatInt(a.Aid, 10)
	i.Goto = model.GotoAv
	i.URI = model.FillURI(i.Goto, i.Param, 0, 0, model.FeedAvPlayHandler(a))
	i.Cid = a.FirstCid
	i.Rid = a.TypeID
	i.TName = a.TypeName
	i.Desc = strconv.Itoa(int(a.Stat.Danmaku)) + "弹幕"
	i.fillFeedArcStat(a)
	i.Duration = a.Duration
	i.Mid = a.Author.Mid
	i.Name = a.Author.Name
	i.Face = a.Author.Face
	i.CTime = a.PubDate
	i.Autoplay = a.Rights.Autoplay
}

func (i *Item) FromLive(r *live.Room) {
	if r.LiveStatus != 1 || r.Title == "" || r.Cover == "" {
		return
	}
	i.Title = r.Title
	i.Cover = r.Cover
	i.Goto = model.GotoLive
	i.Param = strconv.FormatInt(r.RoomID, 10)
	i.URI = model.FillURI(i.Goto, i.Param, 0, 0, model.LiveRoomHandler(r))
	i.Name = r.Uname
	i.Mid = r.UID
	i.Face = r.Face
	i.Online = int(r.Online)
	i.Autoplay = 1
	// i.Area = r.Area
	// i.AreaID = r.AreaID
	i.Area2 = &Area2{ID: r.AreaV2ParentID, Name: r.AreaV2ParentName, Children: &Area2{ID: r.AreaV2ID, Name: r.AreaV2Name}}
	i.Autoplay = 1
}

func (i *Item) FromSeason(b *episodegrpc.EpisodeCardsProto) {
	if b.Season == nil {
		return
	}
	if i.Title == "" {
		i.Title = b.Season.Title
	}
	if i.Cover == "" {
		i.Cover = b.Cover
	}
	i.Goto = model.GotoBangumi
	i.Param = strconv.Itoa(int(b.EpisodeId))
	i.URI = model.FillURI(i.Goto, i.Param, 0, 0, nil)
	if b.Season.Stat != nil {
		i.Play = int(b.Season.Stat.View)
		i.Fav = int(b.Season.Stat.Follow)
	}
	i.Type = int8(b.Season.SeasonType)
	i.Badge = b.Season.SeasonTypeName
	i.Desc = b.Season.NewEpShow
	i.Face = b.Season.Cover
	i.Square = b.Season.Cover
}

func (i *Item) FromPGCSeason(s *episodegrpc.EpisodeCardsProto) {
	if i.Title == "" {
		i.Title = s.Season.Title
	}
	if i.Cover == "" {
		i.Cover = s.Cover
	}
	i.Goto = model.GotoBangumi
	i.Param = strconv.Itoa(int(s.EpisodeId))
	i.URI = model.FillURI(model.GotoBangumi, i.Param, 0, 0, nil)
	i.Index = s.Title
	i.IndexTitle = s.LongTitle
	i.Status = int8(s.Season.SeasonStatus)
	i.CoverMark = s.Season.Badge
	if s.Season.Stat != nil {
		i.Play = int(s.Season.Stat.View)
		i.Fav = int(s.Season.Stat.Follow)
	}
	i.Type = int8(s.Season.SeasonType)
	i.Badge = s.Season.SeasonTypeName
	if s.Season.IsFinish == 1 {
		i.Finish = true
	}
	i.Count = int(s.Season.TotalCount)
	i.LatestIndex = s.Title
	i.Desc = s.Season.NewEpShow
	i.Face = s.Season.Cover
	i.Square = s.Season.Cover
}

func (i *Item) FromLogin() {
	if i.Param == "0" {
		i.Param = "1"
	}
	i.Goto = model.GotoLogin
}

func (i *Item) FromAdAv(adInfo *cm.AdInfo, a *arcgrpc.Arc) {
	// ad
	i.RequestID = adInfo.RequestID
	i.CreativeID = adInfo.CreativeID
	i.SrcID = adInfo.Source
	i.IsAdLoc = adInfo.IsAdLoc
	i.IsAd = adInfo.IsAd
	i.AdCb = adInfo.AdCb
	i.CmMark = adInfo.CmMark
	i.AdIndex = adInfo.Index
	c := adInfo.CreativeContent
	i.Title = c.Title
	i.Desc = c.Desc
	i.Cover = c.ImageURL
	i.Goto = model.GotoAdAv
	i.Name = a.Author.Name
	i.Face = c.LogURL
	i.ShowURL = c.ShowURL
	i.ClickURL = c.ClickURL
	// archive
	i.Param = strconv.FormatInt(a.Aid, 10)
	i.URI = model.FillURI(i.Goto, i.Param, 0, 0, nil)
	if a.TypeName != "广告" {
		i.TName = a.TypeName
	}
	i.fillArcStat(a)
	i.Duration = a.Duration
	i.Mid = a.Author.Mid
	if i.Name == "" {
		i.Name = a.Author.Name
	}
	if i.Face == "" {
		i.Face = a.Author.Face
	}
	i.CTime = a.Ctime
	i.Extra = adInfo.Extra
	i.CardIndex = adInfo.CardIndex
}

func (i *Item) FromAdWeb(adInfo *cm.AdInfo) {
	i.RequestID = adInfo.RequestID
	i.CreativeID = adInfo.CreativeID
	i.SrcID = adInfo.Source
	i.IsAdLoc = adInfo.IsAdLoc
	i.IsAd = adInfo.IsAd
	i.AdCb = adInfo.AdCb
	i.CmMark = adInfo.CmMark
	i.AdIndex = adInfo.Index
	c := adInfo.CreativeContent
	i.Title = c.Title
	i.Desc = c.Desc
	i.Cover = c.ImageURL
	i.Goto = model.GotoAdWeb
	i.URI = model.FillURI(i.Goto, c.URL, 0, 0, nil)
	i.ShowURL = c.ShowURL
	i.ClickURL = c.ClickURL
	i.Extra = adInfo.Extra
	i.CardIndex = adInfo.CardIndex
}

func (i *Item) FromAdLarge(adInfo *cm.AdInfo) {
	i.RequestID = adInfo.RequestID
	i.CreativeID = adInfo.CreativeID
	i.SrcID = adInfo.Source
	i.IsAdLoc = adInfo.IsAdLoc
	i.IsAd = adInfo.IsAd
	i.AdCb = adInfo.AdCb
	i.CmMark = adInfo.CmMark
	i.AdIndex = adInfo.Index
	c := adInfo.CreativeContent
	i.Title = c.Title
	i.Desc = c.Desc
	i.Goto = model.GotoAdLarge
	i.URI = model.FillURI(i.Goto, c.URL, 0, 0, nil)
	i.ShowURL = c.ShowURL
	i.ClickURL = c.ClickURL
	i.Extra = adInfo.Extra
	i.CardIndex = adInfo.CardIndex
}

func (i *Item) FromAdWebS(adInfo *cm.AdInfo) {
	i.RequestID = adInfo.RequestID
	i.CreativeID = adInfo.CreativeID
	i.SrcID = adInfo.Source
	i.IsAdLoc = adInfo.IsAdLoc
	i.IsAd = adInfo.IsAd
	i.AdCb = adInfo.AdCb
	i.CmMark = adInfo.CmMark
	i.AdIndex = adInfo.Index
	c := adInfo.CreativeContent
	i.Title = c.Title
	i.Desc = c.Desc
	i.Cover = c.ImageURL
	i.Goto = model.GotoAdWebS
	i.URI = model.FillURI(i.Goto, c.URL, 0, 0, nil)
	i.ShowURL = c.ShowURL
	i.ClickURL = c.ClickURL
	i.Extra = adInfo.Extra
	i.CardIndex = adInfo.CardIndex
}

func (i *Item) FromSpecial(id int64, title, cover, desc, url string, typ int, badge string, size string) {
	if title == "" || cover == "" {
		return
	}
	i.Title = title
	i.Cover = cover
	i.Goto = model.GotoSpecial
	i.URI = model.FillURI(model.OperateType[typ], url, 0, 0, nil)
	i.Redirect = model.FillRedirect(i.Goto, typ)
	i.Desc = desc
	i.Param = strconv.FormatInt(id, 10)
	i.HideBadge = true
	i.Badge = badge
	var ratio float64
	if size == "1020x300" {
		ratio = 34
	} else if size == "1020x378" {
		ratio = 27
	}
	i.Ratio = ratio
}

func (i *Item) FromSpecialS(id int64, title, cover, square, desc, url string, typ int, badge string) {
	if title == "" || cover == "" {
		return
	}
	i.Title = title
	i.Cover = cover
	// 活不过一个版的单列封面
	if square != "" {
		i.Square = square
	} else {
		i.Square = cover
	}
	i.Goto = model.GotoSpecialS
	i.URI = model.FillURI(model.OperateType[typ], url, 0, 0, nil)
	i.Redirect = model.FillRedirect(i.Goto, typ)
	i.Desc = desc
	i.Param = strconv.FormatInt(id, 10)
	i.Badge = badge
}

func (i *Item) FromRank(ranks []*rank.Rank, am map[int64]*arcgrpc.Arc) {
	if len(ranks) < _rankCount {
		return
	}
	if a, ok := am[ranks[0].Aid]; ok {
		i.Cover1 = a.Pic
	} else {
		return
	}
	if a, ok := am[ranks[1].Aid]; ok {
		i.Cover2 = a.Pic
	} else {
		return
	}
	if a, ok := am[ranks[2].Aid]; ok {
		i.Cover3 = a.Pic
	} else {
		return
	}
	ris := make([]*Item, 0, _rankCount)
	for _, rank := range ranks[:_rankCount] {
		if a, ok := am[rank.Aid]; ok {
			ri := &Item{
				Title: a.Title,
				Cover: a.Pic,
				Goto:  model.GotoAv,
				Param: strconv.FormatInt(a.Aid, 10),
			}
			ri.fillArcStat(a)
			ri.Duration = a.Duration
			ri.URI = model.FillURI(ri.Goto, ri.Param, 0, 0, nil)
			score := int64(rank.Score)
			//nolint:gomnd
			if score < 10000 {
				ri.Score = model.Rounding(score, 0)
			} else if score >= 10000 && score < 100000000 {
				ri.Score = model.Rounding(score, 10000) + "万"
			} else if score >= 100000000 {
				ri.Score = model.Rounding(score, 100000000) + "亿"
			}
			if ri.Score != "" {
				ri.Score = "综合评分:" + ri.Score
			} else {
				ri.Score = "综合评分:-"
			}
			ris = append(ris, ri)
		} else {
			return
		}
	}
	i.Title = _rankTitle
	i.Goto = model.GotoRank
	i.Item = ris
	i.Param = "0"
	i.URI = model.FillURI(i.Goto, _rankURI, 0, 0, nil)
}

func (i *Item) FromBangumiRcmd(u *bangumi.Update) {
	i.Cover = u.SquareCover
	i.Goto = model.GotoBangumiRcmd
	i.Desc = u.Title
	if u.Updates > _bangumiRcmdUpdates {
		i.Updates = _bangumiRcmdUpdates
	} else {
		i.Updates = u.Updates
	}
}

func (i *Item) FromBanner(bs []*banner.Banner, hash string) {
	i.Goto = model.GotoBanner
	i.Hash = hash
	i.BannerItem = bs
}

func (i *Item) FromPlayer(a *arcgrpc.Arc) {
	if a == nil || !a.IsNormal() {
		return
	}
	title := i.Title
	if title == "" {
		title = a.Title
	}
	cover := i.Cover
	if cover == "" {
		cover = a.Pic
	}
	item := &Item{Title: title, Cover: cover, Param: strconv.FormatInt(a.Aid, 10), Goto: model.GotoAv}
	item.URI = model.FillURI(item.Goto, item.Param, 0, 0, nil)
	item.fillArcStat(a)
	i.Item = []*Item{item}
	i.Cid = a.FirstCid
	i.Rid = a.TypeID
	i.TName = a.TypeName
	i.Mid = a.Author.Mid
	i.Goto = model.GotoPlayer
	i.Name = a.Author.Name
	i.Face = a.Author.Face
	i.Duration = a.Duration
	i.Autoplay = a.Rights.Autoplay
}

func (i *Item) FromPlayerLive(r *live.Room) {
	if r.LiveStatus != 1 || r.Title == "" || r.Cover == "" {
		return
	}
	i.Name = r.Uname
	i.Mid = r.UID
	i.Face = r.Face
	item := &Item{Title: r.Title, Cover: r.Cover, Param: strconv.FormatInt(r.RoomID, 10), Goto: model.GotoLive}
	item.URI = model.FillURI(item.Goto, item.Param, 0, 0, model.LiveRoomHandler(r))
	item.Online = int(r.Online)
	item.Area2 = &Area2{ID: r.AreaV2ParentID, Name: r.AreaV2ParentName, Children: &Area2{ID: r.AreaV2ID, Name: r.AreaV2Name}}
	i.Item = []*Item{item}
	i.Goto = model.GotoPlayer
	i.Autoplay = 1
}

func (i *Item) FromRcmdReason(r *ai.RcmdReason) {
	if r != nil {
		if r.Style != 0 {
			i.RcmdReason = &RcmdReason{ID: r.Style, Content: r.Content, BgColor: r.Grounding, IconLocation: r.Position}
			//nolint:gomnd
			if r.Style == 3 {
				i.RcmdReason.Message = i.Name
			}
		} else {
			i.RcmdReason = &RcmdReason{ID: r.ID, Content: r.Content}
		}
	}
}

func (i *Item) FromGameDownloadS(d *operate.Download, plat int8, build int) {
	i.Title = d.Title
	i.Cover = d.DoubleCover
	i.BigCover = d.Cover
	i.Goto = model.GotoGameDownloadS
	i.Desc = d.Desc
	i.URI = model.FillURI(model.OperateType[d.URLType], d.URLValue, plat, build, nil)
	i.Redirect = model.FillRedirect(i.Goto, d.URLType)
	i.Face = d.Icon
	i.Param = strconv.FormatInt(d.ID, 10)
	i.Download = d.Number
	if d.Icon != "" {
		i.Square = d.Icon
	} else {
		i.Square = d.Cover
	}
}

func (i *Item) FromShoppingS(c *show.Shopping) {
	if c.Name == "" || c.URL == "" {
		return
	}
	i.Title = c.Name
	i.STime = c.STime
	i.ETime = c.ETime
	i.City = c.CityName
	if len(c.Tags) != 0 {
		i.PType = c.Tags[0].TagName
	}
	i.Param = strconv.FormatInt(c.ID, 10)
	// 双列封面
	if strings.HasPrefix(c.PerformanceImage, "http:") || strings.HasPrefix(c.PerformanceImage, "https:") {
		i.Cover = c.PerformanceImage
	} else {
		i.Cover = "http:" + c.PerformanceImage
	}
	// 单列封面
	if strings.HasPrefix(c.PerformanceImageP, "http:") || strings.HasPrefix(c.PerformanceImageP, "https:") {
		i.Square = c.PerformanceImageP
	} else {
		i.Square = "http:" + c.PerformanceImageP
	}
	if i.Cover == "" {
		i.Cover = i.Square
	}
	if i.Cover == "" {
		return
	}
	i.Goto = model.GotoShoppingS
	i.URI = model.FillURI(i.Goto, c.URL, 0, 0, nil)
	i.Type = c.Type
	i.Subtitle = c.Subname
	// 漫展需加羊角符
	if i.Type == 1 {
		i.Price = "￥" + c.Pricelt
	} else {
		i.Price = c.Pricelt
	}
	i.Desc = c.Want
}

func (i *Item) FromAudio(a *audio.Audio) {
	i.Title = a.Title
	i.Cover = a.CoverURL
	i.Param = strconv.FormatInt(a.MenuID, 10)
	i.Goto = model.GotoAudio
	i.URI = model.FillURI(i.Goto, i.Param, 0, 0, model.AudioHandler)
	i.Play = int(a.PlayNum)
	i.Count = a.RecordNum
	i.Fav = int(a.FavoriteNum)
	i.Face = a.Face
	// old
	titles := make([]string, 0, len(a.Songs))
	for index, song := range a.Songs {
		if song != nil || song.Title != "" {
			titles = append(titles, song.Title)
			if index == 0 {
				i.SongTitle = song.Title
			}
		}
	}
	i.Desc = strings.Join(titles, ",")
	// new
	for _, ctg := range a.Ctgs {
		tag := &taggrpc.Tag{Id: ctg.ItemID, Name: ctg.ItemVal}
		i.Tags = append(i.Tags, tag)
		//nolint:gomnd
		if len(i.Tags) == 2 {
			break
		}
	}
	// new
	if len(a.Ctgs) != 0 {
		id := a.Ctgs[0].ItemID
		name := a.Ctgs[0].ItemVal
		if len(a.Ctgs) > 1 {
			id = a.Ctgs[1].ItemID
			name += "·" + a.Ctgs[1].ItemVal
		}
		i.Tag = &Tag{Name: name, URI: model.FillURI(model.GotoAudioTag, strconv.FormatInt(id, 10), 0, 0, model.AudioHandler)}
	}
	//nolint:gomnd
	if a.Type == 5 {
		i.Badge = "专辑"
		i.Type = 2
	} else {
		i.Badge = "歌单"
		i.Type = 1
	}
	i.CTime = xtime.Time(a.PaTime)
}

func (i *Item) FromConverge(c *operate.Converge, am map[int64]*arcgrpc.Arc, rm map[int64]*live.Room, artm map[int64]*article.Meta) {
	if len(c.Items) < _convergeMinCount {
		return
	}
	cis := make([]*Item, 0, len(c.Items))
	for _, content := range c.Items {
		ci := &Item{Title: content.Title}
		switch content.Goto {
		case model.GotoAv:
			if a, ok := am[content.Pid]; ok && a != nil && a.IsNormal() {
				if ci.Title == "" {
					ci.Title = a.Title
				}
				ci.Cover = a.Pic
				ci.Goto = model.GotoAv
				ci.Param = strconv.FormatInt(a.Aid, 10)
				ci.URI = model.FillURI(ci.Goto, ci.Param, 0, 0, nil)
				ci.fillArcStat(a)
				ci.Duration = a.Duration
				cis = append(cis, ci)
			}
		case model.GotoLive:
			if r, ok := rm[content.Pid]; ok {
				if r.LiveStatus == 0 || r.Title == "" || r.Cover == "" {
					continue
				}
				if ci.Title == "" {
					ci.Title = r.Title
				}
				ci.Cover = r.Cover
				ci.Goto = model.GotoLive
				ci.Param = strconv.FormatInt(r.RoomID, 10)
				ci.Online = int(r.Online)
				ci.URI = model.FillURI(ci.Goto, ci.Param, 0, 0, model.LiveRoomHandler(r))
				ci.Badge = "直播"
				cis = append(cis, ci)
			}
		case model.GotoArticle:
			if art, ok := artm[content.Pid]; ok {
				ci.Title = art.Title
				ci.Desc = art.Summary
				if len(art.ImageURLs) != 0 {
					ci.Cover = art.ImageURLs[0]
				}
				ci.Goto = model.GotoArticle
				ci.Param = strconv.FormatInt(art.ID, 10)
				ci.URI = model.FillURI(ci.Goto, ci.Param, 0, 0, nil)
				if art.Stats != nil {
					ci.fillArtStat(art)
				}
				ci.Badge = "文章"
				cis = append(cis, ci)
			}
		default:
		}
	}
	if len(cis) < _convergeMinCount {
		return
	}
	i.Item = cis
	i.Goto = model.GotoConverge
	i.URI = model.FillURI(model.OperateType[c.ReType], c.ReValue, 0, 0, nil)
	i.Redirect = model.FillRedirect(i.Goto, c.ReType)
	i.Title = c.Title
	i.Cover = c.Cover
	i.Param = strconv.FormatInt(c.ID, 10)
}

func (i *Item) FromUpBangumi(p *feed.Bangumi) {
	i.Title = p.Title
	i.Cover = p.NewEp.Cover
	i.Goto = model.GotoUpBangumi
	i.Param = strconv.FormatInt(p.EpisodeID, 10)
	i.URI = model.FillURI(i.Goto, i.Param, 0, 0, nil)
	i.Status = int8(p.IsFinish)
	i.Index = p.NewEp.Index
	i.IndexTitle = p.NewEp.IndexTitle
	i.Play = int(p.NewEp.Play)
	i.Danmaku = int(p.NewEp.Dm)
	i.Type = int8(p.BgmType)
	i.Count = int(p.TotalCount)
	i.Updates = int(p.NewEp.EpisodeID)
	i.CTime = xtime.Time(p.Ts)
}

func (i *Item) FromUpLive(f *livemdl.Feed) {
	i.Cover = f.Face
	i.Param = strconv.FormatInt(f.RoomID, 10)
	i.URI = model.FillURI(model.GotoLive, i.Param, 0, 0, nil)
}

func (i *Item) FromUpArticle(m *feedArtApi.Meta) {
	i.Title = m.Title
	i.Desc = m.Summary
	i.Covers = m.ImageURLs
	i.Goto = model.GotoUpArticle
	i.Param = strconv.FormatInt(m.ID, 10)
	i.URI = model.FillURI(i.Goto, i.Param, 0, 0, nil)
	if m.Author != nil {
		i.Mid = m.Author.Mid
		i.Name = m.Author.Name
		i.Face = m.Author.Face
	}
	if m.Category != nil {
		i.Category = &Category{ID: m.Category.ID, Name: m.Category.Name}
	}
	if m.Stats != nil {
		i.Play = int(m.Stats.View)
		i.Reply = int(m.Stats.Reply)
	}
	i.Temple = int(m.TemplateID)
	//nolint:gomnd
	if i.Temple == 4 {
		i.Temple = 1
	}
	i.Template = int(m.TemplateID)
	i.BannerURL = m.BannerURL
	i.CTime = m.PublishTime
}

func (i *Item) FromArticleS(m *article.Meta) {
	if m.State < 0 {
		return
	}
	i.Title = m.Title
	i.Desc = m.Summary
	i.Covers = m.ImageURLs
	i.Goto = model.GotoArticleS
	i.Param = strconv.FormatInt(m.ID, 10)
	i.URI = model.FillURI(i.Goto, i.Param, 0, 0, nil)
	if m.Author != nil {
		i.Mid = m.Author.Mid
		i.Name = m.Author.Name
		i.Face = m.Author.Face
	}
	if len(m.Categories) >= 2 && m.Categories[0] != nil && m.Categories[1] != nil {
		i.Category = &Category{ID: m.Categories[0].ID, Name: m.Categories[0].Name}
		i.Category.Children = &Category{ID: m.Categories[1].ID, Name: m.Categories[1].Name}
	}
	if m.Stats != nil {
		i.fillArtStat(m)
	}
	i.Temple = int(m.TemplateID)
	//nolint:gomnd
	if i.Temple == 4 {
		i.Temple = 1
	}
	i.Template = int(m.TemplateID)
	i.BannerURL = m.BannerURL
	i.CTime = m.PublishTime
}
func (i *Item) FromLiveUpRcmd(id int64, cs []*live.Card, card map[int64]*accountgrpc.Card) {
	//nolint:gomnd
	if len(cs) < 2 {
		return
	}
	is := make([]*Item, 0, 2)
	for _, c := range cs[:2] {
		if c.LiveStatus != 1 {
			return
		}
		it := &Item{}
		it.Title = c.Title
		it.Cover = c.ShowCover
		it.Goto = model.GotoLive
		it.Param = strconv.FormatInt(c.RoomID, 10)
		it.URI = model.FillURI(it.Goto, it.Param, 0, 0, model.LiveUpHandler(c))
		it.Fans = int64(c.Online)
		it.Mid = c.UID
		it.Name = c.Uname
		it.Badge = "直播"
		if card, ok := card[it.Mid]; ok {
			if card.Official.Role != 0 {
				role := card.Official.Role
				//nolint:gomnd
				if card.Official.Role == 7 {
					role = 1
				}
				it.Official = &OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
			}
		}
		is = append(is, it)
	}
	i.Item = is
	i.Goto = model.GotoLiveUpRcmd
	i.Param = strconv.FormatInt(id, 10)
}

func (i *Item) FromWeb(title, cover, uri string) {
	i.Title = title
	i.Cover = cover
	i.Goto = model.GotoWeb
	i.URI = model.FillURI(i.Goto, uri, 0, 0, nil)
	i.Redirect = model.FillRedirect(i.Goto, 0)
}

func (i *Item) FromDislikeReason(plat int8, build int) {
	const (
		_seasonNoSeason = 1
		_seasonRegion   = 2
		_seasonTag      = 3
		_seasonUpper    = 4
		_channelIPhone  = 6720
		_channelAndroid = 5270000
	)
	var reasonName string
	if (plat == model.PlatIPhone && build > _channelIPhone) || (plat == model.PlatAndroid && build >= _channelAndroid) || plat == model.PlatIPhoneB {
		reasonName = "频道:"
	} else {
		reasonName = "标签:"
	}
	if i.Tag != nil {
		i.DisklikeReasons = []*dislike.DisklikeReason{
			{ReasonID: _seasonUpper, ReasonName: "UP主:" + i.Name},
			{ReasonID: _seasonRegion, ReasonName: "分区:" + i.TName},
			{ReasonID: _seasonTag, ReasonName: reasonName + i.Tag.TagName},
			{ReasonID: _seasonNoSeason, ReasonName: "不感兴趣"},
		}
	} else {
		i.DisklikeReasons = []*dislike.DisklikeReason{
			{ReasonID: _seasonUpper, ReasonName: "UP主:" + i.Name},
			{ReasonID: _seasonRegion, ReasonName: "分区:" + i.TName},
			{ReasonID: _seasonNoSeason, ReasonName: "不感兴趣"},
		}
	}
}

func (i *Item) fillArcStat(a *arcgrpc.Arc) {
	if a == nil {
		return
	}
	if a.Access == 0 {
		i.Play = int(a.Stat.View)
	}
	i.Danmaku = int(a.Stat.Danmaku)
	i.Reply = int(a.Stat.Reply)
	i.Fav = int(a.Stat.Fav)
	i.Coin = int(a.Stat.Coin)
	i.Share = int(a.Stat.Share)
	i.Like = int(a.Stat.Like)
	i.Dislike = int(a.Stat.DisLike)
}

func (i *Item) fillFeedArcStat(a *feedArcApi.Arc) {
	if a == nil {
		return
	}
	if a.Access == 0 {
		i.Play = int(a.Stat.View)
	}
	i.Danmaku = int(a.Stat.Danmaku)
	i.Reply = int(a.Stat.Reply)
	i.Fav = int(a.Stat.Fav)
	i.Coin = int(a.Stat.Coin)
	i.Share = int(a.Stat.Share)
	i.Like = int(a.Stat.Like)
	i.Dislike = int(a.Stat.DisLike)
}

func (i *Item) fillArtStat(m *article.Meta) {
	if m == nil {
		return
	}
	i.Play = int(m.Stats.View)
	i.Reply = int(m.Stats.Reply)
}

func (i *Item) FromTabCards(r *operate.Active, am map[int64]*arcgrpc.Arc, downm map[int64]*operate.Download, sm map[int32]*episodegrpc.EpisodeCardsProto, rm map[int64]*live.Room, metam map[int64]*article.Meta, spm map[int64]*operate.Special) {
	items := make([]*Item, 0, len(r.Items))
	for _, r := range r.Items {
		item := &Item{}
		switch r.Goto {
		case model.GotoWeb:
			item.FromWeb(r.Title, r.Cover, model.FillURI(model.GotoWeb, r.Param, 0, 0, nil))
		case model.GotoGame:
			if d, ok := downm[r.Pid]; ok {
				item.FromGameDownloadS(d, 0, 0)
			}
		case model.GotoAv:
			if a, ok := am[r.Pid]; ok {
				item.FromAv(a)
			}
		case model.GotoBangumi:
			if b, ok := sm[int32(r.Pid)]; ok {
				item.FromSeason(b)
			}
		case model.GotoLive:
			if r, ok := rm[r.Pid]; ok {
				item.FromLive(r)
			}
		case model.GotoArticle:
			if m, ok := metam[r.Pid]; ok {
				item.FromArticleS(m)
			}
		case model.GotoSpecial:
			if sc, ok := spm[r.Pid]; ok {
				item.FromSpecialS(sc.ID, sc.Title, sc.Cover, sc.SingleCover, sc.Desc, sc.ReValue, sc.ReType, sc.Badge)
			}
		default:
		}
		if item.Goto != "" {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return
	}
	i.Item = items
	i.Title = r.Title
	i.URI = model.FillURI(model.GotoWeb, r.Param, 0, 0, nil)
	i.Subtitle = r.Subtitle
	i.Goto = r.Type
}

func (i *Item) FromTabTags(r *operate.Active, am map[int64]*arcgrpc.Arc, tagm map[int64]*taggrpc.Tag) {
	items := make([]*Item, 0, len(r.Items))
	for _, r := range r.Items {
		if r == nil {
			continue
		}
		item := &Item{}
		switch r.Goto {
		case model.GotoAv:
			if a, ok := am[r.Pid]; ok {
				item.FromAv(a)
			}
		default:
		}
		if item.Goto != "" {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return
	}
	i.Item = items
	i.Param = strconv.FormatInt(r.Pid, 10)
	if t, ok := tagm[r.Pid]; ok {
		i.Title = t.Name
	}
	i.Goto = r.Type
}

func (i *Item) FromTabBanner(r *operate.Active) {
	i.BannerItem = make([]*banner.Banner, 0, len(r.Items))
	for _, item := range r.Items {
		banner := &banner.Banner{ID: item.Pid, Title: item.Title, Image: item.Cover, URI: cdm.FillURI(item.Goto, 0, 0, item.Param, nil)}
		i.BannerItem = append(i.BannerItem, banner)
	}
	i.Goto = model.GotoBanner
}

func (i *Item) FromNews(r *operate.Active) {
	if r.Desc == "" {
		return
	}
	i.Title = r.Title
	i.Content = r.Desc
	i.Goto = model.GotoTabNews
	i.URI = model.FillURI(model.GotoWeb, r.Param, 0, 0, nil)
}

// 最多配10张卡片 取3个未关注的 不足则不显示该卡片
func (i *Item) FromSubscribe(r *operate.Follow, card map[int64]*accountgrpc.Card, follow map[int64]bool, statm map[int64]*relationgrpc.StatReply, tagm map[int64]*taggrpc.Tag) {
	if r == nil {
		return
	}
	is := make([]*Item, 0, 3)
	switch r.Type {
	case "upper":
		for _, r := range r.Items {
			item := &Item{}
			if card, ok := card[r.Pid]; ok {
				if follow[r.Pid] {
					continue
				}
				item.Name = card.Name
				item.Face = card.Face
				item.Mid = card.Mid
				if card.Official.Role != 0 {
					role := card.Official.Role
					//nolint:gomnd
					if card.Official.Role == 7 {
						role = 1
					}
					item.Official = &OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
				}
				item.IsAtten = 0
				if stat, ok := statm[r.Pid]; ok {
					item.Fans = stat.Follower
				}
				is = append(is, item)
			}
		}
		i.Kind = "upper"
	case "channel_three":
		for _, r := range r.Items {
			item := &Item{}
			tg, ok := tagm[r.Pid]
			if !ok || tg.Attention == 1 {
				continue
			}
			item.Name = tg.Name
			item.Face = tg.Cover
			item.Fans = tg.Sub
			item.IsAtten = int8(tg.Attention)
			item.Param = strconv.FormatInt(tg.Id, 10)
			if item.Face != "" {
				is = append(is, item)
			}
		}
		i.Kind = "channel"
	}
	//nolint:gomnd
	if len(is) < 3 {
		return
	}
	i.Item = is[:3]
	i.Title = r.Title
	i.Param = strconv.FormatInt(r.ID, 10)
	i.Goto = model.GotoSubscribe
}

func (i *Item) FromChannelRcmd(r *operate.Follow, am map[int64]*arcgrpc.Arc, tagm map[int64]*taggrpc.Tag) {
	if r == nil {
		return
	}
	if a, ok := am[r.Pid]; ok {
		i.Goto = model.GotoChannelRcmd
		i.URI = model.FillURI(model.GotoAv, strconv.FormatInt(a.Aid, 10), 0, 0, nil)
		i.Title = a.Title
		i.Cover = a.Pic
		if tag, ok := tagm[r.Tid]; ok {
			i.Tag = &Tag{ID: tag.Id, Name: tag.Name, Face: tag.Cover, Fans: int(tag.Sub), IsAtten: int8(tag.Attention)}
		}
		i.Cid = a.FirstCid
		i.Autoplay = a.Rights.Autoplay
		i.fillArcStat(a)
		i.Duration = a.Duration
		// TODO 等待开启
		// percent := i.Like / (i.Like + i.Dislike) * 100
		// if percent != 0 {
		// 	i.Desc = strconv.Itoa(percent) + "%的人推荐"
		// }
		i.Param = strconv.FormatInt(r.ID, 10)
	}
}

func BuildFeedArc(in *arcgrpc.Arc) *feedArcApi.Arc {
	if in == nil {
		return nil
	}
	var staffs []*feedArcApi.StaffInfo
	for _, v := range in.StaffInfo {
		if v == nil {
			continue
		}
		staffs = append(staffs, &feedArcApi.StaffInfo{
			Mid:       v.Mid,
			Title:     v.Title,
			Attribute: v.Attribute,
		})
	}
	return &feedArcApi.Arc{
		Aid:         in.Aid,
		Videos:      in.Videos,
		TypeID:      in.TypeID,
		TypeName:    in.TypeName,
		Copyright:   in.Copyright,
		Pic:         in.Pic,
		Title:       in.Title,
		PubDate:     in.PubDate,
		Ctime:       in.Ctime,
		Desc:        in.Desc,
		State:       in.State,
		Access:      in.Access,
		Attribute:   in.Attribute,
		Duration:    in.Duration,
		MissionID:   in.MissionID,
		OrderID:     in.OrderID,
		RedirectURL: in.RedirectURL,
		Forward:     in.Forward,
		Rights: feedArcApi.Rights{
			Bp:            in.Rights.Bp,
			Elec:          in.Rights.Elec,
			Download:      in.Rights.Download,
			Movie:         in.Rights.Movie,
			Pay:           in.Rights.Pay,
			HD5:           in.Rights.HD5,
			NoReprint:     in.Rights.NoReprint,
			Autoplay:      in.Rights.Autoplay,
			UGCPay:        in.Rights.UGCPay,
			IsCooperation: in.Rights.IsCooperation,
			UGCPayPreview: in.Rights.UGCPayPreview,
			NoBackground:  in.Rights.NoBackground,
		},
		Author: feedArcApi.Author{
			Mid:  in.Author.Mid,
			Name: in.Author.Name,
			Face: in.Author.Face,
		},
		Stat: feedArcApi.Stat{
			Aid:     in.Stat.Aid,
			View:    in.Stat.View,
			Danmaku: in.Stat.Danmaku,
			Reply:   in.Stat.Reply,
			Fav:     in.Stat.Fav,
			Coin:    in.Stat.Coin,
			Share:   in.Stat.Share,
			NowRank: in.Stat.NowRank,
			HisRank: in.Stat.HisRank,
			Like:    in.Stat.Like,
			DisLike: in.Stat.DisLike,
		},
		ReportResult: in.ReportResult,
		Dynamic:      in.Dynamic,
		FirstCid:     in.FirstCid,
		Dimension: feedArcApi.Dimension{
			Width:  in.Dimension.Width,
			Height: in.Dimension.Height,
			Rotate: in.Dimension.Rotate,
		},
		StaffInfo:   staffs,
		SeasonID:    in.SeasonID,
		AttributeV2: in.AttributeV2,
	}
}

type VerticalTabParam struct {
	Build      int    `form:"build"`
	Platform   string `form:"platform"`
	MobiApp    string `form:"mobi_app"`
	Device     string `form:"device"`
	Network    string `form:"network"`
	Pull       bool   `form:"pull" default:"true"`
	Flush      int    `form:"flush"`
	DeviceType int    `form:"device_type"`
	DeviceName string `form:"device_name"`
	AccessKey  string `form:"access_key"`
	Offset     int32  `form:"offset"`
	ChannelID  int64  `form:"channel_id"`
	Tag        string `form:"tag"`
	SLocale    string `form:"s_locale"`
	CLocale    string `form:"c_locale"`

	Buvid string
	Plat  int8
	Mid   int64
}

type VerticalTag struct {
	Key        string `json:"key"`
	Title      string `json:"title"`
	ServerInfo string `json:"server_info"`
}

type VerticalTagParam struct {
	Build     int    `form:"build"`
	Platform  string `form:"platform"`
	MobiApp   string `form:"mobi_app"`
	Device    string `form:"device"`
	Network   string `form:"network"`
	AccessKey string `form:"access_key"`
	ChannelID int64  `form:"channel_id"`
	SLocale   string `form:"s_locale"`
	CLocale   string `form:"c_locale"`

	Buvid string
	Plat  int8
	Mid   int64
}

type VerticalChannelParam struct {
	ChannelID int64
	Tag       string
	Mid       int64
	Buvid     string
	Offset    int32
	Ps        int32
	Ip        string
}
