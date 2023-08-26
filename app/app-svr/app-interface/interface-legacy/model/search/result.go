package search

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/library/log"
	xtime "go-common/library/time"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/threePointMeta"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	jsonbuilder "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder"
	largecover "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json/builder/large_cover"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	bangumimdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/bangumi"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/banner"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/bplus"
	gmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/game"
	managermodel "go-gateway/app/app-svr/app-interface/interface-legacy/model/manager"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/idsafe/bvid"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	managersearch "git.bilibili.co/bapis/bapis-go/ai/search/mgr/interface"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	esportGRPC "git.bilibili.co/bapis/bapis-go/esports/service"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	livecommon "git.bilibili.co/bapis/bapis-go/live/xroom-gate/common"
	esportsservice "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	mediagrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/media"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	pgcstat "git.bilibili.co/bapis/bapis-go/pgc/service/stat/v1"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	"github.com/pkg/errors"
)

var (
	getHightLight = regexp.MustCompile(`<em.*?em>`)

	videoStrongStyle = &model.ReasonStyle{
		TextColor:        "#FFFFFFFF",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FA8E57",
		BgColorNight:     "#BA6C45",
		BorderColor:      "#FA8E57",
		BorderColorNight: "#BA6C45",
		BgStyle:          model.BgStyleFill,
	}
	videoStrongStyleV2 = &model.ReasonStyle{
		TextColor:        "#FF6633",
		TextColorNight:   "#BF5330",
		BgColor:          "#FFF1ED",
		BgColorNight:     "#3D2D29",
		BorderColor:      "#FFF1ED",
		BorderColorNight: "#3D2D29",
		BgStyle:          model.BgStyleFill,
	}
	videoWeekStyle = &model.ReasonStyle{
		TextColor:        "#FAAB4B",
		TextColorNight:   "#BA833F",
		BgColor:          "",
		BgColorNight:     "",
		BorderColor:      "#FAAB4B",
		BorderColorNight: "#BA833F",
		BgStyle:          model.BgStyleStroke,
	}
	esportButton = map[string]string{
		"booking_text":   "订阅",
		"unbooking_text": "已订阅",
	}
)

// search const
const (
	_emptyLiveCover  = "https://static.hdslb.com/images/transparent.gif"
	_emptyLiveCover2 = "https://i0.hdslb.com/bfs/live/0477300d2adf65062a3d1fb7ef92122b82213b0f.png"

	StarSpace   = 1
	StarChannel = 2

	_styleHorizontal = "horizontal" // 分集展示按照横条样式
	_styleGrid       = "grid"       // 默认宫格

	_channelOfficIconPink  = "https://i0.hdslb.com/bfs/tag/4c0b29e40f239b8093e956ec6623590533ebba1b.png"
	_channelOfficIconWhite = "https://i0.hdslb.com/bfs/tag/3e82aab221dfccab444dafa9e3e95d2953cd4220.png"

	_searchUpBgShow       = 1
	_searchUpLiveFaceShow = 1
	_searchUpSpaceShow    = 1
	_searchUpAvStyleNone  = 0
	_searchUpAvStyleOne   = 1
	_searchUpAvStyleMore  = 2

	_shortLinkHost  = "https://b23.tv"
	_tipsCover      = "https://i0.hdslb.com/bfs/archive/a92eeace0e23e920cd49a888960cc55144567f43.png"
	_tipsCoverNight = "https://i0.hdslb.com/bfs/archive/1e1e7db9b795f9435d9c91873d1279b4e529d2d8.png"

	_sportsStatusReady    = 1
	_sportsStatusStarting = 2
	_sportsStatusFinish   = 3

	// 全文检索类型
	_chapterFullTextType = 1
	_digestFullTextType  = 2

	_100HounourAppBgPicURL = "https://i0.hdslb.com/bfs/ai-search/78ff73a97710a2f3ad681f9de8809b75057cc844.png"
	_100HounourAppFgPicURL = "https://i0.hdslb.com/bfs/ai-search/05150296a9ad8f974135c29fd101740c8aee8c84.png"
)

// Result struct
type Result struct {
	Trackid          string           `json:"trackid,omitempty"`
	Page             int              `json:"page,omitempty"`
	NavInfo          []*NavInfo       `json:"nav,omitempty"`
	Items            ResultItems      `json:"items,omitempty"`
	Item             []*Item          `json:"item,omitempty"`
	OGVCard          *OGVCard         `json:"ogv_card,omitempty"`
	Array            int              `json:"array,omitempty"`
	Attribute        int32            `json:"attribute"`
	EasterEgg        *EasterEgg       `json:"easter_egg,omitempty"`
	ExpStr           string           `json:"exp_str"`
	KeyWord          string           `json:"keyword"`
	ExtraWordList    []string         `json:"extra_word_list,omitempty"`
	OriginExtraWord  string           `json:"org_extra_word,omitempty"`
	SelectBarType    int64            `json:"select_bar_type,omitempty"`
	NewSearchExpNum  int64            `json:"new_search_exp_num,omitempty"`
	AppDisplayOption AppDisplayOption `json:"app_display_option,omitempty"`
}

type AppDisplayOption struct {
	VideoTitleRow        int64 `json:"video_title_row,omitempty"`
	SearchPageVisualOpti int64 `json:"search_page_visual_opti,omitempty"`
}

// ResultItems struct
type ResultItems struct {
	SuggestKeyWord *Item   `json:"suggest_keyword,omitempty"`
	Operation      []*Item `json:"operation,omitempty"`
	Season2        []*Item `json:"season2,omitempty"`
	Season         []*Item `json:"season,omitempty"`
	Upper          []*Item `json:"upper,omitempty"`
	Movie2         []*Item `json:"movie2,omitempty"`
	Movie          []*Item `json:"movie,omitempty"`
	Archive        []*Item `json:"archive,omitempty"`
	LiveRoom       []*Item `json:"live_room,omitempty"`
	LiveUser       []*Item `json:"live_user,omitempty"`
	ESport         []*Item `json:"esport,omitempty"`
}

// NavInfo struct
type NavInfo struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
	Pages int    `json:"pages"`
	Type  int    `json:"type"`
	Show  int    `json:"show_more,omitempty"`
}

// TypeSearch struct
type TypeSearch struct {
	TrackID           string  `json:"trackid"`
	Pages             int     `json:"pages"`
	Total             int     `json:"total"`
	ExpStr            string  `json:"exp_str"`
	KeyWord           string  `json:"keyword"`
	ResultIsRecommend int     `json:"result_is_recommend"`
	Items             []*Item `json:"items,omitempty"`
}

// TypeSearchLiveAll struct
type TypeSearchLiveAll struct {
	TrackID string      `json:"trackid"`
	Pages   int         `json:"pages"`
	Total   int         `json:"total"`
	ExpStr  string      `json:"exp_str"`
	KeyWord string      `json:"keyword"`
	Master  *TypeSearch `json:"live_master,omitempty"`
	Room    *TypeSearch `json:"live_room,omitempty"`
}

// Suggestion struct
type Suggestion struct {
	TrackID string      `json:"trackid"`
	UpUser  interface{} `json:"upuser,omitempty"`
	Bangumi interface{} `json:"bangumi,omitempty"`
	Suggest []string    `json:"suggest,omitempty"`
}

// Suggestion2 struct
type Suggestion2 struct {
	TrackID string  `json:"trackid"`
	List    []*Item `json:"list,omitempty"`
}

// SuggestionResult3 struct
type SuggestionResult3 struct {
	TrackID string  `json:"trackid"`
	ExpStr  string  `json:"exp_str"`
	List    []*Item `json:"list,omitempty"`
}

// RecommendResult struct
type RecommendResult struct {
	TrackID string  `json:"trackid"`
	Title   string  `json:"title,omitempty"`
	Pages   int     `json:"pages"`
	ExpStr  string  `json:"exp_str,omitempty"`
	Items   []*Item `json:"list,omitempty"`
}

// DefaultWordResult struct
type DefaultWordResult struct {
	TrackID string  `json:"trackid"`
	Title   string  `json:"title,omitempty"`
	Pages   int     `json:"pages"`
	Items   []*Item `json:"items,omitempty"`
}

// NoResultRcndResult struct
type NoResultRcndResult struct {
	TrackID string  `json:"trackid"`
	Title   string  `json:"title,omitempty"`
	Pages   int     `json:"pages"`
	Items   []*Item `json:"items,omitempty"`
}

// EasterEgg struct
type EasterEgg struct {
	ID        int64  `json:"id,omitempty"`
	ShowCount int    `json:"show_count,omitempty"`
	EggType   int8   `json:"type,omitempty"` // 1-视频彩蛋 2-跳链彩蛋 3-图片彩蛋(新增)
	URL       string `json:"url,omitempty"`
	// v5.59新增
	CloseCount       int    `json:"close_count,omitempty"`
	MaskTransparency int    `json:"mask_transparency,omitempty"`
	MaskColor        string `json:"mask_color,omitempty"`
	PicType          int    `json:"pic_type,omitempty"` // 图片类型: 1-静态图 2-动态图
	ShowTime         int    `json:"show_time,omitempty"`
	SourceURL        string `json:"source_url,omitempty"`
	SourceMd5        string `json:"source_md5,omitempty"`
	SourceSize       uint   `json:"source_size,omitempty"`
}

// RecommendPreResult struct
type RecommendPreResult struct {
	TrackID string  `json:"trackid"`
	Total   int     `json:"total"`
	Items   []*Item `json:"items,omitempty"`
}

// ResultConverge struct
type ResultConverge struct {
	TrackID    string  `json:"trackid"`
	Pages      int     `json:"pages"`
	Total      int     `json:"total"`
	UserItems  []*Item `json:"user_items,omitempty"`
	VideoItems []*Item `json:"video_items,omitempty"`
	ExpStr     string  `json:"exp_str,omitempty"`
}

// SpaceResult struct
type SpaceResult struct {
	Trackid string  `json:"trackid,omitempty"`
	Page    int     `json:"page,omitempty"`
	Total   int     `json:"total"`
	Item    []*Item `json:"item,omitempty"`
}

type Badge struct {
	Text    string `json:"text,omitempty"`
	BgCover string `json:"bg_cover,omitempty"`
}

type Notice struct {
	Mid            int64  `json:"mid"`
	NoticeID       int64  `json:"notice_id"`
	Content        string `json:"content"`
	URL            string `json:"url"`
	NoticeType     int64  `json:"notice_type"`
	Icon           string `json:"icon"`
	IconNight      string `json:"icon_night"`
	TextColor      string `json:"text_color"`
	TextColorNight string `json:"text_color_night"`
	BGColor        string `json:"bg_color"`
	BGColorNight   string `json:"bg_color_night"`
}

type ExtraLink struct {
	Text string `json:"text,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type RankInfo struct {
	SearchNightIconUrl   string `json:"search_night_icon_url,omitempty"`
	SearchDayIconUrl     string `json:"search_day_icon_url,omitempty"`
	SearchBkgNightColor  string `json:"search_bkg_night_color,omitempty"`
	SearchBkgDayColor    string `json:"search_bkg_day_color,omitempty"`
	SearchFontNightColor string `json:"search_font_night_color,omitempty"`
	SearchFontDayColor   string `json:"search_font_day_color,omitempty"`
	RankContent          string `json:"rank_content,omitempty"`
	RankLink             string `json:"rank_link,omitempty"`
}

type FullTextResult struct {
	Type              int    `json:"type"`
	ShowText          string `json:"show_text"`
	JumpStartProgress int64  `json:"jump_start_progress"`
	JumpUri           string `json:"jump_uri"`
}

// Item struct
type Item struct {
	TrackID        string                   `json:"trackid,omitempty"`
	LinkType       string                   `json:"linktype,omitempty"`
	Position       int                      `json:"position,omitempty"`
	SuggestKeyword string                   `json:"suggest_keyword,omitempty"`
	Title          string                   `json:"title,omitempty"`
	Name           string                   `json:"name,omitempty"`
	Cover          string                   `json:"cover,omitempty"`
	URI            string                   `json:"uri,omitempty"`
	Param          string                   `json:"param,omitempty"`
	Goto           string                   `json:"goto,omitempty"`
	SharePlane     *appcardmodel.SharePlane `json:"share_plane,omitempty"` // 分享面板

	// av
	Play          int                  `json:"play,omitempty"`
	Danmaku       int                  `json:"danmaku,omitempty"`
	Author        string               `json:"author,omitempty"`
	ViewType      string               `json:"view_type,omitempty"`
	PTime         xtime.Time           `json:"ptime,omitempty"`
	RecTags       []string             `json:"rec_tags,omitempty"`
	IsPay         int                  `json:"is_pay,omitempty"`
	NewRecTags    []*model.ReasonStyle `json:"new_rec_tags,omitempty"`
	ShowCardDesc1 string               `json:"show_card_desc_1,omitempty"`
	ShowCardDesc2 string               `json:"show_card_desc_2,omitempty"`
	FullText      *FullTextResult      `json:"full_text,omitempty"`
	// bangumi season
	SeasonID       int64   `json:"season_id,omitempty"`
	SeasonType     int     `json:"season_type,omitempty"`
	SeasonTypeName string  `json:"season_type_name,omitempty"`
	Finish         int8    `json:"finish,omitempty"`
	Started        int8    `json:"started,omitempty"`
	Index          string  `json:"index,omitempty"`
	NewestCat      string  `json:"newest_cat,omitempty"`
	NewestSeason   string  `json:"newest_season,omitempty"`
	CatDesc        string  `json:"cat_desc,omitempty"`
	TotalCount     int     `json:"total_count,omitempty"`
	MediaType      int     `json:"media_type,omitempty"`
	PlayState      int     `json:"play_state,omitempty"`
	Style          string  `json:"style,omitempty"`
	Styles         string  `json:"styles,omitempty"`
	CV             string  `json:"cv,omitempty"`
	Rating         float64 `json:"rating,omitempty"`
	Vote           int     `json:"vote,omitempty"`
	RatingCount    int     `json:"rating_count,omitempty"`
	// BadgeType    int     `json:"badge_type,omitempty"`
	OutName string `json:"out_name,omitempty"`
	OutIcon string `json:"out_icon,omitempty"`
	OutURL  string `json:"out_url,omitempty"`
	// upper
	Sign           string           `json:"sign,omitempty"`
	Fans           int              `json:"fans,omitempty"`
	Level          int              `json:"level,omitempty"`
	Desc           string           `json:"desc,omitempty"`
	OfficialVerify *OfficialVerify  `json:"official_verify,omitempty"`
	Vip            *account.VipInfo `json:"vip,omitempty"`
	FaceNftNew     int32            `json:"face_nft_new,omitempty"`     // face_nft_new 1 nft头像 0 非nft头像
	NftFaceIcon    *NftFaceIcon     `json:"nft_face_icon,omitempty"`    // nft角标展示信息
	IsSeniorMember int32            `json:"is_senior_member,omitempty"` // is_senior_member 1 硬核会员 0 非硬核会员
	NftDamrk       string           `json:"nft_damrk,omitempty"`        // 直播方nft头像角标资源
	AvItems        []*Item          `json:"av_items,omitempty"`
	AvStyle        int              `json:"av_style,omitempty"`
	Item           []*Item          `json:"item,omitempty"`
	CTime          int64            `json:"ctime,omitempty"`
	CTimeLabel     string           `json:"ctime_label,omitempty"`
	IsUp           bool             `json:"is_up,omitempty"`
	LiveURI        string           `json:"live_uri,omitempty"`
	LiveFace       int              `json:"live_face,omitempty"`
	Background     *Background      `json:"background,omitempty"`
	Space          *SpaceEntrance   `json:"space,omitempty"`
	Notice         *Notice          `json:"notice,omitempty"`
	// movie
	ScreenDate string `json:"screen_date,omitempty"`
	Area       string `json:"area,omitempty"`
	CoverMark  string `json:"cover_mark,omitempty"`
	// arc and sp
	Arcs int `json:"archives,omitempty"`
	// arc and movie
	Duration    string `json:"duration,omitempty"`
	DurationInt int64  `json:"duration_int,omitempty"`
	Actors      string `json:"actors,omitempty"`
	Staff       string `json:"staff,omitempty"`
	Length      int    `json:"length,omitempty"`
	Status      int    `json:"status,omitempty"`
	// live
	RoomID      int64                   `json:"roomid,omitempty"`
	Mid         int64                   `json:"mid,omitempty"`
	Type        string                  `json:"type,omitempty"`
	Attentions  int                     `json:"attentions,omitempty"`
	LiveStatus  int                     `json:"live_status,omitempty"`
	Tags        string                  `json:"tags,omitempty"`
	Region      int                     `json:"region,omitempty"`
	Online      int                     `json:"online,omitempty"`
	ShortID     int                     `json:"short_id,omitempty"`
	CateName    string                  `json:"area_v2_name,omitempty"`
	IsSelection int                     `json:"is_selection,omitempty"`
	WatchedShow *livecommon.WatchedShow `json:"watched_show,omitempty"`
	// article
	ID         int64    `json:"id,omitempty"`
	TemplateID int      `json:"template_id,omitempty"`
	ImageUrls  []string `json:"image_urls,omitempty"`
	View       int      `json:"view,omitempty"`
	Like       int      `json:"like,omitempty"`
	Reply      int      `json:"reply,omitempty"`
	// special
	Badge             string          `json:"badge,omitempty"`
	RightTopLiveBadge *card.LiveBadge `json:"right_top_live_badge,omitempty"`
	RcmdReason        *RcmdReason     `json:"rcmd_reason,omitempty"`
	// media bangumi and mdeia ft
	Prompt         string        `json:"prompt,omitempty"`
	Episodes       []*Item       `json:"episodes,omitempty"`
	Label          string        `json:"label,omitempty"`
	WatchButton    *WatchButton  `json:"watch_button,omitempty"`
	FollowButton   *FollowButton `json:"follow_button,omitempty"`
	SelectionStyle string        `json:"selection_style,omitempty"` // grid || horizontal
	IsOut          int           `json:"is_out,omitempty"`          // is all_net_search
	CheckMore      *CheckMore    `json:"check_more,omitempty"`
	EpisodesNew    []*EpisodeNew `json:"episodes_new,omitempty"`
	// game
	Reserve       string    `json:"reserve,omitempty"`
	NoticeName    string    `json:"notice_name,omitempty"`
	NoticeContent string    `json:"notice_content,omitempty"`
	GiftContent   string    `json:"gift_content,omitempty"`
	GiftURL       string    `json:"gift_url,omitempty"`
	ReserveStatus int64     `json:"reserve_status,omitempty"`
	GameRank      int64     `json:"game_rank,omitempty"`
	RankType      int64     `json:"rank_type,omitempty"`
	RankInfo      *RankInfo `json:"rank_info,omitempty"`
	//云游戏
	ShowCloudGameEntry bool             `json:"show_cloud_game_entry,omitempty"`
	CloudGameParams    *CloudGameParams `json:"cloud_game_params,omitempty"`
	// user
	Face string `json:"face,omitempty"`
	// suggest
	From      string  `json:"from,omitempty"`
	KeyWord   string  `json:"keyword,omitempty"`
	CoverSize float64 `json:"cover_size,omitempty"`
	SugType   string  `json:"sug_type,omitempty"`
	TermType  int     `json:"term_type,omitempty"`
	// rcmd query
	List       []*Item `json:"list,omitempty"`
	FromSource string  `json:"from_source,omitempty"`
	// live master
	UCover         string `json:"ucover,omitempty"`
	VerifyType     int    `json:"verify_type,omitempty"`
	VerifyDesc     string `json:"verify_desc,omitempty"`
	LevelColor     int64  `json:"level_color,omitempty"`
	IsAttention    int    `json:"is_atten,omitempty"`
	CateParentName string `json:"cate_parent_name,omitempty"`
	CateNameNew    string `json:"cate_name,omitempty"`
	Glory          *Glory `json:"glory_info,omitempty"`
	LiveLink       string `json:"live_link,omitempty"` // 专门用于透传直播跳转链接用的字段
	CardLeftIcon   int    `json:"card_left_icon,omitempty"`
	CardLeftText   string `json:"card_left_text,omitempty"`
	// inline-live
	IsInlineLive     int64              `json:"is_inline_live,omitempty"`
	IsLiveRoomInline int64              `json:"is_live_room_inline,omitempty"`
	IsUGCInline      int64              `json:"is_ugc_inline,omitempty"`
	IsOGVInline      int64              `json:"is_ogv_inline,omitempty"`
	InlineType       string             `json:"inline_type,omitempty"`
	InlineLive       *SearchEmbedInline `json:"inline_live,omitempty"`      // 用户卡的 inline 字段
	LiveRoomInline   *SearchEmbedInline `json:"live_room_inline,omitempty"` // 直播卡的 inline 字段
	UGCInline        *SearchEmbedInline `json:"ugc_inline,omitempty"`       // UGC 卡的 inline 字段
	OGVInline        *SearchEmbedInline `json:"ogv_inline,omitempty"`       // OGV 卡的 inline 字段
	// twitter
	Covers     []string `json:"covers,omitempty"`
	CoverCount int      `json:"cover_count,omitempty"`
	Upper      *Item    `json:"upper,omitempty"`
	State      *Item    `json:"stat,omitempty"`
	PTimeText  string   `json:"ptime_text,omitempty"`
	// star
	TagItems []*Item `json:"tag_items,omitempty"`
	TagID    int64   `json:"tag_id,omitempty"`
	URIType  int     `json:"uri_type,omitempty"`
	// ticket
	ShowTime      string `json:"show_time,omitempty"`
	City          string `json:"city,omitempty"`
	Venue         string `json:"venue,omitempty"`
	Price         int    `json:"price,omitempty"`
	PriceComplete string `json:"price_complete,omitempty"`
	PriceType     int    `json:"price_type,omitempty"`
	ReqNum        int    `json:"required_number,omitempty"`
	// product
	ShopName string `json:"shop_name,omitempty"`
	// specialer_guide
	Phone    string               `json:"phone,omitempty"`
	Badges   []*model.ReasonStyle `json:"badges,omitempty"`
	ComicURL string               `json:"comic_url,omitempty"`
	// suggest_keyword
	SugKeyWordType int `json:"sugKeyWord_type,omitempty"`
	// operate
	ContentURI string  `json:"content_uri,omitempty"`
	DyTopic    []*Item `json:"dy_topic,omitempty"`
	IsActivity int     `json:"is_activity,omitempty"`
	// ogv card
	SpecialBgColor string             `json:"special_bg_color,omitempty"`
	MoreText       string             `json:"more_text,omitempty"`
	MoreURL        string             `json:"more_url,omitempty"`
	CoverLeftText  string             `json:"cover_left_text,omitempty"`
	Items          []*Item            `json:"items,omitempty"`
	BadgeStyle     *model.ReasonStyle `json:"cover_badge_style,omitempty"`
	ModuleID       int64              `json:"module_id,omitempty"`
	OgvClipInfo    *OgvClipInfo       `json:"ogv_clip_info,omitempty"`
	OgvInlineExp   int64              `json:"ogv_inline_exp,omitempty"`
	// esport
	BgCover     string       `json:"bg_cover,omitempty"`
	MatchTop    *MatchItem   `json:"match_top,omitempty"`
	MatchBottom *MatchItem   `json:"match_bottom,omitempty"`
	Team1       *MatchTeam   `json:"team_1,omitempty"`
	Team2       *MatchTeam   `json:"team_2,omitempty"`
	MatchLabel  *MatchItem   `json:"match_label,omitempty"`
	MatchTime   *MatchItem   `json:"match_time,omitempty"`
	MatchStage  string       `json:"match_stage,omitempty"`
	MatchButton *MatchItem   `json:"match_button,omitempty"`
	IsOlympic   bool         `json:"is_olympic,omitempty"`
	ExtraLink   []*ExtraLink `json:"extra_link,omitempty"`
	Right       bool         `json:"-"`
	// new_channel
	TypeIcon       string        `json:"type_icon,omitempty"`
	ChannelLabel1  *SearchButton `json:"channel_label1,omitempty"`
	ChannelLabel2  *SearchButton `json:"channel_label2,omitempty"`
	ChannelButton  *SearchButton `json:"channel_button,omitempty"`
	DesignType     string        `json:"design_type,omitempty"`
	CoverLeftText1 string        `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 cardmdl.Icon  `json:"cover_left_icon_1,omitempty"`
	Badge2         *Badge        `json:"badge2,omitempty"`
	MediaId        int64         `json:"media_id,omitempty"`
	// tips 副标题
	SubTitle string `json:"sub_title,omitempty"`
	// tips 夜间背景图
	CoverNight string `json:"cover_night,omitempty"`
	// 回粉
	Relation *cardmdl.Relation `json:"relation,omitempty"`
	// 三点字段
	ThreePoint []*ThreePoint `json:"three_point,omitempty"`
	// 分享字段
	Share *Share `json:"share,omitempty"`
	// 卡片业务角标 https://www.tapd.bilibili.co/20055921/prong/stories/view/1120055921002039718
	CardBusinessBadge *CardBusinessBadge `json:"card_business_badge,omitempty"`
	// 聚合卡查看更多是否隐藏
	HideConvergeReadMore bool `json:"hide_coverge_read_more,omitempty"`
	// 百科卡
	ReadMore              *ReadMore     `json:"read_more,omitempty"`
	Navigation            []*Navigation `json:"navigation,omitempty"`
	NavigationModuleCount int64         `json:"navigation_module_count,omitempty"`
	PediaCover            *PediaCover   `json:"pedia_cover,omitempty"`
	// 游戏强化卡
	GameBaseId      int64            `json:"game_base_id,omitempty"`
	GameIcon        string           `json:"game_icon,omitempty"`
	GameStatus      int64            `json:"game_status,omitempty"`
	Score           string           `json:"score,omitempty"`
	VideoCoverImage string           `json:"video_cover_image,omitempty"`
	TopGameUI       *TopGameUI       `json:"top_game_ui,omitempty"`
	ButtonType      int64            `json:"button_type,omitempty"`
	SportsMatchItem *SportsMatchItem `json:"sports_match_item,omitempty"` // 体育卡
	BottomButton    *BottomButton    `json:"bottom_button,omitempty"`     // 合集卡底部按钮
	CollectionIcon  string           `json:"collection_icon,omitempty"`   // 合集卡icon
}

