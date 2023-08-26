package model

import (
	v1 "go-gateway/app/app-svr/archive/service/api"

	gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
)

// CoinArc coin archive.
type CoinArc struct {
	*v1.Arc
	Bvid  string `json:"bvid"`
	Coins int64  `json:"coins"`
	Time  int64  `json:"time"`
	IP    string `json:"ip"`
}

func (gt GaiaResponseType) String() string {
	switch gt {
	case GaiaResponseType_NeedFECheck:
		return "need_fe_check"
	case GaiaResponseType_Reject:
		return "reject"
	case GaiaResponseType_TokenInvalid:
		return "token_invalid"
	default:
		return "default"
	}
}

type AddCoinRes struct {
	Like        bool                    `json:"like"`
	IsRisk      bool                    `json:"is_risk"`
	GaiaResType GaiaResponseType        `json:"gaia_res_type"`
	GaiaData    *gaiamdl.RuleCheckReply `json:"gaia_data"`
}
