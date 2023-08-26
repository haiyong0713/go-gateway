package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-gateway/app/web-svr/web/job/internal/model"

	"github.com/pkg/errors"
)

const _onlineListURI = "/x/internal/chat/num/top/aid"

func (d *dao) OnlineAids(ctx context.Context, num int64) ([]*model.OnlineAid, error) {
	params := url.Values{}
	params.Set("num", strconv.FormatInt(num, 10))
	var res struct {
		Code int                `json:"code"`
		Data []*model.OnlineAid `json:"data"`
	}
	if err := d.httpR.Get(ctx, d.onlineListURL, "", params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, ecode.Int(res.Code)
	}
	if len(res.Data) == 0 {
		return nil, errors.New("online aids is nil")
	}
	return res.Data, nil
}
