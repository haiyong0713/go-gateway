package show

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"

	"go-gateway/app/app-svr/app-show/interface/model/recommend"

	"github.com/pkg/errors"
)

// AIAggregation .
func (d *Dao) AIAggregation(ctx context.Context, hotID int64) (aggSc []*recommend.CardList, err error) {
	params := url.Values{}
	params.Set("hotword_id", strconv.Itoa(int(hotID)))
	var res struct {
		Code int                   `json:"code"`
		List []*recommend.CardList `json:"list"`
	}
	if err = d.client.RESTfulGet(ctx, d.aggURL, "", params, &res, hotID); err != nil {
		err = errors.Wrapf(err, "d.client.Get(%s)", d.aggURL+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Code), "d.client.Get(%s)", d.aggURL+"?"+params.Encode())
		return
	}
	aggSc = res.List
	return
}
