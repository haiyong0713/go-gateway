package model

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	cardlive "go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-channel/interface/model/tab"
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
	// PlatAndroidB is int8 for Android Bule.
	PlatAndroidB = int8(9)
	// PlatIPhoneB is int8 for iphone Bule.
	PlatIPhoneB = int8(10)
	// PlatIPadHD is int8 for ipad.
	PlatIPadHD = int8(20)

	GotoAv            = "av"
	GotoWeb           = "web"
	GotoBangumi       = "bangumi"
	GotoPGC           = "pgc"
	GotoLive          = "live"
	GotoGame          = "game"
	GotoTopic         = "topic"
	GotoActivity      = "activity"
	GotoAdAv          = "ad_av"
	GotoAdWeb         = "ad_web"
	GotoRank          = "rank"
	GotoTag           = "tag"
	GotoBangumiRcmd   = "bangumi_rcmd"
	GotoLogin         = "login"
	GotoUpBangumi     = "up_bangumi"
	GotoBanner        = "banner"
	GotoAdWebS        = "ad_web_s"
	GotoUpArticle     = "up_article"
	GotoGameDownload  = "game_download"
	GotoConverge      = "converge"
	GotoSpecial       = "special"
	GotoArticle       = "article"
	GotoArticleS      = "article_s"
	GotoGameDownloadS = "game_download_s"
	GotoShoppingS     = "shopping_s"
	GotoAudio         = "audio"
	GotoPlayer        = "player"
	GotoAdLarge       = "ad_large"
	GotoSpecialS      = "special_s"
	GotoPlayerLive    = "player_live"
	GotoSong          = "song"
	GotoUpRcmdAv      = "up_rcmd_av"
	GotoSubscribe     = "subscribe"
	GotoLiveUpRcmd    = "live_up_rcmd"
	GotoTopstick      = "topstick"
	GotoChannelRcmd   = "channel_rcmd"
	GotoPgcsRcmd      = "pgcs_rcmd"
	GotoUpRcmdS       = "up_rcmd_s"
	GotoPegasusTab    = "pegasus"
	// audio tag
	GotoAudioTag = "audio_tag"

	// extra tab
	GotoTabBackground    = "background"
	GotoTabEntrance      = "entrance"
	GotoTabContentRcmd   = "content_rcmd"
	GotoTabTagRcmd       = "tag_rcmd"
	GotoTabSignIn        = "sign_in"
	GotoTabNews          = "news"
	GotoChannelNew       = "channel"
	GotoChannelMedia     = "channel_media"
	GotoChannelNewAll    = "channel_all"
	GotoChannelNewSelect = "channel_select"
	GotoChannelNewOP     = "channel_op"
	GotoChannelNewTopic  = "channel_topic"
	GotoChannelCustom    = "channel_custom"
	GotoChannelTab       = "channel_tab"
	GotoActive           = "active"

	// EnvPro is pro.
	EnvPro = "pro"
	// EnvTest is env.
	EnvTest = "test"
	// EnvDev is env.
	EnvDev = "dev"

	OldChanne  = 1
	NewChannel = 2

	// 新稿件禁止属性位
	OverseaBlockKey = "54"
)

var (
	AvHandler = func(a *api.Arc) func(uri string) string {
		return func(uri string) string {
			if a == nil {
				return uri
			}
			if a.Dimension.Height != 0 || a.Dimension.Width != 0 {
				return fmt.Sprintf("%s?player_width=%d&player_height=%d&player_rotate=%d", uri, a.Dimension.Width, a.Dimension.Height, a.Dimension.Rotate)
			}
			return uri
		}
	}
	LiveUpHandler = func(l *cardlive.Card) func(uri string) string {
		return func(uri string) string {
			if l == nil {
				return uri
			}
			return fmt.Sprintf("%s?broadcast_type=%d", uri, l.BroadcastType)
		}
	}
	LiveRoomHandler = func(l *cardlive.Room) func(uri string) string {
		return func(uri string) string {
			if l == nil {
				return uri
			}
			return fmt.Sprintf("%s?broadcast_type=%d", uri, l.BroadcastType)
		}
	}
	PegasusHandler = func(m *tab.Menu) func(uri string) string {
		return func(uri string) string {
			if m == nil {
				return uri
			}
			if m.Title != "" {
				return fmt.Sprintf("%s?name=%s", uri, url.QueryEscape(m.Title))
			}
			return uri
		}
	}
	NewChannelTopic = func(tName string) func(uri string) string {
		return func(uri string) string {
			if tName != "" {
				return fmt.Sprintf("%s&name=%s", uri, url.QueryEscape(tName))
			}
			return uri
		}
	}
	ChannelHandler = func(tab string) func(uri string) string {
		return func(uri string) string {
			return fmt.Sprintf("%s?%s", uri, tab)
		}
	}
)

