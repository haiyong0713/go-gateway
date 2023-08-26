package search

import (
	"encoding/json"

	"go-common/library/log"
	xtime "go-common/library/time"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
)

// search const
const (
	SuggestionJump     = 99
	SuggestionJumpUser = 81
	SuggestionJumpPGC  = 82
	SuggestionAV       = "video"
	SuggestionLive     = "live"
	SuggestionArticle  = "article"

	_MediaCanPlay     = 0
	_MediaIsOutAllNet = 100

	DefaultWordTypeArchive = 1
	DefaultWordTypeArticle = 2
	DefaultWordTypePGC     = 3
	DefaultWordTypeURL     = 4

	OGVCardTypeGame = 1
	OGVCardTypePGC  = 2
	OGVCardTypeMore = 3

	OldChannel    = 1
	OfficeChannel = 2

	MatchStateOver      = 1
	MatchStateAtten     = 3
	MatchStateLive      = 4
	MatchStateBattling  = 5
	MatchStateAttenLive = 6

	NewChannelResourceTypeChildChannel = int32(0)
	NewChannelResourceTypeArchive      = int32(1)
)

// NoResultRcmd no result rcmd
type NoResultRcmd struct {
	Code           int      `json:"code,omitempty"`
	Msg            string   `json:"msg,omitempty"`
	ReqType        string   `json:"req_type,omitempty"`
	Result         []*Video `json:"result,omitempty"`
	NumResults     int      `json:"numResults,omitempty"`
	Page           int      `json:"page,omitempty"`
	Trackid        string   `json:"seid,omitempty"`
	SuggestKeyword string   `json:"suggest_keyword,omitempty"`
	RecommendTips  string   `json:"recommend_tips,omitempty"`
}

// RecommendPre search at pre-page
type RecommendPre struct {
	Code      int    `json:"code,omitempty"`
	Msg       string `json:"msg,omitempty"`
	NumResult int    `json:"numResult,omitempty"`
	Trackid   string `json:"seid,omitempty"`
	Result    []*struct {
		Type  string `json:"type,omitempty"`
		Query string `json:"query,omitempty"`
		List  []*struct {
			Type string `json:"source_type,omitempty"`
			ID   int64  `json:"source_id,omitempty"`
		} `json:"rec_list,omitempty"`
	} `json:"result,omitempty"`
}

// Page struct
type Page struct {
	NumResults int `json:"numResults"`
	Pages      int `json:"pages"`
}

// Bangumi struct
type Bangumi struct {
	Name          string `json:"name,omitempty"`
	SeasonID      int    `json:"season_id,omitempty"`
	Title         string `json:"title,omitempty"`
	Cover         string `json:"cover,omitempty"`
	Evaluate      string `json:"evaluate,omitempty"`
	NewestEpID    int    `json:"newest_ep_id,omitempty"`
	NewestEpIndex string `json:"newest_ep_index,omitempty"`
	IsFinish      int    `json:"is_finish,omitempty"`
	IsStarted     int    `json:"is_started,omitempty"`
	NewestCat     string `json:"newest_cat,omitempty"`
	NewestSeason  string `json:"newest_season,omitempty"`
	TotalCount    int    `json:"total_count,omitempty"`
	Pages         int    `json:"numPages,omitempty"`
	CatList       *struct {
		TV    int `json:"tv"`
		Movie int `json:"movie"`
		Ova   int `json:"ova"`
	} `json:"catlist,omitempty"`
}

// Movie struct
type Movie struct {
	Title      string `json:"title"`
	SpID       string `json:"spid"`
	Type       string `json:"type"`
	Aid        int64  `json:"aid"`
	Desc       string `json:"description"`
	Actors     string `json:"actors"`
	Staff      string `json:"staff"`
	Cover      string `json:"cover"`
	Pic        string `json:"pic"`
	ScreenDate string `json:"screenDate"`
	Area       string `json:"area"`
	Status     int    `json:"status"`
	Length     int    `json:"length"`
	Pages      int    `json:"numPages"`
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
	IsLive       int   `json:"is_live,omitempty"`
	RoomID       int64 `json:"room_id,omitempty"`
	IsUpuser     int   `json:"is_upuser,omitempty"`
	Position     int   `json:"position,omitempty"`
	Version      int   `json:"version,omitempty"`
	IsInlineLive int64 `json:"is_inline_live,omitempty"`
}

