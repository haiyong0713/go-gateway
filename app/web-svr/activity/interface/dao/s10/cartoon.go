package s10

import (
	"bytes"
	"context"
	"encoding/json"
	"go-common/library/ecode"
	"go-common/library/log"
	"net/http"
)

const _cartoonURI = "http://manga.bilibili.co/twirp/activity.v0.Activity/S10Prize"

func (d *Dao) CartoonDiscount(ctx context.Context, mid int64, discount, uniqueID string) (err error) {
	param := &struct {
		Mid   int64  `json:"mid"`
		Prize string `json:"prize"`
		ID    string `json:"id"`
	}{
		Mid:   mid,
		Prize: discount,
		ID:    uniqueID,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Errorc(ctx, "CartoonDiscount json.Marshal param(%+v) error(%v)", param, err)
		return
	}
	var req *http.Request
	if req, err = http.NewRequest("POST", _cartoonURI, bytes.NewReader(paramJSON)); err != nil {
		log.Errorc(ctx, "CartoonDiscount http.NewRequest mid(%d) error(%v)", mid, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err = d.httpClient.Do(ctx, req, &res); err != nil {
		log.Errorc(ctx, "CartoonDiscount d.httpClient.Do mid(%d) res(%v) err(%v)", mid, res, err)
		return
	}
	if res.Code != 0 {
		log.Errorc(ctx, "CartoonDiscount mid(%d) res(%v)", mid, res)
		err = ecode.Int(res.Code)
	}
	return
}
