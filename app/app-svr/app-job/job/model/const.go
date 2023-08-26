package model

import "go-common/library/conf/env"

const (
	TypeArchive    = "archive"
	TypeArchiveHis = "archive_his"

	TypeForView  = "view"
	TypeForDm    = "dm"
	TypeForReply = "reply"
	TypeForFav   = "fav"
	TypeForCoin  = "coin"
	TypeForShare = "share"
	TypeForLike  = "like"
	TypeForRank  = "rank"

	// PlatAndroid is for android.
	PlatAndroid = int32(0)
	// PlatIPhone is for iphone.
	PlatIPhone = int32(1)
	// PlatIPad is for ipad.
	PlatIPad = int32(2)
	// PlatIPadHD is for ipadHD.
	PlatIPadHD = int32(20)
)

// env sh001 run
func EnvRun() (res bool) {
	var _zone = "sh001"
	return env.Zone == _zone
}

// PlatToMobiApp change plat to mobiApp & device
func PlatToMobiApp(plat int32) (mobiApp, device string) {
	switch plat {
	case PlatAndroid:
		return "android", ""
	case PlatIPhone:
		return "iphone", "phone"
	case PlatIPad:
		return "iphone", "pad"
	case PlatIPadHD:
		return "ipad", "pad"
	}
	return "", ""
}
