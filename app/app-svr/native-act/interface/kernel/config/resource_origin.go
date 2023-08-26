package config

import (
	"go-gateway/app/app-svr/native-act/interface/kernel"
)

type ResourceOrigin struct {
	BaseCfgManager

	ResourceCommon
	OriginType         int64
	ShowNum            int64
	Wid                int32
	TabID              int64
	TabList            []*SortListItem
	RoomsByActIdReqID  kernel.RequestID
	ProductDetailReqID kernel.RequestID
	SourceDetailReqID  kernel.RequestID
}
