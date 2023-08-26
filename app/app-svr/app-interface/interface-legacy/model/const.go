package model

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/conf/env"
	"go-common/library/log"
	hisApi "go-gateway/app/app-svr/app-interface/interface-legacy/api/history"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/archive/service/api"

	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	upmdl "git.bilibili.co/bapis/bapis-go/up-archive/service"
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
	PlatAndroidHD   = int8(90)
	PlatMgrIPadHD   = int8(20)
	PlatMgrAndroidB = int8(9)
	PlatMgrIphoneB  = int8(10)

	GotoAv              = "av"
	GotoWeb             = "web"
	GotoBangumi         = "bangumi"
	GotoBangumiV2       = "bangumi_v2"
	GotoMovie           = "movie"
	GotoMovieV2         = "movie_v2"
	GotoBangumiWeb      = "bangumi_web"
	GotoSp              = "sp"
	GotoLive            = "live"
	GotoGame            = "game"
	GotoAuthor          = "author"
	GotoAuthorNew       = "author_new"
	GotoClip            = "clip"
	GotoAlbum           = "album"
	GotoArticle         = "article"
	GotoAudio           = "audio"
	GotoSpecial         = "special"
	GotoBanner          = "banner"
	GotoSpecialS        = "special_s"
	GotoConverge        = "converge"
	GOtoRecommendWord   = "recommend_word"
	GotoPGC             = "pgc"
	GotoSuggestKeyWord  = "suggest_keyword"
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
	GotoSearchUpper     = "search_upper"
	GotoConvergeContent = "converge_content"
	GotoOGVCard         = "ogv_card"
	GotoBangumiRelates  = "bangumi_relates"
	GotoFindMore        = "find_more"
	GotoSearch          = "search"
	GotoNewGame         = "new_game"
	GotoCheese          = "cheese"
	GotoChannelNew      = "channel_new"
	GotoChannelMedia    = "ogv_channel"
	GotoESports         = "esport"
	GotoLiveWeb         = "live_web"
	GotoFullScreen      = "fullscreen_video"
	GotoTips            = "tips"
	GotoGameAD          = "game_ad"
	GotoUGCInline       = "ugc_inline"
	GotoLiveInline      = "live_inline"
	GotoOGVInline       = "ogv_inline"
	GotoPediaCard       = "pedia_card"
	GotoBrandAdAv       = "brand_ad_av"
	GotoBrandAdLocalAv  = "brand_ad_local_av"
	GotoBrandAdLive     = "brand_ad_live"
	GotoSportsVersus    = "sports_versus"
	GotoSports          = "sports"
	GotoPediaInlineCard = "pedia_card_inline"
	GotoRecommendTips   = "recommend_tips"
	GotoCollectionCard  = "collection_card"
	GotoEsportsInline   = "esports_inline"

	// EnvPro is pro.
	EnvPro = "pro"
	EnvHK  = "hk"
	// EnvTest is env.
	EnvTest = "test"
	// EnvDev is env.
	EnvDev = "dev"
	// ForbidCode is forbid by law
	ForbidCode   = -110
	NoResultCode = -111

	CoverIng      = "即将上映"
	CoverPay      = "付费观看"
	CoverFree     = "免费观看"
	CoverVipFree  = "付费观看"
	CoverVipOnly  = "专享"
	CoverVipFirst = "抢先"

	Hans = "hans"
	Hant = "hant"

	// AttrNo attribute no
	AttrNo = int32(0)
	// AttrYes attribute yes
	AttrYes = int32(1)

	AttrBitArchive = uint32(0)
	AttrBitArticle = uint32(1)
	AttrBitClip    = uint32(2)
	AttrBitAlbum   = uint32(3)
	AttrBitAudio   = uint32(34)
	AttrBitComic   = uint32(4)
	AttrBitIsPGC   = uint32(9)
	// badge type
	BgStyleFill              = int8(1)
	BgStyleStroke            = int8(2)
	BgStyleFillAndStroke     = int8(3)
	BgStyleNoFillAndNoStroke = int8(4)

	// 大会员铭牌过期图标
	VipLabelExpire  = "https://i0.hdslb.com/bfs/vip/label_overdue.png"
	VipStatusNormal = 1
	VipStatusExpire = 0

	// 剧集状态
	UGCSeasonUnsigned = 0
	UGCSeasonSole     = 1
	UGCSeasonStarting = 2

	// 角标颜色
	BgColorRed    = int8(1)
	BgColorYellow = int8(2)
	BgColorBlue   = int8(3)

	// bvid开关
	BvOpen = 1

	// 游戏二级页面
	ButtonReserve  = 0
	ButtonDownload = 1
	ButtonEnter    = 4
	GameNotice     = "公告"
	GameGift       = "礼包"
	ReserveName    = "预约"
	DownloadName   = "下载"
	EnterName      = "进入"

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
	// Device Unknown
	DeviceUnknown = int8(0)
	// 车载
	DeviceCar = int8(8)
	// 物联网设备
	DeviceIoT = int8(9)
	// 安卓pad
	DeviceAndPad = int8(10)

	AnswerSourceMyinfo = "myinfo"

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
	SearchEsCard         = "search_es_card"

	// 新用户实验period
	NewUserOgvExperimentPeriod = "0-24"
)

