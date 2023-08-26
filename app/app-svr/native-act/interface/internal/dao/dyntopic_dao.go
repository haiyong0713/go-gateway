package dao

import (
	"context"
	"net/url"
	"strconv"

	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	xhttp "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

const (
	_briefDyns   = "/topic_svr/v0/topic_svr/brief_dyns"
	_activeUsers = "/topic_svr/v1/topic_svr/get_active_users"
)

type dyntopicDao struct {
	client     dyntopicgrpc.TopicClient
	host       string
	httpClient *xhttp.Client
}

func (d *dyntopicDao) HasDyns(c context.Context, req *dyntopicgrpc.HasDynsReq) (*dyntopicgrpc.HasDynsRsp, error) {
	rly, err := d.client.HasDyns(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *dyntopicDao) ListDyns(c context.Context, req *dyntopicgrpc.ListDynsReq) (*dyntopicgrpc.ListDynsRsp, error) {
	rly, err := d.client.ListDyns(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *dyntopicDao) BriefDyns(c context.Context, req *model.BriefDynsReq) (*model.BriefDynsRly, error) {
	params := url.Values{}
	if req.TopicID > 0 {
		params.Set("topic_id", strconv.FormatInt(req.TopicID, 10))
	}
	params.Set("from", req.From)
	params.Set("offset", req.Offset)
	if req.Types != "" {
		params.Set("types", req.Types)
	}
	params.Set("page_size", strconv.FormatInt(req.Ps, 10))
	params.Set("mid", strconv.FormatInt(req.Mid, 10))
	params.Set("sortby", strconv.FormatInt(req.SortBy, 10))
	var res struct {
		Code int                 `json:"code"`
		Data *model.BriefDynsRly `json:"data"`
	}
	ip := metadata.String(c, metadata.RemoteIP)
	if err := d.httpClient.Get(c, d.host+_briefDyns, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.host+_briefDyns+"?"+params.Encode())
	}
	return res.Data, nil
}

func (d *dyntopicDao) ActiveUsers(c context.Context, req *model.ActiveUsersReq) (*model.ActiveUsersRly, error) {
	params := url.Values{}
	params.Set("topic_id", strconv.FormatInt(req.TopicID, 10))
	params.Set("no_limit", strconv.FormatInt(req.NoLimit, 10))
	var res struct {
		Code int                   `json:"code"`
		Data *model.ActiveUsersRly `json:"data"`
	}
	ip := metadata.String(c, metadata.RemoteIP)
	if err := d.httpClient.Get(c, d.host+_activeUsers, ip, params, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.host+_activeUsers+"?"+params.Encode())
	}
	return res.Data, nil
}
