package push

import (
	"fmt"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-job/job/conf"

	broadcastApi "git.bilibili.co/bapis/bapis-go/push/service/broadcast/v2"
)

const _pushURL = "/x/internal/push-strategy/task/add"

// Dao dao
type Dao struct {
	c *conf.Config
	// http client
	client *bm.Client
	// push service URL
	pushURL  string
	bcClient broadcastApi.BroadcastAPIClient
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:       c,
		client:  bm.NewClient(c.HTTPClient),
		pushURL: c.Host.APICo + _pushURL,
	}
	var err error
	if dao.bcClient, err = broadcastApi.NewClient(c.BroadCastClient); err != nil {
		panic(fmt.Sprintf("broadcast new client err(%+v)", err))
	}
	return
}
