package message

// 生成 mock 代码
//go:generate mockgen -source dao.go  -destination dao.mock.go -package brand
import (
	"go-gateway/app/web-svr/activity/interface/conf"

	httpx "go-common/library/net/http/blademaster"
)

// Dao dao interface
type Dao interface {
	Close()
}

const (
	msgURL = "/api/notify/send.user.notify.do"
)

// Dao dao.
type dao struct {
	c      *conf.Config
	msgURL string
	client *httpx.Client
}

// New init
func newDao(c *conf.Config) (newdao Dao) {
	newdao = &dao{
		c:      c,
		msgURL: c.Host.Message + msgURL,
		client: httpx.NewClient(c.HTTPClient),
	}
	return
}

// New new a dao and return.
func New(c *conf.Config) (d Dao) {
	return newDao(c)
}

// Close Dao
func (dao *dao) Close() {
}
