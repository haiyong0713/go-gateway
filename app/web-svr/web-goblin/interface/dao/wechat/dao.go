package wechat

import (
	"go-common/library/cache/redis"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/web-goblin/interface/conf"
)

// Dao dao struct.
type Dao struct {
	// config
	c *conf.Config
	// redis
	redis *redis.Pool
	// httpClient
	httpClient *bm.Client
	// url
	wxAccessTokenURL string
	wxQrcodeURL      string
	wxSendMsgURL     string
	cache            *fanout.Fanout
}

// New new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// config
		c:          c,
		redis:      redis.NewPool(c.Redis.Config),
		httpClient: bm.NewClient(c.HTTPClient),
		cache:      fanout.New("goblin wechat cache", fanout.Worker(1), fanout.Buffer(1024)),
	}
	d.wxAccessTokenURL = d.c.Host.Wechat + _accessTokenURI
	d.wxQrcodeURL = d.c.Host.Wechat + _qrcodeURI
	d.wxSendMsgURL = d.c.Host.Wechat + _sendMsgURI
	return
}
