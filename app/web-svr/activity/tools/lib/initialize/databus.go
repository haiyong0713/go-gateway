package initialize

import (
	"fmt"
	"go-common/library/queue/databus"
	databusV2 "go-common/library/queue/databus.v2"
)

func NewDatabusV1(cfg *databus.Config) *databus.Databus {
	if _, open := IsOpen(fmt.Sprintf("databus.%s", cfg.Topic)); open {
		return databus.New(cfg)
	}
	return nil
}

func NewProducer(client databusV2.Client, topic string) (p databusV2.Producer) {
	if client != nil {
		NewE(fmt.Sprintf("databus.producer.%s", topic), func() (err error) {
			p, err = client.NewProducer(topic)
			return
		})
	}
	return
}

func NewConsumer(client databusV2.Client, topic string) (p databusV2.Consumer) {
	if client != nil {
		NewE(fmt.Sprintf("databus.consumer.%s", topic), func() (err error) {
			p, err = client.NewConsumer(topic)
			return
		})
	}
	return
}
