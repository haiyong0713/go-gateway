package service

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/topic/card/model"

	pgcCardGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcInlineGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcEpisodeGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	pgcSeasonGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

func (s *Service) getEpList(c context.Context, epids []int32, general *model.GeneralParam) (map[int32]*pgcInlineGrpc.EpisodeCard, error) {
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
	reply, err := s.pgcInlineGRPC.EpCard(c, arg)
	if err != nil {
		return nil, err
	}
	return reply.Infos, nil
}

func (s *Service) seasons(c context.Context, ssids []int32) (map[int32]*pgcSeasonGrpc.CardInfoProto, error) {
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
			ss, err := s.seasonsSlice(ctx, partSSids)
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
		log.Error("seasons ssids(%+v) eg.wait(%+v)", ssids, err)
		return nil, err
	}
	return res, nil
}

func (s *Service) seasonsSlice(c context.Context, ssids []int32) (map[int32]*pgcSeasonGrpc.CardInfoProto, error) {
	args := &pgcSeasonGrpc.SeasonInfoReq{SeasonIds: ssids, Type: 3} // type: 0-pgc上架的 1-全部非删除的 2-ott上架的 3-所有上架的
	resTmp, err := s.pgcSeasonGRPC.Cards(c, args)
	if err != nil {
		log.Error("s.pgcSeasonGRPC.Cards error=%+v", err)
		return nil, err
	}
	return resTmp.GetCards(), nil
}

func (s *Service) episodes(c context.Context, epids []int32) (map[int32]*pgcEpisodeGrpc.EpisodeCardsProto, error) {
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
			eps, err := s.episodeSlice(ctx, partEpids)
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
		log.Error("episodes epids(%+v) eg.wait(%+v)", epids, err)
		return nil, err
	}
	return res, nil
}

func (s *Service) episodeSlice(c context.Context, epids []int32) (map[int32]*pgcEpisodeGrpc.EpisodeCardsProto, error) {
	args := &pgcEpisodeGrpc.EpReq{Epids: epids, Type: 3} // type: 0-pgc上架的 1-全部非删除的 2-ott上架的 3-所有上架的
	resTmp, err := s.pgcEpisodeGRPC.Cards(c, args)
	if err != nil {
		log.Error("s.pgcEpisodeGRPC.Cards error=%+v", err)
		return nil, err
	}
	return resTmp.GetCards(), nil
}
