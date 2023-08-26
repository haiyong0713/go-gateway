package ad

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/cm"

	"github.com/pkg/errors"
)

const (
	_bce            = "/bce/sunspot/facade/api/upzone/ad"
	_bceV2          = "/bce/sunspot/facade/api/upzone/ad/v2"
	_topicList      = "/basc/api/open_api/v1/topic/list"
	_pickupEntrance = "/commercialorder/api/open_api/v1/up/space/pickup/entrance"
)

// Dao is ad dao.
type Dao struct {
	// http client
	client         *httpx.Client
	topicClient    *httpx.Client
	entranceClient *httpx.Client
	// ad
	bce string
	// ad v2
	bceV2 string
	// space created list
	topicList string
	// space pickup entrance
	pickupEntrance string
}

// New new a ad dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		client:         httpx.NewClient(c.HTTPAd),
		topicClient:    httpx.NewClient(c.HTTPAdTopic),
		entranceClient: httpx.NewClient(c.HTTPEntrance),
		bce:            c.Host.Ad + _bce,
		bceV2:          c.Host.Ad + _bceV2,
		topicList:      c.Host.AdTopic + _topicList,
		pickupEntrance: c.Host.Ad + _pickupEntrance,
	}
	return
}

func (d *Dao) Ad(c context.Context, mid, vmid int64, build int, buvid string, resource []int64, network, mobiApp, device, adExtra, spmid, fromSpmid string) (advert *cm.Ad, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("resource", xstr.JoinInts(resource))
	params.Set("ip", ip)
	params.Set("network", network)
	params.Set("build", strconv.Itoa(build))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("av_up_id", strconv.FormatInt(vmid, 10))
	params.Set("ad_extra", adExtra)
	params.Set("spmid", spmid)
	params.Set("from_spmid", fromSpmid)
	var res struct {
		Code int `json:"code"`
		Data struct {
			ShopEntrance *cm.Ad `json:"shop_entrance"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.bce, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.bce+"?"+params.Encode())
		return
	}
	advert = res.Data.ShopEntrance
	return
}

// AdVTwo .
func (d *Dao) AdVTwo(c context.Context, mid, vmid int64, build int, buvid string, resource []int64, network, mobiApp, device, adExtra, spmid, fromSpmid string) (advert *cm.Ad, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("resource", xstr.JoinInts(resource))
	params.Set("ip", ip)
	params.Set("network", network)
	params.Set("build", strconv.Itoa(build))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("av_up_id", strconv.FormatInt(vmid, 10))
	params.Set("ad_extra", adExtra)
	params.Set("spmid", spmid)
	params.Set("from_spmid", fromSpmid)
	var res struct {
		Code int `json:"code"`
		Data struct {
			ShopEntrance *cm.Ad `json:"shop_entrance"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.bceV2, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.bceV2+"?"+params.Encode())
		return
	}
	advert = res.Data.ShopEntrance
	return
}

func (d *Dao) CreatedTopicList(ctx context.Context, mid int64) ([]*cm.Topic, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	var res struct {
		Code    int         `json:"code"`
		Data    []*cm.Topic `json:"data"`
		Message string      `json:"message"`
	}
	if err := d.topicClient.Get(ctx, d.topicList, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrapf(ecode.Int(res.Code), "url:%s msg:%s", d.topicList+"?"+params.Encode(), res.Message)
	}
	return res.Data, nil
}

func (d *Dao) PickupEntrance(ctx context.Context, mid, vmid int64) (*cm.PickupEntrance, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("up_space_mid", strconv.FormatInt(vmid, 10))
	params.Set("visiting_mid", strconv.FormatInt(mid, 10))
	params.Set("ts", strconv.FormatInt(time.Now().Unix()*1000, 10))
	var res struct {
		Code    int                `json:"code"`
		Message string             `json:"message"`
		Result  *cm.PickupEntrance `json:"result"`
	}
	if err := d.entranceClient.Get(ctx, d.pickupEntrance, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrapf(ecode.Int(res.Code), "url:%s msg:%s", d.pickupEntrance+"?"+params.Encode(), res.Message)
	}
	return res.Result, nil
}
