package dao

import (
	"context"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
)

const (
	_notify = "4"
	_msgURL = "/api/notify/send.user.notify.do"
)

// SendMessage send system notify.
func (d *Dao) SendMessage(c context.Context, mids []int64, mc, title, msg string) (err error) {
	params := url.Values{}
	params.Set("mid_list", xstr.JoinInts(mids))
	params.Set("title", title)
	params.Set("mc", mc)
	params.Set("data_type", _notify)
	params.Set("context", msg)
	var res struct {
		Code int `json:"code"`
	}
	err = d.msgClient.Post(c, d.messageHost+_msgURL, "", params, &res)
	if err != nil {
		log.Error("SendMessage d.client.Post(%s) error(%+v)", _msgURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Error("SendMessage url(%s) res code(%d)", _msgURL+"?"+params.Encode(), res.Code)
		err = ecode.Int(res.Code)
	}
	return

}