// IsAndroid check plat is android or ipad.
func IsAndroid(plat int8) bool {
	return plat == PlatAndroid || plat == PlatAndroidG
}

// IsAndroidAll check plat is android xxx
func IsAndroidAll(plat int8) bool {
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
	return plat == PlatIPad || plat == PlatIPadI
}

// IsOverseas is overseas
func IsOverseas(plat int8) bool {
	return plat == PlatAndroidI || plat == PlatIPhoneI || plat == PlatIPadI
}

// nolint:gocognit
func IsGoto(gt string) bool {
	return gt == GotoAv || gt == GotoWeb || gt == GotoBangumi || gt == GotoLive || gt == GotoGame || gt == GotoTopic || gt == GotoActivity ||
		gt == GotoAdAv || gt == GotoAdWeb || gt == GotoRank || gt == GotoTag || gt == GotoBangumiRcmd || gt == GotoLogin || gt == GotoUpBangumi ||
		gt == GotoBanner || gt == GotoAdWebS || gt == GotoUpArticle || gt == GotoGameDownload || gt == GotoGameDownloadS || gt == GotoConverge ||
		gt == GotoSpecial || gt == GotoArticle || gt == GotoArticleS || gt == GotoShoppingS || gt == GotoAudio || gt == GotoPlayer || gt == GotoAdLarge ||
		gt == GotoSpecialS || gt == GotoPlayerLive || gt == GotoSong
}

// FillURI deal app schema.
func FillURI(gt, param string, typ int, plat int8, build int, f func(uri string) string) (uri string) {
	if param == "" {
		return
	}
	switch gt {
	case GotoAv, GotoAdAv, "":
		uri = "bilibili://video/" + param
	case GotoLive:
		uri = "bilibili://live/" + param
	case GotoBangumi:
		uri = "https://www.bilibili.com/bangumi/play/ep" + param
	case GotoUpBangumi:
		uri = "https://www.bilibili.com/bangumi/play/ss" + param
	case GotoUpArticle, GotoArticle, GotoArticleS:
		uri = "bilibili://article/" + param
	case GotoGame:
		uri = "bilibili://game_center/detail?id=" + param + "&sourceType=adPut"
	case GotoAudio:
		uri = "bilibili://music/menu/detail/" + param
	case GotoSong:
		uri = "bilibili://music/detail/" + param
	case GotoSpecial, GotoGameDownload, GotoConverge, GotoGameDownloadS, GotoSpecialS, GotoTopstick:
		// nolint:gomnd
		switch typ {
		case 11:
			uri = "bilibili://clip/" + param
		case 10:
			uri = "bilibili://album/" + param
		case 9:
			uri = "bilibili://music/detail/" + param
		case 8:
			uri = "bilibili://music/menu/detail/" + param
		case 7:
			uri = "bilibili://pegasus/list/daily/" + param
		case 6:
			uri = "bilibili://article/" + param
		case 5:
			if param != "" {
				uri = "bilibili://category/65541/" + param
			} else {
				uri = "bilibili://category/65541"
			}
		case 4:
			uri = "bilibili://live/" + param
		case 3:
			uri = "https://www.bilibili.com/bangumi/play/ss" + param
		case 2:
			uri = "bilibili://video/" + param
		case 1:
			uri = "bilibili://game_center/detail?id=" + param + "&sourceType=adPut"
		case 0:
			uri = param
		}
	case GotoAudioTag:
		uri = "bilibili://music/categorydetail/" + param
	case GotoWeb, GotoActivity, GotoTopic, GotoAdWeb, GotoRank, GotoAdWebS, GotoShoppingS, GotoAdLarge:
		uri = param
	case GotoTag:
		if param != "" {
			uri = "bilibili://pegasus/channel/" + param
		}
	case GotoPegasusTab:
		uri = "bilibili://pegasus/channel/op/" + param
	case GotoChannelMedia:
		uri = "bilibili://feed/channel"
	case GotoChannelNew:
		uri = "bilibili://pegasus/channel/v2/" + param
	case GotoChannelNewAll:
		uri = "bilibili://pegasus/channel/detail/all/" + param
	case GotoChannelNewSelect:
		uri = "bilibili://pegasus/channel/detail/select/" + param
	case GotoChannelNewOP:
		uri = "bilibili://pegasus/channel/detail/op/" + param
	case GotoChannelNewTopic:
		uri = "bilibili://pegasus/channel/detail/topic?id=" + param
	case GotoChannelTab:
		uri = "bilibili://pegasus/channel/find?tab_id=" + param
	case GotoActive: // 活动话题
		uri = "bilibili://following/activity_landing/" + param
	}
	if f != nil {
		uri = f(uri)
	}
	return
}

