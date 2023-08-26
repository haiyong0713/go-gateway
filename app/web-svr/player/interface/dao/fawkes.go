package dao

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	fkmdl "go-gateway/app/app-svr/fawkes/service/model/business"

	"github.com/pkg/errors"
)

const (
	_getVersion = "/x/admin/fawkes/business/config/version"
)

// FawkesVersion get fawkes version.
func (d *Dao) FawkesVersion(c context.Context) (map[string]map[string]*fkmdl.Version, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	var res struct {
		Code int                                  `json:"code"`
		Data map[string]map[string]*fkmdl.Version `json:"data"`
	}
	err := d.client.Get(c, d.getVersionURL, ip, nil, &res)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.getVersionURL)
		return nil, err
	}
	return res.Data, nil
}
