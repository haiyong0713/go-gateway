package service

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	"go-gateway/app/app-svr/archive/service/api"
	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
	dygrpc "go-gateway/app/web-svr/dynamic/service/api/v1"
	"go-gateway/app/web-svr/web/interface/model"

	articlemdl "git.bilibili.co/bapis/bapis-go/article/model"
)

const _allRank = 0

func (s *Service) Information(c context.Context, rid int32, tp int8, pn, ps int, mid int64) (*model.Informations, error) {
	// 判断是否资讯区
	var (
		resManagers []*model.Information
		resOrigins  []*model.BvArc
		count       int
	)
	g, ctx := errgroup.WithContext(c)
	// 资讯区获取数据
	if pn == 1 {
		g.Go(func() (err error) {
			if resManagers, err = s.InformationCard(c, rid, mid); err != nil {
				log.Error("%v", err)
				return nil
			}
			return
		})
	}
	g.Go(func() (err error) {
		if resOrigins, count, err = s.NewList(ctx, rid, tp, pn, ps); err != nil {
			log.Error("%v", err)
		}
		return
	})
	if err := g.Wait(); err != nil {
		log.Error("%v", err)
		return nil, err
	}
	var (
		res   = new(model.Informations)
		items []*model.Information
	)
	res.Page = &model.Page{Num: pn, Size: ps, Count: count}
	for _, reOrigin := range resOrigins {
		if reOrigin == nil || !reOrigin.IsNormal() {
			continue
		}
		item := &model.Information{
			CardType: "archive",
			ID:       strconv.FormatInt(reOrigin.GetAid(), 10),
			Title:    reOrigin.GetTitle(),
			Cover:    reOrigin.GetPic(),
			Duration: reOrigin.GetDuration(),
			Stat:     &model.Stat{View: reOrigin.Stat.View, Danmaku: reOrigin.Stat.Danmaku},
			Author:   &model.Author{MID: reOrigin.Author.Mid, Face: reOrigin.Author.Face, Name: reOrigin.Author.Name},
			BVID:     reOrigin.Bvid,
			Cid:      reOrigin.GetFirstCid(),
		}
		items = append(items, item)
	}
	// 资讯区插入运营数据
	for _, resManager := range resManagers {
		if resManager.Position > len(items) {
			items = append(items, resManager)
		} else {
			items = append(items[:resManager.Position-1], append([]*model.Information{resManager}, items[resManager.Position-1:]...)...)
		}
	}
	// 根据最大长度裁剪
	var cutCount = len(items) - len(resManagers)
	res.Items = items[:cutCount]
	return res, nil
}

