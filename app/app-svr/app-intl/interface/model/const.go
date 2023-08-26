package model

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-gateway/app/app-svr/archive/service/api"
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
	// PlatAndroidG is int8 for Android Global.
	PlatAndroidG = int8(4)
	// PlatIPhoneI is int8 for Iphone Global.
	PlatIPhoneI = int8(5)
	// PlatIPadI is int8 for IPAD Global.
	PlatIPadI = int8(6)
	// PlatAndroidTV is int8 for AndroidTV Global.
	PlatAndroidTV = int8(7)
	// PlatAndroidI is int8 for Android Global.
	PlatAndroidI = int8(8)

	GotoAv              = "av"
	GotoWeb             = "web"
	GotoBangumi         = "bangumi"
	GotoPGC             = "pgc"
	GotoLive            = "live"
	GotoGame            = "game"
	GotoAdAv            = "ad_av"
	GotoAdWeb           = "ad_web"
	GotoRank            = "rank"
	GotoBangumiRcmd     = "bangumi_rcmd"
	GotoLogin           = "login"
	GotoBanner          = "banner"
	GotoAdWebS          = "ad_web_s"
	GotoConverge        = "converge"
	GotoSpecial         = "special"
	GotoArticle         = "article"
	GotoArticleS        = "article_s"
	GotoGameDownloadS   = "game_download_s"
	GotoShoppingS       = "shopping_s"
	GotoAudio           = "audio"
	GotoPlayer          = "player"
	GotoAdLarge         = "ad_large"
	GotoSpecialS        = "special_s"
	GotoPlayerLive      = "player_live"
	GotoSong            = "song"
	GotoLiveUpRcmd      = "live_up_rcmd"
	GotoUpRcmdAv        = "up_rcmd_av"
	GotoSubscribe       = "subscribe"
	GotoSearchSubscribe = "search_subscribe"
	GotoChannelRcmd     = "channel_rcmd"
	GotoMoe             = "moe"
	GotoHotPage         = "hot_page"
	GotoConvergeContent = "converge_content"
	GotoPlayerOGV       = "player_ogv"
	GotoVip             = "vip"

	GotoTabTagRcmd     = "tag_rcmd"
	GotoTabContentRcmd = "content_rcmd"
	GotoTabNews        = "news"
	GotoTabEntrance    = "entrance"

	// GotoAuthor is search
	GotoAuthor         = "author"
	GotoSp             = "sp"
	GotoMovie          = "movie"
	GotoEP             = "ep"
	GotoSuggestKeyWord = "suggest_keyword"
	GotoRecommendWord  = "recommend_word"
	GotoTwitter        = "twitter"
	GotoChannel        = "channel"

	FromOrder     = "order"
	FromOperation = "operation"
	FromRcmd      = "recommend"

	CoverIng      = "即将上映"
	CoverPay      = "付费观看"
	CoverFree     = "免费观看"
	CoverVipFree  = "付费观看"
	CoverVipOnly  = "专享"
	CoverVipFirst = "抢先"

	// movie status
	MovieStatusIng      = 0
	MovieStatusPay      = 1
	MovieStatusFree     = 2
	MovieStatusVipFree  = 3
	MovieStatusVipOnly  = 4
	MovieStatusVipFirst = 5

	Hans = "hans"
	Hant = "hant"

	// ForbidCode is forbid by law
	ForbidCode   = -110
	NoResultCode = -111

	// badge type
	BgStyleFill              = int8(1)
	BgStyleStroke            = int8(2)
	BgStyleFillAndStroke     = int8(3)
	BgStyleNoFillAndNoStroke = int8(4)

	// staff attribute
	StaffLabelAd = int32(1)

	// media type
	MediaTypeBangumi        = 1
	MediaTypeMovie          = 2
	MediaTypeDocumentary    = 3
	MediaTypeGuoChuang      = 4
	MediaTypeTvSeries       = 5
	MediaTypeComic          = 6
	MediaTypeShow           = 7
	MediaTypeTvSeriesNew    = 123
	MediaTypeShowNew        = 124
	MediaTypeDocumentaryNew = 125
	MediaTypeMovieNew       = 126
	MediaTypeAnimation      = 127

	// channel type
	ChannelCtypeNew = 2
	// 海外禁止项
	OverseaBlockKey = "54"
)

