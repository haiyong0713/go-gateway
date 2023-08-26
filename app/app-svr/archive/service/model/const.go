package model

import (
	"fmt"

	"go-common/component/metadata/network"
)

const (
	//playurl attribute
	AttrIsHDR      = 0
	AttrIsDolbyHDR = 1

	//PlayerIOSBuild  is (player_info 修改层级版本限制)
	PlayerIOSBuild = 8400
	//PlayerIOSBBuild  is
	PlayerIOSBBuild = 7370
	//PlayerIPadHDBuild  is
	PlayerIPadHDBuild = 12080
	//PlayerAndroidBuild  is
	PlayerAndroidBuild = 5385000
	//PlayerAndroidIBuild  is
	PlayerAndroidIBuild = 2020000

	//QnIOSBuild is (540开始取消写死qn=480p)
	QnIOSBuild = 8430
	//QnAndroidBuild is
	QnAndroidBuild = 5395000

	//ArcsWithPlayurl from
	PlayurlFromStory = "story"

	//Qn desc
	QnHDR      = 125
	Qn1080Plus = 112
	Qn1080     = 80
	//原先720p和720p60合并为新的720p了，qn采用64，定为非会员清晰度
	QnFlv720   = 64
	Qn480      = 32
	QnDolbyHDR = 126

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
	// PlatIPhoneB is int8 for Iphone Blue
	PlatIPhoneB = int8(10)
	// PlatIpadHD is int8 for IpadHD
	PlatIpadHD = int8(20)
)

func RedirectKey(aid int64) string {
	return fmt.Sprintf("redirect_%d", aid)
}

func DescKeyV2(aid int64) string {
	return fmt.Sprintf("desc_v2_%d", aid)
}

func ArcKey(aid int64) string {
	return fmt.Sprintf("a3p_%d", aid)
}

func SimpleArcKey(aid int64) string {
	return fmt.Sprintf("sac_%d", aid)
}

func InternalArcKey(aid int64) string {
	return fmt.Sprintf("innera_%d", aid)
}

func PageKey(aid int64) string {
	return fmt.Sprintf("psb_%d", aid)
}

func VideoKey(aid, cid int64) string {
	return fmt.Sprintf("psb_%d_%d", aid, cid)
}

// 包含高清缩略图
func NewVideoShotKey(cid int64) string {
	return fmt.Sprintf("nvst_%d", cid)
}

// IsAndroid check plat is android or ipad.
func IsAndroid(plat int8) bool {
	return plat == PlatAndroid
}

// IsIOS check plat is iphone or ipad.
func IsIOS(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone || plat == PlatIPadI || plat == PlatIPhoneI
}

// IsOverseas is overseas
func IsOverseas(plat int8) bool {
	return plat == PlatAndroidI || plat == PlatIPhoneI || plat == PlatIPadI
}

// IsAndroidI check plat is android_i.
func IsAndroidI(plat int8) bool {
	return plat == PlatAndroidI
}

// IsIPhoneB check plat is iphone_b
func IsIPhoneB(plat int8) bool {
	return plat == PlatIPhoneB
}

// IsIPadHD check plat is iPadHD
func IsIPadHD(plat int8) bool {
	return plat == PlatIpadHD
}

func IsIPad(plat int8) bool {
	return plat == PlatIpadHD || plat == PlatIPad || plat == PlatIPadI
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
	case "iphone_b":
		return PlatIPhoneB
	case "android_b":
		return PlatAndroidB
	}
	return PlatIPhone
}

// PlayerInfoNew is
func PlayerInfoNew(plat int8, build int64) bool {
	return (IsIOS(plat) && build > PlayerIOSBuild) || (IsIPhoneB(plat) && build > PlayerIOSBBuild) || (IsAndroid(plat) && build > PlayerAndroidBuild) ||
		(IsIPadHD(plat) && build > PlayerIPadHDBuild) || (IsAndroidI(plat) && build > PlayerAndroidIBuild)
}

func TrafficFree(xTfIsp string) (netType, tfType int32) {
	switch xTfIsp {
	case "ct":
		return int32(network.TypeCellular), int32(network.TFTCard)
	case "cu":
		return int32(network.TypeCellular), int32(network.TFUCard)
	case "cm":
		return int32(network.TypeCellular), int32(network.TFCCard)
	}
	return int32(network.TypeUnknown), int32(network.TypeUnknown)
}
