package fawkes

import (
	"context"
	"sync"

	"go-gateway/app/app-svr/fawkes/service/conf"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	databusV2 "go-common/library/queue/databus.v2"
)

var (
	once          sync.Once
	databusClient databusV2.Client
)

func NewDatabus(c *conf.Databus) databusV2.Client {
	once.Do(func() {
		var err error
		databusClient, err = databusV2.NewClient(context.Background(), databusV2.WithAppID(c.AppID), databusV2.WithToken(c.Token))
		if err != nil {
			log.Errorc(context.Background(), "NewDatabus error: %v", err)
			return
		}
	})
	return databusClient
}

func (d *Dao) DatabusSubscribe(c context.Context, group, topic string, handler func(ctx context.Context, msg databusV2.Message) error) (consumer databusV2.Consumer, err error) {
	if d.databusClient == nil {
		log.Error("Databus client is nil")
		return
	}
	consumer, err = d.databusClient.NewConsumer(group, topic)
	if err != nil {
		log.Error("DatabusSubscribe error: %v", err)
		return
	}
	defer consumer.Close()
	err = consumer.Handle(context.Background(), handler)
	if err != nil {
		log.Error("DatabusSubscribe error: %v", err)
	}
	return
}

func (d *Dao) NewProducer(c context.Context, group, topic string) (p databusV2.Producer) {
	if d.databusClient != nil {
		var err error
		if p, err = d.databusClient.NewProducer(group, topic); err != nil {
			log.Errorc(c, "NewProducer error: %v", err)
			return nil
		}
	}
	return
}