// OfficialVerify struct
type OfficialVerify struct {
	Type int    `json:"type"`
	Desc string `json:"desc,omitempty"`
}

// Video struct
type Video struct {
	Type        string      `json:"type"`
	ID          int64       `json:"id"`
	Author      string      `json:"author"`
	Title       string      `json:"title"`
	Pic         string      `json:"pic"`
	Description string      `json:"description"`
	Play        interface{} `json:"play"`
	Danmaku     int         `json:"video_review"`
	Duration    string      `json:"duration"`
	Pages       int         `json:"numPages"`
	ViewType    string      `json:"view_type"`
	RecTags     []string    `json:"rec_tags"`
	PubDate     int64       `json:"pubdate"`
	IsPay       int         `json:"is_pay"`
	NewRecTags  []*RecTag   `json:"new_rec_tags"`
	IsUGCInline int64       `json:"is_ugc_inline"`
	Mid         int64       `json:"mid"`
	ExtraInfo   ExtraInfo   `json:"extra_info,omitempty"`
	Cover       string      `json:"cover,omitempty"`
	RecReason   string      `json:"rec_reason,omitempty"`
	Corner      string      `json:"corner,omitempty"`
	URL         string      `json:"url,omitempty"`
	Desc        string      `json:"desc,omitempty"`
	Fulltext    []*FullText `json:"fulltext,omitempty"`
}

// 全文检索
type FullText struct {
	Type        int    `json:"type"`
	Text        string `json:"text"`
	Abstract    string `json:"abstract"`
	StartSecond int64  `json:"start_second"`
}

// RecTag from video
type RecTag struct {
	Name  string `json:"tag_name"`
	Style int8   `json:"tag_style"`
}

type WikiExtraInfo struct {
	CornerType     int64  `json:"corner_type"`
	CornerText     string `json:"corner_text"`
	CornerSunURL   string `json:"corner_sun_url"`
	CornerNightURL string `json:"corner_night_url"`
	CornerHeight   int64  `json:"corner_height"`
	CornerWidth    int64  `json:"corner_width"`
}

type ExtraInfo struct {
	Title   string         `json:"title"`
	ImgURL  string         `json:"img_url"`
	ReType  int64          `json:"re_type"`
	ReValue string         `json:"re_value"`
	Wiki    *WikiExtraInfo `json:"wiki"`
}

var _searchInlineReType = map[int64]appcardmodel.Gt{
	1: appcardmodel.GotoWeb,
	2: appcardmodel.GotoAv,
	3: appcardmodel.GotoPGC,
	4: appcardmodel.GotoBangumi,
	5: appcardmodel.GotoLive,
	6: appcardmodel.GotoArticle,
}

func genGotoURI(reType int64, reValue string) (string, bool) {
	typeGoto, ok := _searchInlineReType[reType]
	if !ok {
		return "", false
	}
	uri := appcardmodel.FillURI(typeGoto, 0, 0, reValue, nil)
	return uri, true
}

func (i ExtraInfo) GotoURI() (string, bool) {
	return genGotoURI(i.ReType, i.ReValue)
}

// Live struct
type Live struct {
	Total            int       `json:"total,omitempty"`
	Pages            int       `json:"pages"`
	UID              int64     `json:"uid,omitempty"`
	RoomID           int64     `json:"roomid,omitempty"`
	Type             string    `json:"type,omitempty"`
	Title            string    `json:"title,omitempty"`
	LiveStatus       int       `json:"live_status,omitempty"`
	ShortID          int       `json:"short_id,omitempty"`
	Uname            string    `json:"uname,omitempty"`
	Uface            string    `json:"uface,omitempty"`
	Cover            string    `json:"cover,omitempty"`
	Online           int       `json:"online,omitempty"`
	Attentions       int       `json:"attentions,omitempty"`
	Tags             string    `json:"tags,omitempty"`
	Area             int       `json:"area,omitempty"`
	CateName         string    `json:"cate_name,omitempty"`
	CateParentName   string    `json:"cate_parent_name,omitempty"`
	UserCover        string    `json:"user_cover,omitempty"`
	VerifyType       int       `json:"verify_type,omitempty"`
	VerifyDesc       string    `json:"verify_desc,omitempty"`
	Fans             int       `json:"fans,omitempty"`
	IsLiveRoomInline int64     `json:"is_live_room_inline,omitempty"`
	ExtraInfo        ExtraInfo `json:"extra_info,omitempty"`
}

