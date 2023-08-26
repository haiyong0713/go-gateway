package model

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
)

const (
	// PlatAndroid is int8 for android.
	PlatAndroid = int8(0)
	// PlatIPhone is int8 for iphone.
	PlatIPhone = int8(1)
	// PlatIPad is int8 for ipad.
	PlatIPad = int8(2)
	// PlatWPhone is int8 for wphone.
	PlatWPhone = int8(3)
	// PlatAndroidG is int8 for Android Googleplay.
	PlatAndroidG = int8(4)
	// PlatIPhoneI is int8 for Iphone Global.
	PlatIPhoneI = int8(5)
	// PlatIPadI is int8 for IPAD Global.
	PlatIPadI = int8(6)
	// PlatAndroidTV is int8 for AndroidTV Global.
	PlatAndroidTV = int8(7)
	// PlatAndroidI is int8 for Android Global.
	PlatAndroidI = int8(8)
	// PlatIpadHD is int8 for IpadHD
	PlatIpadHD = int8(9)
	// PlatAndroidB is int8 for Android Blue.
	PlatAndroidB = int8(10)
	// PlatIPhoneB is int8 for Android Blue.
	PlatIPhoneB = int8(11)
	// PlatAndroidHD is int8 for android_hd
	PlatAndroidHD = int8(90)

	GotoAv              = "av"
	GotoWeb             = "web"
	GotoBangumi         = "bangumi"
	GotoMovie           = "movie"
	GotoBangumiWeb      = "bangumi_web"
	GotoSp              = "sp"
	GotoLive            = "live"
	GotoGame            = "game"
	GotoAuthor          = "author"
	GotoAuthorNew       = "author_new"
	GotoClip            = "clip"
	GotoAlbum           = "album"
	GotoArticle         = "article"
	GotoConverge        = "converge"
	GOtoRecommendWord   = "recommend_word"
	GotoPGC             = "pgc"
	GotoComic           = "comic"
	GotoChannel         = "channel"
	GotoEP              = "ep"
	GotoTwitter         = "twitter"
	GotoStar            = "star"
	GotoTicket          = "ticket"
	GotoProduct         = "product"
	GotoSpace           = "space"
	GotoSpecialerGuide  = "special_guide"
	GotoDynamic         = "dynamic"
	GotoConvergeContent = "converge_content"
	GotoOGVCard         = "ogv_card"
	GotoBangumiRelates  = "bangumi_relates"
	GotoFindMore        = "find_more"
	GotoSearch          = "search"
	GotoNewGame         = "new_game"
	GotoChannelNew      = "channel_new"
	GotoChannelMedia    = "ogv_channel"
	GotoESports         = "esport"
	GotoLiveWeb         = "live_web"
	GotoFullScreen      = "fullscreen_video"
	GotoTips            = "tips"
	GotoGameAD          = "game_ad"
	GotoOGVInline       = "ogv_inline"
	GotoBrandAdAv       = "brand_ad_av"
	GotoBrandAdLocalAv  = "brand_ad_local_av"
	GotoBrandAdLive     = "brand_ad_live"
	GotoSportsVersus    = "sports_versus"
	GotoSports          = "sports"
	GotoRecommendTips   = "recommend_tips"
	GotoCollectionCard  = "collection_card"
	GotoComicCard       = "comic_card"

	// ForbidCode is forbid by law
	ForbidCode   = -110
	NoResultCode = -111

	CoverIng      = "即将上映"
	CoverPay      = "付费观看"
	CoverFree     = "免费观看"
	CoverVipFree  = "付费观看"
	CoverVipOnly  = "专享"
	CoverVipFirst = "抢先"

	// badge type
	BgStyleFill   = int8(1)
	BgStyleStroke = int8(2)

	// 角标颜色
	BgColorRed    = int8(1)
	BgColorYellow = int8(2)
	BgColorBlue   = int8(3)

	// 播放历史 设备类型
	// DeviceIphone iphoneTV
	DeviceIphone = int8(1)
	// DevicePC PC
	DevicePC = int8(2)
	// DeviceAndroid android
	DeviceAndroid = int8(3)
	// DeviceAndroidTV android TV
	DeviceAndroidTV = int8(33)
	// DeviceIpad ipad
	DeviceIpad = int8(4)
	// DeviceWP8 WP8
	DeviceWP8 = int8(5)
	// DeviceUWP UWP
	DeviceUWP = int8(6)
	// 车载
	DeviceCar = int8(8)
	// 物联网设备
	DeviceIoT = int8(9)
	// 安卓pad
	DeviceAndPad = int8(10)

	// 我的页模块老对应关系
	SelfCenter           = 6
	MyService            = 7
	Creative             = 11
	IPadSelfCenter       = 12
	IPadCreative         = 13
	AndroidSelfCenter    = 14
	AndroidCreative      = 15
	AndroidMyService     = 16
	OpModule             = 17
	AndroidBSelfCenter   = 24
	AndroidBCreative     = 25
	AndroidBMyService    = 26
	AndroidISelfCenter   = 27
	AndroidIMyService    = 29
	IPhoneBselfCenter    = 19
	IPhoneBmyService     = 20
	IPhoneBcreative      = 21
	IPadHDSelfCenter     = 48
	IPadHDCreative       = 47
	AndroidPadSelfCenter = 49
	// 创作中心和直播中心为接入外部业务方
	AndroidLive = 31
	IPhoneLive  = 30

	// 	live entry
	DefaultLiveEntry     = "NONE"
	BrandADLiveEntry     = "search_commerce_card"
	SearchInlineCard     = "search_inline_card"
	SearchLiveInlineCard = "search_live_inline_card" // 搜索直播用单列卡的 inline 卡
	SearchEsInlineCard   = "search_es_inline_card"
	HotSearchLiveCard    = "hot_search_list_live_card"
)

