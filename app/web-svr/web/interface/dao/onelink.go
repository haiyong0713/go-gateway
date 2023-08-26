package dao

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"
	"net/url"
	"strconv"
)

const _h5Onelink = "/x/location/zlimit/group"

func (d *Dao) GetOnelinkAccess(c context.Context, mid int64, gid int64) (res *model.H5Onelink, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("gid", strconv.FormatInt(gid, 10))
	params.Set("ip", ip)
	var rs struct {
		Code int              `json:"code"`
		Data *model.H5Onelink `json:"data"`
	}
	if err = d.httpR.Get(c, d.c.Host.API+_h5Onelink, ip, params, &rs); err != nil {
		log.Error("GetOnelinkAccess d.httpR.Get(%s, %s, %v) error(%v)", d.c.Host.API+_h5Onelink, ip, params, err)
		return
	}
	if rs.Code != 0 {
		return nil, errors.Wrap(ecode.Int(rs.Code), d.c.Host.API+_h5Onelink+"?"+params.Encode())
	}
	return rs.Data, nil
}
