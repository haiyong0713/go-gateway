package live

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-wall/interface/conf"

	xgift "git.bilibili.co/bapis/bapis-go/live/xgift/v1"
	"github.com/pkg/errors"
)

const (
	_addVipURL = "/user/v0/Vip/addVip"
)

// Dao is live dao
type Dao struct {
	client    *httpx.Client
	addVipURL string
	xgiftGRPC xgift.GiftClient
}

// New live dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:    httpx.NewClient(c.HTTPClient),
		addVipURL: c.Host.APILive + _addVipURL,
	}
	var err error
	if d.xgiftGRPC, err = xgift.NewClientGift(c.LiveGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Pack(ctx context.Context, mid, cardType int64) (err error) {
	args := &xgift.GiftSmsRewardReq{
		Uid:      mid,
		CardType: cardType,
	}
	reply, err := d.xgiftGRPC.SmsReward(ctx, args)
	if err != nil {
		err = errors.Wrapf(err, "%v", args)
		return
	}
	log.Warn("live pack success mid:%v,gift:%+v", mid, reply.GetGift())
	return
}

func (d *Dao) AddVIP(c context.Context, mid int64, day int) (msg string, err error) {
	params := url.Values{}
	params.Set("vip_type", "1")
	params.Set("day", strconv.Itoa(day))
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("platform", "main")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err = d.client.Post(c, d.addVipURL, "", params, &res); err != nil {
		err = errors.Wrap(err, d.addVipURL+"?"+params.Encode())
		return "", err
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), d.addVipURL+"?"+params.Encode())
		return res.Msg, err
	}
	return res.Msg, nil
}