type NftFaceIcon struct {
	RegionType int32  `json:"region_type"` // nft所属区域 0 默认 1 大陆 2 港澳台
	Icon       string `json:"icon"`        // 角标链接
	ShowStatus int32  `json:"show_status"` // 展示状态 0:默认 1:放大20% 2:原图大小
}

type BottomButton struct {
	Desc string `json:"desc,omitempty"`
	Link string `json:"link,omitempty"`
}

type OgvClipInfo struct {
	PlayStartTime int64 `json:"play_start_time,omitempty"`
	PlayEndTime   int64 `json:"play_end_time,omitempty"`
}

type CloudGameParams struct {
	SourceFrom int64  `json:"source_from,omitempty"`
	Scene      string `json:"scene,omitempty"`
}

type SportsMatchItem struct {
	MatchId         int64  `json:"match_id,omitempty"`
	SeasonId        int64  `json:"season_id,omitempty"`
	MatchName       string `json:"match_name,omitempty"`
	Img             string `json:"img,omitempty"`
	BeginTimeDesc   string `json:"begin_time_desc,omitempty"`
	MatchStatusDesc string `json:"match_status_desc,omitempty"`
	SubContent      string `json:"sub_content,omitempty"`
	SubExtraIcon    string `json:"sub_extra_icon,omitempty"`
}

type TopGameUI struct {
	BackgroundImage   string `json:"background_image,omitempty"`
	CoverDefaultColor string `json:"cover_default_color,omitempty"`
	GaussianBlurValue string `json:"gaussian_blur_value,omitempty"`
	MarkColorValue    string `json:"mask_color_value,omitempty"`
	MaskOpacity       string `json:"mask_opacity,omitempty"`
	ModuleColor       string `json:"module_color,omitempty"`
}

type PediaCover struct {
	CoverType     int64  `json:"cover_type,omitempty"`
	CoverSunURL   string `json:"cover_sun_url,omitempty"`
	CoverNightURL string `json:"cover_night_url,omitempty"`
	CoverWidth    int64  `json:"cover_width,omitempty"`
	CoverHeight   int64  `json:"cover_height,omitempty"`
}

type Navigation struct {
	ID             int64             `json:"id,omitempty"`
	Children       []*Navigation     `json:"children,omitempty"`
	InlineChildren []*Navigation     `json:"inline_children,omitempty"`
	Title          string            `json:"title,omitempty"`
	URI            string            `json:"uri,omitempty"`
	Button         *NavigationButton `json:"button,omitempty"`
}

type NavigationButton struct {
	Type int64  `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type ReadMore struct {
	Text string `json:"text,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type CardBusinessBadge struct {
	GotoIcon   *jsoncard.GotoIcon    `json:"goto_icon,omitempty"`
	BadgeStyle *jsoncard.ReasonStyle `json:"badge_style,omitempty"`
}

type ThreePoint struct {
	Type  string `json:"type"`
	Icon  string `json:"icon"`
	Title string `json:"title"`
}

type Share struct {
	Type string `json:"type"`
	//nolint:staticcheck
	Video *ShareVideo `json:"video,omitempt"`
}

type ShareVideo struct {
	Bvid          string `json:"bvid"`
	CID           int64  `json:"cid"`
	ShareSubtitle string `json:"share_subtitle"`
	IsHotLabel    bool   `json:"is_hot_label"`
	Page          int    `json:"page"`
	PageCount     int64  `json:"page_count"`
	ShortLink     string `json:"short_link"`
}

func (sv *ShareVideo) FormShareVideo(ap *api.ArcPlayer, ishot bool) {
	if ap.Arc == nil {
		return
	}
	a := ap.Arc
	sv.CID = a.FirstCid
	//nolint:gomnd
	if a.Stat.View > 100000 {
		tmp := strconv.FormatFloat(float64(a.Stat.View)/10000, 'f', 1, 64)
		sv.ShareSubtitle = "已观看" + strings.TrimSuffix(tmp, ".0") + "万次"
	}
	sv.IsHotLabel = ishot
	sv.Page = 1
	sv.PageCount = a.Videos
	sv.ShortLink = fmt.Sprintf(_shortLinkHost+"/av%d", a.Aid)
	bvid, err := bvid.AvToBv(a.Aid)
	if err == nil {
		sv.ShortLink = fmt.Sprintf(_shortLinkHost+"/%s", bvid)
	}
	sv.Bvid = bvid
}

type Background struct {
	Show     int    `json:"show"`
	BgPicURL string `json:"bg_pic_url"`
	FgPicURL string `json:"fg_pic_url"`
}

type SpaceEntrance struct {
	Show           int    `json:"show"`
	Test           string `json:"text"`
	TextColor      string `json:"text_color"`
	TextColorNight string `json:"text_color_night"`
	SpaceURL       string `json:"space_url"`
}

type MatchItem struct {
	State          int               `json:"state,omitempty"`
	Text           string            `json:"text,omitempty"`
	TextColor      string            `json:"text_color,omitempty"`
	TextColorNight string            `json:"text_color_night,omitempty"`
	URI            string            `json:"uri,omitempty"`
	LiveLink       string            `json:"live_link,omitempty"`
	Texts          map[string]string `json:"texts,omitempty"`
}

