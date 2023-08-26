package model

import (
	pcdnAccgrpc "git.bilibili.co/bapis/bapis-go/vas/pcdn/account/service"
	pcdnRewgrpc "git.bilibili.co/bapis/bapis-go/vas/pcdn/reward/service"
)

/******************************************** req *********************************************/
type PcdnV1Req struct {
	Mid          int64  `form:"-"`
	FawkesAppKey string `form:"fawkes_app_key" default:"web_main"`
	FawkesEnv    string `form:"fawkes_env" default:"prod"`
}

type OperatePCDNReq struct {
	Mid     int64 `form:"-"`
	Operate int   `form:"operate" validate:"min=1,max=2"` // 操作，1:开启、2:关闭
	Level   int   `form:"level" validate:"min=1,max=3"`
}

type PcdnRewardExchangeReq struct {
	Mid  int64 `form:"-"`
	Type int64 `form:"type" validate:"min=1,max=2"` // 兑换货币类型，1B币，2能量
	Num  int64 `form:"num" validate:"min=100"`      // 兑换数额
}

/******************************************** rep *********************************************/

// 用户设置信息
type PcdnUserSettingRep struct {
	Joined  bool `json:"joined"`  // 是否加入流量计划
	Started bool `json:"started"` // 是否开启流量贡献
	Level   int  `json:"level"`   // 流量贡献档位，1低、2中、3高。第一次返回0
	// IsFreeze bool `json:"is_freeze"` // 是否禁止加入计划【quit冻结期等等】
	LeftTime int64 `json:"rejoin_left_time"` // 退出后再加入剩余时间
	// Quit    bool `json:"quit"`    // 是否退出了流量计划
}

// fawkes相关
type Fawkes struct {
	ConfigVersion int64 `json:"config_version"`
	FFVersion     int64 `json:"ff_version"`
}

// pcdn v1返回
type PcdnV1Rep struct {
	IsInTopList  bool                `json:"is_in_top_list"` // 是否命中顶导灰度包
	IsInPopList  bool                `json:"is_in_pop_list"` // 是否命中弹窗灰度包
	Fawkes       *Fawkes             `json:"fawkes"`         // fawkes 配置
	UserSettings *PcdnUserSettingRep `json:"user_settings"`  // 用户状态
}

// pcdn pages返回
type PcdnPagesRep struct {
	Notification  []*pcdnAccgrpc.UserNotification `json:"notification"`        // 小黄条
	AccountInfo   []*pcdnAccgrpc.CurrencyInfo     `json:"account_info"`        // 用户资产信息
	DigitalReward *pcdnRewgrpc.DigitalRewardResp  `json:"digital_reward_info"` // 兑换商城信息
}
