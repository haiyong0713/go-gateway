package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type NewactAward struct {
	BaseCfgManager

	Sid   int64 //数据源id
	ReqID kernel.RequestID
}
