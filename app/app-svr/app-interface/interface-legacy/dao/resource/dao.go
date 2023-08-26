package resource

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	searchadm "go-gateway/app/app-svr/app-feed/admin/model/search"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	managermodel "go-gateway/app/app-svr/app-interface/interface-legacy/model/manager"
	resmdl "go-gateway/app/app-svr/resource/service/model"
	resrpc "go-gateway/app/app-svr/resource/service/rpc/client"

	newmontapi "git.bilibili.co/bapis/bapis-go/newmont/service/v1"
	resApi "git.bilibili.co/bapis/bapis-go/resource/service/v1"
	resgrpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"

	"github.com/pkg/errors"
)

const (
	_isup  = 1
	_notup = 0
)

type Dao struct {
	c *conf.Config
	// rpc
	resRPC        *resrpc.Service
	resClient     resApi.ResourceClient
	resGRPC       resgrpc.ResourceClient
	newmontClient newmontapi.NewmontClient
	httpClient    *bm.Client
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// rpc
		resRPC:     resrpc.New(c.ResourceRPC),
		httpClient: bm.NewClient(c.HTTPFeedAdmin, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
	}
	var err error
	if d.resClient, err = resApi.NewClient(c.ResClient); err != nil {
		panic(fmt.Sprintf("resApi.NewClient err(%+v)", err))
	}
	if d.resGRPC, err = resgrpc.NewClient(c.ResourceGRPC); err != nil {
		panic(err)
	}
	if d.newmontClient, err = newmontapi.NewClientNewmont(c.NewmontClient); err != nil {
		panic(err)
	}
	return
}

// Banner get search banner
func (d *Dao) Banner(c context.Context, mobiApp, device, network, channel, buvid, adExtra, resIDStr string, build int, plat int8, mid int64) (res map[int][]*resmdl.Banner, err error) {
	var (
		bs *resmdl.Banners
		ip = metadata.String(c, metadata.RemoteIP)
	)
	arg := &resmdl.ArgBanner{
		MobiApp: mobiApp,
		Device:  device,
		Network: network,
		Channel: channel,
		IP:      ip,
		Buvid:   buvid,
		AdExtra: adExtra,
		ResIDs:  resIDStr,
		Build:   build,
		Plat:    plat,
		MID:     mid,
		IsAd:    true,
	}
	if bs, err = d.resRPC.Banners(c, arg); err != nil || bs == nil {
		log.Error("d.resRPC.Banners(%v) error(%v) or bs is nil", arg, err)
		return
	}
	if len(bs.Banner) > 0 {
		res = bs.Banner
	}
	return
}

// EntrancesIsHidden is
func (d *Dao) EntrancesIsHidden(ctx context.Context, oids []int64, build int, plat int8, channel string) (map[int64]bool, error) {
	req := &resApi.EntrancesIsHiddenRequest{
		Oids:    oids,
		Otype:   2, // 侧边栏入口对应cid,
		Build:   int64(build),
		Plat:    int32(plat),
		Channel: channel,
	}
	res, err := d.resClient.EntrancesIsHidden(ctx, req)
	if err != nil {
		log.Error("d.resClient.EntrancesIsHidden err(%+v) req(%+v)", err, req)
		return nil, err
	}
	if res == nil {
		return nil, err
	}
	return res.Infos, nil
}

// MngIcon is
func (d *Dao) MngIcon(ctx context.Context, oids []int64, mid int64, plat int8) (map[int64]*resApi.MngIcon, error) {
	req := &resApi.MngIconRequest{
		Oids: oids,
		Plat: int32(plat),
		Mid:  mid,
	}
	res, err := d.resClient.MngIcon(ctx, req)
	if err != nil {
		log.Error("d.resClient.MngIcon err(%+v) req(%+v)", err, req)
		return nil, err
	}
	if res == nil {
		return nil, err
	}
	return res.Info, nil
}

// MineSections is
func (d *Dao) MineSections(ctx context.Context, mid int64, plat, build int32, channel, lang string, isUp int, firstLiveTime, fansPeak int64, buvid string) ([]*resApi.Section, error) {
	req := &resApi.MineSectionsRequest{
		Plat:       plat,
		Build:      build,
		Mid:        mid,
		Lang:       lang,
		Channel:    channel,
		Ip:         metadata.String(ctx, metadata.RemoteIP),
		IsUploader: isUp == 1,
		IsLiveHost: firstLiveTime > 0,
		FansCount:  fansPeak,
		Buvid:      buvid,
	}
	res, err := d.resClient.MineSections(ctx, req)
	if err != nil {
		log.Error("d.resClient.MineSections err(%+v) req(%+v)", err, req)
		return nil, err
	}
	return res.GetSections(), nil
}

func (d *Dao) IsUp(ctx context.Context, mid int64) (int, error) {
	req := &resApi.IsUploaderReq{Mid: mid}
	res, err := d.resClient.IsUploader(ctx, req)
	if err != nil {
		log.Error("d.resClient.IsUploader err(%+v) req(%+v)", err, req)
		return _notup, err
	}
	if res == nil {
		log.Error("d.resClient.IsUploader res is nil req(%+v)", req)
		return _notup, ecode.NothingFound
	}
	if res.IsUploader {
		return _isup, nil
	}
	return _notup, nil
}

