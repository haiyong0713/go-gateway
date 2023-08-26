package dynamicV2

import (
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

/* ************************************
	视频feed上滑: FromVideoNew
	视频feed下滑: FromVideoHistory
	视频页获取详情: FromDynBriefs
	视频页快速消费: FromVideoPersonal
*************************************** */

func (list *DynListRes) FromVideoNew(new *dyngrpc.VideoNewRsp, uid int64) {
	list.UpdateNum = new.UpdateNum
	list.HistoryOffset = new.HistoryOffset
	list.UpdateBaseline = new.UpdateBaseline
	list.HasMore = new.HasMore
	for _, item := range new.Dyns {
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
	if new.FoldInfo != nil {
		fold := &FoldInfo{}
		fold.FromFold(new.FoldInfo)
		list.FoldInfo = fold
	}
}

func (list *DynListRes) FromVideoHistory(history *dyngrpc.VideoHistoryRsp, uid int64) {
	list.HistoryOffset = history.HistoryOffset
	list.HasMore = history.HasMore
	for _, item := range history.Dyns {
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

func (list *DynListRes) FromDynBriefs(briefs *dyngrpc.DynBriefsRsp, uid int64) {
	for _, item := range briefs.Dyns {
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
}

// 视频页个人feed流列表信息
type VideoPersonal struct {
	HasMore    bool
	Offset     string
	Dynamics   []*Dynamic
	FoldInfo   *FoldInfo
	ReadOffset string
}

func (list *VideoPersonal) FromVideoPersonal(personal *dyngrpc.VideoPersonalRsp, uid int64) {
	list.HasMore = personal.HasMore
	list.Offset = personal.Offset
	list.ReadOffset = personal.ReadOffset
	if personal.FoldInfo != nil {
		fo := &FoldInfo{}
		fo.FromFold(personal.FoldInfo)
		list.FoldInfo = fo
	}
	for _, item := range personal.Dyns {
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
}

// VdUpListRsp 视频页-最近访问-up主列表
type VdUpListRsp struct {
	Items       []UpListItem `json:"items"`
	ModuleTitle string       `json:"module_title"`
	ShowAll     string       `json:"show_all"`
	Footprint   string       `json:"footprint"`
}

type UpListItem struct {
	HasUpdate       int   `json:"has_update"`
	UID             int64 `json:"uid"`
	IsReserveRecall bool  `json:"is_reserve_recall"` // 是否是预约召回
}
