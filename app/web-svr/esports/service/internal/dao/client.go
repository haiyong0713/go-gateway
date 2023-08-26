package dao

import (
	"context"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	databusv2 "go-common/library/queue/databus.v2"
	activityapi "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	espClient "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/service/conf"

	favClient "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	liveRoom "git.bilibili.co/bapis/bapis-go/live/xroom"
	bGroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	tunnelV2 "git.bilibili.co/bapis/bapis-go/platform/service/tunnel/v2"
)

func initGrpcClients(d *dao) {
	initBGroupClient(d, d.conf)
	initLiveRpcClient(d, d.conf)
	initTunnelClient(d, d.conf)
	initFavClient(d, d.conf)
	initEspClient(d, d.conf)
	initActivityClient(d, d.conf)
	initAccountClient(d, d.conf)
	initDataBusV2Client(d, d.conf)
	initBGroupDataBusPub(d)
}

func initDataBusV2Client(d *dao, conf *conf.Config) {
	dataBusClient, err := databusv2.NewClient(
		context.Background(),
		conf.DataBus.Target,
		databusv2.WithAppID(conf.DataBus.AppID),
		databusv2.WithToken(conf.DataBus.Token),
	)
	if err != nil {
		panic(err)
	}
	d.DataBusV2Client = dataBusClient
}

func initBGroupDataBusPub(d *dao) {
	d.BGroupMessagePub = initialize.NewProducer(d.DataBusV2Client, d.conf.DataBus.Topic.BGroupMessage)
}

func initFavClient(d *dao, conf *conf.Config) {
	if grpcClient, err := favClient.NewClient(conf.FavClient); err != nil {
		log.Error("Init initFavClient favClient.NewClient() Error, err: %+v", err)
		panic(err)
	} else {
		d.favoriteClient = grpcClient
	}
}

func initTunnelClient(d *dao, conf *conf.Config) {
	if grpcClient, err := tunnelV2.NewClient(conf.TunnelV2Client); err != nil {
		log.Error("Init initTunnelClient tunnelV2.NewClient() Error, err: %+v", err)
		panic(err)
	} else {
		d.tunnelV2Client = grpcClient
	}
}

func initBGroupClient(d *dao, conf *conf.Config) {
	if grpcClient, err := bGroup.NewClient(conf.BGroupClient); err != nil {
		log.Error("Init initBGroupClient BGroupGrpcClient Error, err: %+v", err)
		panic(err)
	} else {
		d.bGroupClient = grpcClient
	}
}

func initLiveRpcClient(d *dao, conf *conf.Config) {
	if grpcClient, err := liveRoom.NewClient(conf.LiveRoomGrpc); err != nil {
		log.Error("Init initLiveRpcClient Error, err: %+v", err)
		panic(err)
	} else {
		d.liveRoomClient = grpcClient
	}
}

func initEspClient(d *dao, conf *conf.Config) {
	if grpcClient, err := espClient.NewClient(conf.EspClient); err != nil {
		log.Error("Init initEsportsRpcClient Error, err: %+v", err)
		panic(err)
	} else {
		d.espClient = grpcClient
	}
}

func initActivityClient(d *dao, conf *conf.Config) {
	if grpcClient, err := activityapi.NewClient(conf.ActivityClient); err != nil {
		log.Error("Init initActivityClient Error, err: %+v", err)
		panic(err)
	} else {
		d.activityClient = grpcClient
	}
}

func initAccountClient(d *dao, conf *conf.Config) {
	if grpcClient, err := accapi.NewClient(conf.AccClient); err != nil {
		log.Error("Init initAccountClient Error, err: %+v", err)
		panic(err)
	} else {
		d.accountClient = grpcClient
	}
}
