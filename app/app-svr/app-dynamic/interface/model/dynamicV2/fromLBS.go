package dynamicV2

import (
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

func (list *DynListRes) FromLBS(dyn *dyngrpc.LbsPoiListRsp) {
	list.HistoryOffset = dyn.Offset
	if dyn.HasMore == 1 {
		list.HasMore = true
	}
	for _, item := range dyn.Dyns {
		if item == nil || item.Type == 0 {
			continue
		}
		if item.Type == 1 && item.Origin == nil {
			continue
		}
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
}
