package message

import (
	"context"
	"fmt"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	//type = 4 业务通知
	_notify = "4"
)

// Notify .
func (d *Dao) Notify(mids []int64, c *conf.MesConfig) (err error) {
	params := url.Values{}
	params.Set("mid_list", xstr.JoinInts(mids))
	params.Set("title", c.Title)
	params.Set("mc", c.MC)
	params.Set("data_type", _notify)
	params.Set("context", c.Msg)
	var res struct {
		Code int `json:"code"`
	}
	if err = d.messageHTTPClient.Post(context.Background(), d.c.Message.URL, "", params, &res); err != nil {
		log.Error("Notify d.messageHTTPClient.Post(%s) error(%+v)", d.c.Message.URL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Error("Notify url(%s) res code(%d)", d.c.Message.URL+"?"+params.Encode(), res.Code)
		err = ecode.Int(res.Code)
		return
	}
	log.Info("Notify Success! url(%s) res code(%d)", d.c.Message.URL+"?"+params.Encode(), res.Code)
	return
}

// NotifyTianma send tianma notify.
func (d *Dao) NotifyTianma(mids []int64, title, msg string) (err error) {
	config := d.c.Message.Tianma
	tmp := &conf.MesConfig{
		MC:    config.MC,
		Title: fmt.Sprintf(config.Title, title),
		Msg:   fmt.Sprintf(config.Msg, title, msg, util.CTimeDay()),
	}
	return d.Notify(mids, tmp)
}

// NotifyPopular send popular notify.
func (d *Dao) NotifyPopular(mids []int64) (err error) {
	config := d.c.Message.Popular
	tmp := &conf.MesConfig{
		MC:    config.MC,
		Title: config.Title,
		Msg:   fmt.Sprintf(config.Msg, util.CTimeDay()),
	}
	return d.Notify(mids, tmp)
}
