package dao

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-feed/interface-ng/internal/model"

	"github.com/pkg/errors"
)

const (
	_dynamicDetail = "/dynamic_detail/v0/dynamic/details"
	_fromFeed      = "tianma"
)

type dynamicConfig struct {
	Host string
}

type dynamicDao struct {
	client *bm.Client
	cfg    dynamicConfig
}

func (d *dynamicDao) DynamicDetail(ctx context.Context, req *model.DynamicDetailReq, ids ...int64) (map[int64]*bplus.Picture, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Add("from", _fromFeed)
	for _, id := range ids {
		params.Add("dynamic_ids[]", strconv.FormatInt(id, 10))
	}
	pb, _ := json.Marshal(req)
	params.Add("meta", string(pb))
	var res struct {
		Code int `json:"code"`
		Data *struct {
			List []*bplus.Picture `json:"list"`
		} `json:"data"`
	}
	if err := d.client.Get(ctx, d.cfg.Host+_dynamicDetail, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.cfg.Host+_dynamicDetail+"?"+params.Encode())
	}
	if res.Data == nil {
		return nil, errors.New("empty list")
	}
	out := make(map[int64]*bplus.Picture, len(res.Data.List))
	for _, pic := range res.Data.List {
		out[pic.DynamicID] = pic
	}
	return out, nil
}
