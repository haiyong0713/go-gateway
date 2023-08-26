package model

import (
	"strconv"

	"go-common/component/metadata/device"
	"go-common/component/metadata/restriction"
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"

	dynamicCommon "git.bilibili.co/bapis/bapis-go/dynamic/common"
)

type GeneralParam struct {
	Restriction *restriction.Restriction
	Device      *device.Device
	Mid         int64
	IP          string
	LocalTime   int32
	Source      string
	FromSpmid   string
}

func (g *GeneralParam) SetLocalTime(lo int32) {
	g.LocalTime = lo
	if g.LocalTime < -12 || g.LocalTime > 14 {
		g.LocalTime = 8
	}
}

func (g *GeneralParam) SetFromSpmid(spmid string) {
	g.FromSpmid = spmid
}

func (g *GeneralParam) ToDynCmnMetaData() *dynamicCommon.CmnMetaData {
	return &dynamicCommon.CmnMetaData{
		Build:    g.GetBuildStr(),
		Platform: g.GetPlatform(),
		MobiApp:  g.GetMobiApp(),
		Device:   g.GetDevice(),
		Version:  g.GetVersion(),
		Buvid:    g.GetBuvid(),
	}
}

type DynRawList struct {
	List []*DynRawItem
}

type DynRawItem struct {
	Item *dynamicapi.DynamicItem
}

func (g *GeneralParam) GetDisableRcmdInt() int {
	if g.Restriction == nil {
		return 0
	}
	if g.Restriction.DisableRcmd {
		return 1
	}
	return 0
}

func (g *GeneralParam) GetTeenagerInt() int {
	if g.Restriction == nil {
		return 0
	}
	if g.Restriction.IsTeenagers {
		return 1
	}
	return 0
}

func (g *GeneralParam) GetBuvid() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.Buvid
}

func (g *GeneralParam) GetNetWork() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.Network
}

func (g *GeneralParam) GetMobiApp() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.RawMobiApp
}

func (g *GeneralParam) GetBuild() int64 {
	if g.Device == nil {
		return 0
	}
	return g.Device.Build
}

func (g *GeneralParam) GetBuildStr() string {
	build := g.GetBuild()
	if build != 0 {
		return strconv.FormatInt(build, 10)
	}
	return ""
}

func (g *GeneralParam) GetPlatform() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.RawPlatform
}

func (g *GeneralParam) GetDevice() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.Device
}

func (g *GeneralParam) GetVersion() string {
	if g.Device == nil {
		return ""
	}
	return g.Device.VersionName
}

func (g *GeneralParam) IsPadHD() bool {
	if g.Device == nil {
		return false
	}
	return g.Device.RawMobiApp == "ipad"
}

func (g *GeneralParam) IsPad() bool {
	return g.GetMobiApp() == "iphone" && g.GetDevice() == "pad"
}

func (g *GeneralParam) IsAndroidPick() bool {
	return g.GetMobiApp() == "android"
}

func (g *GeneralParam) IsIPhonePick() bool {
	return g.GetMobiApp() == "iphone" && g.GetDevice() == "phone"
}

func (g *GeneralParam) IsAndroidHD() bool {
	if g.Device == nil {
		return false
	}
	return g.Device.RawMobiApp == "android_hd"
}

func (g *GeneralParam) IsOverseas() bool {
	if g.Device == nil {
		return false
	}
	return g.Device.RawMobiApp == "android_i" || g.Device.RawMobiApp == "iphone_i" || g.Device.RawMobiApp == "ipad_i"
}
