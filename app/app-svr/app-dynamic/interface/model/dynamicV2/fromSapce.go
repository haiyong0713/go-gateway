package dynamicV2

import (
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

func (list *DynListRes) FromSpaceHistory(history *dyngrpc.SpaceHistoryRsp, uid int64) {
	list.HistoryOffset = history.HistoryOffset
	list.HasMore = history.HasMore
	for _, item := range history.Dyns {
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
	if history.FoldInfo != nil {
		fold := &FoldInfo{}
		fold.FromFold(history.FoldInfo)
		list.FoldInfo = fold
	}
}
