package ad

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/splash"

	adgrpc "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/api"
	advo "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/vo"
)

const (
	_splashListURL = "/bce/api/splash/list"
	_splashShowURL = "/bce/api/splash/show"
)

// Dao is advertising dao.
type Dao struct {
	client        *httpx.Client
	splashListURL string
	splashShowURL string
	adClient      adgrpc.SunspotClient
}

// New advertising dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:        httpx.NewClient(c.HTTPClient, httpx.SetResolver(resolver.New(nil, discovery.Builder()))),
		splashListURL: c.HostDiscovery.AdDiscovery + _splashListURL,
		splashShowURL: c.HostDiscovery.AdDiscovery + _splashShowURL,
	}
	var err error
	if d.adClient, err = adgrpc.NewClient(c.ADClient); err != nil {
		panic(fmt.Sprintf("ad new client err(%+v)", err))
	}
	return
}

// SplashList ad splash list
func (d *Dao) SplashList(c context.Context, mobiApp, device, buvid, birth, adExtra string, height, width, build int, mid, bannerResource int64, userAgent, network, loadedCreativeList, clientKeepIds string) (res []*splash.List, config *splash.CmConfig, topview map[int64]int64, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("build", strconv.Itoa(build))
	params.Set("buvid", buvid)
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("ip", ip)
	params.Set("height", strconv.Itoa(height))
	params.Set("width", strconv.Itoa(width))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("network", network)
	params.Set("loaded_creative_list", loadedCreativeList)
	params.Set("client_keep_ids", clientKeepIds)
	if bannerResource > 0 {
		params.Set("banner_resource", strconv.FormatInt(bannerResource, 10))
	}
	if birth != "" {
		params.Set("birth", birth)
	}
	if adExtra != "" {
		params.Set("ad_extra", adExtra)
	}
	var data *SplashListData
	var req *http.Request
	if req, err = d.client.NewRequest(http.MethodGet, d.splashListURL, ip, params); err != nil {
		log.Error("cpm splash url(%s) error(%v)", d.splashListURL+"?"+params.Encode(), err)
		return
	}
	req.Header.Set("User-Agent", userAgent)
	if err = d.client.Do(c, req, &data); err != nil {
		log.Error("cpm splash url(%s) error(%v)", d.splashListURL+"?"+params.Encode(), err)
		return
	}
	b, _ := data.MarshalJSON()
	log.Info("cpm splash list url(%s) response(%s)", d.splashListURL+"?"+params.Encode(), b)
	if data.Code != 0 {
		err = ecode.Int(data.Code)
		log.Error("cpm splash url(%s) code(%d)", d.splashListURL+"?"+params.Encode(), data.Code)
		return
	}
	topview = map[int64]int64{}
	for _, t := range data.Data {
		s := &splash.List{}
		*s = *t
		if s.IsTopview {
			topview[s.ID] = s.TopViewID //用于替换show里面的id
			s.Type = 4                  //topview闪屏
			s.ID = s.TopViewID
		}
		s.RequestID = data.RequestID
		s.ClientIP = ip
		s.IsAdLoc = true
		res = append(res, s)
	}
	config = data.CmConfig
	return
}

// SplashShow ad splash show
func (d *Dao) SplashShow(c context.Context, mobiApp, device, buvid, birth, adExtra string, height, width, build int, mid int64, userAgent, network string) (res []*splash.Show, requestID string, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("build", strconv.Itoa(build))
	params.Set("buvid", buvid)
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("ip", ip)
	params.Set("height", strconv.Itoa(height))
	params.Set("width", strconv.Itoa(width))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("network", network)
	if birth != "" {
		params.Set("birth", birth)
	}
	if adExtra != "" {
		params.Set("ad_extra", adExtra)
	}
	var data struct {
		Code            int            `json:"code"`
		Data            []*splash.Show `json:"data"`
		SplashRequestId string         `json:"splash_request_id"`
	}
	var req *http.Request
	if req, err = d.client.NewRequest(http.MethodGet, d.splashShowURL, ip, params); err != nil {
		log.Error("cpm splash url(%s) error(%v)", d.splashShowURL+"?"+params.Encode(), err)
		return
	}
	req.Header.Set("User-Agent", userAgent)
	if err = d.client.Do(c, req, &data); err != nil {
		log.Error("cpm splash url(%s) error(%v)", d.splashShowURL+"?"+params.Encode(), err)
		return
	}
	if data.Code != 0 {
		err = ecode.Int(data.Code)
		log.Error("cpm splash url(%s) code(%d)", d.splashShowURL+"?"+params.Encode(), data.Code)
		return
	}
	res = data.Data
	requestID = data.SplashRequestId
	return
}

func (d *Dao) SplashShowSearch(ctx context.Context, param *advo.SspSplashRequestVo) (*advo.SspSplashShowResponseVo, error) {
	out, err := d.adClient.SplashShowSearch(ctx, param)
	if err != nil {
		return nil, err
	}
	return out, nil
}
