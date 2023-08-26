package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	thumbmdl "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

func (s *Service) checkBnjAccess(mid int64) bool {
	return true
}

// Bnj2019Aids get bnj aids.3
func (s Service) Bnj2019Aids(c context.Context) []int64 {
	aids := s.c.Bnj2019.BnjListAids
	aids = append(aids, s.c.Bnj2019.BnjMainAid)
	return aids
}

// Timeline get timeline.
func (s *Service) Timeline(c context.Context, mid int64) (data []*model.Timeline, err error) {
	if !s.checkBnjAccess(mid) {
		err = ecode.AccessDenied
		return
	}
	for _, v := range s.c.Bnj2019.Timeline {
		data = append(data, &model.Timeline{
			Name:    v.Name,
			Start:   v.Start.Unix(),
			End:     v.End.Unix(),
			Cover:   v.Cover,
			H5Cover: v.H5Cover,
		})
	}
	return
}

// Bnj2019 get bnj2019 data.
func (s *Service) Bnj2019(c context.Context, mid int64) (data *model.Bnj2019, err error) {
	if !s.checkBnjAccess(mid) {
		err = ecode.AccessDenied
		return
	}
	if s.bnj2019View == nil || !s.bnj2019View.Arc.IsNormal() {
		err = ecode.NothingFound
		return
	}
	data = &model.Bnj2019{
		Bnj2019View: s.bnj2019View,
		Elec:        s.bnjElecInfo,
		Related:     s.bnj2019List,
		ReqUser:     &model.ReqUser{},
	}
	if len(data.Related) == 0 {
		data.Related = make([]*model.Bnj2019Related, 0)
	}
	if mid > 0 {
		authorMid := s.bnj2019View.Author.Mid
		aid := s.bnj2019View.Aid
		ip := metadata.String(c, metadata.RemoteIP)
		group, errCtx := errgroup.WithContext(c)
		// attention
		group.Go(func() error {
			if resp, e := s.accGRPC.Relation3(errCtx, &accmdl.RelationReq{Mid: mid, Owner: authorMid, RealIp: ip}); e != nil {
				log.Error("Bnj2019 s.accGRPC.Relation3(%d,%d,%s) error(%v)", mid, authorMid, ip, e)
			} else if resp != nil {
				data.ReqUser.Attention = resp.Following
			}
			return nil
		})
		// favorite
		group.Go(func() error {
			if resp, e := s.favGRPC.IsFavored(errCtx, &favgrpc.IsFavoredReq{Typ: int32(favmdl.TypeVideo), Mid: mid, Oid: aid}); e != nil {
				log.Error("Bnj2019 s.fav.IsFav(%d,%d,%s) error(%v)", mid, aid, ip, e)
			} else if resp != nil {
				data.ReqUser.Favorite = resp.Faved
			}
			return nil
		})
		// like
		group.Go(func() error {
			if resp, e := s.thumbupGRPC.HasLike(errCtx, &thumbmdl.HasLikeReq{Business: _businessLike, MessageIds: []int64{aid}, Mid: mid, IP: ip}); e != nil {
				log.Error("Bnj2019 s.thumbupGRPC.HasLike(%d,%d,%s) error %v", aid, mid, ip, e)
			} else if resp != nil && resp.States != nil {
				if v, ok := resp.States[aid]; ok {
					switch v.State {
					case thumbmdl.State_STATE_LIKE:
						data.ReqUser.Like = true
					case thumbmdl.State_STATE_DISLIKE:
						data.ReqUser.Dislike = true
					default:
					}
				}
			}
			return nil
		})
		// coin
		group.Go(func() error {
			if resp, e := s.coinGRPC.ItemUserCoins(errCtx, &coinmdl.ItemUserCoinsReq{Mid: mid, Aid: aid, Business: model.CoinArcBusiness}); e != nil {
				log.Error("Bnj2019 s.coinGRPC.ItemUserCoins(%d,%d,%s) error %v", mid, aid, ip, e)
			} else if resp != nil {
				data.ReqUser.Coin = resp.Number
			}
			return nil
		})
		if err := group.Wait(); err != nil {
			log.Error("%+v", err)
		}
	}
	return
}

func (s *Service) loadBnj2019MainArc() {
	if s.bnj19MainRunning {
		return
	}
	s.bnj19MainRunning = true
	defer func() {
		s.bnj19MainRunning = false
	}()
	if s.c.Bnj2019.BnjMainAid == 0 {
		return
	}
	viewReply, err := s.arcGRPC.View(context.Background(), &arcmdl.ViewRequest{Aid: s.c.Bnj2019.BnjMainAid})
	if err != nil {
		log.Error("loadBnj2019MainArc main s.arcGRPC.View(%d) error(%v)", s.c.Bnj2019.BnjMainAid, err)
		return
	}
	if viewReply != nil && viewReply.Arc != nil {
		model.ClearAttrAndAccess(viewReply.Arc)
		s.bnj2019View = &model.Bnj2019View{Arc: viewReply.Arc, Pages: viewReply.Pages}
		// elec
		if elec, err := s.ElecShow(context.Background(), viewReply.Arc.Author.Mid, viewReply.Arc.Aid, 0, viewReply.Arc); err == nil {
			s.bnjElecInfo = elec
			s.bnjElecInfo.TotalCount += s.c.Bnj2019.FakeElec
		} else {
			log.Error("loadBnj2019MainArc s.dao.ElecShow(%d,%d) error(%v)", viewReply.Arc.Author.Mid, viewReply.Arc.Aid, err)
		}
	}
}

func (s *Service) loadBnj2019ArcList() {
	if s.bnj19ListRunning {
		return
	}
	s.bnj19ListRunning = true
	defer func() {
		s.bnj19ListRunning = false
	}()
	if len(s.c.Bnj2019.BnjListAids) == 0 {
		return
	}
	viewsReply, err := s.arcGRPC.Views(context.Background(), &arcmdl.ViewsRequest{Aids: s.c.Bnj2019.BnjListAids})
	if err != nil || viewsReply == nil {
		log.Error("loadBnj2019ArcList list s.arcGRPC.Views(%v) error(%v)", s.c.Bnj2019.BnjListAids, err)
		return
	}
	var tmpList []*model.Bnj2019Related
	for _, aid := range s.c.Bnj2019.BnjListAids {
		if view, ok := viewsReply.Views[aid]; ok && view != nil && view.Arc != nil && view.Arc.IsNormal() {
			model.ClearAttrAndAccess(view.Arc)
			item := &model.Bnj2019Related{Arc: view.Arc, Pages: view.Pages}
			tmpList = append(tmpList, item)
		}
	}
	s.bnj2019List = tmpList
}
