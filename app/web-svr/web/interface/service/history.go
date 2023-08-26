package service

import (
	"context"
	"fmt"
	"strconv"

	hisapi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	"go-common/library/ecode"
	"go-common/library/log"

	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"

	api "git.bilibili.co/bapis/bapis-go/account/service"
	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	cheeseapi "git.bilibili.co/bapis/bapis-go/cheese/service/season/episode"
	favapi "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	roomapi "git.bilibili.co/bapis/bapis-go/live/xroom"
	epapi "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"

	"go-common/library/sync/errgroup.v2"
)

const (
	_hisBusinessPgc     = "pgc"
	_hisBusinessArt     = "article"
	_hisBusinessArc     = "archive"
	_hisBusinessArtList = "article-list"
	_hisBusinessCheese  = "cheese"
	_hisBusinessLive    = "live"
	_wxHistoryType      = "wx-archive"
	_hisIsFav           = 1
)

var (
	hisBadge = map[int32]string{
		1: "番剧",
		2: "电影",
		3: "纪录片",
		4: "国创",
		5: "电视剧",
		7: "综艺",
	}
	hisTab = []*model.HisTab{
		{
			Type: "archive",
			Name: "视频",
		},
		{
			Type: "live",
			Name: "直播",
		},
		{
			Type: "article",
			Name: "专栏",
		},
	}
	hisBusiness = map[string][]string{
		"":           {_hisBusinessArc, _hisBusinessPgc, _hisBusinessArt, _hisBusinessArtList, _hisBusinessCheese, _hisBusinessLive},
		"all":        {_hisBusinessArc, _hisBusinessPgc, _hisBusinessArt, _hisBusinessArtList, _hisBusinessCheese, _hisBusinessLive},
		"archive":    {_hisBusinessArc, _hisBusinessPgc, _hisBusinessCheese},
		"wx-archive": {_hisBusinessArc, _hisBusinessPgc},
		"article":    {_hisBusinessArt, _hisBusinessArtList},
		"live":       {_hisBusinessLive},
	}
)