type MatchTeam struct {
	ID    int64  `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Cover string `json:"cover,omitempty"`
	Score int64  `json:"score,omitempty"`
}

type OGVCard struct {
	TrackID        string `json:"trackid,omitempty"`
	LinkType       string `json:"linktype,omitempty"`
	Goto           string `json:"goto,omitempty"`
	Param          string `json:"param,omitempty"`
	Title          string `json:"title,omitempty"`
	Position       int    `json:"position,omitempty"`
	SubTitle1      string `json:"sub_title1,omitempty"`
	SubTitle2      string `json:"sub_title2,omitempty"`
	Cover          string `json:"cover,omitempty"`
	BgCover        string `json:"bg_cover,omitempty"`
	SpecialBgColor string `json:"special_bg_color,omitempty"`
	URI            string `json:"uri,omitempty"`
	CoverURI       string `json:"cover_uri,omitempty"`
}

// EpisodeNew is new structure of episode given by pgc grpc
type EpisodeNew struct {
	Title    string                        `json:"title,omitempty"`
	Uri      string                        `json:"uri,omitempty"`
	Param    string                        `json:"param,omitempty"`
	IsNew    int32                         `json:"is_new"`           // 1=is new, 0=not new
	Badges   []*pgcsearch.SearchBadgeProto `json:"badges,omitempty"` // badges
	Type     int32                         `json:"type,omitempty"`
	Position int                           `json:"position,omitempty"`
}

// CheckMore is displayed only if none of episode has been hit
type CheckMore struct {
	Content string `json:"content"`
	Uri     string `json:"uri"`
}

// WatchButton is the button of watch
type WatchButton struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

// FollowButton is the button of follow
type FollowButton struct {
	Icon         string            `json:"icon"`
	Texts        map[string]string `json:"texts,omitempty"`
	StatusReport string            `json:"status_report"`
}

// FromPGCCard builds the follow button from search card of PGC
func (v *FollowButton) FromPGCCard(card *pgcsearch.SearchFollowProto) {
	if card == nil {
		return
	}
	v.Icon = card.Icon
	v.StatusReport = card.StatusReport
	if len(card.Text) > 0 {
		v.Texts = make(map[string]string, len(card.Text))
		for key, value := range card.Text {
			v.Texts[fmt.Sprintf("%d", key)] = value
		}
	}
}

// Glory live struct
type Glory struct {
	Title string  `json:"title,omitempty"`
	Total int     `json:"total"`
	Items []*Item `json:"items,omitempty"`
}

// RcmdReason struct
type RcmdReason struct {
	Content string `json:"content,omitempty"`
}

// UserResult struct
type UserResult struct {
	Items []*Item `json:"items,omitempty"`
}

// DefaultWords struct
type DefaultWords struct {
	Trackid   string `json:"trackid,omitempty"`
	Param     string `json:"param,omitempty"`
	Show      string `json:"show,omitempty"`
	Word      string `json:"word,omitempty"`
	ShowFront int    `json:"show_front,omitempty"`
	Value     string `json:"value,omitempty"`
	URI       string `json:"uri,omitempty"`
	Goto      string `json:"goto,omitempty"`
	ExpStr    string `json:"exp_str,omitempty"`
}

// FromSeason form func
func (i *Item) FromSeason(b *Bangumi, bangumi string) {
	i.Title = b.Title
	i.Cover = b.Cover
	i.Goto = model.GotoBangumi
	i.Param = strconv.Itoa(int(b.SeasonID))
	i.URI = model.FillURI(bangumi, i.Param, nil)
	i.Finish = int8(b.IsFinish)
	i.Started = int8(b.IsStarted)
	i.Index = b.NewestEpIndex
	i.NewestCat = b.NewestCat
	i.NewestSeason = b.NewestSeason
	i.TotalCount = b.TotalCount
	var buf bytes.Buffer
	if b.CatList.TV != 0 {
		buf.WriteString(`TV(`)
		buf.WriteString(strconv.Itoa(b.CatList.TV))
		buf.WriteString(`) `)
	}
	if b.CatList.Movie != 0 {
		buf.WriteString(`剧场版(`)
		buf.WriteString(strconv.Itoa(b.CatList.Movie))
		buf.WriteString(`) `)
	}
	if b.CatList.Ova != 0 {
		buf.WriteString(`OVA/OAD/SP(`)
		buf.WriteString(strconv.Itoa(b.CatList.Ova))
		buf.WriteString(`)`)
	}
	i.CatDesc = buf.String()
}

// FromUpUserVip is the wrapper of FromUser, dedicated for Phone For 5.43
func (i *Item) FromUpUserVip(u *User, apm map[int64]*api.ArcPlayer, lv *livexroomgate.EntryRoomInfoResp_EntryList, userInfo *account.Card, isBlue, isNewDuration bool) {
	i.FromUpUser(u, userInfo, apm, lv, isBlue, isNewDuration, nil)
	if userInfo != nil {
		i.Vip = &userInfo.Vip
	}
}

func fakeBuilderContext(ctx context.Context, follow map[int64]bool) cardschema.FeedContext {
	attentionStore := make(map[int64]int8, len(follow))
	for fid, followed := range follow {
		if followed {
			attentionStore[fid] = 1
		}
	}
	authn, _ := auth.FromContext(ctx)
	userSession := feedcard.NewUserSession(authn.Mid, attentionStore, &feedcard.IndexParam{})
	dev, _ := device.FromContext(ctx)
	fCtx := feedcard.NewFeedContext(userSession, feedcard.NewCtxDevice(&dev), time.Now())
	return fCtx
}

func isBuild623(ctx cardschema.FeedContext) bool {
	device := ctx.Device()
	if (device.IsAndroid() && device.Build() < 6240000) ||
		(device.IsIOS() && device.Build() < 62400000) {
		return true
	}
	return false
}

func removeAvatar(ctx cardschema.FeedContext) func(in *jsoncard.LargeCoverInline) {
	return func(in *jsoncard.LargeCoverInline) {
		if isBuild623(ctx) {
			return
		}
		// 仅针对非 623 版本去除头像
		in.Avatar = nil
	}
}

func constructOGVInline(ctx context.Context, inlineEP *pgcinline.EpisodeCard, hasLike map[int64]thumbupgrpc.State, follow map[int64]bool, searchMeta *Media) (string, *jsoncard.LargeCoverInline, error) {
	builderCtx := fakeBuilderContext(ctx, follow)
	if inlineEP == nil {
		return "", nil, errors.Errorf("Empty `inlineEP`")
	}

	inlineConfig := &largecover.Inline{
		LikeButtonShowCount:      true,
		LikeResource:             "https://i0.hdslb.com/bfs/archive/b9f49c9b33532c5d05f5ea701ecd063f81910e94.json",
		LikeResourceHash:         "c8b42c2a76890e703b15874175268b4b",
		DisLikeResource:          "https://i0.hdslb.com/bfs/archive/8aee6952487d118b4207c1afa2fd38616bd7545a.json",
		DisLikeResourceHash:      "bdbc35ebc88d178d1f409145dadec806",
		LikeNightResource:        "https://i0.hdslb.com/bfs/archive/3ed718f59e9e9cf1ce148105c9db9559951d5a7d.json",
		LikeNightResourceHash:    "bc9fecf2624a569c05cef8097e20eb37",
		DisLikeNightResource:     "https://i0.hdslb.com/bfs/archive/c9a20055b712068bfe293878639dc9066ba2690b.json",
		DisLikeNightResourceHash: "c370e8d031381f4716d7564956a8b182",
		IconDrag:                 "https://i0.hdslb.com/bfs/archive/c1461e2c6ca97783ac0298b6ebb2d85d94b8f37c.json",
		IconDragHash:             "31df8ce99de871afaa66a7a78f44deec",
		IconStop:                 "https://i0.hdslb.com/bfs/archive/6ee2f9b016f20714705cb5b8f15da1446587d172.json",
		IconStopHash:             "5648c2926c1c93eb2d30748994ba7b96",
		ThreePointPanelType:      1,
	}

	fakeRcmd := &ai.Item{}
	base, err := jsonbuilder.NewBaseBuilder(builderCtx).
		SetParam(strconv.FormatInt(int64(inlineEP.EpisodeId), 10)).
		SetCardType(appcardmodel.LargeCoverSingleV7).
		SetCardGoto(appcardmodel.CardGt(appcardmodel.CardGotoInlinePGC)).
		SetGoto(appcardmodel.GotoBangumi).
		SetMetricRcmd(fakeRcmd).
		Build()
	if err != nil {
		return "", nil, err
	}

	factory := largecover.NewLargeCoverInlineBuilder(builderCtx)
	card, err := factory.DeriveSingleBangumiBuilder().
		SetBase(base).
		SetRcmd(fakeRcmd).
		SetEpisode(inlineEP).
		SetHasLike(castHasLike(hasLike)).
		SetInline(inlineConfig).
		WithAfter(func(in *jsoncard.LargeCoverInline) {
			in.Title = searchMeta.Title // 关键词变红
			if searchMeta.ExtraInfo.Title != "" {
				in.Title = searchMeta.ExtraInfo.Title
			}
			if searchMeta.ExtraInfo.ImgURL != "" {
				in.Cover = searchMeta.ExtraInfo.ImgURL
			}
			extraURI, ok := searchMeta.ExtraInfo.GotoURI()
			if ok {
				in.ExtraURI = extraURI
			}
		}).
		WithAfter(func(in *jsoncard.LargeCoverInline) {
			in.ThreePointMeta.ShareOrigin = "search_inline"
			in.ThreePointMeta.ShareId = "search.search-result.ogv.0"
			in.ThreePointMeta.FunctionalButtons = removeDislike(in.ThreePointMeta.FunctionalButtons)
			in.SharePlane.ShareFrom = "ogv_search_inline_normal_share"
		}).
		Build()
	if err != nil {
		return "", nil, err
	}
	return "ogv_inline", card, nil
}

// 直接按天马卡的模型来输出直播 inline
func constructInlineLive(ctx context.Context, liveRoom *livexroomgate.EntryRoomInfoResp_EntryList, userInfo *account.Card, follow map[int64]bool) (string, *jsoncard.LargeCoverInline, error) {
	builderCtx := fakeBuilderContext(ctx, follow)
	if liveRoom == nil {
		return "", nil, errors.Errorf("Empty `liveRoom`")
	}

	fakeRcmd := &ai.Item{}
	// fake base
	base, err := jsonbuilder.NewBaseBuilder(builderCtx).
		SetParam(strconv.FormatInt(liveRoom.RoomId, 10)).
		SetCardType(appcardmodel.LargeCoverV8).
		SetCardGoto(appcardmodel.CardGt(appcardmodel.CardGotoInlineLive)).
		SetGoto(appcardmodel.GotoLive).
		SetMetricRcmd(fakeRcmd).
		Build()
	if err != nil {
		return "", nil, err
	}

	builder := largecover.NewLargeCoverInlineBuilder(builderCtx).
		DeriveLiveEntryRoomBuilder().
		SetBase(base).
		SetLiveRoom(liveRoom).
		SetInline(&largecover.Inline{}).
		SetAuthorCard(userInfo).
		SetEntryFrom(model.SearchInlineCard).
		SetRcmd(fakeRcmd). // fake rcmd item
		WithAfter(removeAvatar(builderCtx))
	card, err := builder.Build()
	if err != nil {
		return "", nil, err
	}
	return "live_room", card, nil
}

// FromUpUser form func
func (i *Item) FromUpUser(u *User, userInfo *account.Card, apm map[int64]*api.ArcPlayer, lv *livexroomgate.EntryRoomInfoResp_EntryList, isBlue, isNewDuration bool,
	notices map[int64]*managermodel.SystemNotice) {
	i.Title = u.Name
	i.Cover = u.Pic
	i.Goto = model.GotoAuthor
	i.OfficialVerify = u.OfficialVerify
	i.Param = strconv.Itoa(int(u.Mid))
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	i.Mid = u.Mid
	i.Sign = u.Usign
	i.Fans = u.Fans
	i.Level = u.Level
	i.Arcs = u.Videos
	i.AvItems = make([]*Item, 0, len(u.Res))
	for pos, v := range u.Res {
		vi := &Item{}
		vi.Title = v.Title
		vi.Cover = v.Pic
		vi.Goto = model.GotoAv
		vi.Param = strconv.Itoa(int(v.Aid))
		ap, ok := apm[v.Aid]
		if ok && ap.Arc != nil {
			a := ap.Arc
			playInfo := ap.PlayerInfo[ap.DefaultPlayerCid]
			vi.URI = model.FillURI(vi.Goto, vi.Param, model.AvPlayHandlerGRPC(a, playInfo))
			vi.Play = int(a.Stat.View)
			vi.Danmaku = int(a.Stat.Danmaku)
			if a.Rights.UGCPay == 1 {
				vi.Badges = append(vi.Badges, model.PayBadge)
			}
			if a.Rights.IsCooperation == 1 {
				vi.Badges = append(vi.Badges, model.CooperationBadge)
			}
			if isNewDuration {
				vi.Duration = model.DurationString(a.Duration)
			}
		} else {
			switch play := v.Play.(type) {
			case float64:
				vi.Play = int(play)
			case string:
				vi.Play, _ = strconv.Atoi(play)
			}
			vi.URI = model.FillURI(vi.Goto, vi.Param, nil)
			vi.Danmaku = v.Danmaku
		}
		vi.IsPay = v.IsPay
		vi.CTime = v.Pubdate
		if v.Pubdate != 0 {
			vi.CTimeLabel = fmt.Sprintf("%s投递", cardmdl.PubDataString(time.Unix(v.Pubdate, 0)))
		}
		if !isNewDuration {
			vi.Duration = v.Duration
		}
		vi.Position = pos + 1
		i.AvItems = append(i.AvItems, vi)
	}
	if !isBlue {
		i.LiveStatus = u.IsLive
		i.RoomID = u.RoomID
	}
	i.IsUp = u.IsUpuser == 1
	if i.RoomID != 0 && !isBlue {
		i.LiveURI = model.FillURI(model.GotoLive, strconv.Itoa(int(u.RoomID)), model.LiveEntryHandler(lv, ""))
		i.LiveLink = model.FillURI(model.GotoLive, strconv.Itoa(int(u.RoomID)), model.LiveEntryHandler(lv, model.DefaultLiveEntry))
	}
	notice, ok := notices[u.Mid]
	if ok {
		i.Notice = constructNotice(notice)
	}
	if i.Position == 0 {
		i.Position = u.Position
	}
	if userInfo != nil {
		i.FaceNftNew = userInfo.FaceNftNew
		i.IsSeniorMember = userInfo.IsSeniorMember
	}
}

func (i *Item) FromUpUserNewIPadHD(u *User, userInfo *account.Card, apm map[int64]*api.ArcPlayer, lv *livexroomgate.EntryRoomInfoResp_EntryList, isBlue, isNewDuration bool, searchConf *conf.Search,
	userProfile *account.ProfileWithoutPrivacy, extraFunc ...func(*Item)) {
	i.FromUpUserNew(u, userInfo, apm, lv, isBlue, isNewDuration, searchConf, nil, nil, userProfile, extraFunc...)
	i.Space = &SpaceEntrance{}
	i.AvStyle = _searchUpAvStyleNone
	if userProfile != nil && userProfile.IsLatest_100Honour == 1 {
		// 兼容 ipad HD 用的百大背景图
		i.Background.BgPicURL = "https://i0.hdslb.com/bfs/archive/18f630db0fd2e659cfa25f4c4e7ad9b3e34b0229.png"
		i.Background.FgPicURL = "https://i0.hdslb.com/bfs/archive/f34c0ee18c6f0aa112cb5f862310eac3280f2f1d.png"
	}
	//nolint:gomnd
	if len(i.AvItems) >= 5 {
		i.Space = &SpaceEntrance{
			Show:           _searchUpSpaceShow,
			Test:           "查看全部稿件 >",
			TextColor:      searchConf.SpaceEntrance.TextColor,
			TextColorNight: searchConf.SpaceEntrance.TextColorNight,
			SpaceURL:       i.URI,
		}
		i.AvStyle = _searchUpAvStyleMore
	}
}

// 用于用户卡的直播 inline 卡
func OptInlineLiveFn(ctx context.Context, lv *livexroomgate.EntryRoomInfoResp_EntryList, userInfo *account.Card, follow map[int64]bool) func(i *Item) {
	return func(i *Item) {
		inlineType, inlineLive, err := constructInlineLive(ctx, lv, userInfo, follow)
		if err != nil {
			log.Error("Failed to construct inline live: %+v", err)
			return
		}
		i.InlineType = inlineType
		i.InlineLive = newSearchEmbedInline(inlineLive)
		// 有直播 inline 时 av_items 相关字段都设置为空
		i.AvItems = nil
		i.Space = nil
		i.AvStyle = _searchUpAvStyleNone
	}
}

// 按单列直播卡的模型
func constructLargeCoverSingleV8(ctx context.Context, liveRoom *livexroomgate.EntryRoomInfoResp_EntryList, userInfo *account.Card, follow map[int64]bool, searchMeta *Live, entryFrom string, nftRegion map[int64]*gallerygrpc.NFTRegion) (string, *jsoncard.LargeCoverInline, error) {
	builderCtx := fakeBuilderContext(ctx, follow)
	if liveRoom == nil {
		return "", nil, errors.Errorf("Empty `liveRoom`")
	}

	if entryFrom == "" {
		entryFrom = model.SearchLiveInlineCard
	}
	inlineConfig := &largecover.Inline{
		LikeButtonShowCount:      true,
		LikeResource:             "https://i0.hdslb.com/bfs/archive/b9f49c9b33532c5d05f5ea701ecd063f81910e94.json",
		LikeResourceHash:         "c8b42c2a76890e703b15874175268b4b",
		DisLikeResource:          "https://i0.hdslb.com/bfs/archive/8aee6952487d118b4207c1afa2fd38616bd7545a.json",
		DisLikeResourceHash:      "bdbc35ebc88d178d1f409145dadec806",
		LikeNightResource:        "https://i0.hdslb.com/bfs/archive/3ed718f59e9e9cf1ce148105c9db9559951d5a7d.json",
		LikeNightResourceHash:    "bc9fecf2624a569c05cef8097e20eb37",
		DisLikeNightResource:     "https://i0.hdslb.com/bfs/archive/c9a20055b712068bfe293878639dc9066ba2690b.json",
		DisLikeNightResourceHash: "c370e8d031381f4716d7564956a8b182",
		IconDrag:                 "https://i0.hdslb.com/bfs/archive/c1461e2c6ca97783ac0298b6ebb2d85d94b8f37c.json",
		IconDragHash:             "31df8ce99de871afaa66a7a78f44deec",
		IconStop:                 "https://i0.hdslb.com/bfs/archive/6ee2f9b016f20714705cb5b8f15da1446587d172.json",
		IconStopHash:             "5648c2926c1c93eb2d30748994ba7b96",
		ThreePointPanelType:      1,
	}

	fakeRcmd := &ai.Item{}
	// fake base
	base, err := jsonbuilder.NewBaseBuilder(builderCtx).
		SetParam(strconv.FormatInt(liveRoom.RoomId, 10)).
		SetCardType(appcardmodel.LargeCoverSingleV8).
		SetCardGoto(appcardmodel.CardGt(appcardmodel.GotoLive)).
		SetGoto(appcardmodel.GotoLive). // 没啥意义，客户端需要区分罢了
		SetMetricRcmd(fakeRcmd).
		Build()
	if err != nil {
		return "", nil, err
	}

	factory := largecover.NewLargeCoverInlineBuilder(builderCtx)
	card, err := factory.DeriveLiveEntryRoomBuilder().
		SetBase(base).
		SetRcmd(fakeRcmd).
		SetLiveRoom(liveRoom).
		SetAuthorCard(userInfo).
		SetInline(inlineConfig).
		SetEntryFrom(entryFrom).
		WithAfter(largecover.SingleInlineLiveHideMeta()).
		WithAfter(largecover.SingleV8InlineDesc(userInfo)).
		WithAfter(func(in *jsoncard.LargeCoverInline) {
			if searchMeta == nil {
				return
			}
			in.Title = searchMeta.Title // 标题变红
			if searchMeta.ExtraInfo.Title != "" {
				in.Title = searchMeta.ExtraInfo.Title
			}
			if searchMeta.ExtraInfo.ImgURL != "" {
				in.Cover = searchMeta.ExtraInfo.ImgURL
			}
			extraURI, ok := searchMeta.ExtraInfo.GotoURI()
			if ok {
				in.ExtraURI = extraURI
			}
		}).
		WithAfter(func(in *jsoncard.LargeCoverInline) {
			in.ThreePointMeta.ShareOrigin = "search_inline"
			in.ThreePointMeta.ShareId = "search.search-result.live.0"
			in.ThreePointMeta.FunctionalButtons = removeDislike(in.ThreePointMeta.FunctionalButtons)
		}).
		WithAfter(setInNftFaceIcon(nftRegion)).
		Build()
	if err != nil {
		return "", nil, err
	}
	return "live_room_inline", card, nil
}

func setInNftFaceIcon(nftRegion map[int64]*gallerygrpc.NFTRegion) func(*jsoncard.LargeCoverInline) {
	return func(in *jsoncard.LargeCoverInline) {
		if nftRegion == nil || in.Avatar == nil || in.Avatar.FaceNftNew != 1 {
			return
		}
		if v, ok := nftRegion[in.UpArgs.UpID]; ok {
			in.Avatar.NftFaceIcon = &card.NftFaceIcon{
				RegionType: int32(v.Type),
				Icon:       v.Icon,
				ShowStatus: int32(v.ShowStatus),
			}
		}
	}
}

func removeDislike(in []*threePointMeta.FunctionalButton) []*threePointMeta.FunctionalButton {
	const (
		_typeNotInterested = 1
	)
	out := make([]*threePointMeta.FunctionalButton, 0, len(in))
	for _, v := range in {
		if v.Type == _typeNotInterested {
			continue
		}
		out = append(out, v)
	}
	return out
}

func castHasLike(in map[int64]thumbupgrpc.State) map[int64]int8 {
	out := make(map[int64]int8)
	for k, v := range in {
		out[k] = int8(v)
	}
	return out
}

// FromUpUserNew
func (i *Item) FromUpUserNew(u *User, userInfo *account.Card, apm map[int64]*api.ArcPlayer, lv *livexroomgate.EntryRoomInfoResp_EntryList, isBlue, isNewDuration bool,
	searchConf *conf.Search, inlineLiveFn func(*Item), notice *managermodel.SystemNotice, userProfile *account.ProfileWithoutPrivacy, extraFunc ...func(*Item)) {
	i.Title = u.Name
	i.Cover = u.Pic
	i.Goto = model.GotoAuthorNew
	i.OfficialVerify = u.OfficialVerify
	i.Param = strconv.Itoa(int(u.Mid))
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	i.Mid = u.Mid
	i.Sign = u.Usign
	i.Fans = u.Fans
	i.Level = u.Level
	i.Arcs = u.Videos
	if userInfo != nil {
		i.FaceNftNew = userInfo.FaceNftNew
		i.IsSeniorMember = userInfo.IsSeniorMember
		i.Vip = &userInfo.Vip
	}
	i.AvItems = make([]*Item, 0, len(u.Res))
	i.Background = &Background{}
	if searchConf.BackgroundSwitch && userProfile != nil && userProfile.IsLatest_100Honour == 1 {
		i.Background.Show = _searchUpBgShow
		i.Background.BgPicURL = _100HounourAppBgPicURL
		i.Background.FgPicURL = _100HounourAppFgPicURL
	}
	if searchConf.LiveFaceSwitch && u.IsLive == 1 {
		i.LiveFace = _searchUpLiveFaceShow
	}
	for pos, v := range u.Res {
		vi := &Item{}
		vi.Title = v.Title
		vi.Cover = v.Pic
		vi.Goto = model.GotoAv
		vi.Param = strconv.FormatInt(v.Aid, 10)
		if ap, ok := apm[v.Aid]; ok && ap.Arc != nil {
			a := ap.Arc
			vi.Play = int(a.Stat.View)
			vi.Danmaku = int(a.Stat.Danmaku)
			playInfo := ap.PlayerInfo[ap.DefaultPlayerCid]
			vi.URI = model.FillURI(vi.Goto, vi.Param, model.AvPlayHandlerGRPC(a, playInfo))
			if isNewDuration {
				vi.Duration = model.DurationString(a.Duration)
			}
		} else {
			switch play := v.Play.(type) {
			case float64:
				vi.Play = int(play)
			case string:
				vi.Play, _ = strconv.Atoi(play)
			}
			vi.URI = model.FillURI(vi.Goto, vi.Param, nil)
			vi.Danmaku = v.Danmaku
		}
		// vi.IsPay = v.IsPay
		vi.CTime = v.Pubdate
		if v.Pubdate != 0 {
			vi.CTimeLabel = fmt.Sprintf("%s投递", cardmdl.PubDataString(time.Unix(v.Pubdate, 0)))
		}
		if !isNewDuration {
			vi.Duration = v.Duration
		}
		vi.Position = pos + 1
		i.AvItems = append(i.AvItems, vi)
	}
	switch len(i.AvItems) {
	case 0: // 无视频
		i.Space = &SpaceEntrance{}
		i.AvStyle = _searchUpAvStyleNone
	case 1: // 一个视频
		i.Space = &SpaceEntrance{
			Show:           _searchUpSpaceShow,
			Test:           searchConf.SpaceEntrance.TextMore,
			TextColor:      searchConf.SpaceEntrance.TextColor,
			TextColorNight: searchConf.SpaceEntrance.TextColorNight,
			SpaceURL:       i.URI,
		}
		i.AvStyle = _searchUpAvStyleOne
	default: // 多个视频
		i.Space = &SpaceEntrance{
			Show:           _searchUpSpaceShow,
			Test:           fmt.Sprintf(searchConf.SpaceEntrance.TextMoreWithNum, u.Videos),
			TextColor:      searchConf.SpaceEntrance.TextColor,
			TextColorNight: searchConf.SpaceEntrance.TextColorNight,
			SpaceURL:       i.URI,
		}
		i.AvStyle = _searchUpAvStyleMore
	}
	if !isBlue {
		i.LiveStatus = u.IsLive
		i.RoomID = u.RoomID
	}
	i.IsUp = u.IsUpuser == 1
	if i.RoomID != 0 && !isBlue {
		i.LiveURI = model.FillURI(model.GotoLive, strconv.Itoa(int(u.RoomID)), model.LiveEntryHandler(lv, ""))
		i.LiveLink = model.FillURI(model.GotoLive, strconv.Itoa(int(u.RoomID)), model.LiveEntryHandler(lv, model.DefaultLiveEntry))
		i.IsInlineLive = u.IsInlineLive
		if u.IsInlineLive == 1 && inlineLiveFn != nil {
			inlineLiveFn(i)
		}
	}
	if notice != nil {
		i.Notice = constructNotice(notice)
	}
	if i.Position == 0 {
		i.Position = u.Position
	}
	for _, extFunc := range extraFunc {
		extFunc(i)
	}
}

func constructNotice(in *managermodel.SystemNotice) *Notice {
	const (
		// 原样式icon
		_prInfoOldIcon = "https://i0.hdslb.com/bfs/space/7a89f7ed04b98458b23863846bd2539a90ff1153.png"
		// 原样式夜间icon
		_prInfoOldIconNight = "https://i0.hdslb.com/bfs/space/cab669b46fc1bce8b8b2fbd0ce19909f9f2299a4.png"
		// 缅怀提示日间icon
		_prInfoNewIcon = "https://i0.hdslb.com/bfs/space/ca6d0ed2edae23cf348db19cd2c293f2121c1b59.png"
		// 缅怀提示夜间icon
		_prInfoNewIconNight = "https://i0.hdslb.com/bfs/space/e2a4c7bb9297e74d1be7467f96086bf33931f9d0.png"
		// 缅怀样式背景色
		_prInfoNewBgColor = "#F1F2F3"
		// 缅怀样式文字色
		_prInfoNewTextcolor = "#9499A0"
		// 缅怀样式夜间背景色
		_prInfoNewBgColorNight = "#000000"
		// 缅怀样式夜间文字色
		_prInfoNewTextcolorNight = "#757A81"
		// 原样式背景色
		_prInfoOldBgColor = "#FFF6E4"
		// 原样式文字色
		_prInfoOldTextcolor = "#FFB027"
		// 原样式夜间背景色
		_prInfoOLdBgColorNight = "#342410"
		// 原样式夜间文字色
		_prInfoOldTextcolorNight = "#DB8700"
	)

	out := &Notice{
		Mid:        in.Mid,
		NoticeID:   in.NoticeID,
		Content:    in.Content,
		URL:        in.URL,
		NoticeType: in.NoticeType,
		Icon:       in.Icon,
		TextColor:  in.TextColor,
		BGColor:    in.BGColor,
	}
	//nolint:gomnd
	if in.NoticeType == 1 {
		out.Icon = _prInfoOldIcon
		out.IconNight = _prInfoOldIconNight
		out.BGColor = _prInfoOldBgColor
		out.BGColorNight = _prInfoOLdBgColorNight
		out.TextColor = _prInfoOldTextcolor
		out.TextColorNight = _prInfoOldTextcolorNight
	} else if in.NoticeType == 2 {
		out.Icon = _prInfoNewIcon
		out.IconNight = _prInfoNewIconNight
		out.BGColor = _prInfoNewBgColor
		out.BGColorNight = _prInfoNewBgColorNight
		out.TextColor = _prInfoNewTextcolor
		out.TextColorNight = _prInfoNewTextcolorNight
	}
	return out
}

// FromUserVip is the wrapper of FromUser, dedicated for Phone For 5.43
func (i *Item) FromUserVip(u *User, apm map[int64]*api.ArcPlayer, lv *livexroomgate.EntryRoomInfoResp_EntryList, userInfo *account.Card, isBlue bool) {
	i.FromUser(u, userInfo, apm, lv, isBlue)
	if userInfo != nil {
		i.Vip = &userInfo.Vip
	}
}

// FromUser form func
func (i *Item) FromUser(u *User, userInfo *account.Card, apm map[int64]*api.ArcPlayer, lv *livexroomgate.EntryRoomInfoResp_EntryList, isBlue bool) {
	i.Title = u.Name
	i.Cover = u.Pic
	i.Goto = model.GotoAuthor
	i.OfficialVerify = u.OfficialVerify
	i.Param = strconv.Itoa(int(u.Mid))
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	i.Mid = u.Mid
	i.Sign = u.Usign
	i.Fans = u.Fans
	i.Level = u.Level
	if userInfo != nil {
		i.FaceNftNew = userInfo.FaceNftNew
		i.IsSeniorMember = userInfo.IsSeniorMember
	}
	i.Arcs = u.Videos
	i.AvItems = make([]*Item, 0, len(u.Res))
	if !isBlue {
		i.LiveStatus = u.IsLive
		i.RoomID = u.RoomID
		if i.RoomID != 0 {
			i.LiveURI = model.FillURI(model.GotoLive, strconv.Itoa(int(u.RoomID)), model.LiveEntryHandler(lv, ""))
			i.LiveLink = model.FillURI(model.GotoLive, strconv.Itoa(int(u.RoomID)), model.LiveEntryHandler(lv, model.DefaultLiveEntry))
		}
	}
	if u.IsUpuser == 1 {
		for pos, v := range u.Res {
			vi := &Item{}
			vi.Title = v.Title
			vi.Cover = v.Pic
			vi.Goto = model.GotoAv
			vi.Param = strconv.Itoa(int(v.Aid))
			ap, ok := apm[v.Aid]
			if ok && ap.Arc != nil {
				a := ap.Arc
				playInfo := ap.PlayerInfo[ap.DefaultPlayerCid]
				vi.URI = model.FillURI(vi.Goto, vi.Param, model.AvPlayHandlerGRPC(a, playInfo))
				vi.Play = int(a.Stat.View)
				vi.Danmaku = int(a.Stat.Danmaku)
				vi.ShowCardDesc2 = "· " + cardmdl.PubDataString(ap.Arc.PubDate.Time())
				if a.Rights.UGCPay == 1 {
					vi.Badges = append(vi.Badges, model.PayBadge)
				}
				if a.Rights.IsCooperation == 1 {
					vi.Badges = append(vi.Badges, model.CooperationBadge)
				}
			} else {
				switch play := v.Play.(type) {
				case float64:
					vi.Play = int(play)
				case string:
					vi.Play, _ = strconv.Atoi(play)
				}
				vi.Danmaku = v.Danmaku
				vi.URI = model.FillURI(vi.Goto, vi.Param, nil)
			}
			vi.IsPay = v.IsPay
			vi.CTime = v.Pubdate
			if v.Pubdate != 0 {
				vi.CTimeLabel = fmt.Sprintf("%s投递", cardmdl.PubDataString(time.Unix(v.Pubdate, 0)))
			}
			vi.Duration = v.Duration
			vi.Position = pos + 1
			vi.ShowCardDesc2 = "· " + cardmdl.PubDataString(time.Unix(v.Pubdate, 0))
			i.AvItems = append(i.AvItems, vi)
		}
		i.IsUp = true
	}
}

// FromMovie form func
func (i *Item) FromMovie(m *Movie, apm map[int64]*api.ArcPlayer) {
	i.Title = m.Title
	i.Desc = m.Desc
	if m.Type == "movie" {
		i.Cover = m.Cover
		i.Param = strconv.Itoa(int(m.Aid))
		i.Goto = model.GotoAv
		ap, ok := apm[m.Aid]
		if ok && ap.Arc != nil {
			playInfo := ap.PlayerInfo[ap.DefaultPlayerCid]
			i.URI = model.FillURI(i.Goto, i.Param, model.AvPlayHandlerGRPC(ap.Arc, playInfo))
		} else {
			i.URI = model.FillURI(i.Goto, i.Param, nil)
		}
		i.CoverMark = model.StatusMark(m.Status)
	} else if m.Type == "special" {
		i.Param = m.SpID
		i.Goto = model.GotoSp
		i.URI = model.FillURI(i.Goto, i.Param, nil)
		i.Cover = m.Pic
	}
	i.Staff = m.Staff
	i.Actors = m.Actors
	i.Area = m.Area
	i.Length = m.Length
	i.Status = m.Status
	i.ScreenDate = m.ScreenDate
}

func constructFullTextResult(uri string, fulltext []*FullText) *FullTextResult {
	firstFullText := fulltext[0] // 取ai首位的全文检索元素展示
	res := &FullTextResult{
		Type:              firstFullText.Type,
		JumpStartProgress: firstFullText.StartSecond,
		JumpUri:           makeFullTextJumpUri(firstFullText.Type, uri, firstFullText.Abstract, firstFullText.StartSecond),
	}
	switch firstFullText.Type {
	case _chapterFullTextType:
		res.ShowText = fmt.Sprintf("章节 · %s", firstFullText.Text)
	case _digestFullTextType:
		res.ShowText = firstFullText.Text
	default:
		log.Warn("Unexpected full text type=%+v", firstFullText.Type)
	}
	return res
}

func makeFullTextJumpUri(jumpType int, uri, abstract string, jumpSecond int64) string {
	if abstract == "" {
		abstract = "相关片段"
	}
	return fmt.Sprintf("%s&fulltext_jump_type=%d&jump_toast_text=%s&jump_start_progress=%d", uri, jumpType, abstract, jumpSecond)
}

//nolint:unparam
func constructShowCardDesc(order string, pTime xtime.Time) (string, string) {
	const (
		_orderDanmaku = "dm"
	)
	switch order {
	case _orderDanmaku:
		return "", ""
	default:
		return "", fmt.Sprintf("· %s", cardmdl.PubDataString(pTime.Time()))
	}
}

// FromArticle form func
func (i *Item) FromArticle(a *Article, acc *account.ProfileWithoutPrivacy) {
	i.ID = a.ID
	i.Mid = a.Mid
	if acc != nil {
		i.Author = acc.Name
	}
	i.TemplateID = a.TemplateID
	i.Title = a.Title
	i.Desc = a.Desc
	i.ImageUrls = a.ImageUrls
	i.View = a.View
	i.Play = a.View
	i.Like = a.Like
	i.Reply = a.Reply
	i.Badge = "专栏"
	i.Goto = model.GotoArticle
	i.Param = strconv.Itoa(int(a.ID))
	i.URI = model.FillURI(i.Goto, i.Param, nil)
}

// FromOperate form func
func (i *Item) FromOperate(o *Operate, gt string, isNewColor bool) {
	i.Title = o.Title
	i.Cover = o.Cover
	i.URI = o.RedirectURL
	i.Param = strconv.FormatInt(o.ID, 10)
	i.Desc = o.Desc
	i.Badge = o.Corner
	i.Goto = gt
	if o.RecReason != "" {
		i.RcmdReason = &RcmdReason{Content: o.RecReason}
		vs := &model.ReasonStyle{}
		if isNewColor {
			*vs = *videoStrongStyleV2
		} else {
			*vs = *videoStrongStyle
		}
		vs.Text = o.RecReason
		i.NewRecTags = append(i.NewRecTags, vs)
	}
}

func toHTTPS(in string) string {
	return strings.Replace(in, "http://", "https://", 1)
}

func (i *Item) FromVideoSpecial(in *Video) error {
	i.Title = in.Title
	i.Cover = toHTTPS(in.Cover)
	i.URI = in.URL
	i.Param = strconv.FormatInt(in.ID, 10)
	i.Desc = in.Desc
	i.Badge = in.Corner
	i.Goto = "special_s"
	if in.RecReason != "" {
		i.RcmdReason = &RcmdReason{Content: in.RecReason}
		vs := &model.ReasonStyle{}
		*vs = *videoStrongStyleV2
		vs.Text = in.RecReason
		i.NewRecTags = append(i.NewRecTags, vs)
	}
	return nil
}

func buildPgcCardLabel(m *Media, gt string, season *pgcsearch.SearchCardProto) string {
	var hit string
	for _, v := range m.HitColumns {
		if v == "cv" {
			hit = v
			break
		} else if v == "staff" {
			hit = v
		}
	}
	label := ""
	if hit == "cv" {
		for _, v := range getHightLight.FindAllStringSubmatch(m.CV, -1) {
			//nolint:gomnd
			if m.MediaType == 7 {
				label = fmt.Sprintf("嘉宾: %v...", v[0])
				break
			}
			if gt == model.GotoBangumi {
				label = fmt.Sprintf("声优: %v...", v[0])
				break
			} else if gt == model.GotoMovie {
				label = fmt.Sprintf("演员: %v...", v[0])
				break
			}
		}
	} else if hit == "staff" {
		for _, v := range getHightLight.FindAllStringSubmatch(m.Staff, -1) {
			label = fmt.Sprintf("制作人员: %v...", v[0])
			break
		}
	} else if hit == "" {
		label = FormPGCLabel(m.MediaType, season.Style, m.Staff, m.CV)
	}
	return label
}

func asPGCJSONBadges(in []*pgcsearch.SearchBadgeProto) []*model.ReasonStyle {
	out := make([]*model.ReasonStyle, 0, len(in))
	for _, v := range in {
		rs := &model.ReasonStyle{
			Text:             v.Text,
			TextColor:        v.TextColor,
			TextColorNight:   v.TextColorNight,
			BgColor:          v.BgColor,
			BgColorNight:     v.BgColorNight,
			BorderColor:      v.BorderColor,
			BorderColorNight: v.BorderColorNight,
			BgStyle:          int8(v.BgStyle),
		}
		out = append(out, rs)
	}
	return out
}

func asPGCJSONEpisodes(in []*pgcsearch.SearchEpProto) []*Item {
	out := make([]*Item, 0, len(in))
	for pos, v := range in {
		tmp := &Item{
			Param:    strconv.Itoa(int(v.Id)),
			Index:    v.IndexTitle,
			Position: pos + 1,
			URI:      v.Url,
		}
		out = append(out, tmp)
	}
	return out
}

func asPGCJSONEpisodesNew(season *pgcsearch.SearchCardProto, cfg *conf.PgcSearchCard, isIpadDirect bool) []*EpisodeNew {
	out := make([]*EpisodeNew, 0, len(season.Eps))
	isHorizon := season.SelectionStyle == _styleHorizontal
	var pos int
	for _, epGrpc := range season.Eps {
		if isHorizon && ((isIpadDirect && len(out) >= cfg.IpadEpSize) || (!isIpadDirect && len(out) >= cfg.Epsize)) { // ipad垂搜横条最多3条，ipad综合搜索和手机最多2条
			break
		}
		epNew := new(EpisodeNew)
		if canAppend := epNew.FromPgcRes(epGrpc, isHorizon, cfg.GridBadge); canAppend {
			if epNew.Type == 0 { // 0正常ep 1更多链接
				pos++
				epNew.Position = pos
			}
			out = append(out, epNew)
		}
	}
	return out
}

func adjustPGCEpisodesNewAndCheckMore(m *Media, dst *Item, season *pgcsearch.SearchCardProto, cfg *conf.PgcSearchCard, isIpadDirect bool) {
	isHorizon := season.SelectionStyle == _styleHorizontal
	if m.HitEpids == "" && isHorizon && ((isIpadDirect && len(season.Eps) > cfg.IpadEpSize) || (!isIpadDirect && len(season.Eps) > cfg.Epsize)) { // 未召回单集 && 横条 && 长度>2(phone), >3(ipad) 展示 "查看全部.."
		if isIpadDirect && len(dst.EpisodesNew) > cfg.IpadCheckMoreSize { // ipad垂搜超过3条时候压缩为2条+查看更多
			dst.EpisodesNew = dst.EpisodesNew[0:cfg.IpadCheckMoreSize]
		}
		dst.CheckMore = &CheckMore{
			Content: fmt.Sprintf(cfg.CheckMoreContent, season.EpSize),
			Uri:     fmt.Sprintf(cfg.CheckMoreSchema, _styleHorizontal, season.SeasonId), // must be horizontal
		}
	}
}

func RoundHalfUp(val float64, precision int) float64 {
	p := math.Pow10(precision)
	return math.Floor(val*p+0.5) / p
}

func (i *Item) FromMediaPgcCardPureRPC(m *Media, prompt string, gt string, seasonEps map[int32]*pgcsearch.SearchCardProto, cfg *conf.PgcSearchCard, isIpadDirect bool, extraFunc ...func(*Item)) error { // isIpadDirect ipad垂搜
	season, ok := seasonEps[int32(m.SeasonID)]
	if !ok {
		return errors.Errorf("Failed to get season: %+v", m)
	}

	i.Goto = gt
	i.Prompt = prompt
	i.PlayState = 0 // 固定认为可播

	i.Title = m.Title
	if i.Title == "" {
		i.Title = m.OrgTitle
	}
	i.Param = strconv.Itoa(int(m.MediaID))
	i.MediaType = m.MediaType
	i.CV = m.CV
	i.Staff = m.Staff
	i.SeasonID = m.SeasonID
	if i.Position == 0 {
		i.Position = m.Position
	}

	i.URI = season.Url
	i.Cover = toHTTPS(season.SeasonCover)
	i.Styles = season.Styles
	i.Style = season.Style // 兼容老版本
	if season.Rating != nil {
		i.Score = fmt.Sprintf("%.1f", season.Rating.Score)
		i.Rating = RoundHalfUp(float64(season.Rating.Score), 1)
		i.Vote = int(season.Rating.Count)
	}
	i.PTime = xtime.Time(season.PubTime)
	areas := strings.Split(season.Areas, "、")
	if len(areas) != 0 {
		i.Area = areas[0]
	}
	i.Label = buildPgcCardLabel(m, gt, season)
	i.Badge = model.FormMediaType(int(season.SeasonType))
	i.SeasonType = int(season.SeasonType)
	i.SeasonTypeName = season.SeasonTypeName
	i.IsAttention = int(season.IsFollow)
	i.IsSelection = int(season.IsSelection)
	i.Badges = asPGCJSONBadges(season.Badges)
	i.Episodes = asPGCJSONEpisodes(season.Eps)
	i.SelectionStyle = season.SelectionStyle

	if m.IsAllNet() { // 全网搜，下发搜索的out_url和立即观看
		i.WatchButton = &WatchButton{
			Title: cfg.OnlineWatch,
			Link:  m.AllNetURL,
		}
		i.IsOut = 1
		return nil
	}
	i.WatchButton = &WatchButton{ // pgc下发 立即观看按钮
		Title: season.ButtonText,
		Link:  season.Url,
	}
	if season.Follow != nil { // pgc下发 追番/追剧按钮
		i.FollowButton = new(FollowButton)
		i.FollowButton.FromPGCCard(season.Follow)
		i.IsAttention = int(season.IsFollow)
	} else {
		log.Warn("FollowButton Sid %d Missing Follow", m.SeasonID)
	}
	i.EpisodesNew = asPGCJSONEpisodesNew(season, cfg, isIpadDirect)
	adjustPGCEpisodesNewAndCheckMore(m, i, season, cfg, isIpadDirect)
	i.IsOGVInline = m.IsOGVInline
	for _, extFunc := range extraFunc {
		extFunc(i)
	}
	return nil
}

// FromMediaPgcCard def.
func (i *Item) FromMediaPgcCard(m *Media, prompt string, gt string, bangumis map[string]*bangumimdl.Card, seasonEps map[int32]*pgcsearch.SearchCardProto, medisas map[int32]*pgcsearch.SearchMediaProto, cfg *conf.PgcSearchCard, isIpadDirect bool, extraFunc ...func(*Item)) { // isIpadDirect ipad垂搜
	i.FromMedia(m, prompt, gt, bangumis, medisas)
	i.SelectionStyle = _styleGrid // 默认宫格，当且仅当pgc新接口下发并且为横条才会出横条
	if m.IsAllNet() {             // 全网搜，下发搜索的out_url和立即观看
		i.WatchButton = &WatchButton{
			Title: cfg.OnlineWatch,
			Link:  m.AllNetURL,
		}
		i.IsOut = 1
		for _, extFunc := range extraFunc {
			extFunc(i)
		}
		return
	}
	i.WatchButton = &WatchButton{ // 默认watch_button，当该season不可播时候用到
		Title: cfg.OfflineWatch,
		Link:  m.GotoURL,
	}
	if m.Canplay() {
		if seasonEp, ok := seasonEps[int32(m.SeasonID)]; ok { // 使用pgc下发的按钮和链接
			i.Styles = seasonEp.Styles
			i.PTime = xtime.Time(seasonEp.PubTime)
			isHorizon := seasonEp.SelectionStyle == _styleHorizontal
			i.WatchButton = &WatchButton{ // pgc下发 立即观看按钮
				Title: seasonEp.ButtonText,
				Link:  seasonEp.Url,
			}
			if seasonEp.Follow != nil { // pgc下发 追番/追剧按钮
				i.FollowButton = new(FollowButton)
				i.FollowButton.FromPGCCard(seasonEp.Follow)
				i.IsAttention = int(seasonEp.IsFollow)
			} else {
				log.Warn("FollowButton Sid %d Missing Follow", m.SeasonID)
			}
			i.SelectionStyle = seasonEp.SelectionStyle
			i.IsSelection = int(seasonEp.IsSelection) // when pgc gives is_selection in new grpc, use it to replace the old http's
			if len(seasonEp.Eps) == 0 {               // 无选集信息，不处理选集和查看更多
				for _, extFunc := range extraFunc {
					extFunc(i)
				}
				return
			}
			var pos int
			for _, epGrpc := range seasonEp.Eps {
				if isHorizon && ((isIpadDirect && len(i.EpisodesNew) >= cfg.IpadEpSize) || (!isIpadDirect && len(i.EpisodesNew) >= cfg.Epsize)) { // ipad垂搜横条最多3条，ipad综合搜索和手机最多2条
					break
				}
				epNew := new(EpisodeNew)
				if canAppend := epNew.FromPgcRes(epGrpc, isHorizon, cfg.GridBadge); canAppend {
					if epNew.Type == 0 { // 0正常ep 1更多链接
						pos++
						epNew.Position = pos
					}
					i.EpisodesNew = append(i.EpisodesNew, epNew)
				}
			}
			if m.HitEpids == "" && isHorizon && ((isIpadDirect && len(seasonEp.Eps) > cfg.IpadEpSize) || (!isIpadDirect && len(seasonEp.Eps) > cfg.Epsize)) { // 未召回单集 && 横条 && 长度>2(phone), >3(ipad) 展示 "查看全部.."
				if isIpadDirect && len(i.EpisodesNew) > cfg.IpadCheckMoreSize { // ipad垂搜超过3条时候压缩为2条+查看更多
					i.EpisodesNew = i.EpisodesNew[0:cfg.IpadCheckMoreSize]
				}
				i.CheckMore = &CheckMore{
					Content: fmt.Sprintf(cfg.CheckMoreContent, seasonEp.EpSize),
					Uri:     fmt.Sprintf(cfg.CheckMoreSchema, _styleHorizontal, seasonEp.SeasonId), // must be horizontal
				}
			}
		} else { // pgc未下发，使用搜索的goto_url
			i.WatchButton = &WatchButton{
				Title: cfg.OnlineWatch,
				Link:  m.GotoURL,
			}
		}
	}
	for _, extFunc := range extraFunc {
		extFunc(i)
	}
}

func (i *Item) FromRecommendTips(m *Media) { // 无结果推荐卡
	i.Goto = model.GotoRecommendTips
	i.Title = m.Title
	i.Param = strconv.Itoa(int(m.MediaID))
}

// FromMedia form func
func (i *Item) FromMedia(m *Media, prompt string, gt string, bangumis map[string]*bangumimdl.Card, medisas map[int32]*pgcsearch.SearchMediaProto) {
	i.Title = m.Title
	if i.Title == "" {
		i.Title = m.OrgTitle
	}
	i.Cover = m.Cover
	i.Goto = gt
	i.Param = strconv.Itoa(int(m.MediaID))
	i.URI = m.GotoURL
	i.MediaType = m.MediaType
	i.PlayState = m.PlayState
	i.Style = m.Styles
	i.CV = m.CV
	i.Staff = m.Staff
	if m.MediaScore != nil {
		i.Rating = m.MediaScore.Score
		i.Vote = m.MediaScore.UserCount
	}
	i.PTime = m.Pubtime
	areas := strings.Split(m.Areas, "、")
	if len(areas) != 0 {
		i.Area = areas[0]
	}
	i.Prompt = prompt
	i.OutName = m.AllNetName
	i.OutIcon = m.AllNetIcon
	i.OutURL = m.AllNetURL
	if media, ok := medisas[int32(m.MediaID)]; ok && media != nil {
		i.PTime = xtime.Time(media.PubTime)
		i.Styles = media.Styles
	}
	var hit string
	for _, v := range m.HitColumns {
		if v == "cv" {
			hit = v
			break
		} else if v == "staff" {
			hit = v
		}
	}
	if hit == "cv" {
		for _, v := range getHightLight.FindAllStringSubmatch(m.CV, -1) {
			//nolint:gomnd
			if m.MediaType == 7 {
				i.Label = fmt.Sprintf("嘉宾: %v...", v[0])
				break
			}
			if gt == model.GotoBangumi {
				i.Label = fmt.Sprintf("声优: %v...", v[0])
				break
			} else if gt == model.GotoMovie {
				i.Label = fmt.Sprintf("演员: %v...", v[0])
				break
			}
		}
	} else if hit == "staff" {
		for _, v := range getHightLight.FindAllStringSubmatch(m.Staff, -1) {
			i.Label = fmt.Sprintf("制作人员: %v...", v[0])
			break
		}
	} else if hit == "" {
		i.Label = FormPGCLabel(m.MediaType, m.Styles, m.Staff, m.CV)
	}
	// get from PGC API.
	i.SeasonID = m.SeasonID
	ssID := strconv.Itoa(int(m.SeasonID))
	if bgm, ok := bangumis[ssID]; ok {
		i.Badge = model.FormMediaType(bgm.SeasonType)
		i.SeasonTypeName = bgm.SeasonTypeName
		i.IsAttention = bgm.IsFollow
		i.IsSelection = bgm.IsSelection
		i.SeasonType = bgm.SeasonType
		i.Badges = bgm.Badges
		for pos, v := range bgm.Episodes {
			tmp := &Item{
				Param:    strconv.Itoa(int(v.ID)),
				Index:    v.Index,
				Position: pos + 1,
				URI:      v.URL,
			}
			// tmp.URI = model.FillURI(model.GotoEP, tmp.Param, nil)
			i.Episodes = append(i.Episodes, tmp)
		}
	}
	//nolint:gomnd
	if m.MediaType > 100 {
		i.SeasonTypeName = model.FormMediaType(m.MediaType)
	}
	// var (
	//	badges []*model.ReasonStyle
	//	err    error
	// )
	// err = json.Unmarshal(m.DisplayInfo, &badges)
	// if err != nil {
	//	log.Error("%v", err)
	//	return
	// }
	// i.Badges = badges
	if i.Position == 0 {
		i.Position = m.Position
	}
}

// FromPgcRes builds the episode_new structure
func (v *EpisodeNew) FromPgcRes(ep *pgcsearch.SearchEpProto, isHorizon, gridBadge bool) (canAppend bool) {
	if isHorizon && ep.Title == "" { // 横条且pgc数据为空，认为为非法数据
		return false
	}
	if ep.ReleaseDate == "" { // pgc日期为空时只下发标题
		v.Title = ep.Title
	} else { // 否则日期拼到标题前面
		v.Title = fmt.Sprintf("%s %s", ep.ReleaseDate, ep.Title)
	}
	v.Uri = ep.Url
	v.Param = fmt.Sprintf("%d", ep.Id)
	if isHorizon || gridBadge { // 综合搜索+分类搜索
		v.Badges = ep.Badges
	}
	v.Type = ep.Type
	return true
}

func (i *Item) FromCloudGameConfigs() {
	i.ShowCloudGameEntry = true
	i.CloudGameParams = &CloudGameParams{Scene: "bili_search", SourceFrom: 1000040032}
}

// FromGame form func
func (i *Item) FromGame(g *Game, plat int8) {
	i.Title = g.Title
	i.Cover = g.Cover
	i.Desc = g.Desc
	i.Rating = g.View
	i.ReserveStatus = int64(g.Status)
	var reserve string
	if g.Status == 1 || g.Status == 2 {
		//nolint:gomnd
		if g.Like < 10000 {
			reserve = strconv.FormatInt(g.Like, 10) + "人预约"
		} else {
			reserve = strconv.FormatFloat(float64(g.Like)/10000, 'f', 1, 64) + "万人预约"
		}
	}
	i.Reserve = reserve
	i.Goto = model.GotoGame
	i.Param = strconv.FormatInt(g.ID, 10)
	i.URI = g.RedirectURL
	i.Tags = g.Tag
	i.NoticeName = g.NoticeName
	i.NoticeContent = g.NoticeContent
	if model.IsAndroid(plat) {
		i.GiftContent = g.GiftContentAndroid
		i.GiftURL = g.GiftURLAndroid
	} else if model.IsIOS(plat) {
		i.GiftContent = g.GiftContentIOS
		i.GiftURL = g.GiftURLIOS
	}
}

func (i *Item) FromGameBasedOnMultiGameInfos(gameId int64, multiGameInfos map[int64]*gmdl.Game) bool {
	v, ok := multiGameInfos[gameId]
	if !ok {
		return false
	}
	i.Title = v.GameName
	i.Cover = v.GameIcon
	i.Rating = v.Grade
	i.ReserveStatus = int64(v.GameStatus)
	i.Reserve = makeGameCardReserve(v.GameStatus, v.BookNum)
	i.Goto = model.GotoGame
	i.Param = strconv.FormatInt(gameId, 10)
	i.URI = v.GameLink
	i.Tags = v.GameTags
	i.NoticeName = v.NoticeTitle
	i.NoticeContent = v.Notice
	i.GiftContent = v.GiftTitle
	i.GiftURL = v.GiftUrl
	i.GameRank = v.GameRank
	i.RankType = v.RankType
	if v.RankInfo != nil {
		i.RankInfo = &RankInfo{
			SearchNightIconUrl:   v.RankInfo.SearchNightIconUrl,
			SearchDayIconUrl:     v.RankInfo.SearchDayIconUrl,
			SearchBkgNightColor:  v.RankInfo.SearchBkgNightColor,
			SearchBkgDayColor:    v.RankInfo.SearchBkgDayColor,
			SearchFontNightColor: v.RankInfo.SearchFontNightColor,
			SearchFontDayColor:   v.RankInfo.SearchFontDayColor,
			RankContent:          v.RankInfo.RankContent,
			RankLink:             makeGameRankLink(v.RankInfo.RankLink),
		}
	}
	return true
}

func makeGameCardReserve(gameStatus int32, bookNum int64) string {
	if gameStatus == 1 || gameStatus == 2 {
		//nolint:gomnd
		if bookNum < 10000 {
			return strconv.FormatInt(bookNum, 10) + "人预约"
		}
		return strconv.FormatFloat(float64(bookNum)/10000, 'f', 1, 64) + "万人预约"
	}
	return ""
}

func makeGameRankLink(link string) string {
	if link == "" {
		return ""
	}
	if strings.Contains(link, "sourceFrom") {
		return link
	}
	return fmt.Sprintf("%s&sourceFrom=1000040042", link)
}

// fillArcStat fill func
func (i *Item) fillArcStat(a *api.Arc) {
	if a.Access == 0 {
		i.Play = int(a.Stat.View)
	}
	i.Danmaku = int(a.Stat.Danmaku)
	i.Reply = int(a.Stat.Reply)
	i.Like = int(a.Stat.Like)
}

// fillArtStat fill func
func (i *Item) fillArtStat(m *article.Meta) {
	i.Play = int(m.Stats.View)
	i.Reply = int(m.Stats.Reply)
}

// FromSuggest form func
func (i *Item) FromSuggest(st *SuggestTag) {
	i.From = "search"
	if st.SpID == SuggestionJump {
		switch st.Type {
		case SuggestionAV:
			i.Title = st.Value
			i.Goto = model.GotoAv
			i.URI = model.FillURI(i.Goto, strconv.Itoa(int(st.Ref)), nil)
		case SuggestionLive:
			i.Title = st.Value
			i.Goto = model.GotoLive
			i.URI = model.FillURI(i.Goto, strconv.Itoa(int(st.Ref)), nil)
		}
	} else {
		i.Title = st.Value
	}
}

// FromSuggest2 form func
func (i *Item) FromSuggest2(st *SuggestTag, as map[int64]*api.Arc, ls map[int64]*livexroom.Infos) {
	i.From = "search"
	if st.SpID == SuggestionJump {
		switch st.Type {
		case SuggestionAV:
			i.Title = st.Value
			i.Goto = model.GotoAv
			i.URI = model.FillURI(i.Goto, strconv.Itoa(int(st.Ref)), model.AvHandler(as[st.Ref]))
		case SuggestionLive:
			var (
				l  *livexroom.Infos
				ok bool
			)
			i.Title = st.Value
			i.Goto = model.GotoLive
			if l, ok = ls[st.Ref]; !ok {
				for _, v := range ls {
					if v.Show != nil && v.Show.ShortId == st.Ref {
						l = v
						break
					}
				}
			}
			i.URI = model.FillURI(i.Goto, strconv.Itoa(int(st.Ref)), model.LiveHandler(l))
			if strings.Contains(i.URI, "broadcast_type") {
				i.URI += "&extra_jump_from=23004"
			} else {
				i.URI += "?extra_jump_from=23004"
			}
		}
	} else {
		i.Title = st.Value
	}
}

// FromSuggest3 form func
//
//nolint:gocognit
func (i *Item) FromSuggest3(st *Sug, as map[int64]*api.Arc, ls map[int64]*livexroomgate.EntryRoomInfoResp_EntryList, seasonm map[int32]*pgcsearch.SearchCardProto,
	nftRegion map[int64]*gallerygrpc.NFTRegion) {
	i.From = "search"
	i.Title = st.ShowName
	i.KeyWord = st.Term
	i.Position = st.Pos
	i.Cover = st.Cover
	i.CoverSize = st.CoverSize
	i.SugType = st.SubType
	i.TermType = st.TermType
	i.ModuleID = st.Ref
	if st.TermType == SuggestionJump {
		switch st.SubType {
		case SuggestionAV:
			i.Goto = model.GotoAv
			i.URI = model.FillURI(i.Goto, strconv.Itoa(int(st.Ref)), model.AvHandler(as[st.Ref]))
			i.SugType = "视频"
		case SuggestionLive:
			var (
				l  *livexroomgate.EntryRoomInfoResp_EntryList
				ok bool
			)
			i.Goto = model.GotoLive
			if l, ok = ls[st.Ref]; !ok {
				for _, v := range ls {
					if v.ShortId == st.Ref {
						l = v
						break
					}
				}
			}
			i.URI = model.FillURI(i.Goto, strconv.Itoa(int(st.Ref)), model.LiveEntryHandler(l, ""))
			if strings.Contains(i.URI, "broadcast_type") {
				i.URI += "&extra_jump_from=23004"
			} else {
				i.URI += "?extra_jump_from=23004"
			}
			i.LiveLink = model.FillURI(i.Goto, strconv.Itoa(int(st.Ref)), model.LiveEntryHandler(l, model.DefaultLiveEntry))
			i.SugType = "直播"
		case SuggestionArticle:
			i.Goto = model.GotoArticle
			i.URI = model.FillURI(i.Goto, strconv.Itoa(int(st.Ref)), nil)
			if !strings.Contains(i.URI, "column_from") {
				i.URI += "?column_from=search"
			}
			i.SugType = "专栏"
		}
	} else if st.TermType == SuggestionJumpUser && st.User != nil {
		i.Title = st.User.Name
		i.Cover = st.User.Face
		i.Goto = model.GotoAuthor
		i.OfficialVerify = &OfficialVerify{Type: st.User.OfficialVerifyType}
		i.Param = strconv.Itoa(int(st.User.Mid))
		i.URI = model.FillURI(i.Goto, i.Param, nil)
		i.Mid = st.User.Mid
		i.Fans = st.User.Fans
		i.Level = st.User.Level
		i.Arcs = st.User.Videos
		i.FaceNftNew = st.User.FaceNftNew
		if nftRegion != nil && i.FaceNftNew == 1 {
			if v, ok := nftRegion[i.Mid]; ok {
				i.NftFaceIcon = &NftFaceIcon{
					RegionType: int32(v.Type),
					Icon:       v.Icon,
					ShowStatus: int32(v.ShowStatus),
				}
			}
		}
		i.IsSeniorMember = st.User.IsSeniorMember
	} else if st.TermType == SuggestionJumpPGC && st.PGC != nil {
		i.Title = st.PGC.Title
		i.Cover = st.PGC.Cover
		i.PTime = st.PGC.Pubtime
		if ss, ok := seasonm[int32(st.PGC.SeasonID)]; ok && ss != nil {
			i.URI = ss.Url
			i.Styles = ss.Styles
		} else {
			i.URI = st.PGC.GotoURL
		}
		i.SeasonTypeName = model.FormMediaType(st.PGC.MediaType)
		i.Goto = model.GotoPGC
		i.Param = strconv.Itoa(int(st.PGC.MediaID))
		i.Area = st.PGC.Areas
		i.Style = st.PGC.Styles
		if i.Styles == "" {
			log.Warn("sug3 ssid(%v) styles backup logic", st.PGC.SeasonID)
			var styles []string
			if i.PTime != 0 {
				if pt := i.PTime.Time().Format("2006"); pt != "" {
					styles = append(styles, pt)
				}
			}
			if i.SeasonTypeName != "" {
				styles = append(styles, i.SeasonTypeName)
			}
			if i.Area != "" {
				styles = append(styles, i.Area)
			}
			if len(styles) > 0 {
				i.Styles = strings.Join(styles, " | ")
			}
		}
		i.Label = FormPGCLabel(st.PGC.MediaType, st.PGC.Styles, st.PGC.Staff, st.PGC.CV)
		i.Rating = st.PGC.MediaScore
		i.Vote = st.PGC.MediaUserCount
		i.Badges = st.PGC.Badges
	}
}

// FromQuery form func
func (i *Item) FromQuery(qs []*Query) {
	i.Goto = model.GOtoRecommendWord
	for pos, q := range qs {
		i.List = append(i.List, &Item{Param: strconv.FormatInt(q.ID, 10), Title: q.Name, Type: q.Type, FromSource: q.FromSource, Position: pos + 1})
	}
}

// FromComic form func
func (i *Item) FromComic(ctx context.Context, c *Comic) {
	i.ID = c.ID
	i.Title = c.Title
	if len(c.Author) > 0 {
		i.Name = fmt.Sprintf("作者: %v", strings.Join(c.Author, "、"))
	}
	i.Style = c.Styles
	i.Cover = c.Cover
	i.URI = c.URL
	i.ComicURL = c.ComicURL
	i.Param = strconv.FormatInt(c.ID, 10)
	i.Goto = model.GotoComic
	i.Badge = resolveSearchComicBadge(ctx, c.ComicType)
}

func resolveSearchComicBadge(ctx context.Context, comicType int64) string {
	const (
		_defaultComicBadge = "漫画"
	)
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().Or().IsPlatIPhoneI().Or().IsPlatIPhoneB().And().Build("<", int64(66500000))
	}).MustFinish() {
		// ios单端 665版本之前不适配网关下发文字
		return _defaultComicBadge
	}
	switch comicType {
	case 1:
		return "有声漫"
	default:
		return _defaultComicBadge
	}
}

// FromLiveMaster form func
func (i *Item) FromLiveMaster(l *Live, lv *livexroomgate.EntryRoomInfoResp_EntryList, userInfo *account.Card, extraFunc ...func(*Item)) {
	i.Type = l.Type
	i.Name = l.Uname
	i.UCover = l.Uface
	i.Attentions = l.Fans
	i.VerifyType = l.VerifyType
	i.VerifyDesc = l.VerifyDesc
	i.Title = l.Title
	if l.Cover != "" && l.Cover != _emptyLiveCover {
		i.Cover = l.Cover
	} else {
		i.Cover = _emptyLiveCover2
	}
	i.Goto = model.GotoLive
	i.Mid = l.UID
	i.RoomID = l.RoomID
	i.Param = strconv.Itoa(int(i.RoomID))
	i.URI = model.FillURI(i.Goto, i.Param, model.LiveEntryHandler(lv, ""))
	i.LiveLink = model.FillURI(i.Goto, i.Param, model.LiveEntryHandler(lv, model.DefaultLiveEntry))
	i.Online = l.Online
	i.LiveStatus = l.LiveStatus
	i.CateParentName = l.CateParentName
	i.CateNameNew = l.CateName
	if userInfo != nil && userInfo.FaceNftNew == 1 {
		i.FaceNftNew = userInfo.FaceNftNew
		i.NftDamrk = "https://i0.hdslb.com/bfs/live/9f176ff49d28c50e9c53ec1c3297bd1ee539b3d6.gif"
	}
	if lv != nil {
		i.WatchedShow = lv.WatchedShow
	}
	for _, extFunc := range extraFunc {
		extFunc(i)
	}
}

func WithLiveParentArea(mobiApp string, build int) func(*Item) {
	return func(i *Item) {
		if mobiApp == "android" || (mobiApp == "iphone" && build >= 63900000) || mobiApp == "ipad" {
			i.CateParentName = ""
		}
	}
}

// FromTwitter form twitter
func (i *Item) FromTwitter(t *Twitter, details map[int64]*bplus.Detail, dynamicTopic map[int64]*bplus.DynamicTopics, isUP, isCount, isNew bool) {
	var (
		gt, id string
	)
	i.Title = t.Content
	i.Covers = t.Cover
	i.CoverCount = t.CoverCount
	i.Param = strconv.FormatInt(t.ID, 10)
	i.Goto = model.GotoTwitter
	if isNew {
		gt = model.GotoDynamic
		id = i.Param
	} else {
		gt = model.GotoTwitter
		id = strconv.FormatInt(t.PicID, 10)
	}
	i.URI = model.FillURI(gt, id, nil)
	if detail, ok := details[t.ID]; ok {
		if isUP {
			ii := &Item{
				Mid:       detail.Mid,
				Title:     detail.NickName,
				Cover:     detail.FaceImg,
				PTimeText: detail.PublishTimeText,
			}
			i.Upper = ii
		}
		if isCount {
			ii := &Item{
				Play:  detail.ViewCount,
				Like:  detail.LikeCount,
				Reply: detail.CommentCount,
			}
			i.State = ii
		}
	}
	if topic, k := dynamicTopic[t.ID]; k {
		l := len(topic.FromContent)
		if l > 0 {
			i.DyTopic = make([]*Item, 0, l)
			for pos, v := range topic.FromContent {
				temp := &Item{
					Title:      v.TopicName,
					IsActivity: v.IsActivity,
					URI:        v.TopicLink,
					Position:   pos + 1,
				}
				i.DyTopic = append(i.DyTopic, temp)
			}
		}
	}
}

// FromRcmdPre from rcmd pre.
func (i *Item) FromRcmdPre(id int64, a *api.Arc, bangumi *seasongrpc.CardInfoProto) {
	if a != nil {
		i.Title = a.Title
		i.Cover = a.Pic
		i.Author = a.Author.Name
		i.Param = strconv.Itoa(int(id))
		i.Goto = model.GotoAv
		i.URI = model.FillURI(i.Goto, i.Param, model.AvHandler(a))
		i.fillArcStat(a)
		i.Desc = a.Desc
		i.DurationInt = a.Duration
	} else if bangumi != nil {
		i.Title = bangumi.Title
		i.Cover = bangumi.Cover
		i.Param = strconv.Itoa(int(id))
		i.Goto = model.GotoPGC
		i.URI = model.FillURI(i.Goto, i.Param, nil)
		i.Badge = bangumi.SeasonTypeName
		i.Started = int8(bangumi.IsStarted)
		i.Play = int(bangumi.Stat.View)
		if bangumi.Rating != nil {
			i.Rating = float64(bangumi.Rating.Score)
			i.RatingCount = int(bangumi.Rating.Count)
		}
		i.MediaType = int(bangumi.SeasonType) // 1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧
		if bangumi.Stat != nil {
			i.Attentions = int(bangumi.Stat.Follow)
		}
		if bangumi.NewEp != nil {
			i.Label = bangumi.NewEp.IndexShow
		}
	}
}

// FromTicket from ticket
func (i *Item) FromTicket(t *Ticket) {
	i.ID = t.ID
	i.Param = strconv.Itoa(int(t.ID))
	i.Goto = model.GotoTicket
	i.Badge = "展演"
	i.Title = t.Title
	i.Cover = t.Cover
	i.ShowTime = t.ShowTime
	i.City = t.CityName
	i.Venue = t.VenueName
	i.Price = int(math.Ceil(float64(t.PriceLow) / 100))
	i.PriceComplete = strconv.FormatFloat(float64(t.PriceLow)/100, 'f', -1, 64)
	i.PriceType = t.PriceType
	i.ReqNum = t.ReqNum
	i.URI = t.URL
}

// FromProduct from ticket
func (i *Item) FromProduct(p *Product) {
	i.ID = p.ID
	i.Param = strconv.Itoa(int(p.ID))
	i.Goto = model.GotoProduct
	i.Badge = "商品"
	i.Title = p.Title
	i.Cover = p.Cover
	i.ShopName = p.ShopName
	i.Price = int(math.Ceil(float64(p.Price) / 100))
	i.PriceComplete = strconv.FormatFloat(float64(p.Price)/100, 'f', -1, 64)
	i.PriceType = p.PriceType
	i.ReqNum = p.ReqNum
	i.URI = p.URL
}

// FromSpecialerGuide from ticket
func (i *Item) FromSpecialerGuide(sg *SpecialerGuide) {
	i.ID = sg.ID
	i.Param = strconv.Itoa(int(sg.ID))
	i.Goto = model.GotoSpecialerGuide
	i.Title = sg.Title
	i.Cover = sg.Cover
	i.Desc = sg.Desc
	i.Phone = sg.Tel
}

// FromTagPGC from pgc tag.
func (i *Item) FromTagPGC(m *Media, bangumi *seasongrpc.CardInfoProto) {
	if m.SeasonID == 0 {
		return
	}
	ssid := strconv.Itoa(int(m.SeasonID))
	i.Title = bangumi.Title
	i.Cover = bangumi.Cover
	i.Param = strconv.Itoa(int(m.MediaID))
	i.Goto = model.GotoPGC
	i.URI = model.FillURI(i.Goto, ssid, nil)
	i.Badge = bangumi.SeasonTypeName
	i.Started = int8(bangumi.IsStarted)
	i.Play = int(bangumi.Stat.View)
	if bangumi.Rating != nil {
		i.Rating = float64(bangumi.Rating.Score)
		i.RatingCount = int(bangumi.Rating.Count)
	}
	i.MediaType = int(bangumi.SeasonType) // 1：番剧，2：电影，3：纪录片，4：国漫，5：电视剧
	if bangumi.Stat != nil {
		i.Attentions = int(bangumi.Stat.Follow)
	}
	if bangumi.NewEp != nil {
		i.Label = bangumi.NewEp.IndexShow
	}
}

// FormSpace form space search
func (i *Item) FormSpace(v *SpaceValue) {
	i.Title = v.Title
	i.Cover = v.Pic
	i.Param = strconv.FormatInt(v.Aid, 10)
	i.Goto = model.GotoAv
	i.URI = model.FillURI(i.Goto, i.Param, nil)
	switch play := v.Play.(type) {
	case float64:
		i.Play = int(play)
	case string:
		i.Play, _ = strconv.Atoi(play)
	}
	i.Danmaku = int(v.Danmaku)
	i.Duration = v.Duration
	i.PTime = timeStrToInt(v.Created)
}

// timeStrToInt .
func timeStrToInt(timeStr string) (timeInt xtime.Time) {
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(timeLayout, timeStr, loc)
	if err := timeInt.Scan(theTime); err != nil {
		log.Error("timeInt.Scan error(%v)", err)
	}
	return
}

// flowTest form func
// func flowTest(buvid string) (ok bool) {
// 	id := crc32.ChecksumIEEE([]byte(reverseString(buvid))) % 2
// 	if id%2 > 0 {
// 		ok = true
// 	}
// 	return
// }

// reverseString form func
// func reverseString(s string) string {
// 	runes := []rune(s)
// 	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
// 		runes[from], runes[to] = runes[to], runes[from]
// 	}
// 	return string(runes)
// }

// FormPGCLabel from pgc labe.
func FormPGCLabel(mediaType int, styles, staff, cv string) (label string) {
	//nolint:gomnd
	switch mediaType {
	case 1: // 番剧
		label = strings.Replace(styles, "\n", "、", -1)
	case 2: // 电影
		if cv != "" {
			label = "演员：" + strings.Replace(cv, "\n", "、", -1)
		}
	case 3: // 纪录片
		label = strings.Replace(staff, "\n", "、", -1)
	case 4: // 国创
		label = strings.Replace(styles, "\n", "、", -1)
	case 5: // 电视剧
		if cv != "" {
			label = "演员：" + strings.Replace(cv, "\n", "、", -1)
		}
	case 7: // 综艺
		label = strings.Replace(cv, "\n", "、", -1)
	// case 123: // 电视剧
	//	label = "演员：" + strings.Replace(cv, "\n", "、", -1)
	// case 124: // 综艺
	//	label = strings.Replace(cv, "\n", "、", -1)
	// case 125: // 纪录片
	//	label = strings.Replace(staff, "\n", "、", -1)
	// case 126: // 电影
	//	label = "演员：" + strings.Replace(cv, "\n", "、", -1)
	// case 127: // 动漫
	//	label = strings.Replace(styles, "\n", "、", -1)
	default:
		label = strings.Replace(cv, "\n", "、", -1)
	}
	return
}

// FromConverge2 from converge.
func (i *Item) FromConverge2(u *ConvergeUser, v *ConvergeVideo) {
	if u != nil {
		i.Title = u.Name
		i.Cover = u.Face
		i.Goto = model.GotoSpace
		i.OfficialVerify = &OfficialVerify{Type: u.OfficeType}
		i.Param = strconv.Itoa(int(u.Mid))
		i.URI = model.FillURI(i.Goto, i.Param, nil)
		i.Mid = u.Mid
		i.Fans = u.Fans
		i.Arcs = u.Videos
	} else if v != nil {
		i.Title = v.Title
		i.Cover = v.Cover
		i.Param = strconv.Itoa(int(v.Aid))
		i.Goto = model.GotoAv
		i.URI = model.FillURI(i.Goto, i.Param, nil)
		i.Play = v.Play
		i.Danmaku = v.Danmaku
		i.Duration = v.Duration
	}
}

// EpsNewResult def.
type EpsNewResult struct {
	Episodes []*Item `json:"episodes"`
	Title    string  `json:"title"`
	Total    int32   `json:"total"`
}

//nolint:gocognit
func (i *Item) FromOGVCard(card *SearchOGVCard, seasonstat map[int32]*pgcstat.SeasonStatProto, bangumis map[int32]*seasongrpc.CardInfoProto, plat int8) (cardHead *OGVCard, is []*Item, isBroke bool) {
	cardHead = &OGVCard{
		LinkType:       model.GotoOGVCard,
		Goto:           model.GotoOGVCard,
		Position:       i.Position,
		Title:          card.HeadArea.Title,
		SubTitle1:      card.HeadArea.SubTitle,
		Cover:          card.HeadArea.Cover,
		BgCover:        card.HeadArea.BgCover,
		SpecialBgColor: card.SpecialBgColor,
		TrackID:        i.TrackID,
		Param:          strconv.FormatInt(card.ID, 10),
	}
	const (
		_moreTypeSearch  = 0
		_moreTypeAv      = 1
		_moreTypeArticle = 2
		_moreTypeBangumi = 3
		_moreTypeWeb     = 4
	)
	var (
		allStat int64
		uri     string
	)
	for _, module := range card.Modules {
		switch module.Type {
		case OGVCardTypeGame:
			for _, v := range module.Values {
				item := &Item{
					SpecialBgColor: card.SpecialBgColor,
				}
				item.FromGame(v.Game, plat)
				item.Goto = model.GotoNewGame
				item.Position = i.Position
				item.TrackID = i.TrackID
				item.LinkType = module.LinkType
				is = append(is, item)
			}
		case OGVCardTypePGC:
			var bangumiCount int
			item := &Item{
				SpecialBgColor: card.SpecialBgColor,
				Goto:           model.GotoBangumiRelates,
				LinkType:       module.LinkType,
				MoreText:       "更多",
				TrackID:        i.TrackID,
				Position:       i.Position,
			}
			for _, v := range module.Values {
				item.MoreURL = v.MoreURL
				uri = item.MoreURL
				for pos, ssid := range v.SeasonIDList {
					bangumiCard, ok := bangumis[int32(ssid)]
					if !ok || bangumis == nil {
						continue
					}
					items := &Item{
						Title:    bangumiCard.Title,
						Param:    strconv.Itoa(int(bangumiCard.SeasonId)),
						Goto:     model.GotoBangumi,
						Cover:    bangumiCard.Cover,
						Position: pos + 1,
					}
					if bangumiCard.Url != "" {
						items.URI = bangumiCard.Url
					} else {
						items.URI = model.FillURI(model.GotoBangumi, strconv.Itoa(int(bangumiCard.SeasonId)), nil)
					}
					if bangumiCard.BadgeInfo != nil {
						items.BadgeStyle = &model.ReasonStyle{
							Text:      bangumiCard.BadgeInfo.Text,
							TextColor: "#FFFFFFFF",
							BgColor:   bangumiCard.BadgeInfo.BgColor,
						}
					}
					var pgcview int64
					if pgcstat, ok := seasonstat[int32(ssid)]; ok {
						pgcview = pgcstat.View
					}
					if pgcview > 0 {
						items.CoverLeftText = statString(pgcview, "观看")
					}
					item.Items = append(item.Items, items)
					// casrd stat
					allStat += pgcview
				}
				bangumiCount += len(v.SeasonIDList)
			}
			//nolint:gomnd
			switch len(item.Items) {
			case 0, 1:
				isBroke = true
				return
			case 4:
				item.Items = item.Items[:3]
			case 2, 3, 5, 6:
				item.MoreURL = ""
			default:
				// len(item.Items) > 6
				item.Items = item.Items[:6]
			}
			if item.MoreURL == "" {
				item.MoreText = ""
			}
			item.Title = module.Title + "（" + strconv.Itoa(bangumiCount) + "部）"
			if allStat > 0 {
				cardHead.SubTitle2 = "系列播放数：" + statString(allStat, "")
			}
			is = append(is, item)
		case OGVCardTypeMore:
			item := &Item{
				SpecialBgColor: card.SpecialBgColor,
				Goto:           model.GotoFindMore,
				LinkType:       module.LinkType,
				Title:          module.Title,
				TrackID:        i.TrackID,
				Position:       i.Position,
			}
			for _, v := range module.Values {
				items := &Item{
					Title: v.ShowName,
					Param: strconv.FormatInt(v.ID, 10),
				}
				switch v.Type {
				case _moreTypeSearch:
					items.Goto = model.GotoSearch
				case _moreTypeAv:
					items.Goto = model.GotoAv
				case _moreTypeArticle:
					items.Goto = model.GotoArticle
				case _moreTypeBangumi:
					items.Goto = model.GotoBangumi
					//nolint:gosec
					ssid, _ := strconv.Atoi(v.Value)
					if ss, ok := bangumis[int32(ssid)]; ok && ss != nil {
						items.URI = ss.Url
					}
				case _moreTypeWeb:
					items.Goto = model.GotoWeb
				}
				if items.URI == "" {
					items.URI = model.FillURI(items.Goto, v.Value, nil)
				}
				item.Items = append(item.Items, items)
			}
			is = append(is, item)
		}
	}
	cardHead.URI = uri
	cardHead.CoverURI = uri
	return
}

func anyOlympic(in []*Item) bool {
	for _, i := range in {
		if i.IsOlympic {
			return true
		}
	}
	return false
}

func (i *Item) FormESport(es *ESport, localTime int64, mm map[int64]*esportGRPC.Contest, liveEntry map[int64]*livexroomgate.EntryRoomInfoResp_EntryList, extraFunc ...func(*Item)) {
	i.formESport(es, localTime, mm, liveEntry, extraFunc...)
	if anyOlympic(i.Items) {
		i.formESportAsOlympic(es, localTime, mm, liveEntry)
	}
}

func (i *Item) formESportAsOlympic(_ *ESport, _ int64, mm map[int64]*esportGRPC.Contest, _ map[int64]*livexroomgate.EntryRoomInfoResp_EntryList) {
	i.MatchTop.Text = "奥运热点"
	i.MatchBottom.Text = "更多热门赛事"
	i.Cover = ""

	for _, ii := range i.Items {
		if !ii.IsOlympic {
			continue
		}
		match, ok := mm[ii.ID]
		if !ok {
			continue
		}

		//nolint:gomnd
		switch match.GameState {
		case 6:
			ii.Status = 1
			ii.MatchLabel = &MatchItem{
				Text:           "未开始",
				TextColor:      "#999999",
				TextColorNight: "#686868",
			}
		case 5:
			ii.Status = 2
			ii.MatchLabel = &MatchItem{
				Text:           "进行中",
				TextColor:      "#FB7299",
				TextColorNight: "#BB5B76",
			}
		case 1:
			ii.Status = 3
			ii.MatchLabel = &MatchItem{
				Text:           "已结束",
				TextColor:      "#999999",
				TextColorNight: "#686868",
			}
		default:
			log.Warn("Unrecognized match game state: %+v", match)
			continue
		}

		// 未开始或进行中
		if ii.Status == 1 || ii.Status == 2 {
			ii.MatchButton = &MatchItem{
				Text:     "敬请期待",
				URI:      "",
				State:    3,
				LiveLink: "",
			}
		}
		// 已开始或已结束
		if ii.Status == 2 || ii.Status == 3 {
			// 无集锦
			ii.MatchButton = &MatchItem{
				Text:     "敬请期待",
				URI:      "",
				State:    9,
				LiveLink: "",
			}
			// 有集锦
			if match.CollectionURL != "" {
				ii.MatchButton = &MatchItem{
					Text:     "观看集锦",
					URI:      match.CollectionURL,
					State:    7,
					LiveLink: "",
				}
			}
		}

		if match.OlympicShowRule == 0 {
			ii.MatchButton = &MatchItem{
				Text:     "",
				URI:      "",
				State:    3,
				LiveLink: "",
			}
		}
	}
}

func (i *Item) FromSportsVersus(sports *Sports, match *esportsservice.SportsEventMatchItem, localTime int64, liveEntry map[int64]*livexroomgate.EntryRoomInfoResp_EntryList, extraFunc ...func(*Item)) error {
	if match.Home == nil || match.Away == nil {
		return errors.Errorf("UnExpected SportsEventMatchItem match=%+v", match)
	}
	i.ID = sports.ID
	i.Title = sports.Title
	i.Goto = model.GotoSportsVersus
	i.BgCover = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/ol5o0QeoC3.png" // 冬奥头部图片
	i.Param = strconv.FormatInt(sports.SeasonId, 10)                                                            // param上报赛季id用
	ii := &Item{}
	ii.ID = sports.ID
	ii.Param = strconv.FormatInt(sports.ID, 10)

	// 卡片底部引导
	i.MatchBottom = &MatchItem{
		Text: "热门赛程",
		URI:  sports.Url,
	}
	// 主队信息
	ii.Team1 = &MatchTeam{
		ID:    match.Home.ParticipantId,
		Title: match.Home.ParticipantName,
		Cover: match.Home.ParticipantImg,
	}
	if score, err := strconv.ParseInt(match.Home.ParticipantResult, 10, 64); err == nil {
		ii.Team1.Score = score
	}
	// 客队信息
	ii.Team2 = &MatchTeam{
		ID:    match.Away.ParticipantId,
		Title: match.Away.ParticipantName,
		Cover: match.Away.ParticipantImg,
	}
	if score, err := strconv.ParseInt(match.Away.ParticipantResult, 10, 64); err == nil {
		ii.Team2.Score = score
	}
	var (
		labelText      string
		textColor      string
		textColorNight string
		matchState     int
	)
	switch transferSportsMatchStatus(match.MatchStatus) {
	case _sportsStatusStarting:
		matchState = 2
		labelText = "进行中"
		textColor = "#FB7299"
		textColorNight = "#BB5B76"
	case _sportsStatusFinish:
		matchState = 3
		labelText = "已结束"
		textColor = "#999999"
		textColorNight = "#686868"
	default:
		matchState = 1
	}
	// 比赛状态文案(未开始不下发)
	if labelText != "" && matchState != 1 {
		ii.MatchLabel = &MatchItem{
			Text:           labelText,
			TextColor:      textColor,
			TextColorNight: textColorNight,
		}
	}
	// 比赛开始时间文案
	if timeText := formMatchTime(match.BeginTime, localTime); timeText != "" {
		ii.MatchTime = &MatchItem{
			Text: timeText,
		}
	}
	// 比赛引导按钮
	ii.MatchButton = formSportsMatchButton(match, liveEntry[match.QueryCard.GetUpMid()])
	ii.Status = matchState
	ii.MatchStage = match.Name
	i.Items = append(i.Items, ii)
	i.Right = true
	for _, extFunc := range extraFunc {
		extFunc(i)
	}
	return nil
}

func formSportsMatchButton(match *esportsservice.SportsEventMatchItem, liveEntry *livexroomgate.EntryRoomInfoResp_EntryList) *MatchItem {
	const (
		_matchButtonState           = 8
		_matchBeforeNoResourceState = 3
		_matchIngNoResourceState    = 5
		_matchAfterNoResourceState  = 9
	)
	if match.QueryCard == nil {
		return nil
	}
	text, url := match.QueryCard.Content, match.QueryCard.JumpUrl
	var state int
	switch transferSportsMatchStatus(match.MatchStatus) {
	case _sportsStatusReady:
		state = _matchBeforeNoResourceState
		if url != "" {
			state = _matchButtonState // 有按钮样式
		}
	case _sportsStatusStarting:
		state = _matchIngNoResourceState
	case _sportsStatusFinish:
		state = _matchAfterNoResourceState
		if liveEntry != nil || url != "" {
			state = _matchButtonState
		}
	default:
		state = _matchAfterNoResourceState
	}
	return &MatchItem{
		Text:  text,
		URI:   url,
		State: state,
	}
}

func transferSportsMatchStatus(status esportsservice.SportsMatchStatusEnum) int {
	switch status {
	case esportsservice.SportsMatchStatusEnum_MatchStatusScheduled, esportsservice.SportsMatchStatusEnum_MatchStatusRescheduled, esportsservice.SportsMatchStatusEnum_MatchStatusPostponed, esportsservice.SportsMatchStatusEnum_MatchStatusGettingReady:
		return _sportsStatusReady
	case esportsservice.SportsMatchStatusEnum_MatchStatusRunning, esportsservice.SportsMatchStatusEnum_MatchStatusScheduledBreak, esportsservice.SportsMatchStatusEnum_MatchStatusDelayed:
		return _sportsStatusStarting
	case esportsservice.SportsMatchStatusEnum_MatchStatusFinished, esportsservice.SportsMatchStatusEnum_MatchStatusCancelled:
		return _sportsStatusFinish
	default:
		return 0
	}
}

func formMatchStatusDesc(status esportsservice.SportsMatchStatusEnum) string {
	switch transferSportsMatchStatus(status) {
	case _sportsStatusReady:
		return "未开始"
	case _sportsStatusStarting:
		return "进行中"
	case _sportsStatusFinish:
		return "已结束"
	default:
		return ""
	}
}

func (i *Item) FromSports(sports *Sports, match *esportsservice.SportsEventMatchItem, localTime int64, extraFunc ...func(*Item)) error {
	if match.Name == "" {
		return errors.Errorf("UnExpected SportsEventMatchItem match=%+v", match)
	}
	i.ID = sports.ID
	i.Title = sports.Title
	i.Goto = model.GotoSports
	i.BgCover = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/ol5o0QeoC3.png" // 冬奥头部图片
	i.Param = strconv.FormatInt(sports.SeasonId, 10)                                                            // param上报赛季id用
	i.SportsMatchItem = &SportsMatchItem{
		MatchId:      match.Id,
		SeasonId:     sports.SeasonId,
		MatchName:    match.Name,
		Img:          match.Img,
		SubContent:   match.Content,
		SubExtraIcon: match.QueryCard.InlineIcon,
	}
	// 比赛开始时间文案
	if timeText := formMatchTime(match.BeginTime, localTime); timeText != "" {
		i.SportsMatchItem.BeginTimeDesc = timeText
	}
	// 比赛状态字段映射
	if statusDesc := formMatchStatusDesc(match.MatchStatus); statusDesc != "" {
		i.SportsMatchItem.MatchStatusDesc = statusDesc
	}
	for _, extFunc := range extraFunc {
		extFunc(i)
	}
	return nil
}

func (i *Item) formESport(es *ESport, localTime int64, mm map[int64]*esportGRPC.Contest, liveEntry map[int64]*livexroomgate.EntryRoomInfoResp_EntryList, extraFunc ...func(*Item)) {
	var cover, bgcover string
	i.ID = es.ID
	i.Title = es.Title
	i.Goto = model.GotoESports
	i.Param = strconv.FormatInt(es.ID, 10)
	// 右上角顶部引导
	i.MatchTop = &MatchItem{
		Text: "全部赛程",
		URI:  es.UrlTop,
	}
	// 产品说遇到 240 和 215 就改成赛事专题
	if i.ID == 240 || i.ID == 215 {
		i.MatchTop.Text = "赛事专题"
	}
	// 卡片底部引导
	i.MatchBottom = &MatchItem{
		Text: "全部赛程",
		URI:  es.UrlBottom,
	}
	for _, e := range es.MatchList {
		var (
			match *esportGRPC.Contest
			ok    bool
		)
		if match, ok = mm[e.ID]; !ok {
			continue
		}
		if match.HomeTeam == nil {
			continue
		}
		if match.AwayTeam == nil {
			continue
		}
		if match.Season != nil {
			bgcover = match.Season.SearchImage
			cover = match.Season.LogoFull
		}
		ii := &Item{}
		ii.ID = e.ID
		ii.Param = strconv.FormatInt(e.ID, 10)
		ii.MatchStage = match.GameStage
		ii.IsOlympic = match.IsOlympic
		// 主队信息
		ii.Team1 = &MatchTeam{
			ID:    match.HomeTeam.ID,
			Title: match.HomeTeam.Title,
			Cover: match.HomeTeam.LogoFull,
			Score: match.HomeScore,
		}
		// 客队信息
		ii.Team2 = &MatchTeam{
			ID:    match.AwayTeam.ID,
			Title: match.AwayTeam.Title,
			Cover: match.AwayTeam.LogoFull,
			Score: match.AwayScore,
		}
		var (
			labelText      string
			textColor      string
			textColorNight string
			stime          = match.Stime
			matchState     int
		)
		if match.GameState == MatchStateBattling || match.GameState == MatchStateLive {
			matchState = 2
			labelText = "进行中"
			textColor = "#FB7299"
			textColorNight = "#BB5B76"
		} else if match.GameState == MatchStateOver {
			matchState = 3
			labelText = "已结束"
			textColor = "#999999"
			textColorNight = "#686868"
		} else if match.GameState == MatchStateAtten || match.GameState == MatchStateAttenLive { // 剩下的必然就是未开始
			matchState = 1
		}
		ii.Status = matchState
		// 比赛状态文案(未开始不下发)
		if labelText != "" && matchState != 1 {
			ii.MatchLabel = &MatchItem{
				Text:           labelText,
				TextColor:      textColor,
				TextColorNight: textColorNight,
			}
		}
		// 比赛开始时间文案
		if timeText := formMatchTime(stime, localTime); timeText != "" {
			ii.MatchTime = &MatchItem{
				Text: timeText,
			}
		}
		// 比赛引导按钮
		var (
			buttonText, buttonURI, liveLink string
			buttonState                     int
		)
		buttonText, buttonURI, liveLink, buttonState = formMatchState(matchState, es, match, liveEntry[match.LiveRoom])
		ii.MatchButton = &MatchItem{
			Text:     buttonText,
			URI:      buttonURI,
			State:    buttonState,
			LiveLink: liveLink,
		}
		if ii.MatchButton.State == 1 || ii.MatchButton.State == 2 {
			ii.MatchButton.Texts = esportButton
			if ii.MatchButton.URI == "" {
				ii.MatchButton.URI = fmt.Sprintf("https://www.bilibili.com/h5/match/data/detail/%d", match.ID)
			}
		}
		i.Items = append(i.Items, ii)
		i.Right = true
	}
	i.Cover = cover
	i.BgCover = bgcover

	for _, extFunc := range extraFunc {
		extFunc(i)
	}
}

func WithESportConfig(esId int64, textBottom, urlBottom string, esportConfigs map[int64]*managersearch.EsportConfigInfo, plat int8) func(*Item) {
	return func(i *Item) {
		defer func() {
			switch {
			case model.IsIPhone(plat) || model.IsAndroid(plat):
				if i.MatchTop != nil {
					i.MatchTop.Text = fmt.Sprintf("%s>", i.MatchTop.Text)
				}
				if i.MatchBottom != nil {
					i.MatchBottom.Text = fmt.Sprintf("%s>", i.MatchBottom.Text)
				}
				return
			case model.IsIPad(plat):
				i.MatchBottom.Text = fmt.Sprintf("%s>", i.MatchBottom.Text)
				return
			}
		}()
		ec, ok := esportConfigs[esId]
		if !ok {
			return
		}
		extraLink := []*ExtraLink{}
		for _, btn := range ec.BtnList {
			switch {
			case model.IsAndroid(plat) || model.IsIPhone(plat):
				//nolint:gomnd
				switch btn.Pos {
				case 1:
					i.MatchTop = &MatchItem{
						Text: btn.Text,
						URI:  btn.Link,
					}
				case 2, 3, 4:
					extraLink = append(extraLink, &ExtraLink{
						Text: btn.Text,
						URI:  btn.Link,
					})
				default:
					log.Warn("Invalid btn: %+v", btn)
					continue
				}
			case model.IsIPad(plat):
				extraLink = append(extraLink, &ExtraLink{
					Text: btn.Text,
					URI:  btn.Link,
				})
			}
		}

		switch {
		case model.IsAndroid(plat) || model.IsIPhone(plat):
			if len(extraLink) > 0 {
				if urlBottom != "" {
					i.ExtraLink = append(i.ExtraLink, &ExtraLink{
						Text: textBottom,
						URI:  urlBottom,
					})
				}
				i.ExtraLink = append(i.ExtraLink, extraLink...)
				i.MatchBottom = nil
			}
		case model.IsIPad(plat):
			i.ExtraLink = extraLink
		}
	}
}

func formMatchTime(stime, localTime int64) (label string) {
	// 计算时区差值(默认服务端固定东八区)
	// 与客户端约定：东一至东十二区分别1到12; 0时区0; 西一至西十一分别-1到-11
	dd, _ := time.ParseDuration(fmt.Sprintf("%dh", localTime-8))
	// 用户所在地的相对开赛时间
	ls := time.Unix(stime, 0).Add(dd)
	// 用户所在地的标准时间
	lt := time.Now().Add(dd)
	if lt.Year() == ls.Year() {
		if lt.YearDay()-ls.YearDay() == 1 {
			label = fmt.Sprintf("昨天 %v", ls.Format("15:04"))
			return
		} else if lt.YearDay()-ls.YearDay() == 0 {
			label = fmt.Sprintf("今天 %v", ls.Format("15:04"))
			return
		} else if lt.YearDay()-ls.YearDay() == -1 {
			label = fmt.Sprintf("明天 %v", ls.Format("15:04"))
			return
		} else {
			label = ls.Format("01-02 15:04")
		}
	} else {
		label = ls.Format("2006-01-02 15:04")
	}
	return
}

func formMatchState(matchState int, _ *ESport, match *esportGRPC.Contest, liveEntry *livexroomgate.EntryRoomInfoResp_EntryList) (label, uri, liveLink string, state int) {
	//nolint:gomnd
	switch matchState {
	case 1: // 赛前
		if match.GameState == MatchStateAtten {
			state = 1
			label = "已订阅"
		} else if match.LiveRoom != 0 {
			state = 2
			label = "订阅"
			// uri = model.FillURI(model.GotoLiveWeb, strconv.FormatInt(match.LiveRoom, 10), nil)
		} else {
			state = 3
			label = "敬请期待"
		}
	case 2:
		if match.LiveRoom != 0 {
			state = 4
			label = "观看直播"
			uri = model.FillURI(model.GotoLiveWeb, strconv.FormatInt(match.LiveRoom, 10), nil)
			liveLink = model.FillURI(model.GotoLiveWeb, strconv.FormatInt(match.LiveRoom, 10), model.LiveEntryHandler(liveEntry, model.DefaultLiveEntry))
		} else {
			state = 5
			label = "敬请期待"
		}
	case 3:
		if match.Playback != "" {
			state = 6
			label = "观看回放"
			uri = match.Playback
		} else if match.CollectionURL != "" {
			state = 7
			label = "观看集锦"
			uri = match.CollectionURL
		} else if match.LiveRoom != 0 {
			state = 8
			label = "直播间"
			uri = model.FillURI(model.GotoLiveWeb, strconv.FormatInt(match.LiveRoom, 10), nil)
		} else {
			state = 9
			label = "敬请期待"
		}
	}
	return
}

// statString Stat to string
func statString(number int64, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	//nolint:gomnd
	if number < 10000 {
		s = strconv.FormatInt(number, 10) + suffix
		return
	}
	//nolint:gomnd
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

type ChannelResult struct {
	TrackID       string         `json:"trackid"`
	Pages         int            `json:"pages"`
	Total         int            `json:"total"`
	FaildNum      int            `json:"faild_num"`
	ExpStr        string         `json:"exp_str"`
	Items         []*ChannleItem `json:"items,omitempty"`
	Extend        *ChannleItem2  `json:"extend,omitempty"`
	NoSearchLabel string         `json:"no_search_label,omitempty"`
	NoMoreLabel   string         `json:"no_more_label,omitempty"`
}

type SearchChannelConfig struct {
	More string `json:"more"`
	Hot  string `json:"hot"`
}

type ChannleItem struct {
	ID             int64          `json:"id,omitempty"`
	Title          string         `json:"title,omitempty"`
	Cover          string         `json:"cover,omitempty"`
	URI            string         `json:"uri,omitempty"`
	Param          string         `json:"param,omitempty"`
	Goto           string         `json:"goto,omitempty"`
	IsAtten        int            `json:"is_atten"`
	Label          string         `json:"label,omitempty"`
	Label2         string         `json:"label2,omitempty"`
	TypeIcon       string         `json:"type_icon,omitempty"`
	Right          bool           `json:"-"`
	Icon           string         `json:"icon,omitempty"`
	Button         *SearchButton  `json:"button,omitempty"`
	Items          []*ChannleItem `json:"items,omitempty"`
	CoverLeftText1 string         `json:"cover_left_text_1,omitempty"`
	CoverLeftIcon1 int            `json:"cover_left_icon_1,omitempty"`
	Badge          *ChannelBadge  `json:"badge,omitempty"`
	More           *SearchButton  `json:"more,omitempty"`
	ThemeColor     string         `json:"theme_color,omitempty"`
	Alpha          int32          `json:"alpha,omitempty"`
	// 夜间模式颜色，服务端对明暗度做了调整
	ThemeColorNight string `json:"theme_color_night,omitempty"`
}

type ChannleItem2 struct {
	Label     string         `json:"label"`
	ModelType string         `json:"model_type"`
	Items     []*ChannleItem `json:"items"`
}

type SearchButton struct {
	Text string `json:"text,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type ChannelBadge struct {
	Text      string `json:"text,omitempty"`
	IconBgURL string `json:"icon_bg_url,omitempty"`
}

