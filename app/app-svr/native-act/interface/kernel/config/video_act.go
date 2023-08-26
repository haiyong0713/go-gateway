package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type VideoAct struct {
	BaseCfgManager

	VideoCommon
	SortType      int64
	SortList      []*SortListItem
	ActLikesReqID kernel.RequestID
}