// nolint: gocognit,gomnd
func (s *Service) InformationCard(c context.Context, rid int32, mid int64) ([]*model.Information, error) {
	var (
		idxs       []int
		rcm        map[int]*resourcegrpc.InformationRegionCard
		aids       []int64
		dynamicIDs []int64
		articleIDs []int64
	)
	rcm = make(map[int]*resourcegrpc.InformationRegionCard)
	for _, rc := range s.informationRegionCardCache[rid] {
		if rc == nil {
			continue
		}
		idxs = append(idxs, int(rc.GetPositionIdx()))
		rcm[int(rc.GetPositionIdx())] = rc
		switch rc.GetCardType() {
		case 1: // 稿件
			if rc.GetRid() != 0 {
				aids = append(aids, rc.GetRid())
			}
		case 2: // 动态
			if rc.GetRid() != 0 {
				dynamicIDs = append(dynamicIDs, rc.GetRid())
			}
		case 3: // 专栏
			if rc.GetRid() != 0 {
				articleIDs = append(articleIDs, rc.GetRid())
			}
		}
	}
	var (
		arcm     map[int64]*api.Arc
		articlem map[int64]*articlemdl.Meta
		dynamicm map[int64]*model.DynamicCard
	)
	g, ctx := errgroup.WithContext(c)
	// 获取稿件详情
	if len(aids) > 0 {
		g.Go(func() (err error) {
			if arcm, err = s.dao.Arcs(ctx, aids); err != nil {
				log.Error("%v", err)
			}
			return nil
		})
	}
	if len(dynamicIDs) > 0 {
		g.Go(func() (err error) {
			if dynamicm, err = s.dao.DrawInfos(ctx, mid, dynamicIDs); err != nil {
				log.Error("%v", err)
			}
			return nil
		})
	}
	if len(articleIDs) > 0 {
		g.Go(func() (err error) {
			if articlem, err = s.dao.Articles(ctx, articleIDs); err != nil {
				log.Error("%v", err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%v", err)
		return nil, err
	}
	sort.Ints(idxs)
	var res []*model.Information
	for _, idx := range idxs {
		rc := rcm[idx]
		var item *model.Information
		switch rc.GetCardType() {
		case 1: // 稿件
			aid := rc.GetRid()
			if arc, ok := arcm[aid]; ok {
				item = &model.Information{
					CardType: "archive",
					ID:       strconv.FormatInt(aid, 10),
					Title:    arc.GetTitle(),
					Cover:    arc.GetPic(),
					Duration: arc.GetDuration(),
					Stat:     &model.Stat{View: arc.Stat.View, Danmaku: arc.Stat.Danmaku, Like: arc.Stat.Like, Reply: arc.Stat.Reply},
					Author:   &model.Author{MID: arc.Author.Mid, Face: arc.Author.Face, Name: arc.Author.Name},
					Position: int(rc.GetPositionIdx()),
					BVID:     s.avToBv(aid),
					Cid:      arc.GetFirstCid(),
				}
			}
		case 2: // 动态
			dynamicID := rc.GetRid()
			if dynamic, ok := dynamicm[dynamicID]; ok {
				var DrawDetail *model.DrawDetail
				if err := json.Unmarshal([]byte(dynamic.Card), &DrawDetail); err != nil {
					log.Error("%v", err)
					continue
				}
				title := DrawDetail.Item.Description
				titleRune := []rune(title)
				if len(titleRune) > 50 {
					title = string(titleRune[:50])
				}
				item = &model.Information{
					CardType: "dynamic",
					ID:       strconv.FormatInt(dynamicID, 10),
					Title:    title,
					Stat:     &model.Stat{View: dynamic.Desc.View, Like: dynamic.Desc.Like, Reply: int32(DrawDetail.Item.Reply)},
					Author:   &model.Author{MID: dynamic.Desc.UserProfile.Info.UID, Face: dynamic.Desc.UserProfile.Info.Face, Name: dynamic.Desc.UserProfile.Info.UName},
					Position: int(rc.GetPositionIdx()),
				}
				for _, pic := range DrawDetail.Item.Pictures {
					item.Cover = pic.ImgSrc
					break
				}
			}
		case 3: // 专栏
			articleID := rc.GetRid()
			if article, ok := articlem[articleID]; ok {
				item = &model.Information{
					CardType: "article",
					ID:       strconv.FormatInt(articleID, 10),
					Title:    article.Title,
					Position: int(rc.GetPositionIdx()),
				}
				for _, img := range article.ImageURLs {
					item.Cover = img
					break
				}
				if article.Stats != nil {
					item.Stat = &model.Stat{View: article.Stats.View, Like: int32(article.Stats.Like), Reply: int32(article.Stats.Reply)}
				}
				if article.Author != nil {
					item.Author = &model.Author{MID: article.Author.Mid, Face: article.Author.Face, Name: article.Author.Name}
				}
			}
		}
		if item != nil {
			res = append(res, item)
		}
	}
	return res, nil
}

// NewList get new list by region id.
func (s *Service) NewList(ctx context.Context, rid int32, tp int8, pn, ps int) (res []*model.BvArc, count int, err error) {
	// check first or second region id.
	var firstType bool
	for _, val := range s.c.Rule.Rids {
		if rid == val {
			firstType = true
			break
		}
	}
	// get from dynamic-service
	ip := metadata.String(ctx, metadata.RemoteIP)
	var allArcs []*api.Arc
	if firstType {
		allArcs, err = func() ([]*api.Arc, error) {
			listArcs, topErr := s.rankTopArcs(ctx, rid, _firstPage, ps, ip)
			if topErr != nil {
				log.Error("NewList rankTopArcs rid:%d max:%d error:%v", rid, ps, err)
				return nil, topErr
			}
			// 特殊一级分区数据不够从二级分区补充数据
			func() {
				if tid, ok := model.NewListRid[rid]; ok && len(listArcs) < s.c.Rule.MinNewListCnt {
					moreArcs, _, moreErr := s.rankArcs(ctx, tid, tp, _samplePn, s.c.Rule.MinNewListCnt)
					if moreErr != nil {
						log.Error("NewList rankArcs tid:%d tp:%d pn:%d ps:%d error:%+v", tid, tp, _samplePn, s.c.Rule.MinNewListCnt, moreErr)
						return
					}
					allAidMap := make(map[int64]int64, len(listArcs))
					for _, arc := range listArcs {
						allAidMap[arc.Aid] = arc.Aid
					}
					for _, v := range moreArcs {
						if _, ok := allAidMap[v.Aid]; ok {
							continue
						}
						listArcs = append(listArcs, v)
					}
					log.Info("NewList more arcs rid(%d) tid(%d) len allArcs(%d) len moreArcs(%d)", tid, rid, len(listArcs), len(moreArcs))
				}
			}()
			return listArcs, nil
		}()
	} else {
		allArcs, count, err = func() ([]*api.Arc, int, error) {
			listArcs, allCount, rankErr := s.rankArcs(ctx, rid, tp, pn, ps)
			if err != nil {
				log.Error("NewList rankArcs rid:%d tp:%d pn:%d ps:%d error:%v", rid, tp, pn, ps, err)
				return nil, 0, rankErr
			}
			return listArcs, int(allCount), nil
		}()
	}
	if err != nil {
		// 第一页数据降级
		if pn == _firstPage {
			res, count, err = s.dao.NewListBakCache(ctx, rid, tp)
			if err != nil || len(res) == 0 {
				log.Error("NewList rid:%d pn:%d res nil", rid, pn)
				res = _emptyBvArc
			}
			if len(res) > ps {
				res = res[:ps]
			}
		}
		return res, count, nil
	}
	res = s.fmtArcs3(allArcs, nil)
	return res, count, nil
}

func (s *Service) LpNewlist(ctx context.Context, mid int64, req *model.LpNewlistReq) (*model.LpArc, error) {
	lp, ok := s.c.LandingPage[req.Business]
	if !ok {
		return nil, ecode.NothingFound
	}
	cid, ok := lp.Newlist[strconv.FormatInt(req.Rid, 10)]
	if !ok {
		return nil, ecode.NothingFound
	}
	aids, count, err := s.dao.RecentChannelWebFeed(ctx, cid, mid, req.Pn, req.Ps)
	if err != nil {
		return nil, err
	}
	reply, err := s.batchArchives(ctx, aids)
	if err != nil {
		return nil, err
	}
	var arcs []*api.Arc
	for _, aid := range aids {
		if a, ok := reply[aid]; ok && a.IsNormal() {
			arcs = append(arcs, a)
		}
	}
	page := &model.LpPage{
		Count: count,
		Num:   req.Pn,
		Size:  req.Ps,
	}
	return &model.LpArc{
		Archives: s.fmtArcs3(arcs, nil),
		Page:     page,
	}, nil
}

func (s *Service) rankArcs(c context.Context, rid int32, tp int8, pn, ps int) ([]*api.Arc, int64, error) {
	if rid == _allRank {
		res, err := s.dyGRPC.RecentWeeklyArc(c, &dygrpc.RecentWeeklyArcReq{Pn: int64(pn), Ps: int64(ps)})
		if err != nil {
			log.Error("arcrpc.RankAllArcs2(%d,%d) error(%v)", pn, ps, err)
			return nil, 0, err
		}
		return res.Archives, res.Count, nil
	}
	regRes, err := s.dyGRPC.RegAllArcs(c, &dygrpc.RegAllReq{Rid: int64(rid), Type: int32(tp), Pn: int64(pn), Ps: int64(ps)})
	if err != nil {
		log.Error("s.dyGRPC.RegAllArcs(%d,%d,%d) error(%v)", rid, pn, ps, err)
		return nil, 0, err
	}
	return regRes.Archives, regRes.Count, nil
}

func (s *Service) rankTopArcs(c context.Context, rid int32, pn, ps int, ip string) (arcs []*api.Arc, err error) {
	var res *dygrpc.RecentThrdRegArcReply
	arg := &dygrpc.RecentThrdRegArcReq{Rid: rid, Pn: int64(pn), Ps: int64(ps)}
	if res, err = s.dyGRPC.RecentThrdRegArc(c, arg); err != nil {
		log.Error("arcrpc.RankTopArcs3(%d,%d,%d,%s) error(%v)", rid, pn, ps, ip, err)
		return nil, err
	}
	return res.Archives, nil
}
