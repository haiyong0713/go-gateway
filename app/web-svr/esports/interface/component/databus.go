package component

import (
	"context"

	databusv2 "go-common/library/queue/databus.v2"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"go-gateway/app/web-svr/esports/interface/conf"
)

var (
	DatabusClient    databusv2.Client
	BGroupMessagePub databusv2.Producer
)

func InitClient() (err error) {
	DatabusClient, err = databusv2.NewClient(
		context.Background(),
		conf.Conf.Databus.Target,
		databusv2.WithAppID(conf.Conf.Databus.AppID),
		databusv2.WithToken(conf.Conf.Databus.Token),
	)
	if err != nil {
		panic(err)
	}
	return
}

func InitProducer() (err error) {
	BGroupMessagePub = initialize.NewProducer(DatabusClient, conf.Conf.Databus.Topic.BGroupMessage)
	return
}
