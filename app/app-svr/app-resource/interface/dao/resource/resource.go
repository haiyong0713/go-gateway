package resource

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	resource "go-gateway/app/app-svr/resource/service/model"
	resrpc "go-gateway/app/app-svr/resource/service/rpc/client"

	resApi "git.bilibili.co/bapis/bapis-go/resource/service"

	"github.com/pkg/errors"
)

type Dao struct {
	c *conf.Config
	// rpc
	resRpc     *resrpc.Service
	resClient  resApi.ResourceClient
	httpClient *bm.Client
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// rpc
		resRpc:     resrpc.New(c.ResourceRPC),
		httpClient: bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.resClient, err = resApi.NewClient(c.ResourceClient); err != nil {
		panic(err)
	}
	return
}

// SkinConf .
func (d *Dao) SkinConf(c context.Context) (rly []*resApi.SkinInfo, err error) {
	var (
		resSkin *resApi.SkinConfReply
	)
	if resSkin, err = d.resClient.SkinConf(c, &resApi.NoArgRequest{}); err != nil {
		log.Error("d.resClient.SkinConf error(%v)", err)
		return
	}
	if resSkin == nil || len(resSkin.List) == 0 {
		return
	}
	for _, val := range resSkin.List {
		if val.Info == nil || len(val.Limit) == 0 {
			continue
		}
		rly = append(rly, val)
	}
	return
}

// ResSideBar resource ressidebar
func (d *Dao) ResSideBar(ctx context.Context) (res *resource.SideBars, err error) {
	if res, err = d.resRpc.SideBars(ctx); err != nil {
		log.Error("resource d.resRpc.SideBars error(%v)", err)
		return
	}
	return
}

// AbTest resource abtest
func (d *Dao) AbTest(ctx context.Context, groups string) (res map[string]*resource.AbTest, err error) {
	arg := &resource.ArgAbTest{
		Groups: groups,
	}
	if res, err = d.resRpc.AbTest(ctx, arg); err != nil {
		log.Error("resource d.resRpc.AbTest error(%v)", err)
		return
	}
	return
}

// EntrancesIsHidden is
func (d *Dao) EntrancesIsHidden(ctx context.Context, oids []int64, build int, plat int8, channel string) (*resApi.EntrancesIsHiddenReply, error) {
	req := &resApi.EntrancesIsHiddenRequest{
		Oids:    oids,
		Otype:   0, //首页入口对应sidebar.id,
		Build:   int64(build),
		Plat:    int32(plat),
		Channel: channel,
	}
	res, err := d.resClient.EntrancesIsHidden(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// MenuExtVer .
func (d *Dao) MenuExtVer(c context.Context, id int64, buvid, ver string) (*resApi.MenuExtVerReply, error) {
	rly, err := d.resClient.MenuExtVer(c, &resApi.MenuExtVerReq{Id: id, Buvid: buvid, Ver: ver})
	if err != nil {
		log.Error("d.resClient.MenuExtVer(%d,%s,%s) error(%v)", id, buvid, ver, err)
		return nil, err
	}
	return rly, nil
}

// AddMenuExtVer .
func (d *Dao) AddMenuExtVer(c context.Context, id int64, buvid, ver string) error {
	_, err := d.resClient.AddMenuExtVer(c, &resApi.AddMenuExtVerReq{Id: id, Buvid: buvid, Ver: ver})
	if err != nil {
		log.Error("d.resClient.AddMenuExtVer(%d,%s,%s) error(%v)", id, buvid, ver, err)
		return err
	}
	return nil
}

// TopActivity is
func (d *Dao) TopActivity(ctx context.Context, build int64, plat int8) (*resApi.GetAppEntryStateV2Rep, error) {
	req := &resApi.GetAppEntryStateReq{
		Build: int32(build),
		Plat:  int32(plat),
	}
	reply, err := d.resClient.GetAppEntryStateV2(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) PopUp(ctx context.Context, mid int64, buvid string, plat, build int32) (*resApi.PopUpsReply, error) {
	reply, err := d.resClient.PopUps(ctx, &resApi.PopUpsReq{
		Mid:   mid,
		Buvid: buvid,
		Plat:  plat,
		Build: build,
	})
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, ecode.NothingFound
	}
	return reply, nil
}

// GetTabExt .
func (d *Dao) GetTabExt(ctx context.Context, plat, build int64, buvid string, tabs []*resApi.Tab) ([]*resApi.TabExt, error) {
	req := &resApi.GetTabExtReq{
		Build: build,
		Plat:  plat,
		Buvid: buvid,
		Tabs:  tabs,
	}
	reply, err := d.resClient.GetTabExt(ctx, req)
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, ecode.NothingFound
	}
	return reply.TabExts, nil
}

func (d *Dao) HomeSections(ctx context.Context, mid int64, plat, build int32, lang, channel, buvid string) (*resApi.HomeSectionsReply, error) {
	reply, err := d.resClient.HomeSections(ctx, &resApi.HomeSectionsRequest{Mid: mid, Plat: plat, Build: build, Lang: lang, Channel: channel, Ip: metadata.String(ctx, metadata.RemoteIP), Buvid: buvid})
	if err != nil {
		return nil, errors.Wrapf(err, "d.resClient.HomeSections error mid(%d)", mid)
	}
	return reply, nil
}

func unifyCheckURL(checkURL string, extParams url.Values) (string, url.Values) {
	u, err := url.Parse(checkURL)
	if err != nil {
		return checkURL, extParams
	}
	merged := url.Values{}
	for k, vs := range extParams {
		merged[k] = vs
	}
	for k, vs := range u.Query() {
		merged[k] = vs
	}
	u.RawQuery = ""
	return u.String(), merged
}

// UserCheck 各种入口白名单
// https://www.tapd.cn/20055921/prong/stories/view/1120055921001066980  动态互推TAPD在此！！
func (d *Dao) UserCheck(c context.Context, mid int64, checkURL string) (ok bool, err error) {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Status int `json:"status"`
		} `json:"data"`
	}
	pureURL, pureParams := unifyCheckURL(checkURL, params)
	if err = d.httpClient.Get(c, pureURL, "", pureParams, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), pureURL+"?"+pureParams.Encode())
		return
	}
	if res.Data.Status == 1 {
		ok = true
	}
	return
}
