package es

import (
	xhttp "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
)

type Dao struct {
	c             *conf.Config
	client        *xhttp.Client
	searchChannel string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: xhttp.NewClient(c.HTTPClient),
		// url
		searchChannel: c.Host.APICo + _searchChannel,
	}
	return
}
