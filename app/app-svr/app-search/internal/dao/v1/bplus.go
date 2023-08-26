package v1

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-search/internal/model/search"

	"github.com/pkg/errors"
)

// DynamicDetails get dynamic details by ids.
func (d *dao) DynamicDetails(c context.Context, ids []int64, from string) (details map[int64]*search.Detail, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("from", from)
	for _, id := range ids {
		params.Add("dynamic_ids[]", strconv.FormatInt(id, 10))
	}
	var res struct {
		Code int `json:"code"`
		Data *struct {
			List []*search.Detail `json:"list"`
		} `json:"data"`
	}
	details = make(map[int64]*search.Detail)
	if err = d.client.Get(c, d.dynamicDetail, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.dynamicDetail+"?"+params.Encode())
		return
	}
	if res.Data != nil {
		for _, detail := range res.Data.List {
			if detail.ID != 0 {
				details[detail.ID] = detail
			}
		}
	}
	return
}

// DynamicTopics .
func (d *dao) DynamicTopics(c context.Context, dynamicIDs []int64, platform, mobiApp string, build int) (cs map[int64]*search.DynamicTopics, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("dynamic_ids", xstr.JoinInts(dynamicIDs))
	params.Set("platform", platform)
	params.Set("build", strconv.Itoa(build))
	params.Set("mobi_app", mobiApp)
	var res struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		Message string `json:"message"`
		Data    struct {
			Items []*search.DynamicTopics `json:"items"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.dynamicTopics, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.dynamicTopics+"?"+params.Encode())
		return
	}
	cs = make(map[int64]*search.DynamicTopics, len(res.Data.Items))
	for _, v := range res.Data.Items {
		cs[v.DynamicID] = v
	}
	return
}
