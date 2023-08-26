package manager

import (
	"go-gateway/app/app-svr/fawkes/service/model"
)

// log const.
const (
	ModelCI          = "ci"
	ModelCD          = "cd"
	ModelHotpatch    = "Hotpatch"
	ModelChannelPack = "渠道包"
	ModelBizApk      = "biz_apk"

	OperationCIPush               = "推送到CD"
	OperationCDPushProd           = "推送到正式"
	OperationCDFlowConfig         = "流量配置"
	OperationCDUpgradeConfig      = "应用配置"
	OperationCDFilterConfig       = "配置/修改"
	OperationCDSyncMacross        = "同步到Macross"
	OperationCDSyncManager        = "同步到Manager"
	OperationHotpatchPushProd     = "推送到正式"
	OperationHotpatchSwitch       = "生效开关"
	OperationHotpatchConfigModify = "配置修改"
	OperationChannelPackPushCDN   = "推送到CDN"
	OperationChannelPackTest      = "测试通过"
	OperationChannelPackPushProd  = "发布到线上"

	OperationBizApkFlowConfig   = "流量配置"
	OperationBizApkFilterConfig = "配置/修改"
)

// LogResult struct log list.
type LogResult struct {
	PageInfo *model.PageInfo `json:"page"`
	Items    []*Log          `json:"items"`
}

// Log struct for log model.
type Log struct {
	ID        int64  `json:"id"`
	AppKey    string `json:"app_key"`
	Env       string `json:"env"`
	Model     string `json:"model"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
	Operator  string `json:"operator"`
	CTime     int64  `json:"ctime"`
	MTime     int64  `json:"mtime"`
}
