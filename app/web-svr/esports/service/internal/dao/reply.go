package dao

import (
	"context"
	"net/url"
	"strconv"

	xecode "go-common/library/ecode"
	"go-common/library/log"
)

const (
	_replyState   = "0" // 0: open, 1: close
	_gameOfficial = 32708316
)

var res struct {
	Code int `json:"code"`
}

// RegReply opens eports's reply.
func (d *dao) RegisterReply(ctx context.Context, maid, adid int64, replyType string) (err error) {
	params := url.Values{}
	params.Set("adid", strconv.FormatInt(adid, 10))
	params.Set("mid", strconv.FormatInt(_gameOfficial, 10))
	params.Set("oid", strconv.FormatInt(maid, 10))
	params.Set("type", replyType)
	params.Set("state", _replyState)

	if err = d.replyClient.Post(ctx, d.replyURL, "", params, &res); err != nil {
		log.Errorc(ctx, "d.replyClient.Post(%s) error(%v)", d.replyURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != xecode.OK.Code() {
		log.Errorc(ctx, "d.replyClient.Post(%s), res:%+v, error(%v)", d.replyURL+"?"+params.Encode(), res, err)
		err = xecode.Int(res.Code)
	}
	return
}