// for FillURI
var (
	AvHandler = func(a *arcgrpc.Arc) func(uri string) string {
		return func(uri string) string {
			if a == nil {
				return uri
			}
			if a.Dimension.Height != 0 || a.Dimension.Width != 0 {
				return fmt.Sprintf("%s?player_width=%d&player_height=%d&player_rotate=%d", uri, a.Dimension.Width, a.Dimension.Height, a.Dimension.Rotate)
			}
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("ParamHandler url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("ParamHandler url.ParseQuery error(%v)", err)
				return uri
			}
			// 拜年祭活动合集
			if a.AttrValV2(arcgrpc.AttrBitV2ActSeason) == arcgrpc.AttrYes && a.SeasonTheme != nil {
				params.Set("is_festival", "1")
				params.Set("bg_color", a.SeasonTheme.BgColor)
				params.Set("selected_bg_color", a.SeasonTheme.SelectedBgColor)
				params.Set("text_color", a.SeasonTheme.TextColor)
			}
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}
	LiveHandler = func(l *livexroom.Infos) func(uri string) string {
		return func(uri string) string {
			if l == nil || l.Status == nil {
				return uri
			}
			if l.Status.LiveScreenType == 0 || l.Status.LiveScreenType == 1 {
				return fmt.Sprintf("%s?broadcast_type=%d", uri, l.Status.LiveScreenType)
			}
			return uri
		}
	}
	LiveEntryHandler = func(l *livexroomgate.EntryRoomInfoResp_EntryList, entryFrom string) func(uri string) string {
		return func(uri string) string {
			if l == nil {
				return uri
			}
			if entryFrom != "" {
				entryURI, ok := l.JumpUrl[entryFrom]
				if ok {
					return entryURI
				}
			}
			if l.LiveScreenType == 0 || l.LiveScreenType == 1 {
				return fmt.Sprintf("%s?broadcast_type=%d", uri, l.LiveScreenType)
			}
			return uri
		}
	}
	AvPlayHandlerGRPC = func(a *arcgrpc.Arc, playInfo *arcgrpc.PlayerInfo) func(uri string) string {
		var player string
		if playInfo.GetPlayurl() != nil {
			bs, _ := json.Marshal(playInfo.GetPlayurl())
			player = string(bs)
		}
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				log.Error("AvPlayHandlerGRPC url.Parse error(%v)", err)
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				log.Error("AvPlayHandlerGRPC url.ParseQuery error(%v)", err)
				return uri
			}
			if player != "" {
				params.Set("cid", strconv.FormatInt(int64(playInfo.GetPlayurl().GetCid()), 10))
				params.Set("player_preload", player)
			}
			if playInfo.GetPlayerExtra() != nil {
				params.Set("history_progress", strconv.FormatInt(playInfo.GetPlayerExtra().GetProgress(), 10))
				if playInfo.GetPlayerExtra().GetDimension().GetHeight() != 0 || playInfo.GetPlayerExtra().GetDimension().GetWidth() != 0 {
					params.Set("player_width", strconv.FormatInt(playInfo.GetPlayerExtra().GetDimension().GetWidth(), 10))
					params.Set("player_height", strconv.FormatInt(playInfo.GetPlayerExtra().GetDimension().GetHeight(), 10))
					params.Set("player_rotate", strconv.FormatInt(playInfo.GetPlayerExtra().GetDimension().GetRotate(), 10))
				}
			}
			// 拜年祭活动合集
			if a.AttrValV2(arcgrpc.AttrBitV2ActSeason) == arcgrpc.AttrYes && a.SeasonTheme != nil {
				params.Set("is_festival", "1")
				params.Set("bg_color", a.SeasonTheme.BgColor)
				params.Set("selected_bg_color", a.SeasonTheme.SelectedBgColor)
				params.Set("text_color", a.SeasonTheme.TextColor)
			}
			paramStr := params.Encode()
			// 重新encode的时候空格变成了+号问题修复
			if strings.IndexByte(paramStr, '+') > -1 {
				paramStr = strings.Replace(paramStr, "+", "%20", -1)
			}
			u.RawQuery = paramStr
			return u.String()
		}
	}
	ChannelHandler = func(tab string) func(uri string) string {
		return func(uri string) string {
			return fmt.Sprintf("%s?%s", uri, tab)
		}
	}
	ArcLayerHandler = func(layer int64) func(uri string) string {
		return func(uri string) string {
			return fmt.Sprintf("%s?auto_float_layer=%d", uri, layer)
		}
	}
	// 付费角标
	PayBadge = &ReasonStyle{
		Text:             "付费",
		TextColor:        "#FFFFFFFF",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FAAB4B",
		BgColorNight:     "#BA833F",
		BorderColor:      "#FAAB4B",
		BorderColorNight: "#BA833F",
		BgStyle:          BgStyleFill,
	}
	// 合作角标
	CooperationBadge = &ReasonStyle{
		Text:             "合作",
		TextColor:        "#FFFFFFFF",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FB7299",
		BgColorNight:     "#BB5B76",
		BorderColor:      "#FB7299",
		BorderColorNight: "#BB5B76",
		BgStyle:          BgStyleFill,
	}
	// 付费实验新角标
	PayBadgeV2 = &ReasonStyle{
		Text:             "付费",
		TextColor:        "#FFB027",
		TextColorNight:   "#DB8700",
		BgColor:          "",
		BgColorNight:     "",
		BorderColor:      "#FFC65D",
		BorderColorNight: "#AD6800",
		BgStyle:          BgStyleStroke,
	}
	// 合作实验新角标
	CooperationBadgeV2 = &ReasonStyle{
		Text:             "合作",
		TextColor:        "#FF6699",
		TextColorNight:   "#D44E7D",
		BgColor:          "",
		BgColorNight:     "",
		BorderColor:      "#FF8CB0",
		BorderColorNight: "#A73E65",
		BgStyle:          BgStyleStroke,
	}
	// 黄色空心标
	YellowBadgeV2 = &ReasonStyle{
		Text:             "",
		TextColor:        "#FFB027",
		TextColorNight:   "#DB8700",
		BgColor:          "",
		BgColorNight:     "",
		BorderColor:      "#FFC65D",
		BorderColorNight: "#AD6800",
		BgStyle:          BgStyleStroke,
	}
	// 蓝色空心标
	BlueBadgeV2 = &ReasonStyle{
		Text:             "",
		TextColor:        "#00AEEC",
		TextColorNight:   "#0087BD",
		BgColor:          "",
		BgColorNight:     "",
		BorderColor:      "#40C5F1",
		BorderColorNight: "#006996",
		BgStyle:          BgStyleStroke,
	}
	// 粉色空心标
	PinkBadgeV2 = &ReasonStyle{
		Text:             "",
		TextColor:        "#FF6699",
		TextColorNight:   "#D44E7D",
		BgColor:          "",
		BgColorNight:     "",
		BorderColor:      "#FF8CB0",
		BorderColorNight: "#A73E65",
		BgStyle:          BgStyleStroke,
	}
)

