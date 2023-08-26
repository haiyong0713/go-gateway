package model

import gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"

type GaiaResponseType int

const (
	// 无风控或者风控通过
	GaiaResponseType_Default GaiaResponseType = 0
	// 风控决策需要前端校验
	GaiaResponseType_NeedFECheck GaiaResponseType = 1
	// 请求被风控直接阻止
	GaiaResponseType_Reject GaiaResponseType = 2
	// 风控校验token无效
	GaiaResponseType_TokenInvalid GaiaResponseType = 3
)

// 风控参数：https://info.bilibili.co/pages/viewpage.action?pageId=539671805
type RiskManagement struct {
	Mid         int64  `json:"mid"`
	Buvid       string `json:"buvid"`
	Ip          string `json:"ip"`
	Platform    string `json:"platform"`
	Ctime       string `json:"ctime"`
	Action      string `json:"action"`
	Api         string `json:"api"`
	Origin      string `json:"origin"`
	Referer     string `json:"referer"`
	Ua          string `json:"user_agent"`
	Host        string `json:"host"`
	Query       string `json:"query"`
	Header      string `json:"header"`
	Cookie      string `json:"cookie"`
	Token       string `json:"token"`
	VisitRecord int64  `json:"visit_record"`
	Scene       string `json:"scene"`
}

type RiskResult struct {
	IsRisk      bool                    `json:"is_risk"`
	GaiaResType GaiaResponseType        `json:"gaia_res_type"`
	GaiaData    *gaiamdl.RuleCheckReply `json:"gaia_data"`
}
