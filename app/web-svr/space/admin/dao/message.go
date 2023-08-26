package dao

import (
	"context"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
)

const (
	_sysMessageURI = "/api/notify/send.user.notify.do"
	_notify        = "4"
)

// SendMessage send system notify.
func (d *Dao) SendSystemMessage(c context.Context, mids []int64, mc, title, msg string) (err error) {
	params := url.Values{}
	params.Set("mid_list", xstr.JoinInts(mids))
	params.Set("title", title)
	params.Set("mc", mc)
	params.Set("data_type", _notify)
	params.Set("context", msg)
	var res struct {
		Code int `json:"code"`
	}
	err = d.http.Post(c, d.messageURL, "", params, &res)
	if err != nil {
		log.Error("SendSystemMessage d.client.Post(%s) error(%+v)", d.messageURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Error("SendSystemMessage url(%s) res code(%d)", d.messageURL+"?"+params.Encode(), res.Code)
		err = ecode.Int(res.Code)
	}
	return
}
