package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type EditorOrigin struct {
	BaseCfgManager

	Position          Position
	DisplayMoreButton bool
	BgColor           string
	RdbType           int64
	IsFeed            bool
	PageSize          int64
	PageArcsReqID     kernel.RequestID
	GetHisReqID       kernel.RequestID
	MixExtsReqID      kernel.RequestID
	MixExtReqID       kernel.RequestID
	RankRstReqID      kernel.RequestID
	SelSerieReqID     kernel.RequestID
	ChannelFeedReqID  kernel.RequestID
}
