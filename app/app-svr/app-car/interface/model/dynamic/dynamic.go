package dynamic

import (
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	followgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
)

const (
	RefreshNew            = 0
	RefreshHistory        = 1
	DynTypeVideo          = 8
	DynTypeBangumi        = 512
	DynTypePGCBangumi     = 4097
	DynTypePGCMovie       = 4098
	DynTypePGCGuoChuang   = 4100
	DynTypePGCDocumentary = 4101
)

type DynamicParam struct {
	model.DeviceInfo
	LocalTime      int    `form:"local_time" default:"8"`
	UpdateBaseline string `form:"update_baseline" default:"0"`
	Offset         string `form:"offset"`
	Page           int64  `form:"pn" default:"1" validate:"min=1"`
	RefreshType    int    `form:"refresh_type"`
	AssistBaseline string `form:"assist_baseline" default:"20"`
	ParamStr       string `form:"param"`
	FromType       string `form:"from_type"`
	TopParamStr    string `form:"top_param_str"`
}

// 动态列表资源
type DynVideoListRes struct {
	UpdateNum      int64      `json:"update_num"`
	HistoryOffset  string     `json:"history_offset"`
	UpdateBaseline string     `json:"update_baseline"`
	HasMore        bool       `json:"has_more"`
	Dynamics       []*Dynamic `json:"dynamics"`
}

type Dynamic struct {
	DynamicID int64 `json:"dynamic_id"`
	Type      int64 `json:"type"`
	Rid       int64 `json:"rid"`
	UID       int64 `json:"uid"`
	Ctime     int64 `json:"ctime"`
}

func (list *DynVideoListRes) FromVideoNew(new *dyngrpc.VideoNewRsp) {
	list.UpdateNum = new.UpdateNum
	list.HistoryOffset = new.HistoryOffset
	list.UpdateBaseline = new.UpdateBaseline
	list.HasMore = new.HasMore
	for _, item := range new.Dyns {
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
}

func (list *DynVideoListRes) FromVideoHistory(history *dyngrpc.VideoHistoryRsp) {
	list.HistoryOffset = history.HistoryOffset
	list.HasMore = history.HasMore
	for _, item := range history.Dyns {
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
}

func (list *DynVideoListRes) FromVideoPersonal(history *dyngrpc.VideoPersonalRsp) {
	list.HistoryOffset = history.Offset
	list.HasMore = history.HasMore
	for _, item := range history.Dyns {
		dynTmp := &Dynamic{}
		dynTmp.FromDynamic(item)
		list.Dynamics = append(list.Dynamics, dynTmp)
	}
}

func (dyn *Dynamic) FromDynamic(d *dyncommongrpc.DynBrief) {
	dyn.DynamicID = d.DynId
	dyn.Type = d.Type
	dyn.Rid = d.Rid
	dyn.UID = d.Uid
	dyn.Ctime = d.Ctime
}

func GetAttentionsParams(mid int64, follows []*relationgrpc.FollowingReply, pgcs []*followgrpc.FollowSeasonProto) *dyncommongrpc.AttentionInfo {
	res := &dyncommongrpc.AttentionInfo{}
	for _, item := range follows {
		res.AttentionList = append(res.AttentionList, &dyncommongrpc.Attention{
			Uid:       item.Mid,
			UidType:   1,
			IsSpecial: Int32ToBool(item.Special),
		})
	}
	for _, item := range pgcs {
		res.AttentionList = append(res.AttentionList, &dyncommongrpc.Attention{
			Uid:     int64(item.SeasonId),
			UidType: 2,
		})
	}
	// 赋自己
	res.AttentionList = append(res.AttentionList, &dyncommongrpc.Attention{
		Uid:       mid,
		UidType:   1,
		IsSpecial: false,
	})
	return res
}

func Int32ToBool(v int32) bool {
	return v != 0
}

func (d *Dynamic) DynamicCardChange() *ai.Item {
	r := &ai.Item{
		ID:       d.Rid,
		DynCtime: d.Ctime,
	}
	switch d.Type {
	case DynTypeVideo:
		r.Goto = model.GotoAv
	case DynTypeBangumi, DynTypePGCBangumi, DynTypePGCMovie, DynTypePGCGuoChuang, DynTypePGCDocumentary:
		r.Goto = model.GotoPGC
	}
	return r
}

func (d *Dynamic) DynamicCardChangeV2() *ai.Item {
	r := &ai.Item{
		ID: d.Rid,
	}
	switch d.Type {
	case DynTypeVideo:
		r.Goto = model.GotoAv
	case DynTypeBangumi, DynTypePGCBangumi, DynTypePGCMovie, DynTypePGCGuoChuang, DynTypePGCDocumentary:
		r.Goto = model.GotoPGCEp
	}
	return r
}

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
