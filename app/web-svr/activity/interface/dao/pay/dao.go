package pay

import (
	"context"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/pay"
)

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// Dao dao interface
type Dao interface {
	PayTransferInner(ctx context.Context, pt *pay.TransferInner) (res *pay.ResultInner, err error)
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:       c,
		pay:     newPayClient(bm.NewClient(c.HTTPClient)),
		payHost: c.Host.Pay,
	}
	return
}

type dao struct {
	c       *conf.Config
	pay     *Pay
	payHost string
}
