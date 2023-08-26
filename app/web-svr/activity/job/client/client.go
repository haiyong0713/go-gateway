package client

import (
	"go-common/library/log"

	audit "git.bilibili.co/bapis/bapis-go/aegis/strategy/service"
	fligrpc "git.bilibili.co/bapis/bapis-go/filter/service"
	liveapi "git.bilibili.co/bapis/bapis-go/live/xroom"
	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/conf"

	arcapi "go-gateway/app/app-svr/archive/service/api"

	topic "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	esportsPB "git.bilibili.co/bapis/bapis-go/esports/service"
	garb "git.bilibili.co/bapis/bapis-go/garb/service"
	bGroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	tunnel "git.bilibili.co/bapis/bapis-go/platform/service/tunnel/v2"
	gaiaApi "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	videoup "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

var (
	ActivityClient api.ActivityClient
	GaiaClient     gaiaApi.GaiaClient
	EsportsClient  esportsPB.EsportsClient
	BGroupClient   bGroup.BGroupServiceClient
	TunnelClient   tunnel.TunnelClient
	GarbClient     garb.GarbClient
	VideoClient    videoup.VideoUpOpenClient
	FilterClient   fligrpc.FilterClient
	AuditClient    audit.AegisStrategyServiceClient
	LiveClient     liveapi.RoomClient
	ActplatClient  actplatapi.ActPlatClient
	ArcClient      arcapi.ArchiveClient
	TopicClient    topic.TopicClient
)

func InitClients(cfg *conf.Config) (err error) {
	ActivityClient, err = api.NewClient(cfg.ActClient)
	if err != nil {
		log.Error("InitClients: init activity client err:", err)
		return
	}
	if GaiaClient, err = gaiaApi.NewClient(cfg.ActClient); err != nil {
		log.Error("InitClients: init silverbullet GaiaClient err:", err)
		return
	}
	EsportsClient, err = esportsPB.NewClient(cfg.EsportsClient)
	if err != nil {
		log.Error("InitClients: init esports client err:", err)
		return
	}
	if BGroupClient, err = bGroup.NewClient(cfg.BGroupClient); err != nil {
		log.Error("InitClients: init BGroupClient err:", err)
		return
	}
	if TunnelClient, err = tunnel.NewClient(cfg.TunnelClient); err != nil {
		log.Error("InitClients: init TunnelClient err:", err)
		return
	}
	if GarbClient, err = garb.NewClient(cfg.GarbClient); err != nil {
		log.Error("InitClients: init GarbClient err:", err)
		return
	}
	if VideoClient, err = videoup.NewClient(cfg.GarbClient); err != nil {
		log.Error("InitClients: init VideoClient err:", err)
		return
	}
	if FilterClient, err = fligrpc.NewClient(cfg.FliClient); err != nil {
		log.Error("InitClients: init fliClient err:", err)
		return
	}
	if AuditClient, err = audit.NewClient(cfg.FliClient); err != nil {
		log.Error("InitClients: init auditClient err:", err)
		return
	}
	if ActplatClient, err = actplatapi.NewClient(cfg.ActPlatClient); err != nil {
		log.Error("InitClients: init actplatClient err:", err)
		return
	}
	if ArcClient, err = arcapi.NewClient(cfg.ArcClient); err != nil {
		log.Error("InitClients: init arcClient err:", err)
		return
	}
	if LiveClient, err = liveapi.NewClient(cfg.LiveXRoomClient); err != nil {
		log.Error("InitClients: init LiveClient err:", err)
	}
	if TopicClient, err = topic.NewClient(cfg.TopicClient); err != nil {
		log.Error("InitClients: init TopicClient err:", err)
		return
	}
	return
}
