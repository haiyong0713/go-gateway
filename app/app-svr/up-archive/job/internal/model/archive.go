package model

import (
	"encoding/json"

	"go-gateway/app/app-svr/up-archive/service/api"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

const (
	ActionInsert = "insert"
	ActionUpdate = "update"
)

type CanalMsg struct {
	Action string          `json:"action"`
	Table  string          `json:"table"`
	New    json.RawMessage `json:"new"`
	Old    json.RawMessage `json:"old"`
}

type ArchiveCanalMsg struct {
	Action string      `json:"action"`
	Table  string      `json:"table"`
	New    *ArchiveSub `json:"new"`
	Old    *ArchiveSub `json:"old"`
}

type ArchiveSub struct {
	Aid         int64  `json:"aid"`
	Mid         int64  `json:"mid"`
	PubTime     string `json:"pubtime"`
	State       int    `json:"state"`
	Copyright   int8   `json:"copyright"`
	Attribute   int32  `json:"attribute"`
	AttributeV2 int32  `json:"attribute_v2"`
	RedirectURL string `json:"redirect_url"`
}

type ArchiveFlowControlMsg struct {
	Router string `json:"router"`
	Data   *struct {
		Oid          int64 `json:"oid"`
		NewFlowState struct {
			NoSpace   int32 `json:"no_space"`
			UpNoSpace int32 `json:"up_no_space"`
		} `json:"new_flow_state"`
	} `json:"data"`
}

func (a *ArchiveSub) IsNormal() bool {
	return a.State >= api.StateOpen
}

// 审核空间禁止 no_space
func (a *ArchiveSub) IsAllowed(fItem []*cfcgrpc.ForbiddenItem) bool {
	var noSpace bool
	for _, item := range fItem {
		switch item.Key {
		case "no_space":
			noSpace = item.Value == 1
		default:
		}
	}
	return AttrVal(a.Attribute, api.AttrBitIsPUGVPay) == api.AttrNo && !noSpace
}

func (a *ArchiveSub) IsUpNoSpace(fItem []*cfcgrpc.ForbiddenItem) bool {
	var upNoSpace bool
	for _, item := range fItem {
		switch item.Key {
		case "up_no_space":
			upNoSpace = item.Value == 1
		default:
		}
	}
	return upNoSpace
}

func (a *ArchiveSub) IsStory() bool {
	return AttrVal(a.Attribute, api.AttrBitIsPGC) == api.AttrNo && AttrVal(a.Attribute, api.AttrBitSteinsGate) == api.AttrNo &&
		a.RedirectURL == "" && AttrVal(a.AttributeV2, api.AttrBitV2Pay) == api.AttrNo //非付费稿件= 稿件属性位为非付费
}
