package web

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web-goblin/interface/model/web"
	"go-gateway/pkg/idsafe/bvid"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"

	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artgrpc "git.bilibili.co/bapis/bapis-go/article/service"
	cheeseepgrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/episode"
	hisgrpc "git.bilibili.co/bapis/bapis-go/community/interface/history"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	livexroomgrpc "git.bilibili.co/bapis/bapis-go/live/xroom"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

func (s *Service) HisSearch(ctx context.Context, mid int64, buvid, keyword string, pn int64, business string) (*web.HisSearchReply, error) {
	businesses, ok := s.c.HisSearch.Business[business]
	if !ok {
		return nil, ecode.Error(ecode.RequestErr, "business参数值错误")
	}
	ps := s.c.HisSearch.Ps
	in := &hisgrpc.SearchHistoryReq{
		Mid:        mid,
		Buvid:      buvid,
		Key:        keyword,
		Pn:         int32(pn),
		Ps:         int32(ps),
		Businesses: businesses,
	}
	reply, err := s.hisGRPC.SearchHistory(ctx, in)
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, nil
	}
	res := &web.HisSearchReply{
		HasMore: len(reply.Res) >= int(ps),
		Page: &web.Page{
			Pn:    pn,
			Total: int64(reply.Total),
		},
	}
	if len(reply.Res) == 0 {
		return res, nil
	}
	res.List = s.TogetherHistory(ctx, mid, reply.Res, keyword)
	return res, nil
}

