package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-feed/interface-ng/internal/model"

	"github.com/pkg/errors"
)

const _newAD = "/bce/api/bce/feeds/oversaturated"

type adConfig struct {
	Host string
}

type adDao struct {
	client *bm.Client
	cfg    adConfig
}

func (d *adDao) Ad(ctx context.Context, req *model.AdReq) (*cm.NewAd, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(req.Mid, 10))
	params.Set("buvid", req.Buvid)
	params.Set("resource", xstr.JoinInts(req.Resource))
	params.Set("ip", ip)
	params.Set("country", req.Country)
	params.Set("province", req.Province)
	params.Set("city", req.City)
	params.Set("network", req.Network)
	params.Set("build", strconv.Itoa(req.Build))
	params.Set("mobi_app", req.MobiApp)
	params.Set("device", req.Device)
	params.Set("open_event", req.OpenEvent)
	params.Set("ad_extra", req.AdExtra)
	params.Set("may_resist_gif", strconv.Itoa(req.MayResistGif))
	// 老接口做兼容
	if req.Style > 0 {
		//nolint:gomnd
		if req.Style > 3 {
			req.Style = 1
		}
		params.Set("style", strconv.Itoa(req.Style))
	}
	var res struct {
		Code int       `json:"code"`
		Data *cm.NewAd `json:"data"`
	}
	if err := d.client.Get(ctx, d.cfg.Host+_newAD, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.cfg.Host+_newAD+"?"+params.Encode())
	}
	return res.Data, nil
}
