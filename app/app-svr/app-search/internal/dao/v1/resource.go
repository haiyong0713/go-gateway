package v1

import (
	"context"
	"go-gateway/app/app-svr/app-search/internal/model/search"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	searchadm "go-gateway/app/app-svr/app-feed/admin/model/search"
	resmdl "go-gateway/app/app-svr/resource/service/model"

	"github.com/pkg/errors"
)

func (d *dao) SpecialCards(c context.Context) (map[int64]*searchadm.SpreadConfig, error) {
	ps := 50
	req := &searchadm.RecomParam{
		Pn: 1,
		Ps: ps,
	}
	var scList []*searchadm.SpreadConfig

	type recommendResponse struct {
		Code    int64              `json:"code"`
		Message string             `json:"message"`
		Data    searchadm.RecomRes `json:"data"`
	}
	params := cvtRecomParam(req)
	response := &recommendResponse{}
	if err := d.feedAdminClient.Get(c, d.httpClientCfg.Host.FeedAdmin+"/x/admin/feed/open/search/recommend", "", params, response); err != nil {
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
		if err := d.feedAdminClient.Get(c, d.httpClientCfg.Host.FeedAdmin+"/x/admin/feed/open/search/recommend", "", params, response); err != nil {
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

func (d *dao) ALLSearchSystemNotice(ctx context.Context) (map[int64]*search.SystemNotice, error) {
	reply := &struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
		Data    struct {
			List []*search.SystemNotice `json:"list"`
		} `json:"data"`
	}{}
	if err := d.feedAdminClient.Get(ctx, d.httpClientCfg.Host.Manager+"/x/admin/manager/search/internal/system/notice", "", nil, reply); err != nil {
		return nil, err
	}
	if reply.Code != 0 {
		return nil, errors.Errorf("invalid code: %d: %+v", reply.Code, reply)
	}
	out := make(map[int64]*search.SystemNotice, len(reply.Data.List))
	for _, v := range reply.Data.List {
		out[v.Mid] = v
	}
	return out, nil
}

func (d *dao) Banner(c context.Context, mobiApp, device, network, channel, buvid, adExtra, resIDStr string, build int, plat int8, mid int64) (res map[int][]*resmdl.Banner, err error) {
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