// HistoryCursor .
// nolint: gocognit
func (s *Service) HistoryCursor(c context.Context, mid, max, viewAt int64, business, typ string, ps int32) (*model.HisRes, error) {
	businesses, ok := hisBusiness[typ]
	if !ok {
		return nil, ecode.RequestErr
	}
	var (
		aids        []int64
		epids       []int32
		articleIDs  []int64
		cheeseEpids []int32
		roomIDs     []int64
		liveUpMids  []int64
		views       map[int64]*arcapi.ViewReply
		epInfos     map[int32]*epapi.EpisodeCardsProto
		article     map[int64]*artmdl.Meta
		cheeseCards map[int32]*cheeseapi.EpisodeCard
		roomInfos   map[int64]*roomapi.Infos
		isFavs      map[int64]bool
	)
	arg := &hisapi.HistoryCursorReq{
		Mid:        mid,
		Max:        max,
		ViewAt:     viewAt,
		Ps:         ps,
		Business:   business,
		Businesses: businesses,
	}
	data := &model.HisRes{
		Tab:  hisTab,
		List: []*model.HisItem{},
	}
	hisReply, err := s.hisGRPC.HistoryCursor(c, arg)
	if err != nil {
		log.Error("HistoryCursor s.hisGRPC.HistoryCursor(%+v) error(%v)", arg, err)
		return data, nil
	}
	if hisReply == nil || len(hisReply.Res) == 0 {
		return data, nil
	}
	for _, v := range hisReply.Res {
		switch v.Business {
		case _hisBusinessPgc:
			aids = append(aids, v.Oid)
			epids = append(epids, int32(v.Epid))
		case _hisBusinessArt:
			articleIDs = append(articleIDs, v.Oid)
		case _hisBusinessArc:
			aids = append(aids, v.Oid)
		case _hisBusinessArtList:
			articleIDs = append(articleIDs, v.Cid)
		case _hisBusinessCheese:
			aids = append(aids, v.Oid) //用cid拿时长duration
			cheeseEpids = append(cheeseEpids, int32(v.Epid))
		case _hisBusinessLive:
			if v.Oid > 0 {
				roomIDs = append(roomIDs, v.Oid)
			}
		default:
			log.Warn("HistoryCursor unknown type(%s) msg(%+v)", v.Business, v)
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) error {
			if arcsReply, e := s.arcGRPC.Views(ctx, &arcapi.ViewsRequest{Aids: aids}); e != nil {
				log.Error("HistoryCursor s.arcGRPC.Views(%v) error(%v)", aids, e)
			} else {
				views = arcsReply.GetViews()
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			if isFavsReply, e := s.favGRPC.IsFavoreds(ctx, &favapi.IsFavoredsReq{Typ: int32(favmdl.TypeVideo), Mid: mid, Oids: aids}); e != nil {
				log.Error("HistoryCursor s.favGRPC.IsFavoreds(%v) mid:%d error(%v)", aids, mid, e)
			} else {
				isFavs = isFavsReply.GetFaveds()
			}
			return nil
		})
	}
	if len(epids) > 0 {
		group.Go(func(ctx context.Context) error {
			if epReply, e := s.epGRPC.Cards(ctx, &epapi.EpReq{Epids: epids}); e != nil {
				log.Error("HistoryCursor s.epGRPC.Cards(%v) error(%v)", epids, e)
			} else {
				epInfos = epReply.GetCards()
			}
			return nil
		})
	}
	if len(articleIDs) > 0 {
		group.Go(func(ctx context.Context) error {
			if reply, e := s.artGRPC.ArticleMetas(ctx, &artapi.ArticleMetasReq{Ids: articleIDs}); e != nil {
				log.Error("HistoryCursor s.art.ArticleMetas(%v) error(%v)", articleIDs, e)
			} else {
				article = reply.GetRes()
			}
			return nil
		})
	}
	if len(cheeseEpids) > 0 {
		group.Go(func(ctx context.Context) error {
			if reply, e := s.cheeseGRPC.Cards(ctx, &cheeseapi.EpisodeCardsReq{Ids: cheeseEpids}); e != nil {
				log.Error("HistoryCursor s.cheeseGRPC.Cards(%v) error(%v)", cheeseEpids, e)
			} else {
				cheeseCards = reply.GetCards()
			}
			return nil
		})
	}
	if len(roomIDs) > 0 {
		group.Go(func(ctx context.Context) error {
			if reply, e := s.roomGRPC.GetMultiple(ctx, &roomapi.RoomIDsReq{RoomIds: roomIDs, Attrs: []string{"show", "status", "area"}}); e != nil {
				log.Error("HistoryCursor s.roomGRPC.GetMultiple(%v) error(%v)", roomIDs, e)
			} else if reply != nil {
				roomInfos = reply.List
				for _, v := range reply.List {
					if v == nil {
						continue
					}
					liveUpMids = append(liveUpMids, v.Uid)
				}
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("HistoryCursor group.Wait error:%v", err)
	}
	var accInfos map[int64]*api.Info
	if len(liveUpMids) > 0 {
		accInfos = func() map[int64]*api.Info {
			accReply, err := s.accGRPC.Infos3(c, &api.MidsReq{Mids: liveUpMids})
			if err != nil {
				log.Error("HistoryCursor s.accGRPC.Infos3 mids(%v) error(%v)", liveUpMids, err)
				return nil
			}
			return accReply.GetInfos()
		}()
	}
	for _, his := range hisReply.Res {
		tmpInfo := &model.HisItem{
			ViewAt: his.Unix,
			Kid:    his.Kid,
		}
		tmpInfo.History.Oid = his.Oid
		tmpInfo.History.Business = his.Business
		tmpInfo.History.Cid = his.Cid
		tmpInfo.History.Dt = his.Dt
		switch his.Business {
		case _hisBusinessPgc:
			pgc, okPGC := epInfos[int32(his.Epid)]
			arc, okArc := views[his.Oid]
			if !okPGC || !okArc || pgc == nil {
				continue
			}
			tmpInfo.History.Epid = his.Epid
			if pgc.Season != nil {
				tmpInfo.Title = pgc.Season.Title
				tmpInfo.NewDesc = pgc.Season.NewEpShow
				tmpInfo.Total = pgc.Season.TotalCount
				tmpInfo.IsFinish = pgc.Season.IsFinish
			}
			tmpInfo.LongTitle = pgc.LongTitle
			tmpInfo.ShowTitle = pgc.ShowTitle
			tmpInfo.Cover = pgc.Cover
			tmpInfo.Badge = hisBadge[his.Stp]
			tmpInfo.Progress = his.Pro
			tmpInfo.URI = "https://www.bilibili.com/bangumi/play/ss" + strconv.FormatInt(his.Sid, 10)
			for _, p := range arc.GetPages() {
				if p.Cid == his.Cid {
					tmpInfo.Duration = p.Duration
					break
				}
			}
		case _hisBusinessArc:
			arc, ok := views[his.Oid]
			if !ok || arc == nil {
				continue
			}
			tmpInfo.History.Bvid = s.avToBv(arc.Aid)
			tmpInfo.Title = arc.Title
			tmpInfo.Cover = arc.Pic
			tmpInfo.AuthorMid = arc.Author.Mid
			tmpInfo.AuthorName = arc.Author.Name
			tmpInfo.AuthorFace = arc.Author.Face
			tmpInfo.TagName = arc.TypeName
			tmpInfo.Progress = his.Pro
			tmpInfo.Videos = arc.Videos
			if isFav, ok := isFavs[arc.Aid]; ok && isFav {
				tmpInfo.IsFav = _hisIsFav
			}
			for _, p := range arc.Pages {
				if p.Cid == his.Cid {
					tmpInfo.Duration = p.Duration
					tmpInfo.History.Cid = p.Cid
					tmpInfo.History.Page = p.Page
					tmpInfo.History.Part = p.Part
					break
				}
			}
			// 多p稿件
			if tmpInfo.Videos > 1 {
				tmpInfo.NewDesc = fmt.Sprintf("共%dP", tmpInfo.Videos)
				tmpInfo.ShowTitle = tmpInfo.History.Part
			}
		case _hisBusinessArt:
			art, ok := article[his.Oid]
			if !ok || art == nil {
				continue
			}
			tmpInfo.Title = art.Title
			tmpInfo.Covers = art.ImageURLs
			tmpInfo.AuthorMid = art.Author.Mid
			tmpInfo.AuthorName = art.Author.Name
			tmpInfo.AuthorFace = art.Author.Face
			tmpInfo.Badge = articleBadge(art)
		case _hisBusinessArtList:
			art, ok := article[his.Cid]
			if !ok || art == nil {
				continue
			}
			tmpInfo.Title = art.Title
			tmpInfo.Covers = art.ImageURLs
			tmpInfo.AuthorMid = art.Author.Mid
			tmpInfo.AuthorName = art.Author.Name
			tmpInfo.AuthorFace = art.Author.Face
			tmpInfo.Badge = articleBadge(art)
		case _hisBusinessCheese:
			cheese, okCheese := cheeseCards[int32(his.Epid)]
			arc, okArc := views[his.Oid]
			if !okCheese || !okArc || cheese == nil {
				continue
			}
			tmpInfo.Title = cheese.SeasonTitle
			tmpInfo.ShowTitle = cheese.Title
			tmpInfo.Cover = cheese.Cover
			tmpInfo.Progress = his.Pro
			tmpInfo.URI = "https://www.bilibili.com/cheese/play/ep" + strconv.FormatInt(int64(cheese.Id), 10)
			for _, p := range arc.GetPages() {
				if p.Cid == his.Cid {
					tmpInfo.Duration = p.Duration
					break
				}
			}
		case _hisBusinessLive:
			liveInfo, ok := roomInfos[his.Oid]
			if !ok || liveInfo == nil || liveInfo.Show == nil {
				continue
			}
			liveUser, ok := accInfos[liveInfo.Uid]
			if !ok || liveUser == nil {
				continue
			}
			tmpInfo.Title = liveInfo.Show.Title
			tmpInfo.Cover = liveInfo.Show.Cover
			if liveInfo.Show.Cover == "" {
				tmpInfo.Cover = liveInfo.Show.Keyframe
			}
			tmpInfo.AuthorMid = liveInfo.Uid
			tmpInfo.AuthorName = liveUser.Name
			tmpInfo.AuthorFace = liveUser.Face
			if liveInfo.Area != nil {
				tmpInfo.TagName = liveInfo.Area.AreaName
			}
			tmpInfo.Badge = "未开播"
			if liveInfo.Status != nil {
				if liveInfo.Status.LiveStatus == 1 { //1是直播中，0、2是未开播
					tmpInfo.LiveStatus = 1
					tmpInfo.Badge = "直播中"
				}
			}
			tmpInfo.URI = "https://live.bilibili.com/" + strconv.FormatInt(liveInfo.RoomId, 10)
		default:
			continue
		}
		data.List = append(data.List, tmpInfo)
	}
	if len(data.List) > 0 {
		data.Cursor = model.HisCursor{
			Max:      data.List[len(data.List)-1].Kid,
			ViewAt:   data.List[len(data.List)-1].ViewAt,
			Business: data.List[len(data.List)-1].History.Business,
			Ps:       ps,
		}
	}
	return data, nil
}

// WxHistoryCursor .
// nolint: gocognit
func (s *Service) WxHistoryCursor(c context.Context, mid, max, viewAt int64, business string, ps int32, platform string) (data *model.WxHisRes, err error) {
	businesses, ok := hisBusiness[_wxHistoryType]
	if !ok {
		err = ecode.RequestErr
		return
	}
	var (
		hisReply *hisapi.HistoryCursorReply
		aids     []int64
		epids    []int32
		views    map[int64]*arcapi.ViewReply
		epInfos  map[int32]*epapi.EpisodeCardsProto
	)
	arg := &hisapi.HistoryCursorReq{
		Mid:        mid,
		Max:        max,
		ViewAt:     viewAt,
		Ps:         ps,
		Business:   business,
		Businesses: businesses,
	}
	data = &model.WxHisRes{List: []*model.HisItem{}}
	if hisReply, err = s.hisGRPC.HistoryCursor(c, arg); err != nil {
		log.Error("HistoryCursor s.hisGRPC.HistoryCursor(%+v) error(%v)", arg, err)
		err = nil
		return
	}
	if len(hisReply.Res) == 0 {
		return
	}
	for _, v := range hisReply.Res {
		switch v.Business {
		case _hisBusinessPgc:
			if v.Oid > 0 {
				aids = append(aids, v.Oid)
			}
			if v.Epid > 0 {
				epids = append(epids, int32(v.Epid))
			}
		case _hisBusinessArc:
			if v.Oid > 0 {
				aids = append(aids, v.Oid)
			}
		default:
			log.Warn("HistoryCursor unknown type(%s) msg(%+v)", v.Business, v)
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) error {
			if arcsReply, err := s.arcGRPC.Views(ctx, &arcapi.ViewsRequest{Aids: aids}); err != nil {
				log.Error("HistoryCursor s.arcGRPC.Views(%v) error(%v)", aids, err)
			} else {
				views = arcsReply.Views
			}
			return nil
		})
	}
	if len(epids) > 0 {
		group.Go(func(ctx context.Context) error {
			if epReply, err := s.epGRPC.Cards(ctx, &epapi.EpReq{Epids: epids}); err != nil {
				log.Error("HistoryCursor s.epGRPC.Cards(%v) error(%v)", epids, err)
			} else {
				epInfos = epReply.Cards
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	for _, his := range hisReply.Res {
		tmpInfo := &model.HisItem{
			ViewAt: his.Unix,
			Kid:    his.Kid,
		}
		tmpInfo.History.Oid = his.Oid
		tmpInfo.History.Business = his.Business
		switch his.Business {
		case _hisBusinessPgc:
			pgc, okPGC := epInfos[int32(his.Epid)]
			arc, okArc := views[his.Oid]
			if !okPGC || !okArc {
				continue
			}
			if pgc.Season != nil {
				tmpInfo.Title = pgc.Season.Title
			}
			// business pgc use epid
			tmpInfo.History.Oid = his.Epid
			tmpInfo.ShowTitle = pgc.ShowTitle
			tmpInfo.Cover = pgc.Cover
			tmpInfo.Badge = hisBadge[his.Stp]
			tmpInfo.Progress = his.Pro
			for _, p := range arc.Pages {
				if p.Cid == his.Cid {
					tmpInfo.Duration = p.Duration
					break
				}
			}
		case _hisBusinessArc:
			arc, ok := views[his.Oid]
			if !ok {
				continue
			}
			tmpInfo.History.Bvid = s.avToBv(arc.Aid)
			tmpInfo.Title = arc.Title
			tmpInfo.Cover = arc.Pic
			tmpInfo.AuthorMid = arc.Author.Mid
			tmpInfo.AuthorName = arc.Author.Name
			tmpInfo.Progress = his.Pro
			tmpInfo.Videos = arc.Videos
			for _, p := range arc.Pages {
				if p.Cid == his.Cid {
					tmpInfo.Duration = p.Duration
					tmpInfo.History.Cid = p.Cid
					tmpInfo.History.Page = p.Page
					tmpInfo.History.Part = p.Part
					break
				}
			}
		default:
			continue
		}
		data.List = append(data.List, tmpInfo)
	}
	if len(data.List) > 0 {
		data.Cursor = model.HisCursor{
			Max:      data.List[len(data.List)-1].History.Oid,
			ViewAt:   data.List[len(data.List)-1].ViewAt,
			Business: data.List[len(data.List)-1].History.Business,
			Ps:       ps,
		}
		if platform == "wechat" {
			s.historyFilterBindOid(c, &data.List)
		}
	}
	return
}

func (s *Service) historyFilterBindOid(c context.Context, containOidSlice *[]*model.HisItem) {
	if len(*containOidSlice) == 0 {
		return
	}
	var oidList []int64
	for _, v := range *containOidSlice {
		if v != nil {
			oidList = append(oidList, v.History.Oid)
		}
	}

	bindOidList, err := s.dao.TagBind(c, oidList)
	k := 0
	for _, v := range *containOidSlice {
		if err != nil || bindOidList == nil || v == nil || !inIntSlice(bindOidList, v.History.Oid) {
			(*containOidSlice)[k] = v
			k++
		}
	}
	*containOidSlice = (*containOidSlice)[:k]
}

func articleBadge(art *artmdl.Meta) string {
	if art.Type == model.ArticleTypeNote {
		return "笔记"
	}
	return "专栏"
}