type IterationConverge struct {
	Type              string       `json:"type,omitempty"`
	Title             string       `json:"title,omitempty"`
	Data              interface{}  `json:"data,omitempty"`
	SearchRankingMeta *RankingMeta `json:"search_ranking_meta,omitempty"`
}

type RankingMeta struct {
	OpenSearchRanking bool   `json:"open_search_ranking,omitempty"`
	Text              string `json:"text,omitempty"`
	Link              string `json:"link,omitempty"`
}

type BannerList struct {
	List []*banner.Banner `json:"list,omitempty"`
}

type TrafficConfigOption struct {
	ID   int64  `json:"id,omitempty"`
	Text string `json:"text,omitempty"`
}

type TrafficConfig struct {
	Title           string                 `json:"title,omitempty"`
	Options         []*TrafficConfigOption `json:"options,omitempty"`
	DefaultOptionID int64                  `json:"default_option_id,omitempty"`
}

type SearchEmbedInline struct {
	jsoncard.LargeCoverInline
	TrafficConfig *TrafficConfig `json:"traffic_config,omitempty"`
}

func newSearchEmbedInline(in *jsoncard.LargeCoverInline) *SearchEmbedInline {
	return &SearchEmbedInline{
		LargeCoverInline: *in,
		TrafficConfig:    InlineLiveTrafficConfig(),
	}
}

