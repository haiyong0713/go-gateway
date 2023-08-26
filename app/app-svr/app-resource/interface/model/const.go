package model

import (
	"context"
	"fmt"
	"strings"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/tab"
	feature "go-gateway/app/app-svr/feature/service/sdk"
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
	// PlatAndroidB is int8 for Android Blue.
	PlatAndroidB = int8(9)
	// PlatIPhoneB is int8 for Ios Blue
	PlatIPhoneB = int8(10)
	// PlatBilistudio is int8 for bilistudio
	PlatBilistudio = int8(11)
	// PlatAndroidTVYST is int8 for AndroidTV_YST Global.
	PlatAndroidTVYST = int8(12)
	// PlatIPadHD is int8 for ipad.
	PlatIPadHD = int8(20)
	// PlatHTML5PC is int8 html5 pc
	PlatHTML5PC = int8(31)
	// PlatHTML5Mobile is int8 html5 mobile
	PlatHTML5Mobile = int8(32)
	// PlatFlash is int8 html5 pc
	PlatFlash      = int8(33)
	PlatAndroidCar = int8(35)
	//PlatPcClient is int8 for pc
	PlatPcClient = int8(40)

	// PlatAndroidHD is int8 for android_hd
	PlatAndroidHD = int8(90)

	GotoAv         = "av"
	GotoWeb        = "web"
	GotoBangumi    = "bangumi"
	GotoSp         = "sp"
	GotoLive       = "live"
	GotoGame       = "game"
	GotoPegasusTab = "pegasus"
	GotoActPageTab = "act_page"
	GotoPopGame    = "pop_game"
	GotoArticle    = "article"
	GotoPlaylist   = "playlist"
	GotoAudio      = "audio"
	GotoSong       = "song"
	GotoClip       = "clip"
	GotoAlbum      = "album"
	GotoDaily      = "daily"

	BubbleNoExist = -1
	BubblePushing = 0
	BubblePushed  = 1

	ModuleTop     = "top"
	ModuleTab     = "tab"
	ModuleBottom  = "bottom"
	ModuleTopMore = "top_more"
)

var (
	PegasusHandler = func(m *tab.Menu) func(uri string) string {
		return func(uri string) string {
			if m == nil {
				return uri
			}
			if m.Name != "" {
				return fmt.Sprintf("%s?name=%s", uri, m.Name)
			}
			return uri
		}
	}
)

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
	case "iphone_b":
		return PlatIPhoneB
	case "android_tv":
		return PlatAndroidTV
	case "android_tv_yst":
		return PlatAndroidTVYST
	case "bilistudio":
		return PlatBilistudio
	case "biliLink":
		return PlatIPhone
	case "html5_pc":
		return PlatHTML5PC
	case "html5_mobile":
		return PlatHTML5Mobile
	case "flash":
		return PlatFlash
	case "android_car":
		return PlatAndroidCar
	case "pc_client":
		return PlatPcClient
	}
	return PlatIPhone
}

// Plat return plat by platStr or mobiApp
func Plat2(mobiApp, device string) int8 {
	switch mobiApp {
	case "iphone":
		if device == "pad" {
			return PlatIPad
		}
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
	case "iphone_b":
		return PlatIPhoneB
	case "android_tv":
		return PlatAndroidTV
	case "android_tv_yst":
		return PlatAndroidTVYST
	case "bilistudio":
		return PlatBilistudio
	case "biliLink":
		return PlatIPhone
	case "html5_pc":
		return PlatHTML5PC
	case "html5_mobile":
		return PlatHTML5Mobile
	case "flash":
		return PlatFlash
	case "android_hd":
		return PlatAndroidHD
	}
	return PlatIPhone
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

// IsAndroid check plat is android or ipad.
func IsAndroid(plat int8) bool {
	return plat == PlatAndroid || plat == PlatAndroidG || plat == PlatAndroidB || plat == PlatAndroidI ||
		plat == PlatBilistudio || plat == PlatAndroidTV || plat == PlatAndroidTVYST
}

// IsIOS check plat is iphone or ipad.
func IsIOS(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone || plat == PlatIPadI || plat == PlatIPhoneI || plat == PlatIPhoneB
}

// FillURI deal app schema.
func FillURI(gt, param string, f func(uri string) string) (uri string) {
	if param == "" {
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
		uri = "bilibili://game/" + param
	case GotoSp:
		uri = "bilibili://splist/" + param
	case GotoWeb:
		uri = param
	case GotoPegasusTab:
		uri = "bilibili://pegasus/op/" + param
	case GotoActPageTab:
		uri = "bilibili://following/home_activity_tab/" + param
	case GotoPopGame:
		uri = fmt.Sprintf("bilibili://game_center/detail?id=%s&sourceType=adPut", param)
	case GotoArticle:
		uri = "bilibili://article/" + param
	case GotoClip:
		uri = "bilibili://clip/" + param
	case GotoAlbum:
		uri = "bilibili://album/" + param
	case GotoAudio:
		uri = "bilibili://music/menu/detail/" + param
	case GotoSong:
		uri = "bilibili://music/detail/" + param
	case GotoDaily:
		uri = "bilibili://pegasus/list/daily/" + param
	}
	if f != nil {
		uri = f(uri)
	}
	return
}

// MobiAPPBuleChange
func MobiAPPBuleChange(mobiApp string) string {
	switch mobiApp {
	case "android_b":
		return "android"
	case "iphone_b":
		return "iphone"
	}
	return mobiApp
}

func URLHTTPS(uri string) (url string) {
	if strings.HasPrefix(uri, "http://") {
		url = "https://" + uri[7:]
	} else {
		url = uri
	}
	return
}

// IsOverseas is overseas
func IsOverseas(plat int8) bool {
	return plat == PlatAndroidI || plat == PlatIPhoneI || plat == PlatIPadI
}

func PlatAPPBuleChange(plat int8) int8 {
	switch plat {
	case PlatAndroidB:
		return PlatAndroid
	case PlatIPhoneB:
		return PlatIPhone
	}
	return plat
}

func SplashUseBaseDefaultConfig(c context.Context, config *conf.Feature, mobiApp string, build int) bool {
	return splashLessThan610(c, config, mobiApp, build)
}

func SplashRemoveFull(c context.Context, config *conf.Feature, mobiApp string, build int) bool {
	return splashLessThan610(c, config, mobiApp, build)
}

func splashLessThan610(c context.Context, config *conf.Feature, mobiApp string, build int) bool {
	return feature.GetBuildLimit(c, config.FeatureBuildLimit.Splash610, &feature.OriginResutl{
		MobiApp:    mobiApp,
		Build:      int64(build),
		BuildLimit: (mobiApp == "android" && build < 6100000) || (mobiApp == "iphone" && build <= 10270),
	})
}
