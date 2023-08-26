package wechat

import (
	"context"

	xhttp "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/conf"
)

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package wechat

// Dao dao interface
type Dao interface {
	Close()

	SendWeChat(c context.Context, publicKey, title, msg, user string) (err error)
}

// Dao dao.
type dao struct {
	c      *conf.Config
	client *xhttp.Client
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:      c,
		client: xhttp.NewClient(c.HTTPClient),
	}
	return
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// Close Dao
func (d *dao) Close() {

}
