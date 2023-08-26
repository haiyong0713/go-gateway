package dao

import (
	"context"
	ogvEpSvc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	"sync"

	"go-common/library/log"

	ogvEpisodeSvc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	ogvSeasonSvc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-listener/interface/internal/model"
)

type SeasonDetailsOpt struct {
	Ss []int32
}

func (d *dao) SeasonDetails(ctx context.Context, opt SeasonDetailsOpt) (map[int32]model.SeasonDetail, error) {
	eg := errgroup.WithContext(ctx)
	mu := sync.Mutex{}
	ret := make(map[int32]model.SeasonDetail)
	for _, sid := range opt.Ss {
		sCopy := sid
		eg.Go(func(c context.Context) error {
			req := &ogvSeasonSvc.SeasonIdReq{
				SeasonId: sCopy,
				Type:     0,
			}
			resp, err := d.ogvSeasonGRPC.Profile(c, req)
			if err != nil {
				log.Warnc(ctx, "ogvSeasonGRPC.Profile partial failed req(%+v): %+v", req, err)
				return nil
			}
			if resp.GetProfile() != nil {
				mu.Lock()
				ret[sCopy] = model.SeasonDetail{Ss: resp.GetProfile()}
				mu.Unlock()
			}
			return nil
		})
	}
	_ = eg.Wait()
	return ret, nil
}

func (d *dao) Epids2Aids(ctx context.Context, epids []int32) (map[int32]int64, error) {
	req := &ogvEpisodeSvc.AvInfoReq{
		EpisodeId: epids,
		// 获取所有pgc上架内容
		Type: 0,
	}
	resp, err := d.ogvEpisodeGRPC.AvInfos(ctx, req)
	if err != nil {
		return nil, wrapDaoError(err, "ogvEpisodeGRPC.AvInfos", req)
	}

	ret := make(map[int32]int64)

	for _, info := range resp.GetInfo() {
		if info != nil {
			ret[info.EpisodeId] = info.Aid
		}
	}
	return ret, nil
}

type EpisodeDetailsOpt struct {
	Eps []int32
}

func (d *dao) EpisodeDetails(ctx context.Context, opt EpisodeDetailsOpt) (map[int32]model.EpisodeDetail, error) {
	req := &ogvEpisodeSvc.EpisodeInfoReq{
		EpisodeIds: opt.Eps,
		Type:       0,
	}
	resp, err := d.ogvEpisodeGRPC.List(ctx, req)
	if err != nil {
		return nil, wrapDaoError(err, "ogvEpisodeGRPC.List", req)
	}
	ret := make(map[int32]model.EpisodeDetail)
	for i := range resp.GetInfos() {
		ret[i] = model.EpisodeDetail{Ep: resp.GetInfos()[i]}
	}
	return ret, nil
}

type OGVEpCardOpt struct {
	Eps []int32
}

func (d *dao) OGVEpCards(ctx context.Context, opt OGVEpCardOpt) (map[int32]model.EpCard, error) {
	req := &ogvEpSvc.EpCardsReq{
		EpId: opt.Eps,
	}
	resp, err := d.epCardGRPC.EpCards(ctx, req)
	if err != nil {
		return nil, wrapDaoError(err, "EpGRPC.List", req)
	}
	ret := make(map[int32]model.EpCard)
	for k, v := range resp.GetCards() {
		ret[k] = model.EpCard{Ec: v}
	}
	return ret, nil
}
