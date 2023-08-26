package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
)

const _clearCacheURI = "/x/internal/space/clear/msg"

func (d *Dao) ClearCache(c context.Context, mid, id int64, typ int) (err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("id", strconv.FormatInt(id, 10))
	params.Set("type", strconv.Itoa(typ))
	var res struct {
		Code int `json:"code"`
	}
	if err = d.http.Post(c, d.clearMsgURL, "", params, &res); err != nil {
		log.Error("ClearCache d.client.Post(%s) error(%+v)", d.clearMsgURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Error("ClearCache url(%s) res code(%d)", d.clearMsgURL+"?"+params.Encode(), res.Code)
		err = ecode.Int(res.Code)
	}
	return
}