// IsAndroid check plat is android or ipad.
func IsAndroid(plat int8) bool {
	return plat == PlatAndroid || plat == PlatAndroidG || plat == PlatAndroidI || plat == PlatAndroidB
}

// IsIOS check plat is iphone or ipad.
func IsIOS(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone || plat == PlatIPadI || plat == PlatIPhoneI
}

// IsIPhone check plat is iphone.
func IsIPhone(plat int8) bool {
	return plat == PlatIPhone || plat == PlatIPhoneI
}

// IsIPad check plat is pad.
func IsIPad(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPadI || plat == PlatIpadHD
}

// IsPad check plat is pad.
func IsPad(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPadI || plat == PlatIpadHD || plat == PlatAndroidHD
}

// IsIPhoneB check plat is iphone_b.
func IsIPhoneB(plat int8) bool {
	return plat == PlatIPhoneB
}

// Plat return plat by platStr or mobiApp
func Plat(mobiApp, device string) int8 {
	switch mobiApp {
	case "iphone":
		if device == "pad" {
			return PlatIPad
		}
		return PlatIPhone
	case "white":
		return PlatIPhone
	case "ipad":
		return PlatIpadHD
	case "android":
		return PlatAndroid
	case "win", "winphone":
		return PlatWPhone
	case "android_G":
		return PlatAndroidG
	case "android_i":
		return PlatAndroidI
	case "iphone_i":
		if device == "pad" {
			return PlatIPadI
		}
		return PlatIPhoneI
	case "ipad_i":
		return PlatIPadI
	case "android_tv":
		return PlatAndroidTV
	case "android_b":
		return PlatAndroidB
	case "iphone_b":
		return PlatIPhoneB
	case "android_hd":
		return PlatAndroidHD
	}
	return PlatIPhone
}

