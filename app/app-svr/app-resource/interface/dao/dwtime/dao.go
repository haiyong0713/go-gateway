package dwtime

import (
	"context"
	"net/http"
	"net/url"

	"go-common/library/log"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	resMdl "go-gateway/app/app-svr/app-resource/interface/model/resource"
)

const (
	_cdnPeakHours = "/cdn_peak_hours"
)

// Dao is advertising dao.
type Dao struct {
	conf            *conf.Config
	client          *bm.Client
	cdnPeakHoursURL string
}

// New advertising dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		conf:            c,
		client:          httpx.NewClient(c.HTTPClient, httpx.SetResolver(resolver.New(nil, discovery.Builder()))),
		cdnPeakHoursURL: c.HostDiscovery.CommonArch + _cdnPeakHours,
	}
	return
}

// CdnPeakHours 获取cdn高低峰时间段
func (d *Dao) CdnPeakHours(c context.Context, domain string, day string) (res map[string]*resMdl.CdnDwTime, err error) {
	ip := metadata.String(c, metadata.RemoteIP)

	params := url.Values{}
	params.Set("domain", domain)
	params.Set("day", day)

	var req *http.Request
	if req, err = d.client.NewRequest(http.MethodGet, d.cdnPeakHoursURL, ip, params); err != nil {
		log.Error("CdnPeakHours url(%s) d.client.NewRequest error(%v)", d.cdnPeakHoursURL+"?"+params.Encode(), err)
		return
	}

	var resp map[string]*resMdl.CdnDwTime
	if err = d.client.Do(c, req, &resp); err != nil {
		log.Error("CdnPeakHours url(%s) d.client.Do error(%v)", d.cdnPeakHoursURL+"?"+params.Encode(), err)
		return
	}
	return resp, nil
}
