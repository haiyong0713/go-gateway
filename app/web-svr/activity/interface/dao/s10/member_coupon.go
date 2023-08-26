package s10

import (
	"context"
	"go-common/library/log"
	"net/url"
	"strconv"
)

const _memberCouponURI = "/x/internal/coupon/allowance/receive"

func (d *Dao) MemberCoupon(ctx context.Context, mid int64, batchToken, uniqueID string) (err error) {
	midStr := strconv.FormatInt(mid, 10)
	params := url.Values{}
	params.Set("mid", midStr)
	params.Set("batch_token", batchToken)
	params.Set("order_no", uniqueID)
	var res struct {
		Code int    `json:"code"`
		Data string `json:"data"`
	}
	if err = d.httpClient.Post(ctx, d.memberCouponURL, "", params, &res); err != nil {
		log.Errorc(ctx, "s10 MemberCoupon d.httpClient.Post(%s) error(%+v)", d.couponURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Errorc(ctx, "s10 MemberCoupon uniqueID(%s) res(%v)", uniqueID, res)
		return
	}
	return
}