// TrafficConfigOption ID 的枚举定义参考该文档
// https://info.bilibili.co/pages/viewpage.action?pageId=154629883
func InlineLiveTrafficConfig() *TrafficConfig {
	tc := &TrafficConfig{
		Title: "搜索大卡自动播放",
		Options: []*TrafficConfigOption{
			{
				ID:   10,
				Text: "WIFI/免流/移动网络下自动播放",
			},
			{
				ID:   3,
				Text: "仅WIFI下自动播放",
			},
			{
				ID:   4,
				Text: "关闭自动播放",
			},
		},
		DefaultOptionID: 11,
	}
	return tc
}

func (ci *ChannleItem) FormChannelNew(c *channelgrpc.SearchChannel) {
	ci.ID = c.GetID()
	ci.Title = c.GetName()
	ci.Cover = c.GetIcon()
	ci.Goto = model.GotoChannelNew
	ci.Param = strconv.FormatInt(ci.ID, 10)
	var labels []string
	if c.SubscribedCnt > 0 {
		labels = append(labels, statString(c.SubscribedCnt, "订阅"))
	}
	if c.CType == OfficeChannel {
		ci.URI = model.FillURI(model.GotoChannelNew, ci.Param, model.ChannelHandler("tab=select"))
		ci.TypeIcon = _channelOfficIconPink
		if c.ResourceCnt > 0 {
			labels = append(labels, statString(c.ResourceCnt, "个视频"))
		}
		ci.Right = true
	} else if c.CType == OldChannel {
		ci.URI = model.FillURI(model.GotoChannel, ci.Param, nil)
		ci.Right = true
	}
	if len(labels) > 0 {
		ci.Label = strings.Join(labels, "  ")
	}
	if c.Subscribed {
		ci.IsAtten = 1
	}
}