func FillSongTagURI(id int64) (uri string) {
	return fmt.Sprintf("bilibili://music/categorydetail/%d", id)
}

func FillRedirect(gt string, typ int) (redirect string) {
	switch gt {
	case GotoSpecial, GotoGameDownload, GotoConverge, GotoGameDownloadS, GotoSpecialS:
		// nolint:gomnd
		switch typ {
		case 7:
			redirect = "daily"
		case 6:
			redirect = "article"
		case 5:
			redirect = "category/65541"
		case 4:
			redirect = "live"
		case 3:
			redirect = ""
		case 2:
			redirect = "video"
		case 1:
			redirect = "game"
		case 0:
			redirect = ""
		}
	}
	return
}

// nolint:gomnd
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

// CoverURL convert cover url to full url.
func CoverURL(uri string) (cover string) {
	if uri == "" {
		cover = "http://static.hdslb.com/images/transparent.gif"
		return
	}
	if strings.HasPrefix(uri, "http://i0.hdslb.com") || strings.HasPrefix(uri, "http://i1.hdslb.com") || strings.HasPrefix(uri, "http://i2.hdslb.com") {
		uri = uri[19:]
	} else if strings.HasPrefix(uri, "https://i0.hdslb.com") || strings.HasPrefix(uri, "https://i1.hdslb.com") || strings.HasPrefix(uri, "https://i2.hdslb.com") {
		uri = uri[20:]
	}
	cover = uri
	if strings.HasPrefix(uri, "/bfs") {
		cover = "http://i0.hdslb.com" + uri
		return
	}
	if strings.Index(uri, "http://") == 0 {
		return
	}
	if len(uri) >= 10 && uri[:10] == "/templets/" {
		return
	}
	if strings.HasPrefix(uri, "group1") || strings.HasPrefix(uri, "/group1") {
		cover = "http://i0.hdslb.com/" + uri
		return
	}
	if pos := strings.Index(uri, "/uploads/"); pos != -1 && (pos == 0 || pos == 3) {
		cover = uri[pos+8:]
	}
	cover = strings.Replace(cover, "{IMG}", "", -1)
	cover = "http://i0.hdslb.com" + cover
	return
}

func CoverURLHTTPS(uri string) (cover string) {
	if strings.HasPrefix(uri, "http://") {
		cover = "https://" + uri[7:]
	} else {
		cover = uri
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

// InvalidChannel check source channel is not allow by config channel.
func InvalidChannel(plat int8, srcCh, cfgCh string) bool {
	return plat == PlatAndroid && cfgCh != "*" && cfgCh != srcCh
}

// Plat return plat by platStr or mobiApp
func Plat(mobiApp, device string) int8 {
	switch mobiApp {
	case "iphone", "iphone_b":
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

// Plat2 return plat all
func Plat2(mobiApp, device string) int8 {
	switch mobiApp {
	case "iphone":
		if device == "pad" {
			return PlatIPad
		}
		return PlatIPhone
	case "iphone_b":
		return PlatIPhoneB
	case "white":
		return PlatIPhone
	case "ipad":
		return PlatIPadHD
	case "android":
		return PlatAndroid
	case "win":
		return PlatWPhone
	case "android_G":
		return PlatAndroidG
	case "android_i":
		return PlatAndroidI
	case "android_b":
		return PlatAndroidB
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

// StatString Stat to string
// nolint:gomnd
func StatString(number int32, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	if number < 10000 {
		s = strconv.FormatInt(int64(number), 10) + suffix
		return
	}
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}

// StatString Stat to string
// nolint:gomnd
func Stat64String(number int64, suffix string) (s string) {
	if number == 0 {
		s = "-" + suffix
		return
	}
	if number < 10000 {
		s = strconv.FormatInt(number, 10) + suffix
		return
	}
	if number < 100000000 {
		s = strconv.FormatFloat(float64(number)/10000, 'f', 1, 64)
		return strings.TrimSuffix(s, ".0") + "万" + suffix
	}
	s = strconv.FormatFloat(float64(number)/100000000, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "亿" + suffix
}