// Article struct
type Article struct {
	ID         int64    `json:"id"`
	Mid        int64    `json:"mid"`
	Uname      string   `json:"uname"`
	TemplateID int      `json:"template_id"`
	Title      string   `json:"title"`
	Desc       string   `json:"desc"`
	ImageUrls  []string `json:"image_urls"`
	View       int      `json:"view"`
	Like       int      `json:"like"`
	Reply      int      `json:"reply"`
}

// Media struct
type Media struct {
	Type       string `json:"type,omitempty"`
	MediaID    int64  `json:"media_id,omitempty"`
	SeasonID   int64  `json:"season_id,omitempty"`
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
	ExtraInfo   ExtraInfo       `json:"extra_info,omitempty"`
	IsOGVInline int64           `json:"is_ogv_inline,omitempty"`
	EPID        int32           `json:"epid,omitempty"`
	EpClipStart int64           `json:"ep_clip_start,omitempty"`
	EpClipEnd   int64           `json:"ep_clip_end,omitempty"`
}

// Canplay returns whether the bangumi can play or not
func (m *Media) Canplay() bool {
	return m.PlayState == _MediaCanPlay
}

// IsAllNet tells whether the media is all net
func (m *Media) IsAllNet() bool {
	return m.MediaType >= _MediaIsOutAllNet
}

// Query struct
type Query struct {
	Type       string `json:"type,omitempty"`
	Name       string `json:"name,omitempty"`
	ID         int64  `json:"id,omitempty"`
	FromSource string `json:"from_source,omitempty"`
}

// Hot struct
type Hot struct {
	Code    int    `json:"code,omitempty"`
	SeID    string `json:"seid,omitempty"`
	TrackID string `json:"trackid"`
	List    []*struct {
		Keyword      string        `json:"keyword"`
		Status       string        `json:"status"`
		NameType     string        `json:"name_type"`
		ShowName     string        `json:"show_name,omitempty"`
		WordType     int           `json:"word_type,omitempty"`
		Icon         string        `json:"icon,omitempty"`
		GotoType     int           `json:"goto_type,omitempty"`
		GotoValue    string        `json:"goto_value,omitempty"`
		Goto         string        `json:"goto,omitempty"`
		URI          string        `json:"uri,omitempty"`
		Param        string        `json:"param,omitempty"`
		Pos          int           `json:"pos,omitempty"`
		Position     int           `json:"position,omitempty"`
		ID           int64         `json:"id,omitempty"`
		ModuleID     int64         `json:"module_id,omitempty"`
		ResourceID   int64         `json:"resource_id,omitempty"`
		LiveIds      []int64       `json:"live_id,omitempty"`
		ShowLiveIcon bool          `json:"show_live_icon,omitempty"`
		HeatValue    int64         `json:"heat_value,omitempty"`
		HotId        int64         `json:"hot_id,omitempty"`
		Res          []*HotListRes `json:"res,omitempty"`
	} `json:"list"`
	ExpStr string `json:"exp_str,omitempty"`
}

type HotListRes struct {
	Id       int64  `json:"id,omitempty"`
	CardType string `json:"card_type,omitempty"`
}

// Suggest struct
type Suggest struct {
	Code     int         `json:"code"`
	Stoken   string      `json:"stoken"`
	ResultBs interface{} `json:"result"`
	Result   struct {
		Accurate struct {
			UpUser  interface{} `json:"upuser,omitempty"`
			Bangumi interface{} `json:"bangumi,omitempty"`
		} `json:"accurate,omitempty"`
		Tag []*struct {
			Value string `json:"value,omitempty"`
		} `json:"tag,omitempty"`
	} `json:"-"`
}

// Suggest2 struct
type Suggest2 struct {
	Code   int    `json:"code"`
	Stoken string `json:"stoken"`
	Result *struct {
		Tag []*SuggestTag `json:"tag"`
	} `json:"result"`
}