// TogetherHistory always return 0~50
// nolint: gocognit
func (s *Service) TogetherHistory(ctx context.Context, mid int64, res []*hisgrpc.ModelResource, keyword string) (data []*web.HisItem) {
	const (
		_tpOld         = -1
		_tpOffline     = 0
		_tpArchive     = 3
		_tpPGC         = 4
		_tpArticle     = 5
		_tpLive        = 6
		_tpArticleList = 7
		_tpCheese      = 10
	)
	var (
		_badge = map[int32]string{
			1: "番剧",
			2: "电影",
			3: "纪录片",
			4: "国创",
			5: "电视剧",
			7: "综艺",
		}
		aids, articleIDs, roomIDs []int64
		cheeseepIDs, epids        []int32
		archive                   map[int64]*arcgrpc.ViewReply
		isFavs                    map[int64]bool
		pgcInfo                   map[int32]*episodegrpc.EpisodeCardsProto
		article                   map[int64]*artmdl.Meta
		live                      map[int64]*livexroomgrpc.Infos
		cheeseCards               map[int32]*cheeseepgrpc.EpisodeCard
		accCards                  map[int64]*accgrpc.Card
	)
	for _, his := range res {
		switch his.Tp {
		case _tpOld, _tpOffline, _tpArchive:
			if his.Oid <= 0 {
				continue
			}
			aids = append(aids, his.Oid)
		case _tpPGC:
			if his.Oid <= 0 {
				continue
			}
			aids = append(aids, his.Oid) //用cid拿时长duration
			if his.Epid <= 0 {
				continue
			}
			epids = append(epids, int32(his.Epid))
		case _tpArticle:
			if his.Oid <= 0 {
				continue
			}
			articleIDs = append(articleIDs, his.Oid)
		case _tpLive:
			if his.Oid <= 0 {
				continue
			}
			roomIDs = append(roomIDs, his.Oid)
		case _tpArticleList:
			if his.Cid <= 0 {
				continue
			}
			articleIDs = append(articleIDs, his.Cid)
		case _tpCheese:
			if his.Oid <= 0 {
				continue
			}
			aids = append(aids, his.Oid) //用cid拿时长duration
			if his.Epid <= 0 {
				continue
			}
			cheeseepIDs = append(cheeseepIDs, int32(his.Epid))
		default:
			log.Warn("unknown history type(%d) msg(%+v)", his.Tp, his)
		}
	}
	g := errgroup.WithContext(ctx)
	if len(aids) > 0 {
		g.Go(func(ctx context.Context) error {
			reply, err := s.arcGRPC.Views(ctx, &arcgrpc.ViewsRequest{Aids: aids})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			archive = reply.GetViews()
			return nil
		})
		g.Go(func(ctx context.Context) error {
			reply, err := s.favGRPC.IsFavoreds(ctx, &favgrpc.IsFavoredsReq{Typ: 2, Mid: mid, Oids: aids})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			isFavs = reply.GetFaveds()
			return nil
		})
	}
	if len(epids) > 0 {
		g.Go(func(ctx context.Context) error {
			reply, err := s.episodeGRPC.Cards(ctx, &episodegrpc.EpReq{Epids: epids})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			pgcInfo = reply.GetCards()
			return nil
		})
	}
	if len(articleIDs) > 0 {
		g.Go(func(ctx context.Context) error {
			reply, err := s.artGRPC.ArticleMetas(ctx, &artgrpc.ArticleMetasReq{Ids: articleIDs})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			article = reply.GetRes()
			return nil
		})
	}
	if len(roomIDs) > 0 {
		g.Go(func(ctx context.Context) error {
			in := &livexroomgrpc.RoomIDsReq{RoomIds: roomIDs, Attrs: []string{"show", "status", "area"}}
			reply, err := s.livexroomGRPC.GetMultiple(ctx, in)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			if reply == nil {
				return nil
			}
			live = reply.List
			var upIDs []int64
			if len(live) != 0 {
				for _, lv := range live {
					upIDs = append(upIDs, lv.Uid)
				}
			}
			reply1, err := s.accGRPC.Cards3(ctx, &accgrpc.MidsReq{Mids: upIDs})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			accCards = reply1.GetCards()
			return nil
		})
	}
	if len(cheeseepIDs) > 0 {
		g.Go(func(ctx context.Context) error {
			reply, err := s.cheeseepGRPC.Cards(ctx, &cheeseepgrpc.EpisodeCardsReq{Ids: cheeseepIDs})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			cheeseCards = reply.GetCards()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	for _, his := range res {
		item := &web.HisItem{
			ViewAt: his.Unix,
		}
		item.History.Oid = his.Oid
		item.History.Cid = his.Cid
		item.History.Epid = his.Epid
		item.History.Business = his.Business
		item.Kid = his.Kid
		item.History.Dt = his.Dt
		switch his.Tp {
		case _tpOld, _tpOffline, _tpArchive:
			arc, ok := archive[his.Oid]
			if !ok {
				continue
			}
			item.History.Bvid = s.avToBv(arc.Aid)
			item.Title = arc.Title
			item.Cover = arc.Pic
			item.AuthorMid = arc.Author.Mid
			item.AuthorName = arc.Author.Name
			item.AuthorFace = arc.Author.Face
			item.TagName = arc.TypeName
			item.Videos = arc.Videos
			if arc.AttrVal(arcgrpc.AttrBitSteinsGate) == arcgrpc.AttrYes { // 互动视频进度不展示
				item.Progress = 0
				item.History.Page = 1
				item.History.Part = ""
				item.Duration = 0
			} else {
				item.Progress = his.Pro
				for _, p := range arc.Pages {
					if p.Cid == his.Cid {
						item.Duration = p.Duration
						item.History.Cid = p.Cid
						item.History.Page = p.Page
						item.History.Part = p.Part
						break
					}
				}
			}
			// 多p稿件
			if item.Videos > 1 {
				item.NewDesc = fmt.Sprintf("共%dP", item.Videos)
				item.ShowTitle = item.History.Part
			}
			if isFav, ok := isFavs[arc.Aid]; ok && isFav {
				item.IsFav = 1
			}
		case _tpPGC:
			pgc, okPGC := pgcInfo[int32(his.Epid)]
			arc, okArc := archive[his.Oid]
			if !okPGC || !okArc || pgc.Season == nil {
				continue
			}
			item.Title = pgc.Season.Title
			item.Total = pgc.Season.TotalCount
			item.IsFinish = pgc.Season.IsFinish
			item.NewDesc = pgc.Season.NewEpShow
			item.LongTitle = pgc.LongTitle
			item.ShowTitle = pgc.ShowTitle
			item.Cover = pgc.Cover
			item.Badge = _badge[his.Stp]
			item.Progress = his.Pro
			item.URI = pgc.Url // 默认用pgc返回的url
			for _, p := range arc.Pages {
				if p.Cid == his.Cid {
					item.Duration = p.Duration
					break
				}
			}
		case _tpArticle:
			art, ok := article[his.Oid]
			if !ok {
				continue
			}
			item.Title = art.Title
			item.Covers = art.ImageURLs
			if art.Author != nil {
				item.AuthorMid = art.Author.Mid
				item.AuthorName = art.Author.Name
				item.AuthorFace = art.Author.Face
			}
			item.Badge = "专栏"
		case _tpLive:
			lv, ok := live[his.Oid]
			if !ok || lv.Show == nil {
				continue
			}
			accCard, accok := accCards[lv.Uid]
			if !accok {
				continue
			}
			item.Title = lv.Show.Title
			item.Cover = lv.Show.Cover
			if lv.Show.Cover == "" {
				item.Cover = lv.Show.Keyframe
			}
			item.AuthorMid = lv.Uid
			item.AuthorName = accCard.Name
			item.AuthorFace = accCard.Face
			item.URI = web.FillURI(web.GotoLive, strconv.FormatInt(lv.RoomId, 10))
			if lv.Area != nil {
				item.TagName = lv.Area.AreaName
			}
			if lv.Status != nil {
				if lv.Status.LiveStatus == 1 { //1是直播中，0、2是未开播
					item.LiveStatus = 1
				}
			}
		case _tpArticleList:
			art, ok := article[his.Cid]
			if !ok {
				continue
			}
			item.Title = art.Title
			item.Covers = art.ImageURLs
			if art.Author != nil {
				item.AuthorMid = art.Author.Mid
				item.AuthorName = art.Author.Name
				item.AuthorFace = art.Author.Face
			}
			item.Badge = "专栏"
		case _tpCheese:
			cheese, okCheese := cheeseCards[int32(his.Epid)]
			arc, okArc := archive[his.Oid]
			if !okCheese || !okArc {
				continue
			}
			item.Title = cheese.SeasonTitle
			item.ShowTitle = cheese.Title
			item.Cover = cheese.Cover
			item.Progress = his.Pro
			item.URI = cheese.Url
			for _, p := range arc.Pages {
				if p.Cid == his.Cid {
					item.Duration = p.Duration
					break
				}
			}
		default:
			continue
		}
		data = append(data, item)
	}
	return
}

func (s *Service) avToBv(aid int64) (bvID string) {
	var err error
	if bvID, err = bvid.AvToBv(aid); err != nil {
		log.Warn("avToBv(%d) error(%v)", aid, err)
	}
	return
}
