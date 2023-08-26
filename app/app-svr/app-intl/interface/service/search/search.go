package search

import (
	"context"
	"fmt"
	"time"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-intl/interface/conf"
	accdao "go-gateway/app/app-svr/app-intl/interface/dao/account"
	arcdao "go-gateway/app/app-svr/app-intl/interface/dao/archive"
	artdao "go-gateway/app/app-svr/app-intl/interface/dao/article"
	bgmdao "go-gateway/app/app-svr/app-intl/interface/dao/bangumi"
	locdao "go-gateway/app/app-svr/app-intl/interface/dao/location"
	relationdao "go-gateway/app/app-svr/app-intl/interface/dao/relation"
	resdao "go-gateway/app/app-svr/app-intl/interface/dao/resource"
	srchdao "go-gateway/app/app-svr/app-intl/interface/dao/search"
	tagdao "go-gateway/app/app-svr/app-intl/interface/dao/tag"
	"go-gateway/app/app-svr/app-intl/interface/model"
	"go-gateway/app/app-svr/app-intl/interface/model/bangumi"
	"go-gateway/app/app-svr/app-intl/interface/model/search"
	tag "go-gateway/app/app-svr/app-intl/interface/model/tag/legacy"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	"go-gateway/app/app-svr/archive/service/api"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
)

var (
	_emptyResult = &search.Result{
		NavInfo: []*search.NavInfo{},
		Page:    0,
	}
)

// Service is search service
type Service struct {
	c       *conf.Config
	srchDao *srchdao.Dao
	accDao  *accdao.Dao
	arcDao  *arcdao.Dao
	artDao  *artdao.Dao
	// artDao     *artdao.Dao
	resDao      *resdao.Dao
	tagDao      *tagdao.Dao
	bgmDao      *bgmdao.Dao
	relationDao *relationdao.Dao
	// location
	locDao *locdao.Dao
	// config
	seasonNum          int
	movieNum           int
	seasonShowMore     int
	movieShowMore      int
	upUserNum          int
	uvLimit            int
	userNum            int
	userVideoLimit     int
	biliUserNum        int
	biliUserVideoLimit int
	iPadSearchBangumi  int
	iPadSearchFt       int
}

// New is search service initial func
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		srchDao: srchdao.New(c),
		accDao:  accdao.New(c),
		arcDao:  arcdao.New(c),
		artDao:  artdao.New(c),
		// artDao:             artdao.New(c),
		resDao:      resdao.New(c),
		tagDao:      tagdao.New(c),
		bgmDao:      bgmdao.New(c),
		relationDao: relationdao.New(c),
		// location
		locDao:             locdao.New(c),
		seasonNum:          c.Search.SeasonNum,
		movieNum:           c.Search.MovieNum,
		seasonShowMore:     c.Search.SeasonMore,
		movieShowMore:      c.Search.MovieMore,
		upUserNum:          c.Search.UpUserNum,
		uvLimit:            c.Search.UVLimit,
		userNum:            c.Search.UpUserNum,
		userVideoLimit:     c.Search.UVLimit,
		biliUserNum:        c.Search.BiliUserNum,
		biliUserVideoLimit: c.Search.BiliUserVideoLimit,
		iPadSearchBangumi:  c.Search.IPadSearchBangumi,
		iPadSearchFt:       c.Search.IPadSearchFt,
	}
	return
}

