package fawkes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"go-common/library/ecode"
	"go-gateway/app/web-svr/appstatic/admin/model"

	"github.com/pkg/errors"
)

func (d *Dao) BroadcastPush(ctx context.Context, mobiApp string, now time.Time) error {
	params := url.Values{}
	params.Set("operation", fmt.Sprintf("%d", d.c.BroadcastPush.Operation))
	params.Set("speed", fmt.Sprintf("%d", d.c.BroadcastPush.QPS))
	msg := &model.PushMsg{
		Key:       "AppStaticPush",
		Time:      now.Unix(),
		Operation: d.c.BroadcastPush.Operation,
		Qps:       d.c.BroadcastPush.QPS,
	}
	bs, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	params.Set("message", string(bs))
	params.Set("filter", fmt.Sprintf("mobi_app == %s", mobiApp))
	var res struct {
		Code int `json:"code"`
	}
	if err = d.httpClient.Post(ctx, d.c.BroadcastPush.URL, "", params, &res); err != nil {
		return errors.Wrap(err, d.c.BroadcastPush.URL+"?"+params.Encode())
	}
	if res.Code != ecode.OK.Code() {
		return errors.Wrap(ecode.Int(res.Code), d.c.BroadcastPush.URL+"?"+params.Encode())
	}
	return nil
}
