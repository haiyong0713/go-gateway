package bplus

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/bplus"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"

	"github.com/pkg/errors"
)

const (
	_favorPlus     = "/user_ex/v1/Fav/getFavList"
	_groupsCount   = "/link_group/v1/member/created_groups_num"
	_dynamicDetail = "/dynamic_detail/v0/Dynamic/details"
	_dynamicTopics = "/topic_svr/v1/topic_svr/dyn_topics"
)

// DynamicTopics .
func (d *Dao) DynamicTopics(c context.Context, dynamicIDs []int64, platform, mobiApp string, build int) (cs map[int64]*bplus.DynamicTopics, err error) {
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
			Items []*bplus.DynamicTopics `json:"items"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.dynamicTopics, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.dynamicTopics+"?"+params.Encode())
		return
	}
	cs = make(map[int64]*bplus.DynamicTopics, len(res.Data.Items))
	for _, v := range res.Data.Items {
		cs[v.DynamicID] = v
	}
	return
}

// DynamicCount return dynamic count
func (d *Dao) DynamicCount(c context.Context, mid int64) (count int64, err error) {
	reply, err := d.dynGrpc.SpaceNum(c, &dyngrpc.SpaceNumReq{Uid: mid})
	if err != nil {
		return 0, err
	}
	return reply.DynNum, nil
}

// FavClips get fav from B+ api.
func (d *Dao) FavClips(c context.Context, mid int64, accessKey, actionKey, device, mobiApp, platform string, build, pn, ps int) (cs *bplus.Clips, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("access_key", accessKey)
	params.Set("actionKey", actionKey)
	params.Set("build", strconv.Itoa(build))
	params.Set("device", device)
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("biz_type", strconv.Itoa(bplus.CLIPS))
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	var res struct {
		Code    int          `json:"code"`
		Msg     string       `json:"msg"`
		Message string       `json:"message"`
		Data    *bplus.Clips `json:"data"`
	}
	if err = d.client.Get(c, d.favorPlus, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.favorPlus+"?"+params.Encode())
		return
	}
	cs = res.Data
	return
}

// FavAlbums get fav from B+ api.
func (d *Dao) FavAlbums(c context.Context, mid int64, accessKey, actionKey, device, mobiApp, platform string, build, pn, ps int) (as *bplus.Albums, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("access_key", accessKey)
	params.Set("actionKey", actionKey)
	params.Set("build", strconv.Itoa(build))
	params.Set("device", device)
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("biz_type", strconv.Itoa(bplus.ALBUMS))
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	var res struct {
		Code    int           `json:"code"`
		Msg     string        `json:"msg"`
		Message string        `json:"message"`
		Data    *bplus.Albums `json:"data"`
	}
	if err = d.client.Get(c, d.favorPlus, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.favorPlus+"?"+params.Encode())
		return
	}
	as = res.Data
	return
}

// GroupsCount .
func (d *Dao) GroupsCount(c context.Context, mid, vmid int64) (count int, err error) {
	var (
		req *http.Request
		ip  = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("master_uid", strconv.FormatInt(vmid, 10))
	if req, err = d.client.NewRequest(http.MethodGet, d.groupsCount, ip, params); err != nil {
		return
	}
	req.Header.Set("X-BiliLive-UID", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data *struct {
			Num int `json:"num"`
		}
	}
	if err = d.client.Do(c, req, &res); err != nil {
		err = errors.Wrapf(err, "url(%s) header(X-BiliLive-UID:%s)", req.URL.String(), req.Header.Get("X-BiliLive-UID"))
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(res.Code), "url(%s) header(X-BiliLive-UID:%s)", req.URL.String(), req.Header.Get("X-BiliLive-UID"))
		return
	}
	if res.Data != nil {
		count = res.Data.Num
	}
	return
}

// Dynamic .
func (d *Dao) Dynamic(c context.Context, uid int64) (has bool, err error) {
	reply, err := d.dynGrpc.SpaceNum(c, &dyngrpc.SpaceNumReq{Uid: uid})
	if err != nil {
		return false, err
	}
	return reply.DynNum > 0, nil
}

// DynamicDetails get dynamic details by ids.
func (d *Dao) DynamicDetails(c context.Context, ids []int64, from string) (details map[int64]*bplus.Detail, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("from", from)
	for _, id := range ids {
		params.Add("dynamic_ids[]", strconv.FormatInt(id, 10))
	}
	var res struct {
		Code int `json:"code"`
		Data *struct {
			List []*bplus.Detail `json:"list"`
		} `json:"data"`
	}
	details = make(map[int64]*bplus.Detail)
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
