package dynamicsvr

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/net/metadata"
	dymdl "go-gateway/app/app-svr/app-show/interface/model/dynamic"
	arcmid "go-gateway/app/app-svr/archive/middleware"
)

const (
	_dynamicInfoURI = "/dynamic_svr/v0/dynamic_svr/get_dynamic_info"
	_activeUserURI  = "/topic_svr/v1/topic_svr/get_active_users"
	_feedDynamicURI = "/topic_svr/v1/topic_svr/fetch_dynamics_v2"
	_briefDynURI    = "/topic_svr/v0/topic_svr/brief_dyns"
	_hasFeedURI     = "/topic_svr/v0/topic_svr/has_dyns"
)

// Dynamic .
func (d *Dao) Dynamic(c context.Context, resources *dymdl.Resources, platform, remoteFrom, fromSpmid string, mid int64, dyId []int64) (dyResult *dymdl.DyResult, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	needActivity := 0
	params := url.Values{}
	if len(dyId) <= 0 && (resources == nil || len(resources.Array) == 0) {
		return
	}
	if resources != nil && len(resources.Array) != 0 {
		ress, _ := json.Marshal(resources)
		params.Set("resources", string(ress))
	}
	for _, value := range dyId {
		params.Add("dynamic_ids[]", strconv.FormatInt(value, 10))
	}
	batchArg, ok := arcmid.FromContext(c)
	if ok {
		batchStr, _ := json.Marshal(batchArg)
		params.Set("batch_play_arg", string(batchStr))
		params.Set("mobi_app", batchArg.MobiApp)
		params.Set("buvid", batchArg.Buvid)
		params.Set("device", batchArg.Device)
		params.Set("build", strconv.FormatInt(batchArg.Build, 10))
		mid = batchArg.Mid
		ip = batchArg.Ip
	}
	params.Set("ip", ip)
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("platform", platform)
	params.Set("from_spmid", fromSpmid)
	params.Set("need_activity", strconv.Itoa(needActivity))
	if remoteFrom != "" {
		params.Set("from", remoteFrom)
	}
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

// ActiveUsers .
func (d *Dao) ActiveUsers(c context.Context, topicID, noLimit int64) (date *dymdl.TopicCount, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("topic_id", strconv.FormatInt(topicID, 10))
	params.Set("no_limit", strconv.FormatInt(noLimit, 10))
	var res struct {
		Code int               `json:"code"`
		Data *dymdl.TopicCount `json:"data"`
	}
	if err = d.client.Get(c, d.activeUserURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.activeUserURL+"?"+params.Encode())
		return
	}
	date = res.Data
	return
}

// FetchDynamics .
func (d *Dao) FetchDynamics(c context.Context, topicID, mid, pageSize, dySort int64, deviceID, types, platform, offset, remoteFrom, fromSpmid, scenaryFrom string) (reply *dymdl.DyReply, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	batchArg, ok := arcmid.FromContext(c)
	if ok {
		batchStr, _ := json.Marshal(batchArg)
		params.Set("batch_play_arg", string(batchStr))
		params.Set("build", strconv.FormatInt(batchArg.Build, 10))
		params.Set("mobi_app", batchArg.MobiApp)
		params.Set("buvid", batchArg.Buvid)
		params.Set("device", batchArg.Device)
		mid = batchArg.Mid
		ip = batchArg.Ip
	}
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("platform", platform)
	params.Set("topic_id", strconv.FormatInt(topicID, 10))
	params.Set("device_id", deviceID)
	params.Set("from_spmid", fromSpmid)
	params.Set("sortby", strconv.FormatInt(dySort, 10))
	params.Set("scenary_from", scenaryFrom)
	if offset != "" {
		params.Set("offset", offset)
	}
	if types != "" {
		params.Set("types", types)
	}
	params.Set("page_size", strconv.FormatInt(pageSize, 10))
	if remoteFrom != "" {
		params.Set("from", remoteFrom)
	}
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

// BriefDynamics .
func (d *Dao) BriefDynamics(c context.Context, topicID, pageSize, mid int64, types, offset string, dySort int32) (reply *dymdl.BriefReply, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	if topicID != 0 {
		params.Set("topic_id", strconv.FormatInt(topicID, 10))
	}
	params.Set("from", "activity_page")
	params.Set("offset", offset)
	if types != "" {
		params.Set("types", types)
	}
	params.Set("page_size", strconv.FormatInt(pageSize, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("sortby", strconv.FormatInt(int64(dySort), 10))
	var res struct {
		Code int               `json:"code"`
		Data *dymdl.BriefReply `json:"data"`
	}
	if err = d.client.Get(c, d.briefDynURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.briefDynURL+"?"+params.Encode())
		return
	}
	reply = res.Data
	return
}

func (d *Dao) HasFeed(c context.Context, topicID, dySort int64, types string) (rly uint32, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("topic_id", strconv.FormatInt(topicID, 10))
	params.Set("sortby", strconv.FormatInt(dySort, 10))
	if types != "" {
		params.Set("types", types)
	}
	var res struct {
		Code int `json:"code"`
		Data *struct {
			HasDyns uint32 `json:"has_dyns"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.hasFeedURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.hasFeedURL+"?"+params.Encode())
		return
	}
	if res.Data != nil {
		rly = res.Data.HasDyns
	}
	return
}

// ActPromoIconVisible 暂时关闭解决go-main引用问题
//func (d *Dao) ActPromoIconVisible(c context.Context, mid, tagID int64) (*dyngrpc.ActPromoIconVisibleRsp, error) {
//	return d.dynClient.ActPromoIconVisible(c, &dyngrpc.ActPromoIconVisibleReq{Mid: mid, TagId: tagID})
//}
