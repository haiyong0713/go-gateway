package common

import (
	"go-gateway/app/app-svr/app-car/interface/model"
)

const DefaultTab = "dynamic"

var Tabs = map[string]string{
	"dynamic":  "关注",
	"favorite": "收藏",
	"bangumi":  "追番",
	"cinema":   "追剧",
}

type MineTabsReq struct {
	model.DeviceInfo
}

type MineTab struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Id        int64  `json:"id"`
}