func fetchChannelText(plat int8, build int64) string {
	text := "订阅"
	if (plat == model.PlatAndroid && build >= 6470000) || (plat == model.PlatIPhone && build >= 64700000) {
		text = "收藏"
	}
	return text
}

func (ci *ChannleItem) FormChannel2(c *channelgrpc.SearchChannelCard, apm map[int64]*api.ArcPlayer, plat int8, build int64, isHightBuild bool, spmid string) {
	text := fetchChannelText(plat, build)
	ci.ID = c.GetCid()
	ci.Title = c.GetCname()
	ci.Icon = c.GetIcon()
	ci.Cover = c.GetBackground()
	ci.Goto = model.GotoChannelNew
	ci.Param = strconv.FormatInt(ci.ID, 10)
	ci.Label = statString(c.GetSubscribedCnt(), text)
	ci.Alpha = c.GetAlpha()
	ci.ThemeColor = c.GetColor()
	ci.ThemeColorNight = c.GetColorNight()
	var labels []string
	if c.GetResourceCnt() > 0 {
		labels = append(labels, statString(c.GetResourceCnt(), "视频"))
	}
	if c.GetFeaturedCnt() > 0 {
		labels = append(labels, statString(c.GetFeaturedCnt(), "精选视频"))
	}
	if len(labels) > 0 {
		ci.Label2 = strings.Join(labels, "  ")
	}
	if c.Subscribed {
		ci.IsAtten = 1
	}
	if isHightBuild && c.GetBizType() == channelgrpc.ChannelBizlType_MOVIE {
		ci.URI = model.FillURI(model.GotoChannelMedia, fmt.Sprintf("?biz_id=%s&biz_type=0&source=%s", ci.Param, spmid), nil)
	} else {
		ci.URI = model.FillURI(model.GotoChannelNew, ci.Param, model.ChannelHandler("tab=select"))
	}
	ci.TypeIcon = _channelOfficIconWhite
	ci.Button = &SearchButton{Text: text}
	ci.More = &SearchButton{Text: "进入频道查看更多", URI: ci.URI}
	for _, video := range c.GetVideoCards() {
		if video == nil {
			continue
		}
		if ap, ok := apm[video.GetRid()]; ok {
			if ap == nil || ap.Arc == nil {
				continue
			}
			a := ap.Arc
			i := &ChannleItem{}
			i.ID = a.Aid
			i.Title = a.Title
			i.Cover = a.Pic
			i.Goto = model.GotoAv
			i.Param = strconv.FormatInt(i.ID, 10)
			playInfo := ap.PlayerInfo[ap.DefaultPlayerCid]
			i.URI = model.FillURI(i.Goto, i.Param, model.AvPlayHandlerGRPC(a, playInfo))
			i.CoverLeftText1 = statString(int64(a.Stat.View), "")
			i.CoverLeftIcon1 = 1
			if video.GetBadgeTitle() != "" && video.GetBadgeBackground() != "" {
				i.Badge = &ChannelBadge{Text: video.GetBadgeTitle(), IconBgURL: video.GetBadgeBackground()}
			}
			ci.Items = append(ci.Items, i)
		}
	}
}

