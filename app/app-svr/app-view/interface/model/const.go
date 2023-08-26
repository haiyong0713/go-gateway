package model

import (
	"fmt"
)

const (
	// archive attribute_v2
	AttrBitV2CleanMode = uint(5)

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
	// PlatAndroidB is int8 for android_b
	PlatAndroidB = int8(9)
	// PlatIPhoneB is int8 for iphone_b
	PlatIPhoneB = int8(10)
	// PlatAndroidTVYST is int8 for AndroidTV_YST Global.
	PlatAndroidTVYST = int8(12)
	// PlatIpadHD is int8 for IpadHD
	PlatIpadHD    = int8(20)
	PlatAndroidHD = int8(90)

	GotoAv          = "av"
	GotoWeb         = "web"
	GotoBangumi     = "bangumi"
	GotoBangumiEp   = "bangumi-ep"
	GotoLive        = "live"
	GotoGame        = "game"
	GotoArticle     = "article"
	GotoSpecial     = "special"
	GotoCm          = "cm"
	GotoSearchUpper = "search_upper"
	GotoHotPage     = "hot_page"
	GotoEP          = "ep"
	GotoSpace       = "space"
	GotoOrder       = "order"

	// for fill uri
	GotoAudio    = "audio"
	GotoSong     = "song"
	GotoAudioTag = "audio_tag"
	GotoAlbum    = "album"
	GotoClip     = "clip"
	GotoDaily    = "daily"

	// EnvPro is pro.
	EnvPro = "pro"
	EnvHK  = "hk"
	// EnvTest is env.
	EnvTest = "test"
	// EnvDev is env.
	EnvDev = "dev"
	// ForbidCode is forbid by law
	ForbidCode = -110

	StatusIng      = 0
	StatusPay      = 1
	StatusFree     = 2
	StatusVipFree  = 3
	StatusVipOnly  = 4
	StatusVipFirst = 5
	CoverIng       = "即将上映"
	CoverPay       = "付费观看"
	CoverFree      = "免费观看"
	CoverVipFree   = "付费观看"
	CoverVipOnly   = "专享"
	CoverVipFirst  = "抢先"

	Hans = "hans"
	Hant = "hant"

	FromOrder     = "order"
	FromOperation = "operation"
	FromRcmd      = "recommend"

	// Share channel
	ShareDefault = int32(0)
	ShareDynamic = int32(1)
	ShareMessage = int32(2)
	ShareQQ      = int32(3)
	ShareQzone   = int32(4)
	ShareWeChat  = int32(5)
	ShareMoment  = int32(6)
	ShareWeiBo   = int32(7)
	// Share channel string
	ShareDefaultStr = "default"
	ShareDynamicStr = "dynamic"
	ShareMessageStr = "message"
	ShareQQStr      = "qq"
	ShareQzoneStr   = "qzone"
	ShareWeChatStr  = "wechat"
	ShareMomentStr  = "moment"
	ShareWeiBoStr   = "weibo"
	// share type
	ShareTypeAV     = "av"
	ShareTypeLive   = "live"
	ShareTypeMelloi = "melloi"

	RelateCmd       = "related"
	PlayerRelateCmd = "player_rec"
	RelateTabCmd    = "rrtab"
	// reasonStyle color
	BgColorOrange         = int32(0)
	BgColorTransparentRed = int32(1)
	BgLightColoredOrange  = int32(2)

	BgStyleFill   = int32(1)
	BgStyleStroke = int32(2)

	// staff attribute
	StaffLabelAd = int32(1)

	// view & relate
	PathView           = "/x/v2/view"
	PathViewPage       = "/x/v2/view/page"
	PathRelateTab      = "/x/v2/view/relate/tab"
	RPCPathFeedView    = "/bilibili.app.view.v1.View/FeedView"
	PageTypeRelate     = "2"
	PageTypeRelateTab  = "3"
	PathCacheView      = "/bilibili.app.view.v1.View/CacheView"
	PathContinuousPlay = "/bilibili.app.view.v1.View/ContinuousPlay"
	PathRelatesFeed    = "/bilibili.app.view.v1.View/RelatesFeed"
	PathPlayerRelates  = "/bilibili.app.view.v1.View/PlayerRelates"

	// honor rank
	RankIcon          = "https://i0.hdslb.com/bfs/app/7e28ab559dd8d63cc7c827227ce5566f34c577c3.png"
	RankIconNight     = "https://i0.hdslb.com/bfs/app/01311b8346417aac9bee5e8279bb735f87b92e6b.png"
	ActSeasonRankIcon = "https://i0.hdslb.com/bfs/app/5b9e55e5f639d288c0fa121e792ead27ea98401a.png"

	// 风控相关
	SilverSourceLike       = "1"
	SilverSourceTriple     = "2"
	SilverSourceNologin    = "3"
	SilverSourceCoinTolike = "4"
	SilverSceneLike        = "thumbup_video"
	SilverSceneCoin        = "video_coin"
	SilverSceneTriple      = "video_triplelike"
	SilverSceneShare       = "video_share"
	SilverSceneCointolike  = "video_cointolike"
	SilverActionLike       = "like"
	SilverActionCoin       = "video_coin"
	SilverActionTriple     = "video_triplelike"
	SilverActionShare      = "video_share"
	SilverActionCointolike = "video_cointolike"

	// 大型活动页
	LiveBefore       = 1 //直播前 预约模块配置
	LiveAfter        = 3 //直播后 预约模块配置
	PlatActSeasonApp = 1
	PlatActSeasonHD  = 2

	FavTypeVideo  = 2
	FavTypeSeason = 21
	ReplyTypeAv   = 1

	//mobile
	MobileAppIphone  = "iphone"
	MobileAppAndroid = "android"

	// page version
	PageVersionV1 = "v1" // 传统播放页
	PageVersionV2 = "v2" // 可上下滑的播放页

	PaginationTokenSalt = "relate"

	ElecEnableStatus     = 1
	ElecPlusEnableStatus = 2
)

