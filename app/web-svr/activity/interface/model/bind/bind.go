package bind

import v1 "go-gateway/app/web-svr/activity/interface/api"

const (
	BindAccountTrue = 1
	BindExternalTX  = 1 // 绑定腾讯
	IsBindTrue      = 1 // 绑定到账号
	IsBindFalse     = 0 // 无绑定
	IsBindRole      = 2 // 绑定到角色

)

type UserBindInfo struct {
	ConfigInfo *v1.BindConfigInfo `json:"config_info"`
	BindInfo   *BindInfo          `json:"bind_info"`
}

type BindInfo struct {
	BindType    int64        `json:"bind_type"`
	BindPhone   int64        `json:"bind_phone"`
	RoleInfo    *RoleInfo    `json:"role_info"`
	AccountInfo *AccountInfo `json:"account_info"`
}

type RoleInfo struct {
	RoleName      string `json:"role_name"`
	AreaName      string `json:"area_name"`
	PartitionName string `json:"partition_name"`
	PlatName      string `json:"plat_name"`
}

type TencentGameBindInfo struct {
	IsBind  bool `json:"isBind"`
	GameAcc *struct {
		Type string `json:"type"`
	} `json:"gameAcc"`
	GameRole *struct {
		RoleName      string `json:"roleName"`
		AreaName      string `json:"areaName"`
		PartitionName string `json:"partitionName"`
		PlatName      string `json:"platName"`
	} `json:"gameRole"`
}

type AccountInfo struct {
	AccountType string `json:"account_type"`
}

type BindParams struct {
	Sign       string `json:"sign"`
	BasePath   string `json:"base_path"`
	FaceUrl    string `json:"face_url"`
	Code       string `json:"code"`
	T          int64  `json:"t"`
	LivePlatId string `json:"live_plat_id"`
	NickName   string `json:"nick_name"`
	GameIdList string `json:"game_id_list"`
	OriginId   string `json:"origin_id"`
	Nonce      string `json:"nonce"`
}
