package ad

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model/ad"

	adgrpc "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/api"
	advo "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/vo"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	_adURL = "/bce/api/bce/wise"
)

// Dao dao.
type Dao struct {
	client   *bm.Client
	adURL    string
	adClient adgrpc.SunspotClient
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: bm.NewClient(conf.Conf.HTTPAD, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		adURL:  c.HostDiscovery.AD + _adURL,
	}
	var (
		err        error
		opts       []grpc.DialOption
		WindowSize = int32(65535000)
	)
	opts = append(opts, grpc.WithInitialWindowSize(WindowSize))
	opts = append(opts, grpc.WithInitialConnWindowSize(WindowSize))
	if d.adClient, err = adgrpc.NewClient(c.ADClient, opts...); err != nil {
		panic(fmt.Sprintf("ad new client err(%+v)", err))
	}
	return
}

// Ad ad request.
func (d *Dao) Ad(c context.Context, mobiApp, device, buvid string, build int, mid, upperID, aid int64, rid int32, tids []int64, resource []int64,
	network, adExtra, spmid, fromSpmid, from string) (advert *ad.Ad, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("aid", strconv.FormatInt(aid, 10))
	params.Set("build", strconv.Itoa(build))
	params.Set("device", device)
	params.Set("buvid", buvid)
	params.Set("resource", xstr.JoinInts(resource))
	params.Set("mobi_app", mobiApp)
	params.Set("ip", ip)
	params.Set("av_rid", strconv.FormatInt(int64(rid), 10))
	params.Set("av_tid", xstr.JoinInts(tids))
	params.Set("av_up_id", strconv.FormatInt(upperID, 10))
	params.Set("spmid", spmid)
	params.Set("from_spmid", fromSpmid)
	params.Set("from", from)
	if network != "" {
		params.Set("network", network)
	}
	if adExtra != "" {
		params.Set("ad_extra", adExtra)
	}
	var res struct {
		Code int    `json:"code"`
		Data *ad.Ad `json:"data"`
	}
	if err = d.client.Get(c, d.adURL, ip, params, &res); err != nil {
		return
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		err = errors.Wrap(code, d.adURL+"?"+params.Encode())
		return
	}
	if res.Data != nil {
		res.Data.ClientIP = ip
	}
	advert = res.Data
	return
}

// AdGRPC is
func (d *Dao) AdGRPC(c context.Context, mobiApp, buvid, device string, build int, mid, upperID, aid int64, rid int32, tids []int64, resource []int32, network, adExtra, spmid, fromSpmid, from string, adTab bool) (res *advo.SunspotAdReplyForView, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &advo.SunspotAdRequestForView{
		Mid:         mid,
		Aid:         aid,
		Build:       int32(build),
		Buvid:       buvid,
		Resource:    resource,
		MobiApp:     mobiApp,
		Ip:          ip,
		AvRid:       int64(rid),
		AvTid:       xstr.JoinInts(tids),
		AvUpId:      upperID,
		Spmid:       spmid,
		FromSpmid:   fromSpmid,
		Network:     network,
		AdExtra:     adExtra,
		From:        from,
		RequestType: "wise", //app端
		Device:      device,
		AdTab:       adTab,
	}
	res, err = d.adClient.AdSearch(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "d.adClient.AdSearch arg(%+v)", arg)
		return
	}
	return
}
