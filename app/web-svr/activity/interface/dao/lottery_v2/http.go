package lottery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	lottery "go-gateway/app/web-svr/activity/interface/model/lottery_v2"

	"github.com/pkg/errors"
)

// SendSysMsg send sys msg.
func (d *dao) SendSysMsg(c context.Context, uids []int64, mc, title string, context string, ip string) (err error) {
	params := url.Values{}
	params.Set("mc", mc)
	params.Set("title", title)
	params.Set("data_type", "4")
	params.Set("context", context)
	params.Set("mid_list", xstr.JoinInts(uids))
	var res struct {
		Code int `json:"code"`
		Data *struct {
			Status int8   `json:"status"`
			Remark string `json:"remark"`
		} `json:"data"`
	}
	if err = d.client.Post(c, d.msgURL, ip, params, &res); err != nil {
		log.Error("SendSysMsg d.client.Post(%s) error(%+v)", d.msgURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrapf(ecode.Int(res.Code), "SendSysMsg dao.client.Post(%s,%d)", d.msgURL+"?"+params.Encode(), res.Code)
		return
	}
	log.Infoc(c, "send msg ok, resdata=%+v", res.Data)
	return
}

// GetMemberAddress ...
func (d *dao) GetMemberAddress(c context.Context, id, mid int64) (val *lottery.AddressInfo, err error) {
	var res struct {
		Errno int                  `json:"errno"`
		Msg   string               `json:"msg"`
		Data  *lottery.AddressInfo `json:"data"`
	}
	params := url.Values{}
	params.Set("app_id", d.c.Lottery.AppKey)
	params.Set("app_token", d.c.Lottery.AppToken)
	params.Set("id", strconv.FormatInt(id, 10))
	params.Set("uid", strconv.FormatInt(mid, 10))
	if err = d.client.Get(c, d.getAddressURL, "", params, &res); err != nil {
		log.Errorc(c, "GetMemberAddress:dao.client.Get id(%d) mid(%d) error(%v)", id, mid, err)
		return
	}
	if res.Errno != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Errno), d.getAddressURL+"?"+params.Encode())
	}
	val = res.Data
	return
}

func (d *dao) SendVipBuyCoupon(c context.Context, clientIP, couponID, sourceActivityID, sourceBizID, uname string, sourceID, mid int64) (err error) {
	param := &struct {
		MID              int64  `json:"mid"`
		ClientIP         string `json:"clientIp"`
		CouponID         string `json:"couponId"`
		SourceID         int64  `json:"sourceId"`
		SourceActivityID string `json:"sourceActivityId"`
		SourceBizID      string `json:"sourceBizId"`
		Uname            string `json:"uname"`
		NeedAntifraud    bool   `json:"needAntifraud"`
	}{
		MID:              mid,
		CouponID:         couponID,
		SourceID:         sourceID,
		SourceActivityID: sourceActivityID,
		SourceBizID:      sourceBizID,
		Uname:            uname,
		NeedAntifraud:    false,
		ClientIP:         clientIP,
	}
	paramJSON, err := json.Marshal(param)
	log.Infoc(c, "sendVipBuyCoupon Params (%s)", string(paramJSON))
	if err != nil {
		log.Errorc(c, "Mall json.Marshal param(%+v) error(%v)", param, err)
		return
	}
	var req *http.Request
	if req, err = http.NewRequest("POST", d.vipBuyURL, bytes.NewReader(paramJSON)); err != nil {
		log.Errorc(c, "Mall http.NewRequest mid(%d) error(%v)", mid, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err = d.client.Do(context.Background(), req, &res); err != nil {
		log.Errorc(c, "Mall d.client.Do mid(%d) res(%v) err(%v)", mid, res, err)
		return
	}
	if res.Code != 0 {
		log.Errorc(c, "Mall mid(%d) res(%v)", mid, res)
		err = ecode.Int(res.Code)
	}
	return
}

const _cartoonBnj2021URI = "/twirp/activity.v0.Activity/IsRookie"

// Rookie ...
type Rookie struct {
	Rookie int `json:"rookie"`
}

// ComicsIsRookie 是否新人
func (d *dao) ComicsIsRookie(ctx context.Context, mid int64) (isRookie int, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "rewards.comicsBnj2021Coupon mid(%d)  error:%v", mid, err)
		}
	}()
	param := &struct {
		Mid int64 `json:"uid"`
	}{
		Mid: mid,
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		return
	}
	var req *http.Request
	if req, err = http.NewRequest("POST", d.comicBnj2021CouponURL, bytes.NewReader(paramJSON)); err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int     `json:"code"`
		Msg  string  `json:"msg"`
		Data *Rookie `json:"data"`
	}
	if err = d.client.Do(ctx, req, &res); err != nil {
		return
	}
	if res.Code != 0 {
		err = fmt.Errorf("code: %v, message: %v", ecode.Int(res.Code).Error(), res.Msg)
	}
	if res.Data != nil {
		return res.Data.Rookie, err
	}
	return 0, err
}

func (d *dao) GetCluesSrc(c context.Context, uri string, timeStamp int64) (res []*lottery.Item, err error) {
	params := url.Values{}
	params.Set("t", strconv.FormatInt(timeStamp, 10))
	res = make([]*lottery.Item, 0)
	if err = d.client.Get(c, uri, "", params, &res); err != nil {
		log.Errorc(c, "Get CluesSrc(%s) error(%+v)", uri+"?"+params.Encode(), err)
		return
	}
	log.Infoc(c, "Get CluesSrc url(%s) params (%v) res(%v)", uri, params.Encode(), res)
	return
}