func (ci *ChannleItem) FormChannelMore(c *channelgrpc.RelativeChannel, mobiApp, spmid string, build int64, isHightBuild bool) {
	ci.ID = c.GetCid()
	ci.Title = c.GetCname()
	ci.Icon = c.GetIcon()
	ci.Param = strconv.FormatInt(ci.ID, 10)
	ci.Goto = model.GotoChannelNew
	if isHightBuild && c.GetBizType() == channelgrpc.ChannelBizlType_MOVIE {
		ci.URI = model.FillURI(model.GotoChannelMedia, fmt.Sprintf("?biz_id=%s&biz_type=0&source=%s", ci.Param, spmid), nil)
	} else {
		ci.URI = model.FillURI(model.GotoChannelNew, ci.Param, model.ChannelHandler("tab=select"))
	}
	var labels []string
	if c.GetResourceCnt() > 0 {
		labels = append(labels, statString(c.GetResourceCnt(), "投稿"))
	}
	if c.GetFeaturedCnt() > 0 {
		labels = append(labels, statString(c.GetFeaturedCnt(), "个精选视频"))
	}
	if len(labels) > 0 {
		ci.Label = strings.Join(labels, "  ")
	}
	if c.Subscribed {
		ci.IsAtten = 1
	}
	ci.TypeIcon = _channelOfficIconPink
	ci.Button = &SearchButton{Text: statString(c.GetSubscribedCnt(), "订阅")}
	// 新版本收藏替换订阅
	if card.FavTextReplace(mobiApp, build) {
		ci.Button = &SearchButton{Text: statString(c.GetSubscribedCnt(), "收藏")}
	}
}

