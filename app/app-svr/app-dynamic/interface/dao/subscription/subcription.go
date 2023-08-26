package subscription

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"

	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	"go-gateway/app/app-svr/app-dynamic/interface/model/subscription"

	"github.com/pkg/errors"
)

const _dynamicSubCard = "/x/internal/tunnel/materials/card/dynamic"

func (d *Dao) Subscription(c context.Context, ids []int64, mid int64) (map[int64]*subscription.Subscription, error) {
	params := url.Values{}
	params.Set("oids", xstr.JoinInts(ids))
	params.Set("mid", strconv.FormatInt(mid, 10))
	subscriptionURL := d.c.Hosts.ApiCo + _dynamicSubCard
	var ret struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data *struct {
			Info []*subscription.Subscription `json:"info"`
		} `json:"data"`
	}
	if err := d.client.Get(c, subscriptionURL, "", params, &ret); err != nil {
		xmetric.DyanmicItemAPI.Inc(subscriptionURL, "request_error")
		log.Error("Subscription http GET(%s) failed, params:(%s), error(%+v)", subscriptionURL, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		xmetric.DyanmicItemAPI.Inc(subscriptionURL, "reply_code_error")
		log.Error("Subscription http GET(%s) failed, params:(%s), code: %v, msg: %v", subscriptionURL, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "AudioDetail url(%v) code(%v) msg(%v)", subscriptionURL, ret.Code, ret.Msg)
		return nil, err
	}
	if ret.Data == nil || len(ret.Data.Info) == 0 {
		xmetric.DyanmicItemAPI.Inc(subscriptionURL, "reply_data_error")
		err := errors.Errorf("Subscription get nothing ids %v", ids)
		log.Error("Subscription err %v", err)
		return nil, err
	}
	var res = make(map[int64]*subscription.Subscription)
	for _, sub := range ret.Data.Info {
		if sub == nil {
			continue
		}
		res[sub.OID] = sub
	}
	return res, nil
}
