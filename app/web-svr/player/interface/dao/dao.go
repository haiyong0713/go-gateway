package dao

import (
	"context"
	"fmt"
	"net/http"

	xhttp "go-common/library/net/http/blademaster"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
	"go-gateway/app/web-svr/player/interface/conf"
)

// Dao dao.
type Dao struct {
	// config
	c *conf.Config
	// client
	client   *xhttp.Client
	vsClient *http.Client
	// API URL
	blockTimeURL   string
	onlineCountURL string
	pcdnLoaderURL  string
	getVersionURL  string
	// rpc
	playURLRPCV2 v2.PlayURLClient
}

// New return new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:              c,
		client:         xhttp.NewClient(c.HTTPClient),
		vsClient:       http.DefaultClient,
		blockTimeURL:   c.Host.AccCo + _blockTimeURI,
		onlineCountURL: c.Host.APICo + _onlineCountURI,
		pcdnLoaderURL:  c.Host.APICo + _pcdnLoaderURI,
		getVersionURL:  c.Host.Fawkes + _getVersion,
	}
	var err error
	d.playURLRPCV2, err = v2.NewClient(c.PlayURLClient)
	if err != nil {
		panic(fmt.Sprintf("player v2 NewClient error(%v)", err))
	}
	return
}

// Ping check service health
func (d *Dao) Ping(c context.Context) (err error) {
	return
}