var (
	// AvHandler is handler
	AvHandler = func(a *api.Arc, trackid string) func(uri string) string {
		return func(uri string) string {
			if a == nil {
				return uri
			}
			var uriStr string
			if a.Dimension.Height != 0 || a.Dimension.Width != 0 {
				uriStr = fmt.Sprintf("%s?player_width=%d&player_height=%d&player_rotate=%d", uri, a.Dimension.Width, a.Dimension.Height, a.Dimension.Rotate)
			}
			if trackid != "" {
				if uriStr == "" {
					uriStr = fmt.Sprintf("%s?trackid=%s", uri, trackid)
				} else {
					uriStr = fmt.Sprintf("%s&trackid=%s", uriStr, trackid)
				}
			}
			if uriStr != "" {
				return uriStr
			}
			return uri
		}
	}
	// hant
	hantMap = map[string]struct{}{
		"zh-Hant_TW": {},
		"zh-Hant_HK": {},
		"zh-Hant_MO": {},
		"zh_TW":      {},
		"zh_HK":      {},
		"zh_MO":      {},
		"TW":         {},
		"HK":         {},
		"MO":         {},
	}
	// 稿件秒开
	AvPlayHandlerGRPC = func(a *api.Arc, aPlay *api.PlayerInfo, trackID string) func(uri string) string {
		var (
			player     string
			ap         = aPlay.GetPlayurl()
			curDim     = aPlay.GetPlayerExtra().GetDimension()
			currentCid = aPlay.GetPlayurl().GetCid()
		)
		if ap != nil {
			bs, _ := json.Marshal(ap)
			player = string(bs)
		}
		return func(uri string) string {
			u, err := url.Parse(uri)
			if err != nil {
				return uri
			}
			params, err := url.ParseQuery(u.RawQuery)
			if err != nil {
				return uri
			}
			if aPlay != nil { //有返回播放信息
				params.Set("cid", strconv.FormatInt(int64(currentCid), 10))
				if player != "" {
					params.Set("player_preload", player)
				}
				if curDim != nil && (curDim.Height != 0 || curDim.Width != 0) {
					params.Set("player_width", strconv.FormatInt(curDim.Width, 10))
					params.Set("player_height", strconv.FormatInt(curDim.Height, 10))
					params.Set("player_rotate", strconv.FormatInt(curDim.Rotate, 10))
				}
			} else { //无播放信息
				if a != nil && (a.Dimension.Height != 0 || a.Dimension.Width != 0) {
					params.Set("player_width", strconv.FormatInt(a.Dimension.Width, 10))
					params.Set("player_height", strconv.FormatInt(a.Dimension.Height, 10))
					params.Set("player_rotate", strconv.FormatInt(a.Dimension.Rotate, 10))
				}
			}
			if trackID != "" {
				params.Set("trackid", trackID)
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
)

// TWLocale is taiwan locale
func TWLocale(locale string) bool {
	var twLocalem = map[string]struct{}{
		"zh_hk": {},
		"zh_mo": {},
		"zh_tw": {},
	}
	_, ok := twLocalem[strings.ToLower(locale)]
	return ok
}

func IsHant(cLocale, sLocale string) bool {
	locale := cLocale
	if sLocale != "" {
		locale = sLocale
	}
	if _, ok := hantMap[locale]; ok {
		return true
	}
	return false
}

// IsIPhoneB check plat is ios but not iphone_b.
func IsIOSNormal(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone || plat == PlatIPadI || plat == PlatIPhoneI
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
		return PlatIPad
	case "android", "android_b":
		return PlatAndroid
	case "win":
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
	}
	return PlatIPhone
}

// IsAndroid check plat is android or ipad.
func IsAndroid(plat int8) bool {
	return plat == PlatAndroid || plat == PlatAndroidG
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
	return plat == PlatIPad || plat == PlatIPadI
}

// IsOverseas is overseas
func IsOverseas(plat int8) bool {
	return plat == PlatAndroidI || plat == PlatIPhoneI || plat == PlatIPadI
}

// FillURI deal app schema.
func FillURI(gt, param string, f func(uri string) string) (uri string) {
	if param == "" {
		switch gt {
		case GotoHotPage:
			uri = "bilibili://pegasus/hotpage"
		}
		return
	}
	switch gt {
	case GotoAv, "":
		uri = "bilibili://video/" + param
	case GotoLive:
		uri = "bilibili://live/" + param
	case GotoBangumi:
		uri = "bilibili://bangumi/season/" + param
	case GotoGame:
		uri = "bilibili://game_center/detail?id=" + param + "&sourceType=adPut"
	case GotoSp:
		uri = "bilibili://splist/" + param
	case GotoAuthor:
		uri = "bilibili://author/" + param
	case GotoArticle:
		uri = "bilibili://article/" + param
	case GotoWeb:
		uri = param
	case GotoPGC:
		uri = "https://www.bilibili.com/bangumi/play/ss" + param
	case GotoChannel:
		uri = "bilibili://pegasus/channel/" + param + "/"
	case GotoEP:
		uri = "https://www.bilibili.com/bangumi/play/ep" + param
	case GotoTwitter:
		uri = "bilibili://pictureshow/detail/" + param
	case GotoConvergeContent:
		uri = "bilibili://search/converge/" + param
	default:
		return
	}
	if f != nil {
		uri = f(uri)
	}
	return
}

// InvalidBuild check source build is not allow by config build and condition.
// eg: when condition is gt, means srcBuild must gt cfgBuild, otherwise is invalid srcBuild.
func InvalidBuild(srcBuild, cfgBuild int, cfgCond string) bool {
	if cfgBuild != 0 && cfgCond != "" {
		switch cfgCond {
		case "gt":
			if cfgBuild >= srcBuild {
				return true
			}
		case "lt":
			if cfgBuild <= srcBuild {
				return true
			}
		case "eq":
			if cfgBuild != srcBuild {
				return true
			}
		case "ne":
			if cfgBuild == srcBuild {
				return true
			}
		}
	}
	return false
}

// StatusMark cover status mark
func StatusMark(status int) string {
	if status == MovieStatusIng {
		return CoverIng
	} else if status == MovieStatusPay {
		return CoverPay
	} else if status == MovieStatusFree {
		return CoverFree
	} else if status == MovieStatusVipFree {
		return CoverVipFree
	} else if status == MovieStatusVipOnly {
		return CoverVipOnly
	} else if status == MovieStatusVipFirst {
		return CoverVipFirst
	}
	return ""
}

var (
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
	// 剧集独家角标
	UGCSeasonSoleBadge = &ReasonStyle{
		Text:             "独家",
		TextColor:        "#FEFEFE",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FB7299",
		BgColorNight:     "#BB5B76",
		BorderColor:      "#FB7299",
		BorderColorNight: "#BB5B76",
		BgStyle:          BgStyleFill,
	}
	// 剧集首发角标
	UGCSeasonStartingBadge = &ReasonStyle{
		Text:             "首发",
		TextColor:        "#FEFEFE",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FB7299",
		BgColorNight:     "#BB5B76",
		BorderColor:      "#FB7299",
		BorderColorNight: "#BB5B76",
		BgStyle:          BgStyleFill,
	}
)

// FormMediaType media type
func FormMediaType(mediaType int) (mediaName string) {
	switch mediaType {
	case MediaTypeBangumi:
		mediaName = "番剧"
	case MediaTypeMovie:
		mediaName = "电影"
	case MediaTypeDocumentary:
		mediaName = "纪录片"
	case MediaTypeGuoChuang:
		mediaName = "国创"
	case MediaTypeTvSeries:
		mediaName = "电视剧"
	case MediaTypeComic:
		mediaName = "漫画"
	case MediaTypeShow:
		mediaName = "综艺"
	case MediaTypeTvSeriesNew:
		mediaName = "电视剧"
	case MediaTypeShowNew:
		mediaName = "综艺"
	case MediaTypeDocumentaryNew:
		mediaName = "纪录片"
	case MediaTypeMovieNew:
		mediaName = "电影"
	case MediaTypeAnimation:
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
