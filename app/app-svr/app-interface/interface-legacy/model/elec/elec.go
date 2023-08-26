package elec

import (
	"context"
	"encoding/json"

	upr "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	uprmdl "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank/model"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

type Info struct {
	Start   int64           `json:"start,omitempty"`
	Show    bool            `json:"show"`
	Total   int             `json:"total,omitempty"`
	Count   int             `json:"count,omitempty"`
	ElecNum int             `json:"elec_num"`
	State   int             `json:"state,omitempty"`
	List    json.RawMessage `json:"list,omitempty"`
	User    json.RawMessage `json:"user,omitempty"`
	ElecSet json.RawMessage `json:"elec_set,omitempty"`
}

type NewInfo struct {
	Show    bool                           `json:"show"`
	Total   int64                          `json:"total,omitempty"`
	Count   int64                          `json:"count,omitempty"`
	ElecNum int                            `json:"elec_num"` //充电服务不下发,返回值保留兼容客户端老版本
	List    []*uprmdl.RankElecElementProto `json:"list,omitempty"`
	ElecSet *upr.BatterySet                `json:"elec_set,omitempty"`
	// 充电+
	State           int64                       `json:"state"`
	UpowerTitle     string                      `json:"upower_title,omitempty"`
	UpowerJumpUrl   string                      `json:"upower_jump_url,omitempty"`
	UpowerIconUrl   string                      `json:"upower_icon_url,omitempty"`
	UpowerState     int32                       `json:"upower_state"`
	UpowerButtonMap map[int64]*upr.UpowerButton `json:"upower_button_map,omitempty"`
	RankTitle       string                      `json:"rank_title,omitempty"`
	RankUrl         string                      `json:"rank_url,omitempty"`
}

func FormatElec(ctx context.Context, in *upr.UPRankWithPanelReply) *NewInfo {
	rly := &NewInfo{
		Show:    in.Show,
		Total:   in.CountUPTotalElec,
		Count:   in.Count,
		List:    in.List,
		ElecSet: in.ElecSet,
	}
	// 粉版出新充电+
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid()
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhone()
	}).MustFinish() {
		// 充电+
		rly.State = in.State
		rly.UpowerTitle = in.UpowerTitle
		rly.UpowerJumpUrl = in.UpowerJumpUrl
		rly.UpowerIconUrl = in.UpowerIconUrl
		rly.UpowerState = int32(in.UpowerState)
		rly.RankTitle = in.RankTitle
		rly.RankUrl = in.RankUrl
		rly.UpowerButtonMap = in.UpowerButtonMap
	}
	return rly
}
