package fawkes

import (
	"context"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model/business"

	"github.com/pkg/errors"
)

const (
	_getVersion = "/x/admin/fawkes/business/config/version"
)

// Dao is show dao.
type Dao struct {
	// http client
	client *httpx.Client
	// url
	getVersion string
}

// New new a bangumi dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		client:     httpx.NewClient(c.HTTPClient),
		getVersion: c.Host.Fawkes + _getVersion,
	}
	return d
}

// FawkesVersion get fawkes version.
func (d *Dao) FawkesVersion(c context.Context) (re map[string]map[string]*fkmdl.Version, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	var res struct {
		Code int                                  `json:"code"`
		Data map[string]map[string]*fkmdl.Version `json:"data"`
	}
	if err = d.client.Get(c, d.getVersion, ip, nil, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.getVersion)
		return
	}
	re = res.Data
	return
}
