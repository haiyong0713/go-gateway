package model

import (
	"encoding/json"

	payrank "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
)

// ElecInfo elec info.
type ElecInfo struct {
	Start   int64               `json:"start"`
	Show    bool                `json:"show"`
	Total   int64               `json:"total"`
	Count   int64               `json:"count"`
	State   int                 `json:"state"`
	List    []*ElecUserList     `json:"list,omitempty"`
	User    json.RawMessage     `json:"user,omitempty"`
	ElecSet *payrank.BatterySet `json:"elec_set,omitempty"`
}

type ElecUserList struct {
	Mid        int64       `json:"mid"`
	PayMid     int64       `json:"pay_mid"`
	Rank       int64       `json:"rank"`
	Uname      string      `json:"uname"`
	Avatar     string      `json:"avatar"`
	Message    string      `json:"message"`
	MsgDeleted int64       `json:"msg_deleted"`
	VipInfo    ElecVipInfo `json:"vip_info"`
	TrendType  uint32      `json:"trend_type"`
}

type ElecVipInfo struct {
	VipType    int32 `json:"vipType"`
	VipDueMsec int64 `json:"vipDueMsec"`
	VipStatus  int32 `json:"vipStatus"`
}
