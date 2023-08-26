package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type RcmdVerticalOrigin struct {
	BaseCfgManager

	RcmdCommon
	SourceType string
	// 数据源
	UpListReqID kernel.RequestID
}
