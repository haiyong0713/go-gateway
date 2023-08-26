package elec

import (
	"encoding/json"

	upr "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	uprmdl "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank/model"
)

type Info struct {
	Start   int64           `json:"start,omitempty"`
	Show    bool            `json:"show"`
	Total   int             `json:"total,omitempty"`
	Count   int             `json:"count,omitempty"`
	State   int             `json:"state,omitempty"`
	List    json.RawMessage `json:"list,omitempty"`
	User    json.RawMessage `json:"user,omitempty"`
	ElecSet json.RawMessage `json:"elec_set,omitempty"`
}

type NewInfo struct {
	Show    bool                           `json:"show"`
	Total   int64                          `json:"total,omitempty"`
	Count   int64                          `json:"count,omitempty"`
	ElecNum int                            `json:"elec_num"` //充电服务不下发，兼容客户端老版本
	List    []*uprmdl.RankElecElementProto `json:"list,omitempty"`
	ElecSet *upr.BatterySet                `json:"elec_set,omitempty"`
	//开通状态 -1:未开通 1:开通老充电 2:开通新充电，6.88之后使用
	State int64 `json:"state"`
	//充电标题，例：为TA充电/已为TA充电
	UpowerTitle string `json:"upower_title"`
	//跳转链接，老充电无，新充电有
	UpowerJumpUrl string `json:"upower_jump_url"`
	//充电icon链接
	UpowerIconUrl string `json:"upower_icon_url"`
	//充电的一些状态
	UpowerState upr.UpowerState `json:"upower_state"`
	//充电按钮的配置映射
	UpowerButtonMap map[int64]*upr.UpowerButton `json:"upower_button_map"`
}

func FormatElec(in *upr.UPRankWithPanelReply) *NewInfo {
	rly := &NewInfo{
		Show:            in.GetShow(),
		Total:           in.GetCountUPTotalElec(),
		Count:           in.GetCount(),
		List:            in.GetList(),
		ElecSet:         in.GetElecSet(),
		State:           in.GetState(),
		UpowerTitle:     in.GetUpowerTitle(),
		UpowerJumpUrl:   in.GetUpowerJumpUrl(),
		UpowerIconUrl:   in.GetUpowerIconUrl(),
		UpowerState:     in.UpowerState,
		UpowerButtonMap: in.UpowerButtonMap,
	}
	return rly
}
