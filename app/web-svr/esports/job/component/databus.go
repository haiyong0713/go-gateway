package component

import (
	"context"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"

	databusv2 "go-common/library/queue/databus.v2"
	"go-gateway/app/web-svr/esports/job/conf"
)

var (
	DatabusClient      databusv2.Client
	BGroupMessagePub   databusv2.Producer
	ContestSchedulePub databusv2.Producer
)

func InitComponents() (err error) {
	DatabusClient, err = databusv2.NewClient(
		context.Background(),
		conf.Conf.Databus.Target,
		databusv2.WithAppID(conf.Conf.Databus.AppID),
		databusv2.WithToken(conf.Conf.Databus.Token),
	)
	if err != nil {
		panic(err)
	}
	BGroupMessagePub = initialize.NewProducer(DatabusClient, conf.Conf.Databus.Topic.BGroupMessage)
	ContestSchedulePub = initialize.NewProducer(DatabusClient, conf.Conf.Databus.Topic.ContestSchedule)
	return
}
