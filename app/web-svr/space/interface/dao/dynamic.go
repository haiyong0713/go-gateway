package dao

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"go-common/library/ecode"

	dynamicFeed "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"

	"github.com/pkg/errors"
)

const (
	_dynamicInfoURI = "/dynamic_svr/v0/dynamic_svr/get_dynamic_info"
)

func (d *Dao) DynamicSearch(ctx context.Context, mid, vmid int64, keyword string, pn int, ps int) (dynamicIDs []int64, total int32, err error) {
	req := &dynamicFeed.PersonalSearchReq{
		Keywords: keyword,
		Pn:       int32(pn),
		Ps:       int32(ps),
		Mid:      mid,
		UpId:     vmid,
	}
	reply, err := d.dynamicFeedClient.PersonalSearch(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	var ids []int64
	for _, v := range reply.GetDynamics() {
		ids = append(ids, v.DynamicId)
	}
	return ids, reply.GetTotal(), nil
}

func (d *Dao) DynamicDetail(ctx context.Context, mid int64, dynamicIDs []int64) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	for _, id := range dynamicIDs {
		params.Add("dynamic_ids[]", strconv.FormatInt(id, 10))
	}
	var res struct {
		Code int `json:"code"`
		Data struct {
			Cards json.RawMessage `json:"cards"`
		} `json:"data"`
	}
	if err := d.httpDynamic.Get(ctx, d.dynamicInfoURL, "", params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.dynamicInfoURL+"?"+params.Encode())
	}
	return res.Data.Cards, nil
}
