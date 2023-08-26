// Package campus this is used for PC Campus
package campus

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"

	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/model/campus"
	"go-gateway/app/web-svr/web/interface/model/rcmd"

	dycommon "git.bilibili.co/bapis/bapis-go/dynamic/common"
	campusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
	"github.com/pkg/errors"
)

const (
	_campusNearbyRcmdURI = "/recommand"
	_campusNearbyRcmdCmd = "campus_nearby"
	_campusNearby        = "/data/rank/reco-campus-nearby-zb.json" // 灾备
	_webPlat             = 30
	_fromPC              = dycommon.CampusReqFromType_PC // 来源类型
)

// Dao web campus dao.
type Dao struct {
	c            *conf.Config
	cpClient     campusgrpc.CampusSvrClient
	nyClient     *bm.Client
	nyRcmdUri    string
	nyHotRcmdUri string
}

// New get campus dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:            c,
		nyClient:     bm.NewClient(c.HTTPClient.Rcmd, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		nyRcmdUri:    c.Host.CampusRcmdDiscovery + _campusNearbyRcmdURI,
		nyHotRcmdUri: c.Host.CampusRcmdDiscovery + _campusNearby,
	}
	var err error
	if d.cpClient, err = campusgrpc.NewClient(c.CampusGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Pages(c context.Context, req *campus.CampusRcmdReq) (*campusgrpc.PagesReply, error) {
	res, err := d.cpClient.Pages(c, &campusgrpc.PagesReq{
		FromType:   _fromPC,
		Uid:        uint64(req.Mid),
		CampusId:   uint64(req.CampusId),
		CampusName: req.CampusName,
		IpAddr:     metadata.String(c, metadata.RemoteIP),
		Lat:        req.Lat,
		Lng:        req.Lng,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) SchoolSearch(c context.Context, keywords string, ps, offset uint64) (*campusgrpc.SearchReply, error) {
	req := &campusgrpc.SchoolSearchReq{
		Keywords: keywords,
		PageSize: ps,
		Offset:   offset,
		FromType: _fromPC,
	}
	res, err := d.cpClient.SchoolSearch(c, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) SchoolRecommend(c context.Context, mid uint64, lat, lng float32) ([]*campusgrpc.CampusInfo, error) {
	req := &campusgrpc.SchoolRecommendReq{
		Mid:       mid,
		Latitude:  lat,
		Longitude: lng,
		Ip:        metadata.String(c, metadata.RemoteIP),
	}
	res, err := d.cpClient.SchoolRecommend(c, req)
	if err != nil {
		return nil, err
	}
	return res.Results, nil
}

func (d *Dao) OfficialAccounts(c context.Context, req *campus.CampusOfficialReq) (*campusgrpc.OfficialAccountsReply, error) {
	param := &campusgrpc.OfficialAccountsReq{
		FromType:   _fromPC,
		CampusId:   req.CampusId,
		CampusName: req.CampusName,
		Uid:        uint64(req.Mid),
		Offset:     req.Offset,
	}
	reply, err := d.cpClient.OfficialAccounts(c, param)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) TopicList(c context.Context, mid, campusId uint64, offset int64) (*campusgrpc.TopicListReply, error) {
	req := &campusgrpc.TopicListReq{
		FromType: _fromPC,
		Uid:      mid,
		CampusId: campusId,
		Offset:   strconv.FormatInt(offset, 10),
	}
	reply, err := d.cpClient.TopicList(c, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) OfficialDynamics(c context.Context, req *campus.CampusOfficialReq) (*campusgrpc.OfficialDynamicsReply, error) {
	param := &campusgrpc.OfficialDynamicsReq{
		FromType:   _fromPC,
		CampusId:   req.CampusId,
		CampusName: req.CampusName,
		Uid:        uint64(req.Mid),
		Offset:     req.Offset,
	}
	reply, err := d.cpClient.OfficialDynamics(c, param)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) CampusFeedback(c context.Context, req *campus.CampusFeedbackReq) error {
	var list []*campusgrpc.FeedbackInfo
	for _, item := range req.List {
		BizId, err := strconv.ParseInt(item.BizId, 10, 64)
		if err != nil {
			continue
		}
		list = append(list, &campusgrpc.FeedbackInfo{
			BizType:  item.BizType,
			BizId:    BizId,
			CampusId: item.CampusId,
			Reason:   item.Reason,
		})
	}
	param := &campusgrpc.FeedbackReq{
		ReqFromType: _fromPC,
		Mid:         int64(req.Mid),
		List:        list,
		FromType:    req.From,
	}
	_, err := d.cpClient.Feedback(c, param)
	return err
}

func (d *Dao) CampusBillboard(c context.Context, mid, campus_id int64, version_code string) (*campusgrpc.BillboardReply, error) {
	req := &campusgrpc.BillboardReq{
		FromType:    _fromPC,
		Mid:         mid,
		CampusId:    campus_id,
		VersionCode: version_code,
	}
	res, err := d.cpClient.Billboard(c, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("unexpected nil campus Billboard info")
	}
	return res, err
}

func (d *Dao) CampusNearbyRcmd(c context.Context, req *campus.CampusNearbyRcmdReq) (data []*rcmd.AITopRcmd, code int, userFeature string, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	req.Ip = ip
	params := url.Values{}
	timeout := time.Duration(d.c.HTTPClient.Read.Timeout) / time.Millisecond
	params.Set("cmd", _campusNearbyRcmdCmd)
	params.Set("page_type", "pc_nearby")
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Add("mid", strconv.FormatInt(req.Mid, 10))
	params.Add("buvid", req.Buvid)
	params.Add("plat", strconv.FormatInt(_webPlat, 10))
	params.Add("request_cnt", strconv.Itoa(req.Ps))
	params.Add("page_no", strconv.Itoa(req.Pn))
	params.Add("user_campus_id", strconv.Itoa(req.CampusId))
	params.Add("previous_campus_id", strconv.Itoa(req.PreCampusId))
	params.Add("moment_campus_id", strconv.Itoa(0))
	params.Add("fresh_type", strconv.Itoa(req.FreshType))
	params.Add("ip", ip)
	var res struct {
		Code        int               `json:"code"`
		Data        []*rcmd.AITopRcmd `json:"data"` // 目前结构类似，先使用该结构体，后续有必要再修改
		UserFeature string            `json:"user_feature"`
	}
	if err = d.nyClient.Get(c, d.nyRcmdUri, ip, params, &res); err != nil {
		return nil, ecode.ServerErr.Code(), "", err
	}
	if res.Code != ecode.OK.Code() {
		if res.Code == -3 || res.Code == -11 {
			// 稿件不足时，有多少返回多少
			return res.Data, res.Code, res.UserFeature, nil
		}
		return nil, res.Code, "", errors.Wrap(ecode.Int(res.Code), d.nyRcmdUri+"?"+params.Encode())
	}
	return res.Data, res.Code, res.UserFeature, nil
}

// 校园灾备
func (d *Dao) CampusHotRcmd(c context.Context) (data []*rcmd.AITopRcmd, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	var res struct {
		Code int               `json:"code"`
		Data []*rcmd.AITopRcmd `json:"data"` // 目前结构类似，先使用该结构体，后续有必要再修改
	}
	if err = d.nyClient.Get(c, d.nyHotRcmdUri, ip, url.Values{}, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
		return
	}
	for _, v := range res.Data {
		if v.Goto == "av" && v.ID > 0 {
			data = append(data, v)
		}
	}
	return
}

func (d *Dao) RedDot(c context.Context, req *campus.CampusRedDotReq) (*campus.CampusRedDotReply, error) {
	param := &campusgrpc.RedDotReq{
		FromType: _fromPC,
		CampusId: req.CampusId,
		Uid:      uint64(req.Mid),
	}
	reply, err := d.cpClient.RedDot(c, param)
	if err != nil {
		return nil, err
	}
	res := &campus.CampusRedDotReply{
		RedDot: (reply.RedDot != 0),
	}
	return res, nil
}
