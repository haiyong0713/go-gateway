package dao

import (
	"context"
	"strconv"

	"go-common/library/log"
	appmodel "go-gateway/app/app-svr/misaka/interface/internal/model/app"
	webmodel "go-gateway/app/app-svr/misaka/interface/internal/model/web"

	"github.com/Shopify/sarama"
)

func (d *Dao) pubAppProc() {
	for {
		data, ok := <-d.appMessages
		if !ok {
			log.Warn("d.appMessages chan closed")
			return
		}
		var (
			value []byte
			err   error
		)
		if value, err = data.Marshal(); err != nil {
			log.Error("pubAppProc data.Marshal error(%v)", err)
			continue
		}
		msg := &sarama.ProducerMessage{
			Topic: d.appTopic,
			Key:   sarama.ByteEncoder([]byte(data.IP)),
			Value: sarama.ByteEncoder(value),
		}
		if _, _, err := d.producer.SendMessage(msg); err != nil {
			log.Error("d.pubAppProc.SendMessage(%+v) error(%v)", msg, err)
			continue
		}
		d.infoProm.Incr(strconv.FormatInt(data.Data.LogID, 10))
	}
}

func (d *Dao) pubWebProc() {
	for {
		info, ok := <-d.webMessages
		if !ok {
			log.Warn("d.webMessages chan closed")
			return
		}
		var (
			value []byte
			err   error
		)
		if value, err = info.Marshal(); err != nil {
			log.Error("pubWebProc info.Marshal error(%v)", err)
			continue
		}
		msg := &sarama.ProducerMessage{
			Topic: d.webTopic,
			Key:   sarama.ByteEncoder([]byte(info.IP)),
			Value: sarama.ByteEncoder(value),
		}
		if _, _, err := d.producer.SendMessage(msg); err != nil {
			log.Error("d.pubWebProc.SendMessage(%+v) error(%v)", msg, err)
			continue
		}
		d.infoProm.Incr(strconv.FormatInt(info.Data.LogID, 10))
	}
}

// PubApp is
func (d *Dao) PubApp(c context.Context, info *appmodel.Info) (err error) {
	select {
	case d.appMessages <- info:
	default:
		log.Error("d.appMessages chan full")
	}
	return
}

// PubWeb is
func (d *Dao) PubWeb(c context.Context, info *webmodel.Info) (err error) {
	select {
	case d.webMessages <- info:
	default:
		log.Error("d.webMessages chan full")
	}
	return
}
