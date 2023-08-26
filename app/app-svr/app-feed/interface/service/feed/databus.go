package feed

import (
	"context"

	databusv2 "go-common/library/queue/databus.v2"
	"go-gateway/app/app-svr/app-feed/interface/conf"
)

var (
	DatabusClient  databusv2.Client
	FeedAppListPub databusv2.Producer
)

func InitDatabus(c *conf.Config) {
	err := func() error {
		if err := initClient(c); err != nil {
			return err
		}
		if err := initProducer(c); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		panic(err)
	}
}

func initClient(c *conf.Config) (err error) {
	DatabusClient, err = databusv2.NewClient(
		context.Background(),
		databusv2.WithAppID(c.Databus.AppID),
		databusv2.WithToken(c.Databus.Token),
	)
	return
}

func initProducer(c *conf.Config) (err error) {
	FeedAppListPub, err = DatabusClient.NewProducer(c.Databus.Group, c.Databus.Topic)
	return
}
