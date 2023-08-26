package jsonwebcard

import (
	"go-common/component/metadata/device"
	"go-common/component/metadata/restriction"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
)

type MetaContext struct {
	Restriction *restriction.Restriction
	Device      *device.Device
	Mid         int64
	IP          string
	LocalTime   int32
	Config      *Config
}

type Config struct {
	DynCmtTopicControl map[int64]*topiccardmodel.DynCmtMeta
	ItemFromControl    map[int64]string
	HiddenAttached     map[int64]bool // 隐式关联
}
