package pgc

import (
	"context"
	"net/url"
	"strconv"
	"sync"

	"go-common/component/metadata/network"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	mdlpgc "go-gateway/app/app-svr/app-dynamic/interface/model/pgc"
	arcmid "go-gateway/app/app-svr/archive/middleware"

	pgcCardGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcInlineGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcShareGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/share"
	pgcEpisodeGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	pgcSeasonGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	"github.com/pkg/errors"
)

const (
	_batchInfoURL  = "/pugv/internal/dynamic/batch"
	_seasonInfoURL = "/pugv/internal/dynamic/season"
	_fromDynamic   = 1
)

func (d *Dao) EpList(c context.Context, epids []int32, general *mdlv2.GeneralParam, playurlParam *api.PlayurlParam) (map[int32]*pgcInlineGrpc.EpisodeCard, error) {
	arg := &pgcInlineGrpc.EpReq{
		User: &pgcInlineGrpc.UserReq{
			MobiApp:  general.GetMobiApp(),
			Device:   general.GetDevice(),
			Platform: general.GetPlatform(),
			Ip:       general.IP,
			Build:    int32(general.GetBuild()),
			NetType:  pgcCardGrpc.NetworkType(general.Device.NetworkType),
			TfType:   pgcCardGrpc.TFType(general.Device.TfType),
			Buvid:    general.GetBuvid(),
		},
		EpIds: epids,
		SceneControl: &pgcInlineGrpc.SceneControl{
			WasDynamic: true,
		},
		CustomizeReq: &pgcInlineGrpc.CustomizeReq{
			NeedShareCount: true,
		},
	}
	if playurlParam != nil {
		arg.User.Fnver = uint32(playurlParam.Fnver)
		arg.User.Fnval = uint32(playurlParam.Fnval)
		arg.User.Qn = uint32(playurlParam.Qn)
		arg.User.Fourk = playurlParam.Fourk
	}
	reply, err := d.pgcInlineGRPC.EpCard(c, arg)
	if err != nil {
		return nil, err
	}
	return reply.Infos, nil
}

func (d *Dao) MyFollows(c context.Context, mid int64) (*pgcAppGrpc.FollowReply, error) {
	in := &pgcAppGrpc.FollowReq{
		Mid:  mid,
		From: _fromDynamic,
	}
	rsp, err := d.pgcAppGRPC.MyFollows(c, in)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return rsp, nil
}

func (d *Dao) PGCBatch(c context.Context, batch []int64, general *mdlv2.GeneralParam) (map[int64]*mdlpgc.PGCBatch, error) {
	params := url.Values{}
	if batchArg, ok := arcmid.FromContext(c); ok && batchArg != nil {
		params.Set("mobi_app", batchArg.GetMobiApp())
		params.Set("device", batchArg.GetDevice())
		params.Set("build", strconv.FormatInt(batchArg.GetBuild(), 10))
		params.Set("user_ip", batchArg.Ip)
		params.Set("fnver", strconv.FormatInt(batchArg.GetFnver(), 10))
		params.Set("fnval", strconv.FormatInt(batchArg.GetFnval(), 10))
		params.Set("mid", strconv.FormatInt(batchArg.GetMid(), 10))
		params.Set("qn", strconv.FormatInt(batchArg.GetQn(), 10))
		params.Set("fourk", strconv.FormatInt(batchArg.GetFourk(), 10))
		params.Set("buvid", batchArg.GetBuvid())
	}
	params.Set("batch_ids", xstr.JoinInts(batch))
	params.Set("platform", general.GetPlatform())
	if nw, ok := network.FromContext(c); ok {
		params.Set("cdn_ip", nw.WebcdnIP)
	}
	pgcBatch := d.c.Hosts.ApiCo + _batchInfoURL
	var ret struct {
		Code int                        `json:"code"`
		Msg  string                     `json:"message"`
		Data map[int64]*mdlpgc.PGCBatch `json:"data"`
	}
	if err := d.client.Get(c, pgcBatch, "", params, &ret); err != nil {
		xmetric.DyanmicItemAPI.Inc(pgcBatch, "request_error")
		return nil, errors.WithStack(err)
	}
	if ret.Code != 0 {
		xmetric.DyanmicItemAPI.Inc(pgcBatch, "reply_code_error")
		return nil, errors.Wrapf(ecode.Int(ret.Code), "Failed to HTTP GET: %v. params: %v. code: %v. msg: %v", pgcBatch, params.Encode(), ret.Code, ret.Msg)
	}
	return ret.Data, nil
}

