package model

import cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

// archive forbidden
const (
	_NoDynamic = "50" // 分区动态禁止
	_NoWeb     = "51" // 禁止网页端输出
	_NoMobile  = "52" // 禁止客户端列表

)

type ArchiveFlowControlMsg struct {
	Router string `json:"router"`
	Data   *struct {
		Oid          int64 `json:"oid"`
		NewFlowState struct {
			NoWeb    int32 `json:"noweb"`
			NoMobile int32 `json:"nomobile"`
		} `json:"new_flow_state"`
	} `json:"data"`
}

type ArcForbidden struct {
	NoDynamic bool
	NoWeb     bool
	NoMobile  bool
}

func (a *ArcForbidden) AllowShow() bool {
	return !a.NoWeb && !a.NoMobile
}

func ItemToArcForbidden(info *cfcgrpc.FlowCtlInfoV2Reply) *ArcForbidden {
	acrForbidden := &ArcForbidden{}
	if info == nil || len(info.Items) == 0 {
		return acrForbidden
	}
	for _, item := range info.Items {
		if item == nil {
			continue
		}
		switch item.Key {
		case _NoDynamic:
			if item.Value == 1 {
				acrForbidden.NoDynamic = true
			}
		case _NoWeb:
			if item.Value == 1 {
				acrForbidden.NoWeb = true
			}
		case _NoMobile:
			if item.Value == 1 {
				acrForbidden.NoMobile = true
			}
		}
	}
	return acrForbidden
}