// FillURI deal app schema.
func FillURI(gt, param string, f func(uri string) string) (uri string) {
	switch gt {
	case GotoAv, "":
		uri = "bilibili://video/" + param
	case GotoLive:
		uri = "bilibili://live/" + param
	case GotoBangumi:
		uri = "bilibili://bangumi/season/" + param
	case GotoBangumiWeb:
		uri = "http://bangumi.bilibili.com/anime/" + param
	case GotoGame:
		uri = "bilibili://game_center/detail?id=" + param + "&sourceType=adPut"
	case GotoSp:
		uri = "bilibili://splist/" + param
	case GotoAuthor:
		uri = "bilibili://author/" + param
	case GotoClip:
		uri = "bilibili://clip/" + param
	case GotoAlbum:
		uri = "bilibili://album/" + param
	case GotoArticle:
		uri = "bilibili://article/" + param
	case GotoWeb:
		uri = param
	case GotoPGC:
		uri = "https://www.bilibili.com/bangumi/play/ss" + param
	case GotoChannel:
		uri = "bilibili://pegasus/channel/" + param
	case GotoEP:
		uri = "https://www.bilibili.com/bangumi/play/ep" + param
	case GotoTwitter:
		uri = "bilibili://pictureshow/detail/" + param
	case GotoSpace, GotoAuthorNew:
		uri = "bilibili://space/" + param
	case GotoDynamic:
		uri = "bilibili://following/detail/" + param
	case GotoConvergeContent:
		uri = "bilibili://search/converge/" + param
	case GotoSearch:
		uri = "bilibili://search/?keyword=" + url.QueryEscape(param)
	case GotoChannelNew:
		uri = "bilibili://pegasus/channel/v2/" + param
	case GotoChannelMedia:
		uri = "bilibili://feed/channel" + param
	case GotoLiveWeb:
		uri = "https://live.bilibili.com/" + param
	case GotoComic:
		uri = "https://manga.bilibili.com/m/detail/mc" + param
	case GotoFullScreen:
		uri = "bilibili://video/fullscreen/" + param
	}
	if f != nil {
		uri = f(uri)
	}
	return
}

