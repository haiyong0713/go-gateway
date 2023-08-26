package medialist

import (
	"context"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"

	medialistmdl "go-gateway/app/app-svr/app-dynamic/interface/model/medialist"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"

	"github.com/pkg/errors"
)

const (
	_favoriteDetail = "/x/internal/medialist/v1/dynamic/mget"
)

func (d *Dao) FavoriteDetail(c context.Context, ids []int64) (map[int64]*medialistmdl.FavoriteItem, error) {
	params := url.Values{}
	params.Set("rids", xstr.JoinInts(ids))
	favDetailURL := d.c.Hosts.ApiCo + _favoriteDetail
	var ret struct {
		Code int                       `json:"code"`
		Msg  string                    `json:"msg"`
		Data *medialistmdl.FavoriteRes `json:"data"`
	}
	if err := d.client.Get(c, favDetailURL, "", params, &ret); err != nil {
		xmetric.DyanmicItemAPI.Inc(favDetailURL, "request_error")
		log.Error("FavoriteDetail http GET(%s) failed, params:(%s), error(%+v)", favDetailURL, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		xmetric.DyanmicItemAPI.Inc(favDetailURL, "reply_code_error")
		log.Error("FavoriteDetail http GET(%s) failed, params:(%s), code: %v, msg: %v", favDetailURL, params.Encode(), ret.Code, ret.Msg)
		err := errors.Wrapf(ecode.Int(ret.Code), "FavoriteDetail url(%v) code(%v) msg(%v)", favDetailURL, ret.Code, ret.Msg)
		return nil, err
	}
	if ret.Data == nil {
		log.Warn("FavoriteDetail ret.Data nil")
		return nil, nil
	}
	return ret.Data.Cards, nil
}
