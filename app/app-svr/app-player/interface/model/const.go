package model

import (
	"go-common/component/metadata/network"
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
	// PlatIpadHD is int8 for IpadHD
	PlatIpadHD = int8(9)
	// PlatAndroidB is int8 for Android Blue.
	PlatAndroidB = int8(10)
	// PlatIphoneB is int8 for Iphone Blue
	PlatIphoneB = int8(11)
	// DlDash is download dash
	DlDash = 2
	// DlFlv is download flv
	DlFlv = 1
	// qn hdr
	QnHDR = uint32(125)
	// qn dolbyHdr
	QnDolbyHDR = uint32(126)
	// qn 1080p+
	QnPPlus = uint32(112)
	// qn 1080高清
	Qn1080 = uint32(80)
	// qn 480
	Qn480 = uint32(32)
	// code H265
	CodeH265 = uint32(12)
	// code H264
	CodeH264 = uint32(7)
	// code av1
	CodeAV1 = uint32(13)

	//playurl attribute
	AttrIsHDR      = 0
	AttrIsDolbyHDR = 1
)

// IsAndroid check plat is android or ipad.
func IsAndroid(plat int8) bool {
	return plat == PlatAndroid
}

// IsIOS check plat is iphone or ipad.
func IsIOS(plat int8) bool {
	return plat == PlatIPad || plat == PlatIPhone || plat == PlatIPadI || plat == PlatIPhoneI
}

// IsIphone check plat is iphone.
func IsIphone(plat int8) bool {
	return plat == PlatIPhone
}

// IsIPad check plat is pad.
func IsIPad(plat int8) bool {
	return plat == PlatIPad
}

// IsIPadHD check plat is padHD.
func IsIPadHD(plat int8) bool {
	return plat == PlatIpadHD
}

// IsOverseas is overseas
func IsOverseas(plat int8) bool {
	return plat == PlatAndroidI || plat == PlatIPhoneI || plat == PlatIPadI
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
	case "iphone_b":
		return PlatIphoneB
	}
	return PlatIPhone
}

func IsVipQuality(qn uint32) bool {
	return qn >= QnPPlus
}

func IsSubtitleQuality(qn uint32) bool {
	return qn >= Qn1080
}

func IsLoginQuality(qn uint32) bool {
	return qn > Qn480
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
