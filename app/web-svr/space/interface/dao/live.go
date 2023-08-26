package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/space/interface/model"

	"github.com/pkg/errors"
)

const (
	_liveURI = "/room/v1/Room/getRoomInfoOld"
)

// Live is space live data.
func (d *Dao) Live(c context.Context, mid int64, platform string) (live *model.Live, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("platform", platform)
	var res struct {
		Code int         `json:"code"`
		Data *model.Live `json:"data"`
	}
	if err = d.httpR.Get(c, d.liveURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.liveURL+"?"+params.Encode())
		return
	}
	live = res.Data
	return
}