// SuggestTag struct
type SuggestTag struct {
	Value string `json:"value,omitempty"`
	Ref   int64  `json:"ref,omitempty"`
	Name  string `json:"name,omitempty"`
	SpID  int    `json:"spid,omitempty"`
	Type  string `json:"type,omitempty"`
}

// Suggest3 struct
type Suggest3 struct {
	Code    int    `json:"code"`
	TrackID string `json:"trackid"`
	ExpStr  string `json:"exp_str"`
	Result  []*Sug `json:"result"`
}

// Sug struct
type Sug struct {
	ShowName  string          `json:"show_name,omitempty"`
	Term      string          `json:"term,omitempty"`
	Ref       int64           `json:"ref,omitempty"`
	TermType  int             `json:"term_type,omitempty"`
	SubType   string          `json:"sub_type,omitempty"`
	Pos       int             `json:"pos,omitempty"`
	Cover     string          `json:"cover,omitempty"`
	CoverSize float64         `json:"cover_size,omitempty"`
	Value     json.RawMessage `json:"value,omitempty"`
	PGC       *SugPGC         `json:"-"`
	User      *SugUser        `json:"user,omitempty"`
}

// SugPGC fro sug
type SugPGC struct {
	MediaID        int64                `json:"media_id,omitempty"`
	SeasonID       int64                `json:"season_id,omitempty"`
	Title          string               `json:"title,omitempty"`
	MediaType      int                  `json:"media_type,omitempty"`
	GotoURL        string               `json:"goto_url,omitempty"`
	Areas          string               `json:"areas,omitempty"`
	Pubtime        xtime.Time           `json:"pubtime,omitempty"`
	FixPubTime     string               `json:"fix_pubtime_str,omitempty"`
	Styles         string               `json:"styles,omitempty"`
	CV             string               `json:"cv,omitempty"`
	Staff          string               `json:"staff,omitempty"`
	MediaScore     float64              `json:"media_score,omitempty"`
	MediaUserCount int                  `json:"media_user_cnt,omitempty"`
	Cover          string               `json:"cover,omitempty"`
	Badges         []*model.ReasonStyle `json:"badges,omitempty"`
}

// SugUser fro sug
type SugUser struct {
	Mid                int64  `json:"uid,omitempty"`
	Face               string `json:"face,omitempty"`
	Name               string `json:"uname,omitempty"`
	Fans               int    `json:"fans,omitempty"`
	Videos             int    `json:"videos,omitempty"`
	Level              int    `json:"level,omitempty"`
	OfficialVerifyType int    `json:"verify_type,omitempty"`
	FaceNftNew         int32  `json:"face_nft_new,omitempty"`
	IsSeniorMember     int32  `json:"is_senior_member,omitempty"`
}

