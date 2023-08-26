package dynamic

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	dymdl "go-gateway/app/web-svr/activity/interface/model/dynamic"

	"github.com/pkg/errors"
)

const (
	_dynamicInfoURI = "/dynamic_svr/v0/dynamic_svr/get_dynamic_info"
	_feedDynamicURI = "/topic_svr/v1/topic_svr/fetch_dynamics"
)

// Dynamic .
func (d *Dao) Dynamic(c context.Context, resources *dymdl.Resources, mid int64) (dyResult *dymdl.DyResult, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	needActivity := 0
	params := url.Values{}
	if resources == nil || len(resources.Array) == 0 {
		return
	}
	ress, _ := json.Marshal(resources)
	params.Set("resources", string(ress))
	params.Set("need_activity", strconv.Itoa(needActivity))
	params.Set("uid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Cards []*dymdl.DyCard `json:"cards"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.dynamicInfoURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.dynamicInfoURL+"?"+params.Encode())
		return
	}
	dyResult = &dymdl.DyResult{}
	dyResult.Cards = make(map[int64]*dymdl.DyCard, len(res.Data.Cards))
	for _, v := range res.Data.Cards {
		dyResult.Cards[v.Desc.Rid] = v
	}
	return
}

// FetchDynamics .
func (d *Dao) FetchDynamics(c context.Context, topicID, mid, pageSize int64, types, topicName string, dySort int32) (reply *dymdl.DyReply, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	if topicID != 0 {
		params.Set("topic_id", strconv.FormatInt(topicID, 10))
	}
	if topicName != "" {
		params.Set("topic_name", topicName)
	}
	params.Set("uid", strconv.FormatInt(mid, 10))
	if types != "" {
		params.Set("types", types)
	}
	params.Set("page_size", strconv.FormatInt(pageSize, 10))
	params.Set("sortby", strconv.FormatInt(int64(dySort), 10))
	var res struct {
		Code int            `json:"code"`
		Data *dymdl.DyReply `json:"data"`
	}
	if err = d.client.Get(c, d.feedDynamicURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.feedDynamicURL+"?"+params.Encode())
		return
	}
	reply = res.Data
	return
}
