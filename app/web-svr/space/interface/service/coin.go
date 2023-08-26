package service

import (
	"context"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/space/interface/model"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"
	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	pgccardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card"
)

const (
	_coinVideoLimit = 20
	_businessCoin   = "archive"
)

var _emptyCoinArcList = make([]*model.CoinArc, 0)

// CoinVideo get coin archives
func (s *Service) CoinVideo(c context.Context, mid, vmid int64) (list []*model.CoinArc, err error) {
	var (
		coinReply *coinmdl.ListReply
		aids      []int64
		arcReply  *arcmdl.ArcsReply
		epCards   map[int64]*pgccardgrpc.EpisodeCard
	)
	if mid != vmid {
		if err = s.privacyCheck(c, vmid, model.PcyCoinVideo); err != nil {
			return
		}
	}
	if coinReply, err = s.coinClient.List(c, &coinmdl.ListReq{Mid: vmid, Business: _businessCoin, Ts: time.Now().Unix()}); err != nil {
		log.Error("s.coinClient.List(%d) error(%v)", vmid, err)
		err = nil
		list = _emptyCoinArcList
		return
	}
	existAids := make(map[int64]int64, len(coinReply.List))
	afVideos := make(map[int64]*coinmdl.ModelList, len(coinReply.List))
	for _, v := range coinReply.List {
		if len(aids) >= _coinVideoLimit {
			break
		}
		if _, ok := existAids[v.Aid]; ok {
			if v.Aid > 0 {
				afVideos[v.Aid].Number += v.Number
			}
			continue
		}
		if v.Aid > 0 {
			afVideos[v.Aid] = v
			aids = append(aids, v.Aid)
			existAids[v.Aid] = v.Aid
		}
	}
	if len(aids) == 0 {
		list = _emptyCoinArcList
		return
	}
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) error {
		if len(aids) == 0 {
			return nil
		}
		rly, err := s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: aids})
		if err != nil {
			log.Error("Fail to request arcgrpc.Arcs, aids=%+v error=%+v", aids, err)
			return err
		}
		arcReply = rly
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
		return _emptyCoinArcList, nil
	}
	for _, aid := range aids {
		arc, ok := arcReply.Arcs[aid]
		if !ok || !arc.IsNormal() {
			continue
		}
		if arc.Access >= model.ArcAccessVariable {
			arc.Stat.View = -1
		}
		afVideo, ok := afVideos[aid]
		if !ok {
			continue
		}
		coinArc := &model.CoinArc{
			Arc:          arc,
			Bvid:         s.avToBv(arc.Aid),
			Coins:        afVideo.Number,
			Time:         afVideo.Ts,
			InterVideo:   arc.AttrVal(arcapi.AttrBitSteinsGate) == arcapi.AttrYes,
			ResourceType: model.ResourceTypeUGC,
		}
		if ep, ok := epCards[aid]; ok && ep != nil {
			coinArc.FormatAsEpCard(ep)
		}
		model.ClearAttrAndAccess(arc)
		list = append(list, coinArc)
	}
	if len(list) == 0 {
		list = _emptyCoinArcList
	}
	return
}
