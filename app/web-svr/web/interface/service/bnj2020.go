package service

import (
	"context"
	"strconv"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	v1 "go-gateway/app/app-svr/resource/service/api/v1"
	steinsapi "go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/web-svr/web/interface/model"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	payrank "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	prmdl "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank/model"
	coinmdl "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	thumbmdl "git.bilibili.co/bapis/bapis-go/community/service/thumbup"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) checkBnj2020Access(mid int64) bool {
	return true
}

// Timeline get timeline.
func (s *Service) Bnj20Timeline(c context.Context, mid int64) (data []*model.Timeline, err error) {
	if !s.checkBnj2020Access(mid) {
		err = ecode.AccessDenied
		return
	}
	for _, v := range s.c.Bnj2020.Timeline {
		data = append(data, &model.Timeline{
			Name:     v.Name,
			Start:    v.Start.Unix(),
			End:      v.End.Unix(),
			Cover:    v.Cover,
			H5Cover:  v.H5Cover,
			Subtitle: v.Subtitle,
			Type:     v.Type,
			Tag:      v.Tag,
		})
	}
	return
}

// Bnj2020 get bnj 2020 data.
func (s *Service) Bnj2020(c context.Context, mid int64) (data *model.Bnj2020, err error) {
	if !s.checkBnj2020Access(mid) {
		err = ecode.AccessDenied
		return
	}
	if s.bnj20Cache.MainView == nil || !s.bnj20Cache.MainView.Arc.IsNormal() {
		err = ecode.NothingFound
		return
	}
	data = &model.Bnj2020{
		ViewReply: s.bnj20Cache.MainView,
		SpView:    s.bnj20Cache.SpView,
		Related:   s.bnj20Cache.RelatedList,
		ReqUser:   &model.ReqUser{},
	}
	if s.bnj20Cache.ElecInfo != nil && s.bnj20Cache.ElecInfo.RankElecAVProto != nil {
		data.Elec = &model.BnjElec{TotalCount: s.bnj20Cache.ElecInfo.CountUPTotalElec}
	}
	if len(data.Related) == 0 {
		data.Related = make([]*arcmdl.ViewReply, 0)
	}
	if mid > 0 {
		authorMid := s.bnj20Cache.MainView.Author.Mid
		aid := s.bnj20Cache.MainView.Aid
		ip := metadata.String(c, metadata.RemoteIP)
		group := errgroup.WithContext(c)
		// attention
		group.Go(func(ctx context.Context) error {
			resp, e := s.accGRPC.Relation3(ctx, &accmdl.RelationReq{Mid: mid, Owner: authorMid, RealIp: ip})
			if e != nil {
				log.Error("Bnj2020 s.accGRPC.Relation3(%d,%d,%s) error(%v)", mid, authorMid, ip, e)
				return nil
			}
			if resp != nil {
				data.ReqUser.Attention = resp.Following
			}
			return nil
		})
		// favorite
		group.Go(func(ctx context.Context) error {
			resp, e := s.favGRPC.IsFavored(ctx, &favgrpc.IsFavoredReq{Typ: int32(favmdl.TypeVideo), Mid: mid, Oid: aid})
			if e != nil {
				log.Error("Bnj2020 s.fav.IsFav(%d,%d,%s) error(%v)", mid, aid, ip, e)
				return nil
			}
			if resp != nil {
				data.ReqUser.Favorite = resp.Faved
			}
			return nil
		})
		// like
		group.Go(func(ctx context.Context) error {
			resp, e := s.thumbupGRPC.HasLike(ctx, &thumbmdl.HasLikeReq{Business: _businessLike, MessageIds: []int64{aid}, Mid: mid, IP: ip})
			if e != nil {
				log.Error("Bnj2020 s.thumbupGRPC.HasLike(%d,%d,%s) error %v", aid, mid, ip, e)
				return nil
			}
			if resp != nil && resp.States != nil {
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
		group.Go(func(ctx context.Context) error {
			resp, e := s.coinGRPC.ItemUserCoins(ctx, &coinmdl.ItemUserCoinsReq{Mid: mid, Aid: aid, Business: model.CoinArcBusiness})
			if e != nil {
				log.Error("Bnj2020 s.coinGRPC.ItemUserCoins(%d,%d,%s) error %v", mid, aid, ip, e)
				return nil
			}
			if resp != nil {
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

// Bnj2020Item get bnj 2020 item.
// nolint: gocognit
func (s *Service) Bnj2020Item(c context.Context, aid, mid int64) (data *model.Bnj20Item, err error) {
	var (
		arc       *arcmdl.Arc
		bnjArcs   []*arcmdl.ViewReply
		relations map[int64]*accmdl.RelationReply
		staff     []*model.Staff
		staffMids []int64
	)
	data = new(model.Bnj20Item)
	data.ReqUser = new(model.ReqUser)
	bnjArcs = append(bnjArcs, s.bnj20Cache.MainView)
	if s.bnj20Cache.SpView != nil {
		bnjArcs = append(bnjArcs, &arcmdl.ViewReply{Arc: s.bnj20Cache.SpView.Arc, Pages: s.bnj20Cache.SpView.Pages})
	}
	bnjArcs = append(bnjArcs, s.bnj20Cache.RelatedList...)
	for _, v := range bnjArcs {
		if v != nil && v.Aid == aid {
			arc = v.Arc
		}
	}
	if arc == nil {
		err = ecode.NothingFound
		return
	}
	data.Stat = arc.Stat
	authorMid := arc.Author.Mid
	for _, v := range arc.StaffInfo {
		staffMids = append(staffMids, v.Mid)
	}
	staffMids = append(staffMids, authorMid)
	ip := metadata.String(c, metadata.RemoteIP)
	group := errgroup.WithContext(c)
	if mid > 0 {
		// attention
		group.Go(func(ctx context.Context) error {
			reply, e := s.accGRPC.Relations3(ctx, &accmdl.RelationsReq{Mid: mid, Owners: staffMids, RealIp: ip})
			if e != nil {
				log.Error("Bnj2020Item s.accGRPC.Relation3(%d,%v,%s) error(%v)", mid, staffMids, ip, e)
				return nil
			}
			if reply != nil {
				relations = reply.Relations
				if authorRela, ok := relations[authorMid]; ok && authorRela != nil {
					data.ReqUser.Attention = authorRela.Following
				}
			}
			return nil
		})
		// favorite
		group.Go(func(ctx context.Context) error {
			resp, e := s.favGRPC.IsFavored(ctx, &favgrpc.IsFavoredReq{Typ: int32(favmdl.TypeVideo), Mid: mid, Oid: aid})
			if e != nil {
				log.Error("Bnj2020Item s.fav.IsFav(%d,%d,%s) error(%v)", mid, aid, ip, e)
				return nil
			}
			if resp != nil {
				data.ReqUser.Favorite = resp.Faved
			}
			return nil
		})
		// like
		group.Go(func(ctx context.Context) error {
			resp, e := s.thumbupGRPC.HasLike(ctx, &thumbmdl.HasLikeReq{Business: _businessLike, MessageIds: []int64{aid}, Mid: mid, IP: ip})
			if e != nil {
				log.Error("Bnj2020Item s.thumbup.HasLike(%d,%d,%s) error %v", aid, mid, ip, e)
				return nil
			}
			if resp != nil && resp.States != nil {
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
		group.Go(func(ctx context.Context) error {
			resp, e := s.coinGRPC.ItemUserCoins(ctx, &coinmdl.ItemUserCoinsReq{Mid: mid, Aid: aid, Business: model.CoinArcBusiness})
			if e != nil {
				log.Error("Bnj2020Item s.coinGRPC.ItemUserCoins(%d,%d,%s) error %v", mid, aid, ip, e)
				return nil
			}
			if resp != nil {
				data.ReqUser.Coin = resp.Number
			}
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		staff = s.staffInfo(ctx, arc.Author.Mid, arc.StaffInfo)
		return nil
	})
	//resource
	if rid, ok := s.c.Bnj2020.RelateAidToRid[strconv.FormatInt(aid, 10)]; ok && rid > 0 {
		group.Go(func(ctx context.Context) error {
			reply, e := s.resgrpc.Banners2(ctx, &v1.BannersRequest{Aid: aid, Mid: mid, ResIDs: strconv.FormatInt(rid, 10), Ip: ip})
			if e != nil {
				log.Error("Bnj2020Item s.resourceGRPC.Banners2(%d,%d,%s) error %v", mid, aid, ip, e)
				return nil
			}
			if reply != nil {
				if banners, ok := reply.Banners[int32(rid)]; ok && len(banners.Banners) > 0 {
					for _, v := range banners.Banners {
						data.Banner = append(data.Banner, &model.Banner{
							ID:         v.Id,
							Title:      v.Title,
							Image:      v.Image,
							Hash:       v.Hash,
							URI:        v.Value,
							RequestID:  v.RequestId,
							CreativeID: v.CreativeId,
							SrcID:      v.SrcId,
							IsAd:       v.IsAd,
							IsAdLoc:    v.IsAdLoc,
							AdCb:       v.AdCb,
							ShowURL:    v.ShowUrl,
							ClickURL:   v.ClickUrl,
							ClientIP:   v.ClientIp,
							ServerType: v.ServerType,
							ResourceID: v.ResourceId,
							Index:      v.Index,
							CmMark:     v.CmMark,
						})
					}
				}
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	// fill relation data
	for _, v := range staff {
		if v == nil {
			continue
		}
		item := &model.BnjStaff{
			Staff: v,
		}
		if mid > 0 {
			if relation, ok := relations[v.Mid]; ok && relation != nil {
				item.Attention = relation.Following
			}
		}
		data.Staff = append(data.Staff, item)
	}
	return
}

// Bnj20ElecShow get bnj elec show
func (s *Service) Bnj20ElecShow(c context.Context) *payrank.BNJRankWithPanelReply {
	return s.bnj20Cache.ElecInfo
}

func (s *Service) Bnj2020Aids(c context.Context) []int64 {
	aids := s.c.Bnj2020.ListAids
	aids = append(aids, s.c.Bnj2020.MainAid)
	aids = append(aids, s.c.Bnj2020.SpAid)
	return aids
}

func (s *Service) loadBnj2020MainView() {
	if s.bnj20MainRunning {
		return
	}
	s.bnj20MainRunning = true
	defer func() {
		s.bnj20MainRunning = false
	}()
	if s.c.Bnj2020.MainAid == 0 {
		return
	}
	ctx := context.Background()
	viewReply, err := s.arcGRPC.View(ctx, &arcmdl.ViewRequest{Aid: s.c.Bnj2020.MainAid})
	if err != nil {
		log.Error("loadBnj2020MainView main s.arcGRPC.View(%d) error(%v)", s.c.Bnj2020.MainAid, err)
		return
	}
	if viewReply != nil {
		model.ClearAttrAndAccess(viewReply.Arc)
		s.bnj20Cache.MainView = viewReply
		// elec
		elec, err := s.payRankGRPC.BNJRankWithPanel(ctx, &payrank.RankElecAVReq{UPMID: viewReply.Author.Mid, AVID: s.c.Bnj2020.MainAid, RankSize: 30})
		if err != nil || elec == nil {
			log.Error("loadBnj2020MainView s.dao.ElecShow(%d,%d) error(%v) or elec nil", viewReply.Arc.Author.Mid, viewReply.Arc.Aid, err)
			return
		}
		if elec.RankElecAVProto == nil {
			elec.RankElecAVProto = new(prmdl.RankElecAVProto)
		}
		elec.RankElecAVProto.CountUPTotalElec += s.bnj20Cache.LiveGiftCnt
		s.bnj20Cache.ElecInfo = elec
	}
}

func (s *Service) loadBnj2020LiveArc() {
	if s.bnj20LiveRunning {
		return
	}
	s.bnj20LiveRunning = true
	defer func() {
		s.bnj20LiveRunning = false
	}()
	if s.c.Bnj2020.LiveAid == 0 {
		return
	}
	arcReply, err := s.arcGRPC.Arc(context.Background(), &arcmdl.ArcRequest{Aid: s.c.Bnj2020.LiveAid})
	if err != nil {
		log.Error("loadBnj2020LiveArc live arc s.arcGRPC.Arc(%d) error(%v)", s.c.Bnj2020.LiveAid, err)
		return
	}
	if arcReply != nil {
		model.ClearAttrAndAccess(arcReply.Arc)
		s.bnj20Cache.LiveArc = arcReply
	}
}

func (s *Service) loadBnj2020SpView() {
	if s.bnj20SpRunning {
		return
	}
	s.bnj20SpRunning = true
	defer func() {
		s.bnj20SpRunning = false
	}()
	if s.c.Bnj2020.SpAid == 0 {
		return
	}
	viewReply, err := s.arcGRPC.SteinsGateView(context.Background(), &arcmdl.SteinsGateViewRequest{Aid: s.c.Bnj2020.SpAid})
	if err != nil || viewReply == nil {
		log.Error("loadBnj2020SpView sp arc s.arcGRPC.SteinsGateView(%d) error(%v) or viewReply nil", s.c.Bnj2020.SpAid, err)
		return
	}
	steinsView, err := s.steinsGRPC.GraphView(context.Background(), &steinsapi.GraphViewReq{Aid: s.c.Bnj2020.SpAid})
	if err != nil || steinsView == nil {
		log.Error("loadBnj2020SpView sp arc s.arcGRPC.GraphView(%d) error(%v)", s.c.Bnj2020.SpAid, err)
		return
	}
	// replace page and first cid
	if steinsView.Page != nil {
		viewReply.Pages = []*arcmdl.Page{model.ArchivePage(steinsView.Page)}
		viewReply.FirstCid = steinsView.Page.Cid
	}
	model.ClearAttrAndAccess(viewReply.Arc)
	s.bnj20Cache.SpView = viewReply
}

func (s *Service) loadBnj2020ViewList() {
	if s.bnj20ListRunning {
		return
	}
	s.bnj20ListRunning = true
	defer func() {
		s.bnj20ListRunning = false
	}()
	if len(s.c.Bnj2020.ListAids) == 0 {
		return
	}
	viewsReply, err := s.arcGRPC.Views(context.Background(), &arcmdl.ViewsRequest{Aids: s.c.Bnj2020.ListAids})
	if err != nil || viewsReply == nil {
		log.Error("loadBnj2020ViewList list s.arcGRPC.Views(%v) error(%v)", s.c.Bnj2020.ListAids, err)
		return
	}
	var tmpList []*arcmdl.ViewReply
	for _, aid := range s.c.Bnj2020.ListAids {
		if view, ok := viewsReply.Views[aid]; ok && view != nil && view.Arc != nil && view.Arc.IsNormal() {
			model.ClearAttrAndAccess(view.Arc)
			tmpList = append(tmpList, view)
		}
	}
	s.bnj20Cache.RelatedList = tmpList
}
