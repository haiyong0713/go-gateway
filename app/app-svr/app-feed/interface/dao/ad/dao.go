package ad

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-feed/interface/conf"
)

const (
	_bce   = "/bce/api/bce/wise"
	_newAD = "/bce/api/bce/feeds/oversaturated"
)

// Dao is ad dao.
type Dao struct {
	// http client
	client *bm.Client
	// ad
	bce   string
	newAD string
}

// New new a ad dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		client: bm.NewClient(c.HTTPAd, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		bce:    c.HostDiscovery.Ad + _bce,
		newAD:  c.HostDiscovery.Ad + _newAD,
	}
	return
}

func (d *Dao) Ad(c context.Context, mid int64, build int, buvid string, resource []int64, country, province, city, network, mobiApp, device, openEvent, adExtra string, style int, now time.Time) (advert *cm.Ad, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("resource", xstr.JoinInts(resource))
	params.Set("ip", ip)
	params.Set("country", country)
	params.Set("province", province)
	params.Set("city", city)
	params.Set("network", network)
	params.Set("build", strconv.Itoa(build))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("open_event", openEvent)
	params.Set("ad_extra", adExtra)
	// 老接口做兼容
	if style > 0 {
		//nolint:gomnd
		if style > 3 {
			style = 1
		}
		params.Set("style", strconv.Itoa(style))
	}
	var res struct {
		Code int    `json:"code"`
		Data *cm.Ad `json:"data"`
	}
	if err = d.client.Get(c, d.bce, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.bce+"?"+params.Encode())
		return
	}
	if res.Data != nil {
		res.Data.ClientIP = ip
	}
	advert = res.Data
	return
}

// NewAd 接口：https://info.bilibili.co/pages/viewpage.action?pageId=53620629
func (d *Dao) NewAd(c context.Context, mid int64, build int, buvid string, resource []int64, country, province, city, network, mobiApp, device, openEvent, adExtra string, style, mayResistGif int, now time.Time) (advert *cm.NewAd, respCode int, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("resource", xstr.JoinInts(resource))
	params.Set("ip", ip)
	params.Set("country", country)
	params.Set("province", province)
	params.Set("city", city)
	params.Set("network", network)
	params.Set("build", strconv.Itoa(build))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("open_event", openEvent)
	params.Set("ad_extra", adExtra)
	params.Set("may_resist_gif", strconv.Itoa(mayResistGif))
	// 老接口做兼容
	if style > 0 {
		//nolint:gomnd
		if style > 3 {
			style = 1
		}
		params.Set("style", strconv.Itoa(style))
	}
	var res struct {
		Code int       `json:"code"`
		Data *cm.NewAd `json:"data"`
	}
	if err = d.client.Get(c, d.newAD, ip, params, &res); err != nil {
		respCode = ecode.ServerErr.Code()
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.newAD+"?"+params.Encode())
		respCode = res.Code
		return
	}
	advert = res.Data
	return
}
