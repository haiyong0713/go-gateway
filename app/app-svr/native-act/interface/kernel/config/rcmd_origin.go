package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type RcmdOrigin struct {
	BaseCfgManager

	RcmdCommon
	SourceType string
	// 数据源
	UpListReqID kernel.RequestID
	// 排行榜
	DisplayRankScore bool
	MixExtReqID      kernel.RequestID
	RankRstReqID     kernel.RequestID
}
