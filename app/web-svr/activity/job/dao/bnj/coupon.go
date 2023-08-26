package bnj

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
)

const _mallCouponURI = "/mall-marketing/coupon_code/createV2" //会员购

// 会员购优惠券.
func (d *Dao) MallCoupon(c context.Context, mid, sourceID int64, couponID, sourceActivityID string) (err error) {
	param := &struct {
		MID              int64  `json:"mid"`
		CouponID         string `json:"couponId"`
		SourceId         int64  `json:"sourceId"`
		SourceActivityId string `json:"sourceActivityId"`
		SourceBizId      string `json:"sourceBizId"`
	}{
		MID:              mid,
		CouponID:         couponID,
		SourceId:         sourceID,
		SourceActivityId: sourceActivityID,
		SourceBizId:      strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.FormatInt(mid, 10),
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Error("Mall json.Marshal param(%+v) error(%v)", param, err)
		return
	}
	var req *http.Request
	if req, err = http.NewRequest("POST", d.mallCouponURL, bytes.NewReader(paramJSON)); err != nil {
		log.Error("Mall http.NewRequest mid(%d) error(%v)", mid, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("Mall d.client.Do mid(%d) res(%v) err(%v)", mid, res, err)
		return
	}
	if res.Code != 0 {
		log.Error("Mall mid(%d) res(%v)", mid, res)
		err = ecode.Int(res.Code)
	}
	return
}
