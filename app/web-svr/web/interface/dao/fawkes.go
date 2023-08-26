package dao

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	fkmdl "go-gateway/app/app-svr/fawkes/service/model/business"

	"github.com/pkg/errors"
)

const (
	_getFawkesVersion = "/x/admin/fawkes/business/config/version"
)

// FawkesVersion get fawkes version.
func (d *Dao) FawkesVersion(c context.Context) (map[string]map[string]*fkmdl.Version, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	var err error
	var res struct {
		Code int                                  `json:"code"`
		Data map[string]map[string]*fkmdl.Version `json:"data"`
	}

	if err = d.httpR.Get(c, d.fawkesVersionURL, ip, nil, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.fawkesVersionURL)
		return nil, err
	}
	return res.Data, nil
}