func (d *Dao) PGCSeason(c context.Context, season []int64, general *mdlv2.GeneralParam) (map[int64]*mdlpgc.PGCSeason, error) {
	params := url.Values{}
	if batchArg, ok := arcmid.FromContext(c); ok && batchArg != nil {
		params.Set("mobi_app", batchArg.GetMobiApp())
		params.Set("device", batchArg.GetDevice())
		params.Set("build", strconv.FormatInt(batchArg.GetBuild(), 10))
		params.Set("user_ip", batchArg.Ip)
		params.Set("fnver", strconv.FormatInt(batchArg.GetFnver(), 10))
		params.Set("fnval", strconv.FormatInt(batchArg.GetFnval(), 10))
		params.Set("mid", strconv.FormatInt(batchArg.GetMid(), 10))
		params.Set("qn", strconv.FormatInt(batchArg.GetQn(), 10))
		params.Set("fourk", strconv.FormatInt(batchArg.GetFourk(), 10))
		params.Set("buvid", batchArg.GetBuvid())
	}
	if nw, ok := network.FromContext(c); ok {
		params.Set("cdn_ip", nw.WebcdnIP)
	}
	params.Set("season_ids", xstr.JoinInts(season))
	params.Set("platform", general.GetPlatform())
	pgcSeason := d.c.Hosts.ApiCo + _seasonInfoURL
	var ret struct {
		Code int                         `json:"code"`
		Msg  string                      `json:"message"`
		Data map[int64]*mdlpgc.PGCSeason `json:"data"`
	}
	if err := d.client.Get(c, pgcSeason, "", params, &ret); err != nil {
		xmetric.DyanmicItemAPI.Inc(pgcSeason, "request_error")
		return nil, errors.WithStack(err)
	}
	if ret.Code != 0 {
		xmetric.DyanmicItemAPI.Inc(pgcSeason, "reply_code_error")
		return nil, errors.Wrapf(ecode.Int(ret.Code), "Failed to HTTP GET: %v. params: %v. code: %v. msg: %v", pgcSeason, params.Encode(), ret.Code, ret.Msg)
	}
	return ret.Data, nil
}

func (d *Dao) ShareMessage(c context.Context, epids []int32) ([]*pgcShareGrpc.ShareMessageResBody, error) {
	reply, err := d.pgcShareGRPC.QueryShareMessageInfo(c, &pgcShareGrpc.ShareMessageReq{EpId: epids})
	if err != nil {
		return nil, err
	}
	return reply.GetBodys(), nil
}

func (d *Dao) Seasons(c context.Context, ssids []int32) (map[int32]*pgcSeasonGrpc.CardInfoProto, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int32]*pgcSeasonGrpc.CardInfoProto)
	for i := 0; i < len(ssids); i += max50 {
		var partSSids []int32
		if i+max50 > len(ssids) {
			partSSids = ssids[i:]
		} else {
			partSSids = ssids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			ss, err := d.SeasonsSlice(ctx, partSSids)
			if err != nil {
				return err
			}
			mu.Lock()
			for seasonid, s := range ss {
				res[seasonid] = s
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("Seasons ssids(%+v) eg.wait(%+v)", ssids, err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) SeasonsSlice(c context.Context, ssids []int32) (res map[int32]*pgcSeasonGrpc.CardInfoProto, err error) {
	args := &pgcSeasonGrpc.SeasonInfoReq{SeasonIds: ssids, Type: 3} // type: 0-pgc上架的 1-全部非删除的 2-ott上架的 3-所有上架的
	var resTmp *pgcSeasonGrpc.CardsInfoReply
	if resTmp, err = d.pgcSeasonGRPC.Cards(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetCards()
	return
}

func (d *Dao) Episodes(c context.Context, epids []int32) (map[int32]*pgcEpisodeGrpc.EpisodeCardsProto, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int32]*pgcEpisodeGrpc.EpisodeCardsProto)
	for i := 0; i < len(epids); i += max50 {
		var partEpids []int32
		if i+max50 > len(epids) {
			partEpids = epids[i:]
		} else {
			partEpids = epids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			eps, err := d.EpisodeSlice(ctx, partEpids)
			if err != nil {
				return err
			}
			mu.Lock()
			for epid, ep := range eps {
				res[epid] = ep
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("Episodes epids(%+v) eg.wait(%+v)", epids, err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) EpisodeSlice(c context.Context, epids []int32) (res map[int32]*pgcEpisodeGrpc.EpisodeCardsProto, err error) {
	args := &pgcEpisodeGrpc.EpReq{Epids: epids, Type: 3} // type: 0-pgc上架的 1-全部非删除的 2-ott上架的 3-所有上架的
	var resTmp *pgcEpisodeGrpc.EpisodeCardsReply
	if resTmp, err = d.pgcEpisodeGRPC.Cards(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetCards()
	return
}
