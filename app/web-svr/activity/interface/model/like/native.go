package like

import (
	account "git.bilibili.co/bapis/bapis-go/account/service"
)

// RcmdRsp .
type RcmdRsp struct {
	Infos map[int64]*RcmdInfo `json:"infos"`
}

// RcmdInfo .
type RcmdInfo struct {
	Mid      int64                `json:"mid"`
	Face     string               `json:"face"`
	Name     string               `json:"name"`
	IsFav    int                  `json:"is_fav"`
	Vip      account.VipInfo      `json:"vip"`
	Official account.OfficialInfo `json:"official"`
}