func cvtRecomParam(in *searchadm.RecomParam) url.Values {
	out := url.Values{}
	out.Set("ts", strconv.FormatInt(in.Ts, 10))
	out.Set("start_ts", strconv.FormatInt(in.StartTs, 10))
	out.Set("end_ts", strconv.FormatInt(in.EndTs, 10))
	out.Set("ps", strconv.FormatInt(int64(in.Ps), 10))
	out.Set("pn", strconv.FormatInt(int64(in.Pn), 10))
	out.Set("plat", strconv.FormatInt(int64(in.Plat), 10))
	out.Set("pos", strconv.FormatInt(int64(in.Pos), 10))
	out.Set("card_type", xstr.JoinInts([]int64{3, 6}))
	return out
}

func (d *Dao) SpecialCards(c context.Context) (map[int64]*searchadm.SpreadConfig, error) {
	ps := 50
	req := &searchadm.RecomParam{
		Pn: 1,
		Ps: ps,
	}
	scList := []*searchadm.SpreadConfig{}

	type recommendResponse struct {
		Code    int64              `json:"code"`
		Message string             `json:"message"`
		Data    searchadm.RecomRes `json:"data"`
	}
	params := cvtRecomParam(req)
	response := &recommendResponse{}
	if err := d.httpClient.Get(c, d.c.Host.FeedAdmin+"/x/admin/feed/open/search/recommend", "", params, response); err != nil {
		return nil, err
	}
	if response.Code != 0 {
		return nil, ecode.Error(ecode.Code(response.Code), response.Message)
	}
	reply := response.Data
	scList = append(scList, reply.Item...)

	times := reply.Page.Total / ps
	if reply.Page.Total%ps > 0 {
		times += 1
	}
	for i := 0; i < times; i++ {
		req.Pn = req.Pn + 1
		params := cvtRecomParam(req)
		response := &recommendResponse{}
		if err := d.httpClient.Get(c, d.c.Host.FeedAdmin+"/x/admin/feed/open/search/recommend", "", params, response); err != nil {
			log.Warn("Failed to request search special card: %+v: %+v", req, err)
			continue
		}
		if response.Code != 0 {
			log.Warn("Failed to request search special card: %+v: with ecode: %+v", req, ecode.Error(ecode.Code(response.Code), response.Message))
			continue
		}
		reply := response.Data
		scList = append(scList, reply.Item...)
	}
	out := make(map[int64]*searchadm.SpreadConfig, len(scList))
	for _, v := range scList {
		if v.Special == nil {
			log.Warn("Invalid spread config: %+v", v)
			continue
		}
		out[v.Special.ID] = v
	}
	return out, nil
}

func (d *Dao) CheckCommonBWList(ctx context.Context, vmid int64) (bool, error) {
	checkReq := &resgrpc.CheckCommonBWListReq{
		Oid:    strconv.FormatInt(vmid, 10),
		Token:  d.c.LegoToken.SpaceIPLimit,
		UserIp: metadata.String(ctx, metadata.RemoteIP),
	}
	checkReply, err := d.resGRPC.CheckCommonBWList(ctx, checkReq)
	if err != nil {
		return false, err
	}
	return checkReply.GetIsInList(), nil
}

func (d *Dao) ALLSearchSystemNotice(ctx context.Context) (map[int64]*managermodel.SystemNotice, error) {
	reply := &struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []*managermodel.SystemNotice `json:"list"`
		} `json:"data"`
	}{}
	if err := d.httpClient.Get(ctx, d.c.Host.Manager+"/x/admin/manager/search/internal/system/notice", "", nil, reply); err != nil {
		return nil, err
	}
	if reply.Code != 0 {
		return nil, errors.Errorf("invalid code: %d: %+v", reply.Code, reply)
	}
	out := make(map[int64]*managermodel.SystemNotice, len(reply.Data.List))
	for _, v := range reply.Data.List {
		out[v.Mid] = v
	}
	return out, nil
}

func (d *Dao) MineSectionsNewmont(ctx context.Context, mid int64, plat, build int32, channel, lang string, isUp int, firstLiveTime, fansPeak int64, buvid string) ([]*newmontapi.Section, error) {
	req := &newmontapi.MineSectionsRequest{
		Plat:       plat,
		Build:      build,
		Mid:        mid,
		Lang:       lang,
		Channel:    channel,
		Ip:         metadata.String(ctx, metadata.RemoteIP),
		IsUploader: isUp == 1,
		IsLiveHost: firstLiveTime > 0,
		FansCount:  fansPeak,
		Buvid:      buvid,
	}
	res, err := d.newmontClient.MineSections(ctx, req)
	if err != nil {
		log.Error("d.resClient.MineSections err(%+v) req(%+v)", err, req)
		return nil, err
	}
	return res.GetSections(), nil
}
