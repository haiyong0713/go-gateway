package gwconfig

import (
	"go-gateway/app/app-svr/app-gw/management-job/api"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	sdkwarden "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden/server"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster"
)

type LogReply struct {
	Order  string        `json:"order"`
	Sort   string        `json:"sort"`
	Result []ManagerInfo `json:"result"`
}

// ManagerInfo.
type ManagerInfo struct {
	Action    string `json:"action"`
	Business  int    `json:"business"`
	ExtraData string `json:"extra_data"`
	Oid       int64  `json:"oid"`
	Str0      string `json:"str_0"`
	Str1      string `json:"str_1"`
	Str2      string `json:"str_2"`
	Type      int8   `json:"type"`
	UID       int64  `json:"uid"`
	Uname     string `json:"uname"`
}

type ProxyConfig struct {
	DynPath []*sdk.PathMeta `json:"dyn_path"`
}

type ProxyConfigs struct {
	ProxyConfig *ProxyConfig `json:"ProxyConfig"`
}

type GrpcProxyConfig struct {
	ProxyConfig *sdkwarden.Config `json:"ProxyConfig"`
}

type PushConfigReq struct {
	AppID      string
	TreeID     int64
	ConfigMeta *pb.ConfigMeta
	Buffer     []byte
}

type RawConfigReq struct {
	AppID      string
	TreeID     int64
	ConfigMeta *pb.ConfigMeta
}

type Extra struct {
	Result string `json:"result"`
}

type RawLogReq struct {
	Node       string
	Gateway    string
	Order      string
	ObjectType int
	Pn         int
	Ps         int
}

type PushConfigContext struct {
	Sponsor string
	Action  string
	Ctime   int64
	Mtime   int64
	Compare bool
}

func (pc *PushConfigContext) FromTask(in *api.TaskDoReq) {
	pc.Sponsor = in.Sponsor
	pc.Ctime = in.Params.Ctime
	pc.Mtime = in.Params.Mtime
}
