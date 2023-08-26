package model

import (
	"errors"
	"fmt"

	appHistoryApi "go-gateway/app/app-svr/app-interface/interface-legacy/api/history"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/conf"
)

var (
	ErrNoHistoryRecord = errors.New("GetPlayHistoryByIds: no history found")
)

type (
	PlayHistory struct {
		// 稿件类型 UGC/OGV
		ArcType int32
		// aid 或者 sid
		Oid int64
		// 最后播放的cid或者epid
		LastPlay int64
		// 播放进度 秒
		Progress   int64
		DeviceType int64
		Timestamp  int64
	}
)

func (ph PlayHistory) Hash() string {
	if ph.Oid == 0 {
		return ""
	}
	return fmt.Sprintf("%d-%d", ph.ArcType, ph.Oid)
}

func (ph PlayHistory) ToV1PlayItem() *v1.PlayItem {
	return &v1.PlayItem{
		ItemType: ph.ArcType,
		Oid:      ph.Oid,
	}
}

const (
	HistoryDeviceUnknown    int64 = 0
	HistoryDeviceIPhone     int64 = 1
	HistoryDevicePC         int64 = 2
	HistoryDeviceAndroid    int64 = 3
	HistoryDeviceIPad       int64 = 4
	HistoryDeviceWP8        int64 = 5
	HistoryDeviceUWP        int64 = 6
	HistoryDeviceH5         int64 = 7
	HistoryDeviceAndroidCar int64 = 8 // 车载
	HistoryDeviceAndroidIoT int64 = 9 // 物联网
	HistoryDeviceAndroidPad int64 = 10
	HistoryDeviceAndroidTV  int64 = 33
)

func historyDeviceType2Icon(devType int64) (appHistoryApi.DT, string) {
	switch devType {
	case HistoryDeviceIPhone, HistoryDeviceAndroid, HistoryDeviceWP8:
		return appHistoryApi.DT_Phone, conf.C.Res.HistoryIcon.Phone
	case HistoryDeviceIPad:
		return appHistoryApi.DT_Pad, conf.C.Res.HistoryIcon.Pad
	case HistoryDeviceAndroidPad:
		return appHistoryApi.DT_AndPad, conf.C.Res.HistoryIcon.Pad
	case HistoryDevicePC, HistoryDeviceUWP, HistoryDeviceH5:
		return appHistoryApi.DT_PC, conf.C.Res.HistoryIcon.PC
	case HistoryDeviceAndroidCar:
		return appHistoryApi.DT_Car, conf.C.Res.HistoryIcon.Car
	case HistoryDeviceAndroidTV:
		return appHistoryApi.DT_TV, conf.C.Res.HistoryIcon.TV
	case HistoryDeviceAndroidIoT:
		return appHistoryApi.DT_IoT, conf.C.Res.HistoryIcon.Iot
	default:
		return appHistoryApi.DT_Unknown, conf.C.Res.HistoryIcon.PC
	}
}

func (ph PlayHistory) ToAppHistoryDeviceType() *appHistoryApi.DeviceType {
	ret := new(appHistoryApi.DeviceType)
	ret.Type, ret.Icon = historyDeviceType2Icon(ph.DeviceType)
	return ret
}
