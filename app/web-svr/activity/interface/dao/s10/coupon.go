package s10

import (
	"bytes"
	"context"
	"encoding/json"
	"go-common/library/ecode"
	"go-common/library/log"
	"net/http"
)

const _mallCouponURI = "/mall-marketing/coupon_code/createV2" //会员购

// 会员购优惠券.
func (d *Dao) MallCoupon(ctx context.Context, mid int64, couponID, uniqueID string) (err error) {
	param := &struct {
		MID              int64  `json:"mid"`
		CouponID         string `json:"couponId"`
		SourceId         int64  `json:"sourceId"`
		SourceActivityId string `json:"sourceActivityId"`
		SourceBizId      string `json:"sourceBizId"`
	}{
		MID:              mid,
		CouponID:         couponID,
		SourceId:         10,
		SourceActivityId: "2020s10",
		SourceBizId:      uniqueID,
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Errorc(ctx, "MallCoupon json.Marshal param(%+v) error(%v)", param, err)
		return
	}
	var req *http.Request
	if req, err = http.NewRequest("POST", d.couponURL, bytes.NewReader(paramJSON)); err != nil {
		log.Errorc(ctx, "MallCoupon http.NewRequest mid(%d) error(%v)", mid, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err = d.httpClient.Do(ctx, req, &res); err != nil {
		log.Errorc(ctx, "MallCoupon d.httpClient.Do mid(%d) res(%v) err(%v)", mid, res, err)
		return
	}
	if res.Code != 0 {
		log.Errorc(ctx, "MallCoupon mid(%d) res(%v)", mid, res)
		err = ecode.Int(res.Code)
	}
	return
}
