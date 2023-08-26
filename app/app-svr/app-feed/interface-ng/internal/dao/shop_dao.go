package dao

import (
	"context"
	"net/url"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-card/interface/model/card/show"

	"github.com/pkg/errors"
)

const _getCard = "/api/ticket/project/getcard"

type shopConfig struct {
	Host string
}

type shopDao struct {
	client *bm.Client
	cfg    shopConfig
}

func (d *shopDao) Card(ctx context.Context, ids []int64) (map[int64]*show.Shopping, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("id", xstr.JoinInts(ids))
	params.Set("for", "1")
	params.Set("price", "1")
	var res struct {
		Code int              `json:"errno"`
		Data []*show.Shopping `json:"data"`
	}
	if err := d.client.Get(ctx, d.cfg.Host+_getCard, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.New(d.cfg.Host + _getCard + "?" + params.Encode())
	}
	rs := make(map[int64]*show.Shopping, len(res.Data))
	for _, r := range res.Data {
		rs[r.ID] = r
	}
	return rs, nil
}