// Operate struct
type Operate struct {
	ID          int64  `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Cover       string `json:"cover,omitempty"`
	RedirectURL string `json:"redirect_url,omitempty"`
	Desc        string `json:"desc,omitempty"`
	Corner      string `json:"corner,omitempty"`
	RecReason   string `json:"rec_reason,omitempty"`
	CardType    string `json:"card_type,omitempty"`
	ContentList []*struct {
		Type int   `json:"type,omitempty"`
		ID   int64 `json:"id,omitempty"`
	} `json:"content_list,omitempty"`
	ExtraInfo  ExtraInfo `json:"extra_info,omitempty"`
	BtnReType  int64     `json:"btn_re_type"`
	BtnReValue string    `json:"btn_re_value"`
}

// Game struct
type Game struct {
	ID                 int64   `json:"id,omitempty"`
	Title              string  `json:"title,omitempty"`
	Cover              string  `json:"cover,omitempty"`
	Desc               string  `json:"description,omitempty"`
	View               float64 `json:"view,omitempty"`
	Like               int64   `json:"like,omitempty"`
	Status             int     `json:"status,omitempty"`
	RedirectURL        string  `json:"redirect_url,omitempty"`
	Tag                string  `json:"tag,omitempty"`
	NoticeName         string  `json:"notice_name,omitempty"`
	NoticeContent      string  `json:"notice_content,omitempty"`
	GiftContentAndroid string  `json:"gift_content_android,omitempty"`
	GiftURLAndroid     string  `json:"gift_url_android,omitempty"`
	GiftContentIOS     string  `json:"gift_content_ios,omitempty"`
	GiftURLIOS         string  `json:"gift_url_ios,omitempty"`
}

// Comic struct
type Comic struct {
	ID        int64    `json:"id,omitempty"`
	Title     string   `json:"title,omitempty"`
	Author    []string `json:"author,omitempty"`
	Cover     string   `json:"cover,omitempty"`
	Styles    string   `json:"styles,omitempty"`
	URL       string   `json:"url,omitempty"`
	ComicURL  string   `json:"sq_url,omitempty"`
	ComicType int64    `json:"comic_type,omitempty"`
}

// Twitter twitter.
type Twitter struct {
	ID         int64    `json:"id,omitempty"`
	PicID      int64    `json:"pic_id"`
	Cover      []string `json:"cover,omitempty"`
	CoverCount int      `json:"cover_count,omitempty"`
	Content    string   `json:"content,omitempty"`
}

// Star struct
type Star struct {
	ID      int64  `json:"id,omitempty"`
	Cover   string `json:"cover,omitempty"`
	Desc    string `json:"desc,omitempty"`
	Title   string `json:"title,omitempty"`
	MID     int64  `json:"mid,omitempty"`
	TagID   int64  `json:"tag_id,omitempty"`
	TagList []*struct {
		TagName   string `json:"tagname,omitempty"`
		KeyWord   string `json:"searchtagname,omitempty"`
		ValueList []*struct {
			Type  string `json:"type,omitempty"`
			Video *Video `json:"values,omitempty"`
		} `json:"value_list,omitempty"`
	} `json:"tag_list,omitempty"`
}

// Ticket for search.
type Ticket struct {
	ID        int64  `json:"id,omitempty"`
	Title     string `json:"project_name,omitempty"`
	Cover     string `json:"cover,omitempty"`
	ShowTime  string `json:"show_time,omitempty"`
	CityName  string `json:"city_name,omitempty"`
	VenueName string `json:"venue_name,omitempty"`
	PriceLow  int    `json:"price_low,omitempty"`
	PriceType int    `json:"need_up,omitempty"`
	ReqNum    int    `json:"required_number,omitempty"`
	URL       string `json:"url,omitempty"`
}

// Product for search.
type Product struct {
	ID        int64  `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	Cover     string `json:"cover,omitempty"`
	ShopName  string `json:"shop_name,omitempty"`
	Price     int    `json:"price,omitempty"`
	PriceType int    `json:"need_up,omitempty"`
	ReqNum    int    `json:"required_number,omitempty"`
	URL       string `json:"url,omitempty"`
}

