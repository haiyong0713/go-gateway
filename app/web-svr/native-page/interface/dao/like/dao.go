package like

import (
	httpx "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/native-page/interface/conf"
)

// Dao struct
type Dao struct {
	c              *conf.Config
	client         *httpx.Client
	epPlayURL      string
	arcTypeListURL string
}

// New init
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:              c,
		client:         httpx.NewClient(c.HTTPClient),
		epPlayURL:      c.Host.APICo + _epPlayURI,
		arcTypeListURL: c.Host.APICo + _arcTypeListURI,
	}
	return
}
