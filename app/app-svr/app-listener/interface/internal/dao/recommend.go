package dao

import (
	"context"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	"go-common/library/sync/errgroup.v2"

	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
)

type RecommendArchivesOpt struct {
	Dev         *device.Device
	Net         *network.Network
	Mid         int64
	ExtraAids   []int64
	OriginAid   int64
	RcmdFrom    int64
	Page        int64
	SessionId   string
	FromTrackId string
}

const (
	_aiGotoArchive = "av"
)

type RecommendArchivesRes struct {
	Arcs    []model.ArchiveDetail
	TrackID string
}

func (d *dao) RecommendArchives(ctx context.Context, opt RecommendArchivesOpt) (RecommendArchivesRes, error) {
	var bringIn int64
	if len(opt.ExtraAids) > 0 {
		bringIn = opt.ExtraAids[0]
	}
	limit, _ := restriction.FromContext(ctx)
	req := &listenerSvc.RecommendArchivesReq{
		Mid:         opt.Mid,
		Ip:          opt.Net.RemoteIP,
		Lat:         0,
		Lng:         0,
		City:        0,
		Province:    0,
		Aid:         bringIn,
		Network:     toListenSvrNetMeta(opt.Net),
		Device:      toListenSvrDevMeta(opt.Dev),
		DisableRcmd: limit.DisableRcmd,
		FromType:    listenerSvc.RecommendFromType(opt.RcmdFrom),
		Page:        opt.Page,
		SessionId:   opt.SessionId,
		FromTrackid: opt.FromTrackId,
	}
	if req.Aid == 0 && opt.OriginAid != 0 {
		req.Aid = opt.OriginAid
	}
	resp, err := d.listenerGRPC.GetRecommendArchives(ctx, req)
	if err != nil {
		return RecommendArchivesRes{}, wrapDaoError(err, "listenerGRPC.GetRecommendArchives", req)
	}
	if (resp == nil || len(resp.Archives) == 0) && len(opt.ExtraAids) == 0 {
		return RecommendArchivesRes{TrackID: resp.GetTrackId()}, nil
	}
	aidm := make(map[int64][]int64)
	for _, arc := range resp.GetArchives() {
		if arc.Goto == _aiGotoArchive && arc.Aid > 0 {
			aidm[arc.Aid] = nil
		}
	}
	for _, aid := range opt.ExtraAids {
		aidm[aid] = nil
	}
	if len(aidm) <= 0 {
		return RecommendArchivesRes{TrackID: resp.GetTrackId()}, nil
	}
	arcDetails, err := d.ArchiveDetails(ctx, ArcDetailsOpt{
		Aids:     aidm,
		Mid:      opt.Mid,
		RemoteIP: opt.Net.RemoteIP,
		Dev:      opt.Dev,
	})
	if err != nil {
		return RecommendArchivesRes{}, err
	}
	list := make([]model.ArchiveDetail, 0, len(arcDetails))
	deDup := make(map[int64]struct{})
	for _, aid := range opt.ExtraAids {
		arcInfo, ok := arcDetails[aid]
		if ok {
			list = append(list, arcInfo)
		}
		deDup[aid] = struct{}{}
	}
	for _, arc := range resp.GetArchives() {
		if _, ok := deDup[arc.Aid]; ok {
			continue
		}
		arcInfo, ok := arcDetails[arc.Aid]
		if ok {
			list = append(list, arcInfo)
		}
		deDup[arc.Aid] = struct{}{}
	}
	return RecommendArchivesRes{Arcs: list, TrackID: resp.GetTrackId()}, nil
}

type RcmdTopCardsOpt struct {
	Mid int64
	Dev *device.Device
}

func (d *dao) RcmdTopCards(ctx context.Context, opt RcmdTopCardsOpt) ([]model.RcmdTopCard, error) {
	resp, err := d.listenerGRPC.GetHeadCards(ctx, &listenerSvc.GetHeadCardsReq{
		Mid: opt.Mid, Buvid: opt.Dev.Buvid,
	})
	if err != nil {
		return nil, err
	}
	ret := make([]model.RcmdTopCard, 0, len(resp.Cards))
	idx := make(map[int]CardDetailsOpt)
	for i, c := range resp.GetCards() {
		ret = append(ret, model.RcmdTopCard{
			Card: c,
		})
		if ch := c.GetChannel(); ch != nil && ch.GetChannel() == model.TpcdChannelPickToday {
			idx[i] = CardDetailsOpt{
				CardId: ch.GetBizId(), PickId: ch.GetBizType(),
			}
		}
	}
	if len(idx) > 0 {
		eg := errgroup.WithCancel(ctx)
		for i, opt := range idx {
			optCopy := opt
			iCopy := i
			eg.Go(func(ctx context.Context) error {
				res, err := d.CardDetail(ctx, optCopy)
				if err != nil {
					return err
				}
				tmp := ret[iCopy]
				tmp.PickInfo = &res
				ret[iCopy] = tmp
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			return nil, err
		}
	}

	return ret, nil
}
