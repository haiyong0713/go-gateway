package archive

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-intl/interface/conf"
	steinApi "go-gateway/app/app-svr/steins-gate/service/api"
)

// Dao is archive dao.
type Dao struct {
	// grpc
	steinClient steinApi.SteinsGateClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.steinClient, err = steinApi.NewClient(c.SteinClient); err != nil {
		panic("SteinClient not found")
	}
	return
}

// Archive3 get archive.
func (d *Dao) View(c context.Context, aid, mid int64, buvid string) (a *steinApi.ViewReply, err error) {
	arg := &steinApi.ViewReq{Aid: aid, Mid: mid, Buvid: buvid}
	if a, err = d.steinClient.View(c, arg); err != nil {
		log.Error("d.steinClient.View(%v) error(%+v)", arg, err)
		return
	}
	return
}
