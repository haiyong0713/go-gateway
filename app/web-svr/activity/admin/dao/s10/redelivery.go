package s10

import (
	"context"
	"go-common/library/ecode"
	"net/url"
	"strconv"

	"go-common/library/log"
)

const _redeliveryPath = "/x/internal/activity/s10/redelivery"

// act:0-exchange;1-gift
func (d *Dao) Redelivery(ctx context.Context, id, mid, gid, act int64) (err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("id", strconv.FormatInt(id, 10))
	params.Set("gid", strconv.FormatInt(gid, 10))
	params.Set("act", strconv.FormatInt(act, 10))
	var res struct {
		Code int    `json:"code"`
		Data string `json:"data"`
	}
	if err = d.httpClient.Post(ctx, d.redeliveryURL, "", params, &res); err != nil {
		log.Errorc(ctx, "s10 MemberCoupon d.httpClient.Post(%s) error(%+v)", d.redeliveryURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Errorc(ctx, "s10 MemberCoupon uniqueID(%d) res(%v)", id, res)
		err = ecode.New(res.Code)
	}
	return
}
