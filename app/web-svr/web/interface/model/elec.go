package model

import "encoding/json"

// ElecShow elec show
type ElecShow struct {
	ShowInfo   *ShowInfo       `json:"show_info"`         // 开关
	AvCount    int             `json:"av_count"`          // 前端没使用，仅保留
	Count      int64           `json:"count"`             // 前端没使用，仅保留
	TotalCount int64           `json:"total_count"`       // 历史对up主充电次数
	SpecialDay int             `json:"special_day"`       // 前端没使用，仅保留
	DisplayNum int             `json:"display_num"`       // 前端没使用，仅保留
	AvList     json.RawMessage `json:"av_list,omitempty"` // 前端没使用，仅保留
	AvUser     json.RawMessage `json:"av_user,omitempty"` // 前端没使用，仅保留
	List       []*ElecUserList `json:"list,omitempty"`    // 充电人列表
	User       json.RawMessage `json:"user,omitempty"`    // 前端没使用，仅保留
}

// ShowInfo show info
type ShowInfo struct {
	Show    bool   `json:"show"`
	State   int8   `json:"state"`    // -1 未开通 1 老 2新
	Title   string `json:"title"`    // button的文案
	JumpUrl string `json:"jump_url"` // 新充电跳链
	Icon    string `json:"icon"`     // 新充电按钮icon
}

type ElecUserList struct {
	Mid        int64       `json:"mid"`
	PayMid     int64       `json:"pay_mid"`
	Rank       int64       `json:"rank"`
	Uname      string      `json:"uname"`
	Avatar     string      `json:"avatar"`
	Message    string      `json:"message"`
	MsgDeleted int         `json:"msg_deleted"`
	VipInfo    ElecVipInfo `json:"vip_info"`
	TrendType  uint32      `json:"trend_type"`
}

type ElecVipInfo struct {
	VipType    int32 `json:"vipType"`
	VipDueMsec int64 `json:"vipDueMsec"`
	VipStatus  int32 `json:"vipStatus"`
}