var (
	// share
	ShareTypeMap = map[string]int32{
		ShareTypeAV:     3,
		ShareTypeLive:   6,
		ShareTypeMelloi: 8378,
	}
	DisplayHonor = map[int32]int32{
		1: 1, //入站必刷
		2: 2, //每周必看
		3: 3, //日排行
		//4: 4, //热门
		5: 5, //频道精选
	}
	HonorIcon = map[int32]string{
		1: "https://i0.hdslb.com/bfs/activity-plat/static/20190813/4d5f834914c6978cebdcdcecd3eb32b0/hWeqoAoebD.png",
		2: "https://i0.hdslb.com/bfs/activity-plat/static/20190813/4d5f834914c6978cebdcdcecd3eb32b0/8fG5kctAZe.png",
		3: "https://i0.hdslb.com/bfs/activity-plat/static/20190813/4d5f834914c6978cebdcdcecd3eb32b0/AonQf3n9vt.png",
		5: "https://i0.hdslb.com/bfs/tag/fa4b0cf7bb16bd63720aa2ae5347d580db7a6b32.png",
	}
	HonorIconNight = map[int32]string{
		1: "https://i0.hdslb.com/bfs/activity-plat/static/20190813/4d5f834914c6978cebdcdcecd3eb32b0/ZeuBgPHYg.png",
		2: "https://i0.hdslb.com/bfs/activity-plat/static/20190813/4d5f834914c6978cebdcdcecd3eb32b0/S8WzVVsVE1.png",
		3: "https://i0.hdslb.com/bfs/activity-plat/static/20190813/4d5f834914c6978cebdcdcecd3eb32b0/wD3L1sV8zq.png",
		5: "https://i0.hdslb.com/bfs/tag/6677c68ae1fa40c88376bc65dc4bbf50f6df76bf.png",
	}
	HonorTextExtra = map[int32]string{
		1: "收录",
		2: "收录",
		5: "精选",
	}
	HonorTextColor = map[int32]string{
		1: "#F3921F",
		2: "#F7B800",
		3: "#FB7299",
		5: "#FB7299",
	}
	HonorTextColorNight = map[int32]string{
		1: "#BA6C45",
		2: "#BA833F",
		3: "#BB5B76",
		5: "#BB5B76",
	}
	HonorBgColor = map[int32]string{
		1: "#FFF5EA",
		2: "#FFFAE8",
		3: "#FFF0F1",
		5: "#FFF0F1",
	}
	HonorBgColorNight = map[int32]string{
		1: "#332E29",
		2: "#333029",
		3: "#332929",
		5: "#332929",
	}
	HonorURLText = map[int32]string{
		1: "点击查看",
		2: "点击查看",
		3: "当前排行榜",
		5: "点击查看",
	}
	OperateType = map[int]string{
		0:  GotoWeb,
		1:  GotoGame,
		2:  GotoAv,
		3:  GotoEP,
		4:  GotoLive,
		6:  GotoArticle,
		7:  GotoDaily,
		8:  GotoAudio,
		9:  GotoSong,
		10: GotoAlbum,
		11: GotoClip,
		14: GotoBangumi,
	}

	// Share channel
	ShareChannelToString = map[int32]string{
		ShareDefault: ShareDefaultStr,
		ShareDynamic: ShareDynamicStr,
		ShareMessage: ShareMessageStr,
		ShareQQ:      ShareQQStr,
		ShareQzone:   ShareQzoneStr,
		ShareWeChat:  ShareWeChatStr,
		ShareMoment:  ShareMomentStr,
		ShareWeiBo:   ShareWeiBoStr,
	}
	LiveRoomHandler = func(broadcastType int64) func(uri string) string {
		return func(uri string) string {
			return fmt.Sprintf("%s?broadcast_type=%d", uri, broadcastType)
		}
	}
	DynamicCoverTp = map[string]string{
		GotoSpecial: "1",
		GotoAv:      "2",
	}
)

