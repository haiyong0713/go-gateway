package client

import (
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	v1 "go-gateway/app/web-svr/esports/service/api/v1"

	bGroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	tunnelV2 "git.bilibili.co/bapis/bapis-go/platform/service/tunnel/v2"
	"go-gateway/app/web-svr/esports/admin/conf"
	esportsV1 "go-gateway/app/web-svr/esports/interface/api/v1"
)

var (
	BGroupClient          bGroup.BGroupServiceClient
	TunnelV2Client        tunnelV2.TunnelClient
	EsportsGrpcClient     esportsV1.EsportsClient
	EsportsServiceClient  v1.EsportsServiceClient
	ActivityServiceClient api.ActivityClient
)

func InitClients(cfg *conf.Config) (err error) {
	if BGroupClient, err = bGroup.NewClient(cfg.BGroupClient); err != nil {
		log.Error("InitClients: init BGroupClient err:(%+v)", err)
		return
	}
	if TunnelV2Client, err = tunnelV2.NewClient(cfg.TunnelV2Client); err != nil {
		log.Error("InitClients: init TunnelV2Client err:(%+v)", err)
		return
	}
	if EsportsGrpcClient, err = esportsV1.NewClient(cfg.EspClient); err != nil {
		panic(err)
	}
	if EsportsServiceClient, err = v1.NewClient(cfg.EsportsServiceClient); err != nil {
		panic(err)
	}
	if ActivityServiceClient, err = api.NewClient(cfg.ActivityServiceClient); err != nil {
		panic(err)
	}
	return
}
