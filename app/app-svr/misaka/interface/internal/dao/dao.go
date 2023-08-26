package dao

import (
	"context"
	"fmt"
	"runtime"

	"go-common/library/conf/paladin"
	"go-common/library/stat/prom"
	appmodel "go-gateway/app/app-svr/misaka/interface/internal/model/app"
	webmodel "go-gateway/app/app-svr/misaka/interface/internal/model/web"

	"github.com/Shopify/sarama"
)

// Dao dao.
type Dao struct {
	producer    sarama.SyncProducer
	appTopic    string
	webTopic    string
	ac          *paladin.Map
	appMessages chan *appmodel.Info
	webMessages chan *webmodel.Info
	infoProm    *prom.Prom
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// New new a dao and return.
func New(ac *paladin.Map) (dao *Dao) {
	dao = &Dao{
		ac:          ac,
		appMessages: make(chan *appmodel.Info, 10240),
		webMessages: make(chan *webmodel.Info, 10240),
		infoProm:    prom.BusinessInfoCount,
	}
	conf := sarama.NewConfig()
	conf.Producer.Return.Successes = true
	conf.Version = sarama.V1_0_0_0
	appt, err := ac.Get("appTopic").String()
	if err != nil {
		panic("app topic error")
	}
	dao.appTopic = appt
	webt, err := ac.Get("webTopic").String()
	if err != nil {
		panic("web topic error")
	}
	dao.webTopic = webt
	var addrs []string
	ac.Get("addrs").Slice(&addrs)
	if dao.producer, err = sarama.NewSyncProducer(addrs, conf); err != nil {
		panic(fmt.Sprintf("saram.NewSyncProducer error(%+v)", err))
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		go dao.pubAppProc()
		go dao.pubWebProc()
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {
}

// Ping ping the resource.
func (d *Dao) Ping(ctx context.Context) (err error) {
	return nil
}