// Search get all type search data.
// nolint:gocognit,gomnd
func (s *Service) Search(c context.Context, mid int64, mobiApp, device, platform, buvid, keyword, duration, order, filtered, lang, fromSource, recommend string, plat int8, rid, highlight, build, pn, ps, isQuery int, now time.Time) (res *search.Result, err error) {
	var (
		aids                  []int64
		am                    map[int64]*api.ArcPlayer
		owners                []int64
		follows               map[int64]bool
		seasonIDs             []int64
		bangumis              map[string]*bangumi.Card
		seasonNum             int
		movieNum              int
		isNewOrder, newPlayer bool
		// 新订阅关系
		relationm map[int64]*relationgrpc.InterrelationReply
		sepReqs   []*pgcsearch.SeasonEpReq
		seasonEps map[int32]*pgcsearch.SearchCardProto
		medisas   map[int32]*pgcsearch.SearchMediaProto
		ip        = metadata.String(c, metadata.RemoteIP)
		zoneid    int64
	)
	seasonNum = s.seasonNum
	movieNum = s.movieNum
	if plat == model.PlatAndroidI && build > 2033000 {
		if rid != 0 || duration != "0" || order != "totalrank" {
			isNewOrder = true
		}
	}
	if (plat == model.PlatAndroidI && build > s.c.SearchBuildLimit.ArcWithPlayerAndroid) || (plat == model.PlatIPhoneI && build > s.c.SearchBuildLimit.ArcWithPlayerIOS) {
		newPlayer = true
	}
	if ipInfos, _ := s.locDao.Info(c, ip); ipInfos != nil {
		zoneid = ipInfos.ZoneId
	}
	all, code, err := s.srchDao.Search(c, mid, zoneid, mobiApp, device, platform, buvid, keyword, duration, order, filtered, fromSource, recommend, plat, seasonNum, movieNum, s.upUserNum, s.uvLimit, s.userNum, s.userVideoLimit, s.biliUserNum, s.biliUserVideoLimit, rid, highlight, build, pn, ps, isQuery, now, ip)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if code == model.ForbidCode || code == model.NoResultCode {
		res = _emptyResult
		return
	}
	res = &search.Result{}
	res.Trackid = all.Trackid
	res.Page = all.Page
	res.Array = all.FlowPlaceholder
	res.Attribute = all.Attribute
	res.NavInfo = s.convertNav(all, plat, build, lang)
	if len(all.FlowResult) != 0 {
		var item []*search.Item
		for _, v := range all.FlowResult {
			switch v.Type {
			case search.TypeUser, search.TypeBiliUser:
				owners = append(owners, v.User.Mid)
				for _, vr := range v.User.Res {
					aids = append(aids, vr.Aid)
				}
			case search.TypeVideo:
				aids = append(aids, v.Video.ID)
			case search.TypeMediaBangumi, search.TypeMediaFt:
				seasonIDs = append(seasonIDs, v.Media.SeasonID)
				if v.Media.Canplay() {
					sepReqs = append(sepReqs, v.Media.BuildPgcReq())
				}
			}
		}
		g, ctx := errgroup.WithContext(c)
		if len(owners) != 0 {
			if mid > 0 {
				g.Go(func() error {
					follows = s.accDao.Relations3(ctx, owners, mid)
					return nil
				})
				g.Go(func() error {
					relationm, err = s.relationDao.Interrelations(ctx, mid, owners)
					return nil
				})
			}
		}
		if len(aids) != 0 {
			if newPlayer {
				var aidPls []*api.PlayAv
				for _, aVal := range aids {
					aidPls = append(aidPls, &api.PlayAv{Aid: aVal})
				}
				g.Go(func() (err error) {
					if am, err = s.arcDao.ArcsPlayer(ctx, aidPls); err != nil {
						log.Error("%+v", err)
					}
					return nil
				})
			} else {
				g.Go(func() (err error) {
					if am, err = s.arcDao.Arcs(ctx, aids); err != nil {
						log.Error("%+v", err)
					}
					return nil
				})
			}
		}
		if len(seasonIDs) != 0 {
			g.Go(func() (err error) {
				if bangumis, err = s.bgmDao.Card(ctx, mid, seasonIDs); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				return
			})
		}
		if len(sepReqs) != 0 {
			g.Go(func() (err error) {
				batchArg, _ := arcmid.FromContext(ctx)
				if seasonEps, medisas, err = s.bgmDao.SearchPGCCards(ctx, sepReqs, keyword, mobiApp, device, platform, mid, int(batchArg.Fnver), int(batchArg.Fnval), int(batchArg.Qn), int(batchArg.Fourk), build, true); err != nil {
					log.Error("bgmDao SearchPGCCards %v", err)
					err = nil
				}
				return
			})
		}
		if err = g.Wait(); err != nil {
			log.Error("%+v", err)
			return
		}
		if all.SuggestKeyword != "" && pn == 1 {
			i := &search.Item{Title: all.SuggestKeyword, Goto: model.GotoSuggestKeyWord, SugKeyWordType: 1}
			item = append(item, i)
		} else if all.CrrQuery != "" && pn == 1 && !isNewOrder {
			i := &search.Item{Title: fmt.Sprintf("已匹配%q的搜索结果", all.CrrQuery), Goto: model.GotoSuggestKeyWord, SugKeyWordType: 2}
			item = append(item, i)
		}
		for _, v := range all.FlowResult {
			i := &search.Item{TrackID: v.TrackID, LinkType: v.LinkType, Position: v.Position}
			switch v.Type {
			case search.TypeVideo:
				i.FromVideo(v.Video, am[v.Video.ID], false)
			case search.TypeMediaBangumi:
				i.FromMediaPgcCard(v.Media, "", model.GotoBangumi, bangumis, seasonEps, medisas, s.c.Cfg.PgcSearchCard, false) // flow result, not ipad
			case search.TypeMediaFt:
				i.FromMediaPgcCard(v.Media, "", model.GotoMovie, bangumis, seasonEps, medisas, s.c.Cfg.PgcSearchCard, false)
			case search.TypeSpecial:
				i.FromOperate(v.Operate, model.GotoSpecial)
			case search.TypeBanner:
				i.FromOperate(v.Operate, model.GotoBanner)
			case search.TypeUser:
				if follows[v.User.Mid] {
					i.Attentions = 1
				}
				i.Relation = cardmdl.RelationChange(v.User.Mid, relationm)
				i.FromUser(v.User, am)
			case search.TypeBiliUser:
				if follows[v.User.Mid] {
					i.Attentions = 1
				}
				i.Relation = cardmdl.RelationChange(v.User.Mid, relationm)
				i.FromUpUser(v.User, am)
			case search.TypeSpecialS:
				i.FromOperate(v.Operate, model.GotoSpecialS)
			case search.TypeQuery:
				i.Title = v.TypeName
				i.FromQuery(v.Query)
			case search.TypeConverge:
				var (
					avids, artids []int64
					avm           map[int64]*api.Arc
					artm          map[int64]*article.Meta
				)
				for _, c := range v.Operate.ContentList {
					switch c.Type {
					case 0:
						avids = append(avids, c.ID)
					case 2:
						artids = append(artids, c.ID)
					}
				}
				g, ctx := errgroup.WithContext(c)
				if len(aids) != 0 {
					g.Go(func() (err error) {
						if avm, err = s.arcDao.Archives(ctx, avids); err != nil {
							log.Error("%+v", err)
							err = nil
						}
						return
					})
				}
				if len(artids) != 0 {
					g.Go(func() (err error) {
						if artm, err = s.artDao.Articles(ctx, artids); err != nil {
							log.Error("%+v", err)
							err = nil
						}
						return
					})
				}
				if err = g.Wait(); err != nil {
					log.Error("%+v", err)
					continue
				}
				i.FromConverge(v.Operate, avm, artm)
			case search.TypeTwitter:
				i.FromTwitter(v.Twitter)
			}
			if i.Goto != "" {
				item = append(item, i)
			}
		}
		res.Item = item
		if all.EggInfo != nil {
			res.EasterEgg = &search.EasterEgg{ID: all.EggInfo.Source, ShowCount: all.EggInfo.ShowCount}
		}
		return
	}
	var items []*search.Item
	if all.SuggestKeyword != "" && pn == 1 {
		res.Items.SuggestKeyWord = &search.Item{Title: all.SuggestKeyword, Goto: model.GotoSuggestKeyWord}
	}
	// archive
	for _, v := range all.Result.Video {
		aids = append(aids, v.ID)
	}
	if duration == "0" && order == "totalrank" && rid == 0 {
		for _, v := range all.Result.Movie {
			if v.Type == "movie" {
				aids = append(aids, v.Aid)
			}
		}
	}
	if pn == 1 {
		for _, v := range all.Result.User {
			for _, vr := range v.Res {
				aids = append(aids, vr.Aid)
			}
		}
		for _, v := range all.Result.BiliUser {
			for _, vr := range v.Res {
				aids = append(aids, vr.Aid)
			}
			owners = append(owners, v.Mid)
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(owners) != 0 {
		if mid > 0 {
			g.Go(func() error {
				follows = s.accDao.Relations3(ctx, owners, mid)
				return nil
			})
		}
	}
	if len(aids) != 0 {
		if newPlayer {
			var aidPlays []*api.PlayAv
			for _, aVal := range aids {
				aidPlays = append(aidPlays, &api.PlayAv{Aid: aVal})
			}
			g.Go(func() (err error) {
				if am, err = s.arcDao.ArcsPlayer(ctx, aidPlays); err != nil {
					log.Error("%+v", err)
				}
				return nil
			})
		} else {
			g.Go(func() (err error) {
				if am, err = s.arcDao.Arcs(ctx, aids); err != nil {
					log.Error("%+v", err)
				}
				return nil
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	if pn == 1 {
		// upper + user
		tmp := all.Result.BiliUser
		items = make([]*search.Item, 0, len(tmp)+len(all.Result.User))
		for _, v := range all.Result.User {
			si := &search.Item{}
			si.FromUser(v, am)
			if follows[v.Mid] {
				si.Attentions = 1
			}
			items = append(items, si)
		}
		if len(items) == 0 {
			for _, v := range tmp {
				si := &search.Item{}
				si.FromUpUser(v, am)
				if follows[v.Mid] {
					si.Attentions = 1
				}
				items = append(items, si)
			}
		}
		res.Items.Upper = items
	}
	items = make([]*search.Item, 0, len(all.Result.Video))
	for _, v := range all.Result.Video {
		si := &search.Item{}
		si.FromVideo(v, am[v.ID], false)
		items = append(items, si)
	}
	res.Items.Archive = items
	return
}

// SearchByType is tag bangumi movie upuser video search
func (s *Service) SearchByType(c context.Context, mid int64, mobiApp, device, platform, buvid, sType, keyword, filtered, order string, plat int8, build, highlight, categoryID, userType, orderSort, pn, ps, fnver, fnval, qn, fourk int, now time.Time) (res *search.TypeSearch, err error) {
	var (
		zoneid int64
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	if ipInfos, _ := s.locDao.Info(c, ip); ipInfos != nil {
		zoneid = ipInfos.ZoneId
	}
	switch sType {
	case "upper":
		if res, err = s.upper(c, mid, zoneid, keyword, mobiApp, device, platform, buvid, filtered, order, s.biliUserVideoLimit, highlight, build, userType, orderSort, pn, ps, now); err != nil {
			return
		}
	case "article":
		if res, err = s.article(c, mid, zoneid, highlight, keyword, mobiApp, device, platform, buvid, filtered, order, sType, plat, categoryID, build, pn, ps, now); err != nil {
			return
		}
	case "season2":
		if res, err = s.srchDao.Season2(c, mid, zoneid, keyword, mobiApp, device, platform, buvid, highlight, build, pn, ps, fnver, fnval, qn, fourk); err != nil {
			return
		}
	case "movie2":
		if res, err = s.srchDao.MovieByType2(c, mid, zoneid, keyword, mobiApp, device, platform, buvid, highlight, build, pn, ps, fnver, fnval, qn, fourk); err != nil {
			return
		}
	case "tag":
		if res, err = s.channel(c, mid, keyword, mobiApp, platform, buvid, device, order, sType, build, pn, ps, highlight); err != nil {
			return
		}
	}
	if res == nil {
		res = &search.TypeSearch{Items: []*search.Item{}}
	}
	return
}

// Suggest3 for search suggest
func (s *Service) Suggest3(c context.Context, mid int64, platform, buvid, keyword string, build, highlight int, mobiApp string, now time.Time) (res *search.SuggestionResult3) {
	var (
		suggest *search.Suggest3
		err     error
		aids    []int64
		am      map[int64]*api.Arc
	)
	res = &search.SuggestionResult3{}
	if suggest, err = s.srchDao.Suggest3(c, mid, platform, buvid, keyword, build, highlight, mobiApp, now); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, v := range suggest.Result {
		if v.TermType == search.SuggestionJump {
			if v.SubType == search.SuggestionAV {
				aids = append(aids, v.Ref)
			}
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(aids) != 0 {
		g.Go(func() (err error) {
			if am, err = s.arcDao.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, v := range suggest.Result {
		si := &search.Item{}
		si.FromSuggest3(v, am)
		res.List = append(res.List, si)
	}
	res.TrackID = suggest.TrackID
	return
}

// convertNav deal with old search pageinfo to new.
func (s *Service) convertNav(all *search.Search, _ int8, _ int, lang string) (nis []*search.NavInfo) {
	const (
		_showHide = 0
	)
	var (
		season  = "番剧"
		upper   = "用户"
		movie   = "影视"
		article = "专栏"
	)
	if lang == model.Hant {
		season = "番劇"
		upper = "UP主"
		movie = "影視"
		article = "專欄"
	}
	nis = make([]*search.NavInfo, 0, 4)
	// season
	// media season
	if all.PageInfo.MediaBangumi != nil {
		var nav = &search.NavInfo{
			Name:  season,
			Total: all.PageInfo.MediaBangumi.NumResults,
			Pages: all.PageInfo.MediaBangumi.Pages,
			Type:  7,
		}
		if all.PageInfo.MediaBangumi.NumResults > s.seasonNum {
			nav.Show = s.seasonShowMore
		} else {
			nav.Show = _showHide
		}
		nis = append(nis, nav)
	}
	// upper
	if all.PageInfo.BiliUser != nil {
		var nav = &search.NavInfo{
			Name:  upper,
			Total: all.PageInfo.BiliUser.NumResults,
			Pages: all.PageInfo.BiliUser.Pages,
			Type:  2,
		}
		nis = append(nis, nav)
	}
	// media movie
	if all.PageInfo.MediaFt != nil {
		var nav = &search.NavInfo{
			Name:  movie,
			Total: all.PageInfo.MediaFt.NumResults,
			Pages: all.PageInfo.MediaFt.Pages,
			Type:  8,
		}
		if all.PageInfo.MediaFt.NumResults > s.movieNum {
			nav.Show = s.movieShowMore
		} else {
			nav.Show = _showHide
		}
		nis = append(nis, nav)
	}
	if all.PageInfo.Article != nil {
		var nav = &search.NavInfo{
			Name:  article,
			Total: all.PageInfo.Article.NumResults,
			Pages: all.PageInfo.Article.Pages,
			Type:  6,
		}
		nis = append(nis, nav)
	}
	return
}

// upper search for upper
func (s *Service) upper(c context.Context, mid, zoneid int64, keyword, mobiApp, device, platform, buvid, filtered, order string, biliUserVL, highlight, build, userType, orderSort, pn, ps int, now time.Time) (res *search.TypeSearch, err error) {
	var (
		owners  []int64
		follows map[int64]bool
		// 新订阅关系
		relationm map[int64]*relationgrpc.InterrelationReply
	)
	if res, err = s.srchDao.Upper(c, mid, zoneid, keyword, mobiApp, device, platform, buvid, filtered, order, biliUserVL, highlight, build, userType, orderSort, pn, ps, now); err != nil {
		return
	}
	if res == nil || len(res.Items) == 0 {
		return
	}
	owners = make([]int64, 0, len(res.Items))
	for _, item := range res.Items {
		owners = append(owners, item.Mid)
	}
	if len(owners) != 0 {
		g, ctx := errgroup.WithContext(c)
		if mid > 0 {
			g.Go(func() error {
				follows = s.accDao.Relations3(ctx, owners, mid)
				return nil
			})
			g.Go(func() error {
				if relationm, err = s.relationDao.Interrelations(ctx, mid, owners); err != nil {
					log.Error("%v", err)
				}
				return nil
			})
		}
		if err = g.Wait(); err != nil {
			log.Error("%+v", err)
			return
		}
		for _, item := range res.Items {
			if follows[item.Mid] {
				item.Attentions = 1
			}
			item.Relation = cardmdl.RelationChange(item.Mid, relationm)
		}
	}
	return
}

// article search for article
func (s *Service) article(c context.Context, mid, zoneid int64, highlight int, keyword, mobiApp, device, platform, buvid, filtered, order, sType string, plat int8, categoryID, build, pn, ps int, now time.Time) (res *search.TypeSearch, err error) {
	if res, err = s.srchDao.ArticleByType(c, mid, zoneid, keyword, mobiApp, device, platform, buvid, filtered, order, sType, plat, categoryID, build, highlight, pn, ps, now); err != nil {
		log.Error("%+v", err)
		return
	}
	if res == nil || len(res.Items) == 0 {
		return
	}
	var mids []int64
	for _, v := range res.Items {
		mids = append(mids, v.Mid)
	}
	var infom map[int64]*account.Info
	if infom, err = s.accDao.Infos3(c, mids); err != nil {
		log.Error("%+v", err)
		err = nil
		return
	}
	for _, item := range res.Items {
		if info, ok := infom[item.Mid]; ok {
			item.Name = info.Name
		}
	}
	return
}

// channel search for channel
func (s *Service) channel(c context.Context, mid int64, keyword, mobiApp, platform, buvid, device, order, sType string, build, pn, ps, highlight int) (res *search.TypeSearch, err error) {
	var (
		g          *errgroup.Group
		ctx        context.Context
		tags       []int64
		tagMyInfos []*tag.Tag
	)
	if res, err = s.srchDao.Channel(c, mid, keyword, mobiApp, platform, buvid, device, order, sType, build, pn, ps, highlight); err != nil {
		return
	}
	if res == nil || len(res.Items) == 0 {
		return
	}
	tags = make([]int64, 0, len(res.Items))
	for _, item := range res.Items {
		tags = append(tags, item.ID)
	}
	if len(tags) != 0 {
		g, ctx = errgroup.WithContext(c)
		if mid > 0 {
			g.Go(func() error {
				tagMyInfos, _ = s.tagDao.TagInfos(ctx, tags, mid)
				return nil
			})
		}
		if err = g.Wait(); err != nil {
			log.Error("%+v", err)
			return
		}
		for _, item := range res.Items {
			for _, myInfo := range tagMyInfos {
				if myInfo != nil && myInfo.TagID == item.ID {
					item.IsAttention = myInfo.IsAtten
					break
				}
			}
		}
	}
	return
}