// SpecialerGuide for search
type SpecialerGuide struct {
	ID    int64  `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	Desc  string `json:"desc,omitempty"`
	Cover string `json:"cover,omitempty"`
	Tel   string `json:"tel,omitempty"`
}

// Converge for search
type Converge struct {
	Code   int    `json:"code"`
	SeID   string `json:"seid"`
	Total  int    `json:"numResults"`
	Pages  int    `json:"numPages"`
	ExpStr string `json:"exp_str"`
	Result struct {
		User  []*ConvergeUser  `json:"user_infos,omitempty"`
		Video []*ConvergeVideo `json:"video_infos,omitempty"`
	} `json:"result,omitempty"`
}

// ConvergeUser for search
type ConvergeUser struct {
	CardType   string `json:"type,omitempty"`
	Mid        int64  `json:"mid,omitempty"`
	Name       string `json:"uname,omitempty"`
	Face       string `json:"face,omitempty"`
	Fans       int    `json:"fans,omitempty"`
	Videos     int    `json:"videos,omitempty"`
	OfficeType int    `json:"office_type,omitempty"`
}

// ConvergeVideo for search
type ConvergeVideo struct {
	CardType string `json:"type,omitempty"`
	Aid      int64  `json:"aid,omitempty"`
	Mid      int64  `json:"mid,omitempty"`
	Title    string `json:"title,omitempty"`
	Cover    string `json:"cover,omitempty"`
	Play     int    `json:"play,omitempty"`
	Danmaku  int    `json:"dm,omitempty"`
	Duration string `json:"duration,omitempty"`
}

// Space for space.
type Space struct {
	Code    int    `json:"code"`
	Trackid string `json:"seid"`
	Total   int    `json:"total"`
	Page    int    `json:"page"`
	Result  *struct {
		VList []*SpaceValue `json:"vlist"`
	} `json:"result"`
}

// SpaceValue for space search
type SpaceValue struct {
	Play     interface{} `json:"play,omitempty"`
	Danmaku  int         `json:"video_review,omitempty"`
	Created  string      `json:"created,omitempty"`
	Title    string      `json:"title,omitempty"`
	Aid      int64       `json:"aid,omitempty"`
	Pic      string      `json:"pic,omitempty"`
	Duration string      `json:"length,omitempty"`
}

// SearchOGVCard for ogvcard search
type SearchOGVCard struct {
	ID             int64  `json:"id,omitempty"`
	Type           string `json:"type,omitempty"`
	SpecialBgColor string `json:"special_bg_color,omitempty"`
	HeadArea       struct {
		Cover    string `json:"cover,omitempty"`
		BgCover  string `json:"bg_cover,omitempty"`
		Title    string `json:"title,omitempty"`
		SubTitle string `json:"sub_title,omitempty"`
	} `json:"head_area,omitempty"`
	Modules []*SearchModules `json:"modules,omitempty"`
}

// SearchModules search modules
type SearchModules struct {
	Pos      int                   `json:"pos,omitempty"`
	Title    string                `json:"title,omitempty"`
	Type     int                   `json:"type,omitempty"`
	LinkType string                `json:"linktype,omitempty"`
	Values   []*SearchOGVCardItems `json:"values,omitempty"`
}

// SearchOGVCardItems items
type SearchOGVCardItems struct {
	// game card
	*Game
	// season
	SeasonIDList []int64 `json:"season_id_list,omitempty"`
	MoreURL      string  `json:"more_url,omitempty"`
	// more card
	ShowName string `json:"show_name,omitempty"`
	Type     int    `json:"type,omitempty"`
	Value    string `json:"value,omitempty"`
}

type ESport struct {
	ID        int64        `json:"id,omitempty"`
	Title     string       `json:"title,omitempty"`
	UrlTop    string       `json:"url_top,omitempty"`
	UrlBottom string       `json:"url_bottom,omitempty"`
	MatchList []*MatchList `json:"match_list,omitempty"`
}

type Sports struct {
	ID       int64  `json:"id,omitempty"`
	SeasonId int64  `json:"season_id,omitempty"`
	Title    string `json:"title,omitempty"`
	Url      string `json:"url,omitempty"`
}

type CollectionCard struct {
	ID         int64  `json:"id,omitempty"`
	Uid        int64  `json:"uid,omitempty"`
	Author     string `json:"author,omitempty"`
	Play       int64  `json:"play,omitempty"`
	SubNum     int64  `json:"sub_num,omitempty"`
	Title      string `json:"title,omitempty"`
	RecallType int64  `json:"recall_type,omitempty"`
	CateTitle  string `json:"cate_title,omitempty"`
	CateId     int64  `json:"cate_id,omitempty"`
}

type MatchList struct {
	ID          int64 `json:"id,omitempty"`
	HomeTeamID  int64 `json:"home_team_id,omitempty"`
	GuestTeamID int64 `json:"guest_team_id,omitempty"`
}

// SugChange chagne sug value
func (s *Sug) SugChange() {
	var err error
	switch s.TermType {
	case SuggestionJumpPGC:
		err = json.Unmarshal(s.Value, &s.PGC)
	case SuggestionJumpUser:
		err = json.Unmarshal(s.Value, &s.User)
	}
	if err != nil {
		log.Error("SugChange json.Unmarshal(%s) error(%+v)", s.Value, err)
	}
}

// NewChannel struct
type NewChannel struct {
	ID         int64 `json:"id,omitempty"`
	RankOffset int   `json:"rank_offset,omitempty"`
	RankIndex  int   `json:"rank_index,omitempty"`
}

// Tips struct
type Tips struct {
	ID int64 `json:"id,omitempty"`
}

type TopGame struct {
	ID     int64 `json:"id,omitempty"`
	CardId int64 `json:"card_id,omitempty"`
}