// StatusMark cover status mark
func StatusMark(status int) string {
	//nolint:gomnd
	if status == 0 {
		return CoverIng
	} else if status == 1 {
		return CoverPay
		//nolint:gomnd
	} else if status == 2 {
		return CoverFree
		//nolint:gomnd
	} else if status == 3 {
		return CoverVipFree
		//nolint:gomnd
	} else if status == 4 {
		return CoverVipOnly
		//nolint:gomnd
	} else if status == 5 {
		return CoverVipFirst
	}
	return ""
}

// Direction define
type Direction int

// app-interface const
const (
	Upward   Direction = 1
	Downward Direction = 2
)

// Cursor struct
type Cursor struct {
	Current   int64
	Direction Direction
	Size      int
}

// Latest judge cursor Current
func (c *Cursor) Latest() bool {
	return c.Current == 0
}

// MoveUpward judge cursor Direction
func (c *Cursor) MoveUpward() bool {
	return c.Direction == Upward
}

// MoveDownward judge cursor Direction
func (c *Cursor) MoveDownward() bool {
	return c.Direction == Downward
}

// FormMediaType media type
func FormMediaType(mediaType int) (mediaName string) {
	//nolint:gomnd
	switch mediaType {
	case 1:
		mediaName = "番剧"
	case 2:
		mediaName = "电影"
	case 3:
		mediaName = "纪录片"
	case 4:
		mediaName = "国创"
	case 5:
		mediaName = "电视剧"
	case 6:
		mediaName = "漫画"
	case 7:
		mediaName = "综艺"
	case 123:
		mediaName = "电视剧"
	case 124:
		mediaName = "综艺"
	case 125:
		mediaName = "纪录片"
	case 126:
		mediaName = "电影"
	case 127:
		mediaName = "动漫"
	}
	return
}

// ReasonStyle reason style
type ReasonStyle struct {
	Text             string `json:"text,omitempty"`
	TextColor        string `json:"text_color,omitempty"`
	TextColorNight   string `json:"text_color_night,omitempty"`
	BgColor          string `json:"bg_color,omitempty"`
	BgColorNight     string `json:"bg_color_night,omitempty"`
	BorderColor      string `json:"border_color,omitempty"`
	BorderColorNight string `json:"border_color_night,omitempty"`
	BgStyle          int8   `json:"bg_style,omitempty"`
}

// DurationString duration to string
func DurationString(second int64) (s string) {
	var hour, min, sec int
	if second < 1 {
		return
	}
	d, err := time.ParseDuration(strconv.FormatInt(second, 10) + "s")
	if err != nil {
		log.Error("%v", err)
		return
	}
	r := strings.NewReplacer("h", ":", "m", ":", "s", ":")
	ts := strings.Split(strings.TrimSuffix(r.Replace(d.String()), ":"), ":")
	//nolint:gomnd
	if len(ts) == 1 {
		sec, _ = strconv.Atoi(ts[0])
		//nolint:gomnd
	} else if len(ts) == 2 {
		min, _ = strconv.Atoi(ts[0])
		sec, _ = strconv.Atoi(ts[1])
		//nolint:gomnd
	} else if len(ts) == 3 {
		hour, _ = strconv.Atoi(ts[0])
		min, _ = strconv.Atoi(ts[1])
		sec, _ = strconv.Atoi(ts[2])
	}
	if hour == 0 {
		s = fmt.Sprintf("%d:%02d", min, sec)
		return
	}
	s = fmt.Sprintf("%d:%02d:%02d", hour, min, sec)
	return
}
