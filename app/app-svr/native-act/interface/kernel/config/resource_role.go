package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type ResourceRole struct {
	BaseCfgManager

	ResourceCommon
	ShowNum       int64 //当前页面展示数量
	RelInfosReqID kernel.RequestID
}
