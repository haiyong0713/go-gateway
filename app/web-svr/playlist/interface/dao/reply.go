package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
)

const (
	_replyType  = "18"
	_replyState = "0" // 0: open, 1: close
)

// RegReply opens playlist's reply.
func (d *Dao) RegReply(c context.Context, pid, mid int64) (err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("oid", strconv.FormatInt(pid, 10))
	params.Set("type", _replyType)
	params.Set("state", _replyState)
	var res struct {
		Code int `json:"code"`
	}
	if err = d.http.Post(c, d.replyURL, "", params, &res); err != nil {
		log.Error("d.http.Post(%s) error(%v)", d.replyURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("d.http.Post(%s) error(%v)", d.replyURL+"?"+params.Encode(), err)
		err = ecode.Int(res.Code)
	}
	return
}
