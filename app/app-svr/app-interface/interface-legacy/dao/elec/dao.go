package elec

import (
	"context"
	"net/url"
	"strconv"

	upr "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/elec"

	"github.com/pkg/errors"
)

const (
	_elec          = "/api/elec/info/query"
	_elecMonthRank = "1"
)

// Dao is elec dao.
type Dao struct {
	client     *httpx.Client
	elec       string
	ugcpayrank upr.UGCPayRankClient
}

// New elec dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: httpx.NewClient(c.HTTPClient),
		elec:   c.Host.Elec + _elec,
	}
	var err error
	d.ugcpayrank, err = upr.NewClient(c.ElecClient)
	if err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Info(c context.Context, mid, paymid int64) (data *elec.Info, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("pay_mid", strconv.FormatInt(paymid, 10))
	params.Set("type", _elecMonthRank)
	var res struct {
		Code int        `json:"code"`
		Data *elec.Info `json:"data"`
	}
	if err = d.client.Get(c, d.elec, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		//nolint:gomnd
		if res.Code == 500011 {
			return
		}
		err = errors.Wrap(ecode.Int(res.Code), d.elec+"?"+params.Encode())
		return
	}
	data = res.Data
	return
}

// UPRankWithPanelByUPMid .
func (d *Dao) UPRankWithPanelByUPMid(c context.Context, upmid, build, mid int64, mobiApp, platform, device string) (*upr.UPRankWithPanelReply, error) {
	rly, err := d.ugcpayrank.UPRankWithPanelByUPMid(c, &upr.RankElecUPReq{UPMID: upmid, MobiApp: mobiApp, Platform: platform, Build: build, Device: device, Mid: mid})
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	return rly, nil
}
