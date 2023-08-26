package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/appstatic/admin/model"

	"github.com/pkg/errors"
)

// CallPush calls the push server api
func (d *Dao) CallPush(ctx context.Context) (err error) {
	var (
		cfg    = d.c.Cfg.Push
		params = url.Values{}
		bs     []byte
	)
	params.Set("operation", fmt.Sprintf("%d", cfg.Operation))
	params.Set("speed", fmt.Sprintf("%d", cfg.QPS))
	msg := &model.PushMsg{
		Key:       "AppStaticPush",
		Time:      time.Now().Unix(),
		Operation: cfg.Operation,
		Qps:       cfg.QPS,
	}
	if bs, err = json.Marshal(msg); err != nil {
		return
	}
	params.Set("message", string(bs))
	var res struct {
		Code int `json:"code"`
	}
	if err = d.client.Post(ctx, cfg.URL, "", params, &res); err != nil {
		err = errors.Wrap(err, cfg.URL+"?"+params.Encode())
		log.Error("CallPush url(%s) param(%v) error(%v)", cfg.URL, params, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("CallPush url(%s) param(%v) code error code(%d) res(%v)", cfg.URL, params, res.Code, res)
		err = errors.Wrap(ecode.Int(res.Code), cfg.URL+"?"+params.Encode())
	}
	return
}