// IsAndroid check plat is android or ipad.
func IsAndroid(plat int8) bool {
	return plat == PlatAndroid || plat == PlatAndroidG || plat == PlatAndroidI || plat == PlatAndroidB
}

// IsIOS check plat is iphone or ipad.
func IsIOS(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone || plat == PlatIPadI || plat == PlatIPhoneI || plat == PlatIPhoneB
}

// IsIPhone check plat is iphone.
func IsIPhone(plat int8) bool {
	return plat == PlatIPhone || plat == PlatIPhoneI || plat == PlatIPhoneB
}

// IsIPad check plat is pad.
func IsIPad(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPadI || plat == PlatIpadHD
}

// IsIOSNormal check plat is ios except iphone_b
func IsIOSNormal(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone || plat == PlatIPadI || plat == PlatIPhoneI
}

// IsIPhoneB check plat is iphone_b
func IsIPhoneB(plat int8) bool {
	return plat == PlatIPhoneB
}

// IsAndroidB check plat is android_b
func IsAndroidB(plat int8) bool {
	return plat == PlatAndroidB
}

// IsIpadHD check plat is ipadHD
func IsIpadHD(plat int8) bool {
	return plat == PlatIpadHD
}

// IsOverseas is overseas
func IsOverseas(plat int8) bool {
	return plat == PlatAndroidI || plat == PlatIPhoneI || plat == PlatIPadI
}

//nolint:gomnd
func TrafficFree(xTfIsp string) (netType, tfType int32) {
	switch xTfIsp {
	case "ct":
		return 2, 5
	case "cu":
		return 2, 1
	case "cm":
		return 2, 3
	}
	return 0, 0
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
		uri = "https://www.bilibili.com/bangumi/play/ss" + param
	case GotoBangumiEp:
		uri = "https://www.bilibili.com/bangumi/play/ep" + param
	case GotoEP:
		uri = "https://www.bilibili.com/bangumi/play/ep" + param
	case GotoArticle:
		uri = "bilibili://article/" + param
	case GotoGame:
		uri = param
	case GotoAudio:
		uri = "bilibili://music/menu/detail/" + param
	case GotoSong:
		uri = "bilibili://music/detail/" + param
	case GotoAudioTag:
		uri = "bilibili://music/categorydetail/" + param
	case GotoDaily:
		uri = "bilibili://pegasus/list/daily/" + param
	case GotoAlbum:
		uri = "bilibili://album/" + param
	case GotoClip:
		uri = "bilibili://clip/" + param
	case GotoSpace:
		uri = "bilibili://space/" + param
	case GotoWeb:
		uri = param
	}
	if f != nil {
		uri = f(uri)
	}
	return
}

//nolint:gomnd
func StatusMark(status int) string {
	if status == 0 {
		return CoverIng
	} else if status == 1 {
		return CoverPay
	} else if status == 2 {
		return CoverFree
	} else if status == 3 {
		return CoverVipFree
	} else if status == 4 {
		return CoverVipOnly
	} else if status == 5 {
		return CoverVipFirst
	}
	return ""
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

// Platform plat to platform
func Platform(plat int8) string {
	if IsAndroid(plat) {
		return "android"
	} else {
		return "ios"
	}
}

// PlatNew 区分蓝版
func PlatNew(mobiApp, device string) int8 {
	switch mobiApp {
	case "iphone":
		if device == "pad" {
			return PlatIPad
		}
		return PlatIPhone
	case "ipad":
		return PlatIpadHD
	case "android":
		return PlatAndroid
	case "android_b":
		return PlatAndroidB
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
	case "android_tv_yst":
		return PlatAndroidTVYST
	case "iphone_b":
		return PlatIPhoneB
	case "android_hd":
		return PlatAndroidHD
	}
	return PlatIPhone
}