// for FillURI
var (
	AvHandler = func(a *api.Arc) func(uri string) string {
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
			if a.AttrValV2(api.AttrBitV2ActSeason) == api.AttrYes && a.SeasonTheme != nil {
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
	AvPlayHandlerGRPC = func(a *api.Arc, playInfo *api.PlayerInfo) func(uri string) string {
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
			if a.AttrValV2(api.AttrBitV2ActSeason) == api.AttrYes && a.SeasonTheme != nil {
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
	NoteHandler = func(noteId int64) func(uri string) string {
		return func(uri string) string {
			return fmt.Sprintf("%s?cvid=%d&locate_note_editing=true", uri, noteId)
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
	// 热门视频角标
	PopularBadge = &ReasonStyle{
		Text:             "热门",
		TextColor:        "#FEFEFE",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FB7299",
		BgColorNight:     "#BB5B76",
		BorderColor:      "#FB7299",
		BorderColorNight: "#BB5B76",
		BgStyle:          BgStyleFill,
	}
	// 互动视频角标
	SteinsBadge = &ReasonStyle{
		Text:             "互动",
		TextColor:        "#FEFEFE",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FB7299",
		BgColorNight:     "#BB5B76",
		BorderColor:      "#FB7299",
		BorderColorNight: "#BB5B76",
		BgStyle:          BgStyleFill,
	}
	// 直播回放角标
	LivePlaybackBadge = &ReasonStyle{
		Text:             "直播回放",
		TextColor:        "#FEFEFE",
		TextColorNight:   "#E5E5E5",
		BgColor:          "#FB7299",
		BgColorNight:     "#BB5B76",
		BorderColor:      "#FB7299",
		BorderColorNight: "#BB5B76",
		BgStyle:          BgStyleFill,
	}
	// 付费UGC新角标
	NewPayBadge = &ReasonStyle{
		Text:             "付费",
		TextColor:        "#FFFFFF",
		TextColorNight:   "#FFFFFF",
		BgColor:          "#FF6699",
		BgColorNight:     "#D44E7D",
		BorderColor:      "#FF6699",
		BorderColorNight: "#D44E7D",
		BgStyle:          BgStyleFill,
	}

	// 我的页模块对应关系
	IPhoneMenu = map[int]string{
		OpModule:          "宅家挑战赛",
		Creative:          "创作中心",
		IPhoneBcreative:   "创作中心",
		IPhoneLive:        "直播中心",
		SelfCenter:        "个人中心",
		IPhoneBselfCenter: "个人中心",
		MyService:         "我的服务",
		IPhoneBmyService:  "我的服务",
	}
	IPhoneMenuTp = map[int]int{
		Creative:          1,
		IPhoneBcreative:   1,
		IPhoneLive:        2,
		SelfCenter:        3,
		IPhoneBselfCenter: 3,
		MyService:         4,
		IPhoneBmyService:  4,
		OpModule:          5,
	}
	IPadNormalMenu = map[int8][]int{
		PlatIPad:      {IPadCreative, IPadSelfCenter},
		PlatIpadHD:    {IPadHDCreative, IPadHDSelfCenter},
		PlatAndroidHD: {AndroidPadSelfCenter},
	}
	IPadFilterMenu = map[int8][]int{
		PlatIPad:   {IPadSelfCenter},
		PlatIpadHD: {IPadHDSelfCenter},
	}
	AndroidMenu = map[int8][]int{
		PlatAndroid:  {AndroidSelfCenter, OpModule, AndroidCreative, AndroidLive, AndroidMyService},
		PlatAndroidI: {AndroidISelfCenter, AndroidIMyService},
		PlatAndroidB: {AndroidBSelfCenter, AndroidBCreative, AndroidBMyService},
	}
	IPhoneNormalMenu = map[int8][]int{
		PlatIPhone:  {Creative, OpModule, IPhoneLive, SelfCenter, MyService},
		PlatIPhoneB: {IPhoneBcreative, IPhoneBselfCenter, IPhoneBmyService},
	}
	IPhoneFilterMenu = map[int8][]int{
		PlatIPhone:  {SelfCenter, MyService},
		PlatIPhoneB: {IPhoneBselfCenter, IPhoneBmyService},
	}
	CreativeModules = map[int64]int64{
		Creative:         Creative,
		IPhoneBcreative:  IPhoneBcreative,
		AndroidCreative:  AndroidCreative,
		AndroidBCreative: AndroidBCreative,
		IPadCreative:     IPadCreative,
		IPadHDCreative:   IPadHDCreative,
	}
	LiveModules = map[int64]int64{
		AndroidLive: AndroidLive,
		IPhoneLive:  IPhoneLive,
	}
)

// IsAndroid check plat is android or ipad.
func IsAndroid(plat int8) bool {
	return plat == PlatAndroid || plat == PlatAndroidG || plat == PlatAndroidI || plat == PlatAndroidB
}

// IsAndroidPick check plat is android pick
func IsAndroidPick(plat int8) bool {
	return plat == PlatAndroid
}

// IsIOSPick check plat is iphone or ipad pick
func IsIOSPick(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone
}

func IsIPadPink(plat int8) bool {
	return plat == PlatIPad
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

func IsAndroidB(plat int8) bool {
	return plat == PlatAndroidB
}

func IsBlue(plat int8) bool {
	return plat == PlatIPhoneB || plat == PlatAndroidB
}

func IsAndroidHD(plat int8) bool {
	return plat == PlatAndroidHD
}

func IsIPadHD(plat int8) bool {
	return plat == PlatIpadHD
}

func IsPinkAndBlue(plat int8) bool {
	return plat == PlatIPhone || plat == PlatAndroid || plat == PlatIPhoneB || plat == PlatAndroidB
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

// IsOverseas is overseas
func IsOverseas(plat int8) bool {
	return plat == PlatAndroidI || plat == PlatIPhoneI || plat == PlatIPadI
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

// AttrVal get attribute value
func AttrVal(attr int32, bit uint32) (v int32) {
	v = (attr >> bit) & int32(1)
	return
}

// AttrSet set attribute value
func AttrSet(attr int32, v int32, bit uint32) int32 {
	return attr&(^(1 << bit)) | (v << bit)
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

// NewCursor judge cuser
func NewCursor(maxID int64, minID int64, size int) (cuser *Cursor, err error) {
	if maxID < 0 || minID < 0 {
		err = fmt.Errorf("either max_id(%d) or min_id(%d) < 0", maxID, minID)
		return
	}
	if (minID * maxID) != 0 {
		err = fmt.Errorf("both max_id(%d) and max_id(%d) > 0", maxID, minID)
		return
	}
	if minID == 0 && maxID == 0 {
		cuser = &Cursor{Current: 0, Direction: Downward, Size: size}
	} else if maxID > 0 {
		cuser = &Cursor{Current: maxID, Direction: Downward, Size: size}
	} else {
		cuser = &Cursor{Current: minID, Direction: Upward, Size: size}
	}
	return
}

// InvalidBuild invalid build
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

// env sh001 run
func EnvRun() (res bool) {
	var _zone = "sh001"
	//nolint:gosimple
	if env.Zone == _zone {
		return true
	}
	return false
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

func BadgeStyleFrom(style int8, text string) (res *ReasonStyle) {
	res = &ReasonStyle{
		Text: text,
	}
	switch style {
	case BgColorYellow:
		res.TextColor = "#FFFFFFFF"
		res.BgColor = "#FFF9AC4B"
	case BgColorBlue:
		res.TextColor = "#FF23ADE5"
		res.BgColor = "#3323ADE5"
	case BgColorRed:
		res.TextColor = "#FFFFFFFF"
		res.BgColor = "#FFFB7299"
	}
	return
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

func HistoryDt(dt int8, icon *conf.HisIcon) *hisApi.DeviceType {
	if icon == nil {
		return nil
	}
	switch dt {
	case DeviceIphone, DeviceAndroid, DeviceWP8, DeviceUWP:
		return &hisApi.DeviceType{
			Type: hisApi.DT_Phone,
			Icon: icon.Phone,
		}
	case DevicePC:
		return &hisApi.DeviceType{
			Type: hisApi.DT_PC,
			Icon: icon.PC,
		}
	case DeviceIpad:
		return &hisApi.DeviceType{
			Type: hisApi.DT_Pad,
			Icon: icon.Pad,
		}
	case DeviceAndroidTV:
		return &hisApi.DeviceType{
			Type: hisApi.DT_TV,
			Icon: icon.TV,
		}
	case DeviceCar:
		return &hisApi.DeviceType{
			Type: hisApi.DT_Car,
			Icon: icon.Car,
		}
	case DeviceIoT:
		return &hisApi.DeviceType{
			Type: hisApi.DT_IoT,
			Icon: icon.Iot,
		}
	case DeviceAndPad:
		return &hisApi.DeviceType{
			Type: hisApi.DT_AndPad,
			Icon: icon.Pad,
		}
	default:
		return nil
	}
}

func CopyFromArc(from *upmdl.Arc) *api.Arc {
	to := &api.Arc{
		Aid:         from.Aid,
		Videos:      from.Videos,
		TypeID:      from.TypeID,
		TypeName:    from.TypeName,
		Copyright:   from.Copyright,
		Pic:         from.Pic,
		Title:       from.Title,
		PubDate:     from.PubDate,
		Ctime:       from.Ctime,
		Desc:        from.Desc,
		State:       from.State,
		Access:      from.Access,
		Attribute:   from.Attribute,
		Tag:         from.Tag,
		Tags:        from.Tags,
		Duration:    from.Duration,
		MissionID:   from.MissionID,
		OrderID:     from.OrderID,
		RedirectURL: from.RedirectURL,
		Forward:     from.Forward,
		Rights: api.Rights{
			Bp:              from.Rights.Bp,
			Elec:            from.Rights.Elec,
			Download:        from.Rights.Download,
			Movie:           from.Rights.Movie,
			Pay:             from.Rights.Pay,
			HD5:             from.Rights.HD5,
			NoReprint:       from.Rights.NoReprint,
			Autoplay:        from.Rights.Autoplay,
			UGCPay:          from.Rights.UGCPay,
			IsCooperation:   from.Rights.IsCooperation,
			UGCPayPreview:   from.Rights.UGCPayPreview,
			NoBackground:    from.Rights.NoBackground,
			ArcPay:          from.Rights.ArcPay,
			ArcPayFreeWatch: from.Rights.ArcPayFreeWatch,
		},
		Author: api.Author{
			Mid:  from.Author.Mid,
			Name: from.Author.Name,
			Face: from.Author.Face,
		},
		Stat: api.Stat{
			Aid:     from.Stat.Aid,
			View:    from.Stat.View,
			Danmaku: from.Stat.Danmaku,
			Reply:   from.Stat.Reply,
			Fav:     from.Stat.Fav,
			Coin:    from.Stat.Coin,
			Share:   from.Stat.Share,
			NowRank: from.Stat.NowRank,
			HisRank: from.Stat.HisRank,
			Like:    from.Stat.Like,
			DisLike: from.Stat.DisLike,
			Follow:  from.Stat.Follow,
		},
		ReportResult: from.ReportResult,
		Dynamic:      from.Dynamic,
		FirstCid:     from.FirstCid,
		Dimension: api.Dimension{
			Width:  from.Dimension.Width,
			Height: from.Dimension.Height,
			Rotate: from.Dimension.Rotate,
		},
		SeasonID:    from.SeasonID,
		AttributeV2: from.AttributeV2,
	}
	for _, v := range from.StaffInfo {
		if v == nil {
			continue
		}
		to.StaffInfo = append(to.StaffInfo, &api.StaffInfo{
			Mid:       v.Mid,
			Title:     v.Title,
			Attribute: v.Attribute,
		})
	}
	return to
}

// 和app/app-svr/app-dynamic/interface/model/const.go StatString方法逻辑保持一致
// nolint:gomnd
func StatNumberToString(number int64, suffix string) string {
	if number < 10000 {
		return strconv.FormatInt(number, 10) + suffix
	}
	var rawFormat string
	if number < 100000000 {
		rawFormat = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(rawFormat, ".0") + "万" + suffix
	}
	rawFormat = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(rawFormat, ".0") + "亿" + suffix
}

// 和app/app-svr/app-dynamic/interface/model/const.go UpPubDataString方法逻辑保持一致
func PubTimeToString(t time.Time) string {
	now := time.Now()
	if now.Year() == t.Year() {
		if now.Month() == t.Month() && now.Day() == t.Day() {
			return "今天 " + t.Format("15:04")
		}
		if now.Month() == t.Month() && now.Day() < t.Day() && (t.Day()-now.Day() == 1) {
			return "明天 " + t.Format("15:04")
		}
		return t.Format("01-02 15:04")
	}
	return t.Format("2006-01-02 15:04")
}
