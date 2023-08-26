package v1

import (
	"context"
	"net/url"
	"strconv"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-search/internal/model/search"
	arcmid "go-gateway/app/app-svr/archive/middleware"

	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	pgcstat "git.bilibili.co/bapis/bapis-go/pgc/service/stat/v1"

	"github.com/pkg/errors"
)

func (d *dao) SeasonsStatGRPC(ctx context.Context, seasonIds []int32) (result map[int32]*pgcstat.SeasonStatProto, err error) {
	var (
		req   = &pgcstat.SeasonStatsReq{SeasonIds: seasonIds}
		reply *pgcstat.SeasonStatsReply
	)
	if reply, err = d.pgcstatClient.Seasons(ctx, req); err != nil {
		log.Error("SeasonsStatGRPC seasons error(%v)", err)
		return
	}
	if reply != nil {
		result = reply.Infos
	}
	return
}

func (d *dao) BangumiCard(c context.Context, mid int64, sids []int64) (s map[string]*search.Card, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("season_ids", xstr.JoinInts(sids))
	var res struct {
		Code   int                     `json:"code"`
		Result map[string]*search.Card `json:"result"`
	}
	if err = d.bangumiClient.Get(c, d.bangumiCard, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.bangumiCard+"?"+params.Encode())
		return
	}
	s = res.Result
	return
}

// SearchPGCCards returns the pgc cards dedicated to highlight the episodes
func (d *dao) SearchPGCCards(ctx context.Context, seps []*pgcsearch.SeasonEpReq, query, mobiApp, device_, platform string, mid int64, fnver, fnval, qn, fourk, build int64, isWithPlayURL bool) (result map[int32]*pgcsearch.SearchCardProto, medias map[int32]*pgcsearch.SearchMediaProto, err error) {
	var (
		reply *pgcsearch.SearchCardReply
		req   = &pgcsearch.SearchCardReq{
			Seps:  seps,
			Query: query,
			User: &pgcsearch.UserReq{
				Mid:      mid,
				MobiApp:  mobiApp,
				Device:   device_,
				Platform: platform,
				Ip:       metadata.String(ctx, metadata.RemoteIP),
				Fnver:    uint32(fnver),
				Fnval:    uint32(fnval),
				Qn:       uint32(qn),
				Build:    int32(build),
				Fourk:    int32(fourk),
			},
		}
	)

	dev, ok := device.FromContext(ctx)
	if ok {
		req.User.Buvid = dev.Buvid
	}
	bpa, ok := arcmid.FromContext(ctx)
	if ok {
		req.User.NetType = pgccardgrpc.NetworkType(bpa.NetType)
		req.User.TfType = pgccardgrpc.TFType(bpa.TfType)
	}
	if !isWithPlayURL {
		req.User.LimitPlay = true
	}
	if reply, err = d.pgcsearchClient.Card(ctx, req); err != nil {
		log.Error("SearchCards Query %s, Mid %d, Err %v", query, mid, err)
		return
	}
	result = reply.Cards
	medias = reply.Medias
	return
}

func (d *dao) InlineCards(c context.Context, epIDs []int32, mobiApp, platform, device string, build int, mid int64) (map[int32]*pgcinline.EpisodeCard, error) {
	batchArg, _ := arcmid.FromContext(c)
	arg := &pgcinline.EpReq{
		EpIds: epIDs,
		User: &pgcinline.UserReq{
			Mid:      mid,
			MobiApp:  mobiApp,
			Device:   device,
			Platform: platform,
			Ip:       metadata.String(c, metadata.RemoteIP),
			Fnver:    uint32(batchArg.Fnver),
			Fnval:    uint32(batchArg.Fnval),
			Qn:       uint32(batchArg.Qn),
			Build:    int32(build),
			Fourk:    int32(batchArg.Fourk),
			NetType:  pgccardgrpc.NetworkType(batchArg.NetType),
			TfType:   pgccardgrpc.TFType(batchArg.TfType),
		},
	}
	info, err := d.pgcinlineClient.EpCard(c, arg)
	if err != nil {
		log.Error("pgc inline error(%v) arg(%v)", err, arg)
		return nil, err
	}
	return info.Infos, nil
}

func (d *dao) SeasonCards(ctx context.Context, seasonIds []int32) (res map[int32]*seasongrpc.CardInfoProto, err error) {
	arg := &seasongrpc.SeasonInfoReq{
		SeasonIds: seasonIds,
	}
	info, err := d.seasonRpcClient.Cards(ctx, arg)
	if err != nil {
		log.Error("d.rpcClient.Cards error(%v)", err)
		return nil, err
	}
	res = info.Cards
	return
}

func (d *dao) SearchEpsGrpc(ctx context.Context, req *search.EpisodesNewReq) (reply *pgcsearch.SearchEpReply, err error) {
	var pgcReq = &pgcsearch.SearchEpReq{
		SeasonId: req.SeasonId,
		Ps:       req.Ps,
		Pn:       req.Pn,
	}
	if reply, err = d.pgcsearchClient.Ep(ctx, pgcReq); err != nil {
		log.Error("SearchCards Ssid %d, Pn %d, Ps %d, Err %v", req.SeasonId, req.Pn, req.Ps, err)
	}
	return
}

func (d *dao) SugOGV(c context.Context, ssids []int32) (res map[int32]*pgcsearch.SearchCardProto, err error) {
	var (
		args   = &pgcsearch.SugReq{SeasonIds: ssids}
		resTmp *pgcsearch.SearchCardReply
	)
	if resTmp, err = d.pgcsearchClient.Sug(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetCards()
	return
}
