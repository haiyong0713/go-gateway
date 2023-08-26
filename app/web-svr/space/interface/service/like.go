package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/space/interface/model"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
)

func (s *Service) LikeVideo(c context.Context, req *model.LikeVideoReq, mid int64) (*model.LikeVideoRly, error) {
	if mid != req.Vmid {
		if err := s.privacyCheck(c, req.Vmid, model.PcyLikeVideo); err != nil {
			return nil, err
		}
	}
	likes, err := s.thumbupClient.UserLikes(c, &thumbupgrpc.UserLikesReq{
		Business: model.BusinessLike,
		Mid:      req.Vmid,
		Pn:       1,
		Ps:       20,
		IP:       metadata.String(c, metadata.RemoteIP),
	})
	if err != nil {
		log.Errorc(c, "Fail to request thumbupgrpc.UserLikes, mid=%d error=%+v", req.Vmid, err)
		return nil, err
	}
	var aids []int64
	for _, v := range likes.Items {
		aids = append(aids, v.MessageID)
	}
	var (
		arcs    map[int64]*arcgrpc.Arc
		epCards map[int64]*pgccardgrpc.EpisodeCard
	)
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) error {
		if len(aids) == 0 {
			return nil
		}
		rly, err := s.arcClient.Arcs(ctx, &arcgrpc.ArcsRequest{Aids: aids})
		if err != nil {
			log.Error("Fail to request arcgrpc.Arcs, aids=%+v error=%+v", aids, err)
			return err
		}
		arcs = rly.GetArcs()
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if len(aids) == 0 {
			return nil
		}
		rly, err := s.pgcCardClient.EpCards(ctx, &pgccardgrpc.EpCardsReq{Aid: aids})
		if err != nil {
			log.Error("Fail to request pgccardgrpc.EpCards, aids=%+v error=%+v", aids, err)
			return nil
		}
		epCards = rly.GetAidCards()
		return nil
	})
	if err := eg.Wait(); err != nil {
		return &model.LikeVideoRly{}, nil
	}
	likeArcs := make([]*model.LikeVideoItem, 0, len(aids))
	for _, aid := range aids {
		arc, ok := arcs[aid]
		if !ok || !arc.IsNormal() {
			continue
		}
		if arc.Access >= model.ArcAccessVariable {
			arc.Stat.View = -1
		}
		likeArc := &model.LikeVideoItem{
			Arc:          arc,
			Bvid:         s.avToBv(arc.Aid),
			InterVideo:   arc.AttrVal(arcapi.AttrBitSteinsGate) == arcapi.AttrYes,
			ResourceType: model.ResourceTypeUGC,
		}
		if ep, ok := epCards[aid]; ok && ep != nil {
			likeArc.FormatAsEpCard(ep)
		}
		model.ClearAttrAndAccess(arc)
		likeArcs = append(likeArcs, likeArc)
	}
	return &model.LikeVideoRly{List: likeArcs}, nil
}