func (ci *ChannleItem) FormChannelHot(ctx context.Context, c *channelgrpc.ChannelCard, isHightBuild bool, spmid string) {
	ci.ID = c.GetChannelId()
	ci.Title = c.GetChannelName()
	ci.Icon = c.GetIcon()
	ci.Param = strconv.FormatInt(ci.ID, 10)
	ci.Goto = model.GotoChannelNew
	if isHightBuild && c.GetBizType() == channelgrpc.ChannelBizlType_MOVIE {
		ci.URI = model.FillURI(model.GotoChannelMedia, fmt.Sprintf("?biz_id=%s&biz_type=0&source=%s", ci.Param, spmid), nil)
	} else {
		ci.URI = model.FillURI(model.GotoChannelNew, ci.Param, model.ChannelHandler("tab=select"))
	}
	var labels []string
	if c.GetRCnt() > 0 {
		labels = append(labels, statString(int64(c.GetRCnt()), "投稿"))
	}
	if c.GetFeaturedCnt() > 0 {
		labels = append(labels, statString(int64(c.GetFeaturedCnt()), "个精选视频"))
	}
	if len(labels) > 0 {
		ci.Label = strings.Join(labels, "  ")
	}
	if c.Subscribed {
		ci.IsAtten = 1
	}
	ci.TypeIcon = _channelOfficIconPink
	ci.Button = &SearchButton{Text: statString(int64(c.GetSubscribedCnt()), "订阅")}
	dev, _ := device.FromContext(ctx)
	if card.FavTextReplace(dev.MobiApp(), dev.Build) {
		ci.Button = &SearchButton{Text: statString(int64(c.GetSubscribedCnt()), "收藏")}
	}
}

// nolint:gomnd
func (i *Item) fromOgvCardStyles(info *mediagrpc.MediaBizInfoGetReply) {
	var styles []string
	if info.FirstReleaseDate != 0 {
		if pt := time.Unix(info.FirstReleaseDate, 0).Format("2006"); pt != "" {
			styles = append(styles, pt)
		}
	}
	if info.CategoryDesc != "" {
		switch info.CategoryId {
		case 1, 4:
			// 番剧和国产动画统一显示为动画
			styles = append(styles, "动画")
		default:
			styles = append(styles, info.CategoryDesc)
		}
	}
	if len(info.Areas) > 0 {
		i.Area = info.Areas[0]
		styles = append(styles, i.Area)
	}
	if len(styles) > 0 {
		i.Styles = strings.Join(styles, " | ")
	}
}

func (i *Item) FormNewChannel(ctx context.Context, nc *NewChannel, cs map[int64]*channelgrpc.SearchChannelInHome, apm map[int64]*api.ArcPlayer) {
	c, ok := cs[nc.ID]
	if !ok {
		return
	}
	if c.GetCid() == 0 || c.GetCname() == "" {
		return
	}
	i.ID = c.GetCid()
	i.Param = strconv.FormatInt(i.ID, 10)
	i.Title = c.GetCname()
	i.Cover = c.GetIcon()
	i.Goto = model.GotoChannelNew
	i.URI = model.FillURI(model.GotoChannelNew, i.Param, model.ChannelHandler("tab=select&from=search.search-result.0.0"))
	i.TypeIcon = _channelOfficIconPink
	if c.GetViewCnt() > 0 {
		i.ChannelLabel1 = &SearchButton{
			Text: statString(int64(c.GetViewCnt()), "播放>"),
			URI:  model.FillURI(model.GotoChannelNew, i.Param, model.ChannelHandler("tab=select&from=search.search-result.0.0")),
		}
	}
	if c.GetFeaturedCnt() > 0 {
		i.ChannelLabel2 = &SearchButton{
			Text: statString(int64(c.GetFeaturedCnt()), "精选视频>"),
			URI:  model.FillURI(model.GotoChannelNew, i.Param, model.ChannelHandler("tab=select&from=search.search-result.0.0")),
		}
	}
	i.ChannelButton = &SearchButton{
		Text: "进入频道",
		URI:  model.FillURI(model.GotoChannelNew, i.Param, model.ChannelHandler("tab=select&from=search.search-result.0.0")),
	}
	switch c.GetResourceType() {
	case NewChannelResourceTypeArchive:
		i.DesignType = "archive"
		for _, video := range c.GetVideoCards() {
			if ap, ok := apm[video.GetRid()]; ok {
				if ap.GetArc().GetAid() == 0 {
					continue
				}
				a := ap.Arc
				ii := &Item{
					ID:             a.GetAid(),
					Title:          a.GetTitle(),
					Cover:          a.GetPic(),
					Goto:           model.GotoAv,
					Param:          strconv.FormatInt(a.GetAid(), 10),
					CoverLeftText1: statString(int64(a.Stat.View), ""),
					CoverLeftIcon1: cardmdl.IconPlay,
				}
				playInfo := ap.PlayerInfo[ap.DefaultPlayerCid]
				ii.URI = model.FillURI(ii.Goto, ii.Param, model.AvPlayHandlerGRPC(a, playInfo))
				i.Items = append(i.Items, ii)
			}
		}
		//nolint:gomnd
		if len(i.Items) >= 2 {
			i.Right = true
		}
	case NewChannelResourceTypeChildChannel:
		i.DesignType = "channel"
		for _, child := range c.GetChildren() {
			if child.GetCid() == 0 {
				continue
			}
			ii := &Item{
				ID:    child.GetCid(),
				Title: child.GetCname(),
				Cover: child.GetIcon(),
				Goto:  model.GotoChannelNew,
				Param: strconv.FormatInt(child.GetCid(), 10),
			}
			ii.URI = model.FillURI(ii.Goto, ii.Param, model.ChannelHandler("tab=select&from=search.search-result.0.0"))
			if pd.WithContext(ctx).Where(func(pdContext *pd.PDContext) {
				pdContext.IsPlatAndroid().And().Build(">=", 6750000)
			}).OrWhere(func(pdContext *pd.PDContext) {
				pdContext.IsPlatIPhone().And().Build(">=", 67500000)
			}).MustFinish() && child.BizType == channelgrpc.ChannelBizlType_MOVIE {
				// 新版本电影频道新跳链
				ii.URI = model.FillURI(model.GotoChannelMedia, fmt.Sprintf("?biz_id=%d&biz_type=0&source=search.search-result.0.0", child.Cid), nil)
			}
			i.Items = append(i.Items, ii)
		}
		//nolint:gomnd
		if len(i.Items) >= 3 {
			i.Right = true
		}
	}
}
