package dynamic

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	dymdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"

	"github.com/pkg/errors"
)

const (
	_feedDynamicURI = "/topic_svr/v1/topic_svr/fetch_dynamics_v2"
	_briefDynURI    = "/topic_svr/v0/topic_svr/brief_dyns"
)

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

func (d *Dao) BriefDynamics(c context.Context, topicID, pageSize, mid int64, types, topicName string, dySort int32) (reply *dymdl.BriefReply, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	if topicID != 0 {
		params.Set("topic_id", strconv.FormatInt(topicID, 10))
	}
	if topicName != "" {
		params.Set("topic_name", topicName)
	}
	params.Set("from", "activity_page")
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
