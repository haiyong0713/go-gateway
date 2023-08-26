package bangumi

import (
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/job/conf"
)

type Dao struct {
	// api
	client                  *httpx.Client
	channelcontent          string
	channelcontentchange    string
	channelcontentoffshelve string
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		client:                  httpx.NewClient(c.HTTPPGC),
		channelcontent:          c.Host.Bangumi + _channelcontent,
		channelcontentchange:    c.Host.Bangumi + _channelcontentchange,
		channelcontentoffshelve: c.Host.Bangumi + _channelcontentoffshelve,
	}
	return d
}
