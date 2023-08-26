package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type ResourceAct struct {
	BaseCfgManager

	ResourceCommon
	SortType      int64
	SortList      []*SortListItem
	ActLikesReqID kernel.RequestID
}
