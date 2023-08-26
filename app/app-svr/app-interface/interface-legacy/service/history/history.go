package history

import (
	"context"
	"hash/crc32"
	"strconv"
	"sync"
	"time"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup"

	arcApi "go-gateway/app/app-svr/archive/service/api"

	cardm "go-gateway/app/app-svr/app-card/interface/model"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/history"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	artm "go-gateway/app/app-svr/app-interface/interface-legacy/model/article"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/history"
	livemdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/live"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	cheeseEp "git.bilibili.co/bapis/bapis-go/cheese/service/season/episode"
	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

const (
	_tpOld            = -1
	_tpOffline        = 0
	_tpArchive        = 3
	_tpPGC            = 4
	_tpArticle        = 5
	_tpLive           = 6
	_tpCorpus         = 7
	_tpCheese         = 10
	_tpCheeseIPad     = 12
	_androidBroadcast = 5305000
	_allStr           = "all"
	_arcStr           = "archive"
	_artStr           = "article"
	_corpusStr        = "article-list"
	_liveStr          = "live"
	_pgcStr           = "pgc"
	_cheeseStr        = "cheese"
	_cheeseIPadStr    = "cheese-ipad"
	_mallGoods        = "mall-goods"
	_mallShow         = "mall-show"
	_goods            = "goods"
	_game             = "game"
	_show             = "show"
	_homePage         = "homepage"
	_hdExpRtime       = "rtime"
	_hdExpLtime       = "ltime"
	_hisCursorV2      = "/bilibili.app.interface.v1.History/CursorV2"
	_hisCursor        = "/bilibili.app.interface.v1.History/Cursor"
	_hisSearch        = "/bilibili.app.interface.v1.History/Search"
	_hisLiveList      = "/x/v2/history/liveList"
	_liveWifi         = "1"
	_liveMobile       = "2"
	_liveOther        = "3"
)

var (
	gotoDesc = map[int32]string{
		_tpOld:        model.GotoAv,
		_tpOffline:    model.GotoAv,
		_tpArchive:    model.GotoAv,
		_tpPGC:        model.GotoPGC,
		_tpArticle:    model.GotoArticle,
		_tpLive:       model.GotoLive,
		_tpCorpus:     model.GotoArticle,
		_tpCheese:     model.GotoCheese,
		_tpCheeseIPad: model.GotoCheese,
	}
	badge = map[int32]string{
		1: "番剧",
		2: "电影",
		3: "纪录片",
		4: "国创",
		5: "电视剧",
		7: "综艺",
	}
	busTab = []*history.BusTab{
		{
			Business: "all",
			Name:     "全部",
		},
		{
			Business: "archive",
			Name:     "视频",
		},
		{
			Business: "live",
			Name:     "直播",
		},
		{
			Business: "article",
			Name:     "专栏",
		},
	}
	busTabOversea = []*history.BusTab{
		{
			Business: "all",
			Name:     "全部",
		},
		{
			Business: "archive",
			Name:     "视频",
		},
		{
			Business: "article",
			Name:     "专栏",
		},
	}
	businessMap = map[int32]string{
		_tpOld:        _arcStr,
		_tpOffline:    _arcStr,
		_tpArchive:    _arcStr,
		_tpPGC:        _pgcStr,
		_tpArticle:    _artStr,
		_tpLive:       _liveStr,
		_tpCorpus:     _corpusStr,
		_tpCheese:     _cheeseStr,
		_tpCheeseIPad: _cheeseIPadStr,
	}
	busTabMap = map[string][]string{
		"all":     {_arcStr, _pgcStr, _artStr, _liveStr, _corpusStr},
		"archive": {_arcStr, _pgcStr},
		"article": {_artStr, _corpusStr},
		"live":    {_liveStr},
		"goods":   {_mallGoods},
		"show":    {_mallShow},
		"game":    {_game},
	}
	busTabNewOverseaMap = map[string][]string{
		"all":     {_arcStr, _pgcStr, _artStr, _corpusStr, _liveStr},
		"live":    {_liveStr},
		"archive": {_arcStr, _pgcStr},
		"article": {_artStr, _corpusStr},
	}
	busTabOverseaMap = map[string][]string{
		"all":     {_arcStr, _pgcStr, _artStr, _corpusStr},
		"archive": {_arcStr, _pgcStr},
		"article": {_artStr, _corpusStr},
	}
)

// List history list
func (s *Service) List(c context.Context, mid, build int64, pn, ps int32, mobiApp string, plat int8) (data []*history.ListRes, err error) {
	res, err := s.historyDao.History(c, mid, pn, ps)
	//nolint:gomnd
	if len(res) > 50 {
		log.Warn("history lens(%d) mid(%d) pn(%d) ps(%d)", len(res), mid, pn, ps)
	}
	if err != nil {
		log.Error("%+v ", err)
		return
	}
	if len(res) == 0 {
		data = []*history.ListRes{}
		return
	}
	data = s.TogetherHistory(c, res, mid, build, plat, mobiApp, false, nil)
	return
}

// Live get live for history
func (s *Service) Live(c context.Context, roomIDs []int64) (res []*livemdl.RoomInfo, err error) {
	live, err := s.liveDao.GetMultiple(c, roomIDs)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if len(live) == 0 {
		res = []*livemdl.RoomInfo{}
		return
	}
	for _, lv := range live {
		item := &livemdl.RoomInfo{
			RoomID: lv.RoomId,
			URI:    model.FillURI("live", strconv.FormatInt(lv.RoomId, 10), model.LiveHandler(lv)),
		}
		if lv.Status != nil && lv.Status.LiveStatus == 1 {
			item.Status = int(lv.Status.LiveStatus)
		}
		res = append(res, item)
	}
	return
}

// LiveList get live list for history
func (s *Service) LiveList(c context.Context, param *history.HisParam, plat int8) (data []*history.ListRes, err error) {
	res, err := s.historyDao.HistoryByTP(c, param.Mid, param.Pn, param.Ps, businessMap[_tpLive])
	if err != nil {
		log.Error("%+v ", err)
		return
	}
	if len(res) == 0 {
		data = []*history.ListRes{}
		return
	}
	liveParam := &history.LiveParam{
		Uid:      param.Mid,
		Platform: param.Platform,
		ReqBiz:   _hisLiveList,
		Build:    param.Build,
	}
	data = s.TogetherHistory(c, res, param.Mid, param.Build, plat, param.MobiApp, false, liveParam)
	return
}

// Cursor for history
func (s *Service) Cursor(c context.Context, param *history.HisParam, curPs int32, plat int8, needPlayUrl bool, liveParam *history.LiveParam) (data *history.ListCursor, hasMore bool, err error) {
	data = &history.ListCursor{
		List: []*history.ListRes{},
		Tab:  busTab,
	}
	hisBusTabMap := busTabMap
	if model.IsOverseas(plat) {
		// 新版繁体版出直播
		if needShowLiveContent(c) {
			hisBusTabMap = busTabNewOverseaMap
		} else {
			data.Tab = busTabOversea
			hisBusTabMap = busTabOverseaMap
		}
	}
	businesses, ok := hisBusTabMap[param.Business]
	if !ok {
		err = ecode.RequestErr
		log.Error("日志告警 历史记录非法business: %s", param.Business)
		return
	}
	if s.cheeseDao.HasCheese(plat, int(param.Build), false) && (param.Business == "all" || param.Business == "archive") {
		businesses = append(businesses, _cheeseStr)
	}
	if ((plat == model.PlatIPad && param.Build > int64(s.c.BuildLimit.IPadCheese)) || plat == model.PlatAndroidHD || (plat == model.PlatIpadHD && param.Build > int64(s.c.BuildLimit.IPadHDCheese))) &&
		(param.Business == "all" || param.Business == "archive") {
		businesses = append(businesses, _cheeseIPadStr)
	}
	var paramMaxBus string
	if _, ok := businessMap[param.MaxTP]; ok {
		paramMaxBus = businessMap[param.MaxTP]
	}
	res, err := s.historyDao.Cursor(c, param.Mid, param.Max, curPs, paramMaxBus, businesses, param.Buvid)
	if err != nil {
		log.Error("s.historyDao.Cursor error(%+v) ", err)
		return
	}
	if len(res) == 0 {
		return
	}
	if len(res) >= int(curPs) {
		hasMore = true
	}
	data.List = s.TogetherHistory(c, res, param.Mid, param.Build, plat, param.MobiApp, needPlayUrl, liveParam)
	if len(data.List) >= int(param.Ps) {
		data.List = data.List[:param.Ps]
	}
	if len(data.List) > 0 {
		data.Cursor = &history.Cursor{
			Max:   data.List[len(data.List)-1].ViewAt,
			MaxTP: data.List[len(data.List)-1].History.Tp,
			Ps:    param.Ps,
		}
	}
	return
}

// TogetherHistory always return 0~50
//
//nolint:gocognit
func (s *Service) TogetherHistory(c context.Context, res []*hisApi.ModelResource, mid, build int64, plat int8, mobiApp string, needPlayUrl bool, liveParam *history.LiveParam) (data []*history.ListRes) {
	var (
		aids, articleIDs, roomIDs, upIDs []int64
		cheeseEpids, epids               []int32
		archive                          map[int64]*arcApi.ViewReply
		pgcInfo                          map[int32]*episodegrpc.EpisodeCardsProto
		article                          map[int64]*article.Meta
		live                             map[int64]*livexroom.Infos
		cheeseCards                      map[int32]*cheeseEp.EpisodeCard
		interrelations                   map[int64]*relationgrpc.InterrelationReply
		accCards                         map[int64]*accountgrpc.Card
		apm                              map[int64]*arcApi.ArcPlayer
		mutex                            sync.Mutex
		playAvs                          []*arcApi.PlayAv
		livep                            map[int64]*livexroom.LivePlayUrlData
	)
	cheeseType := _tpCheese
	if model.IsPad(plat) {
		cheeseType = _tpCheeseIPad
	}
	for _, his := range res {
		switch his.Tp {
		case _tpOld, _tpOffline, _tpArchive:
			playAvs = append(playAvs, &arcApi.PlayAv{Aid: his.Oid})
			aids = append(aids, his.Oid)
		case _tpPGC:
			aids = append(aids, his.Oid) //用cid拿时长duration
			epids = append(epids, int32(his.Epid))
		case _tpArticle:
			articleIDs = append(articleIDs, his.Oid)
		case _tpLive:
			roomIDs = append(roomIDs, his.Oid)
		case _tpCorpus:
			articleIDs = append(articleIDs, his.Cid)
		case int32(cheeseType):
			aids = append(aids, his.Oid) //用cid拿时长duration
			cheeseEpids = append(cheeseEpids, int32(his.Epid))
		default:
			log.Warn("unknown history type(%d) msg(%+v)", his.Tp, his)
		}
	}
	eg, ctx := errgroup.WithContext(c)
	if len(aids) > 0 {
		eg.Go(func() (err error) {
			archive, err = s.historyDao.Archive(ctx, aids)
			if err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			if len(archive) > 0 {
				for _, arc := range archive {
					if len(arc.Pages) > 1 { //多P视频不展示up名字也不展示关注按钮
						continue
					}
					mutex.Lock()
					upIDs = append(upIDs, arc.Author.Mid)
					mutex.Unlock()
				}
			}
			return
		})
	}
	if len(playAvs) > 0 && needPlayUrl && feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.HistoryPlayurl, nil) {
		eg.Go(func() (err error) {
			apm, err = s.arcDao.ArcsPlayer(ctx, playAvs, false)
			if err != nil {
				log.Error("s.historyDao.ArcsWithPlayUrl err(+%v)", err)
				return nil
			}
			return nil
		})
	}
	if len(epids) > 0 {
		eg.Go(func() (err error) {
			pgcInfo, err = s.bangumiDao.EpCards(ctx, epids)
			if err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(articleIDs) > 0 {
		eg.Go(func() (err error) {
			article, err = s.artDao.Articles(ctx, articleIDs)
			if err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			if len(article) > 0 {
				for _, art := range article {
					if art.Author == nil {
						continue
					}
					mutex.Lock()
					upIDs = append(upIDs, art.Author.Mid)
					mutex.Unlock()
				}
			}
			return
		})
	}
	if len(roomIDs) > 0 {
		eg.Go(func() (err error) {
			live, livep, err = s.liveDao.GetMultipleWithPlayUrl(ctx, roomIDs, liveParam)
			if err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			if len(live) > 0 {
				for _, lv := range live {
					mutex.Lock()
					upIDs = append(upIDs, lv.Uid)
					mutex.Unlock()
				}
			}
			return
		})
	}
	if len(cheeseEpids) > 0 {
		eg.Go(func() (err error) {
			cheeseCards, err = s.cheeseDao.EpCards(ctx, cheeseEpids)
			if err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if egErr := eg.Wait(); egErr != nil {
		log.Error("eg1 err(%+v)", egErr)
	}
	eg, ctx = errgroup.WithContext(c)
	if len(upIDs) > 0 {
		eg.Go(func() (err error) {
			if interrelations, err = s.relDao.Interrelations(ctx, mid, upIDs); err != nil {
				log.Error("s.relDao.Interrelations err(%+v)", err)
			}
			return nil
		})
		eg.Go(func() (err error) {
			if accCards, err = s.accDao.Cards3(ctx, upIDs); err != nil {
				log.Error("s.accDao.Cards3 err(%+v)", err)
			}
			return nil
		})
	}
	//nolint:errcheck
	eg.Wait()
	for _, his := range res {
		// 旧版繁体版不出直播
		if his.Tp == _tpLive && model.IsOverseas(plat) && !needShowLiveContent(c) {
			continue
		}
		tmpInfo := &history.ListRes{
			Goto:   gotoDesc[his.Tp],
			ViewAt: his.Unix,
		}
		tmpInfo.History.Oid = his.Oid
		tmpInfo.History.Tp = his.Tp
		tmpInfo.History.Business = his.Business
		tmpInfo.History.Kid = his.Kid
		tmpInfo.History.Dt = his.Dt
		tmpInfo.History.WatchedBuvid = his.Buvid
		switch his.Tp {
		case _tpOld, _tpOffline, _tpArchive:
			arc, ok := archive[his.Oid]
			if !ok {
				continue
			}
			tmpInfo.State = int64(arc.State)
			tmpInfo.Title = arc.Title
			tmpInfo.Cover = arc.Pic
			tmpInfo.Mid = arc.Author.Mid
			tmpInfo.Name = arc.Author.Name
			tmpInfo.Videos = arc.Videos
			tmpInfo.View = int64(arc.Stat.View)
			if arc.AttrVal(arcApi.AttrBitSteinsGate) == arcApi.AttrYes { // 互动视频进度不展示
				tmpInfo.Progress = 0
				tmpInfo.History.Page = 1
				tmpInfo.History.Part = ""
				tmpInfo.Duration = 0
			} else {
				tmpInfo.Progress = his.Pro
				for _, p := range arc.Pages {
					if p.Cid == his.Cid {
						tmpInfo.Duration = p.Duration
						tmpInfo.History.Cid = p.Cid
						tmpInfo.History.Page = p.Page
						tmpInfo.History.Part = p.Part
						break
					}
				}
			}
			if arc.AttrValV2(arcApi.AttrBitV2Pay) == arcApi.AttrYes && arc.Rights.ArcPayFreeWatch == 0 { //付费合集添加角标
				tmpInfo.Badge = "付费"
			}
			tmpInfo.URI = model.FillURI(tmpInfo.Goto, strconv.FormatInt(his.Oid, 10), model.AvHandler(arc.Arc))
			ap, ok := apm[his.Oid]
			if ok && ap != nil {
				playInfo, ok := ap.PlayerInfo[ap.DefaultPlayerCid]
				if ok && playInfo != nil && ap.DefaultPlayerCid == his.Cid { // 621之前DefaultPlayerCid返回的第1p,只有his.cid也是第一p才出秒开
					tmpInfo.URI = model.FillURI(tmpInfo.Goto, strconv.FormatInt(his.Oid, 10), model.AvPlayHandlerGRPC(ap.Arc, playInfo))
				}
			}
		case _tpPGC:
			pgc, okPGC := pgcInfo[int32(his.Epid)]
			if !okPGC || pgc.Season == nil {
				continue
			}
			tmpInfo.State = int64(pgc.IsDelete)
			tmpInfo.Title = pgc.Season.Title
			tmpInfo.ShowTitle = pgc.ShowTitle
			tmpInfo.Cover = pgc.Cover
			tmpInfo.Badge = badge[his.Stp]
			tmpInfo.Progress = his.Pro
			tmpInfo.URI = pgc.Url // 默认用pgc返回的url
			if pgc.Url == "" {
				tmpInfo.URI = model.FillURI(tmpInfo.Goto, strconv.FormatInt(his.Sid, 10), nil)
			}
			tmpInfo.Duration = int64(pgc.Duration / 1000)
		case _tpArticle:
			art, ok := article[his.Oid]
			if !ok {
				continue
			}
			tmpInfo.Title = art.Title
			tmpInfo.Covers = art.ImageURLs
			if art.Author != nil {
				tmpInfo.Mid = art.Author.Mid
				tmpInfo.Name = art.Author.Name
				tmpInfo.Relation = cardm.RelationChange(art.Author.Mid, interrelations)
				follow := cardm.RelationOldChange(art.Author.Mid, interrelations)
				if follow == 0 && art.Author.Mid != mid && mid > 0 {
					tmpInfo.DisAtten = 1
				}
			}
			articleInfo := artm.GetArticleInfo(ctx, int64(art.Type), his.Oid, art.CoverAvid)
			tmpInfo.Badge = articleInfo.Badge
			tmpInfo.URI = articleInfo.Uri
		case _tpLive:
			lv, ok := live[his.Oid]
			if !ok || lv.Show == nil {
				continue
			}
			accCard, accok := accCards[lv.Uid]
			if !accok {
				continue
			}
			tmpInfo.Title = lv.Show.Title
			tmpInfo.Cover = lv.Show.Cover
			if lv.Show.Cover == "" {
				tmpInfo.Cover = lv.Show.Keyframe
			}
			tmpInfo.Mid = lv.Uid
			tmpInfo.Name = accCard.Name
			if lv.Area != nil {
				tmpInfo.TagName = lv.Area.AreaName
				tmpInfo.LiveParentAreaId = lv.Area.ParentAreaId
				tmpInfo.LiveAreaId = lv.Area.AreaId
			}
			if lv.Status != nil {
				if lv.Status.LiveStatus == 1 { //1是直播中，0、2是未开播
					tmpInfo.LiveStatus = 1
				}
			}
			if model.IsAndroid(plat) && build < _androidBroadcast {
				lv = nil
			}
			tmpInfo.URI = model.FillURI(tmpInfo.Goto, strconv.FormatInt(his.Oid, 10), model.LiveHandler(lv))
			lp, ok := livep[his.Oid]
			if ok && lp != nil && lp.Link != "" {
				tmpInfo.URI = lp.Link
			}
			if lv != nil {
				tmpInfo.Relation = cardm.RelationChange(lv.Uid, interrelations)
				follow := cardm.RelationOldChange(lv.Uid, interrelations)
				if follow == 0 && lv.Uid != mid && mid > 0 && !showLiveShare(c) {
					tmpInfo.DisAtten = 1
				}
			}
		case _tpCorpus:
			art, ok := article[his.Cid]
			if !ok {
				continue
			}
			tmpInfo.Title = art.Title
			tmpInfo.Covers = art.ImageURLs
			if art.Author != nil {
				tmpInfo.Mid = art.Author.Mid
				tmpInfo.Name = art.Author.Name
				tmpInfo.Relation = cardm.RelationChange(art.Author.Mid, interrelations)
				follow := cardm.RelationOldChange(art.Author.Mid, interrelations)
				if follow == 0 && art.Author.Mid != mid && mid > 0 {
					tmpInfo.DisAtten = 1
				}
			}
			articleInfo := artm.GetArticleInfo(ctx, int64(art.Type), his.Cid, art.CoverAvid)
			tmpInfo.Badge = articleInfo.Badge
			tmpInfo.URI = articleInfo.Uri
		case int32(cheeseType):
			cheese, okCheese := cheeseCards[int32(his.Epid)]
			arc, okArc := archive[his.Oid]
			if !okCheese || !okArc {
				continue
			}
			tmpInfo.State = int64(cheese.Status)
			tmpInfo.Title = cheese.SeasonTitle
			tmpInfo.ShowTitle = cheese.Title
			tmpInfo.Cover = cheese.Cover
			tmpInfo.Progress = his.Pro
			tmpInfo.URI = cheese.Url
			for _, p := range arc.Pages {
				if p.Cid == his.Cid {
					tmpInfo.Duration = p.Duration
					break
				}
			}
		default:
			continue
		}
		data = append(data, tmpInfo)
	}
	return
}

// Del for history
func (s *Service) Del(c context.Context, mid int64, hisRes []*hisApi.ModelHistory, dev device.Device) (err error) {
	err = s.historyDao.Del(c, mid, hisRes, dev.Buvid)
	if err != nil {
		log.Error("s.historyDao.Del error(%+v)", err)
		return
	}
	return
}

// Clear for history
func (s *Service) Clear(c context.Context, param *history.HisParam, plat int8, dev device.Device) (err error) {
	businesses, ok := busTabMap[param.Business]
	if !ok {
		err = ecode.RequestErr
		log.Error("historyClear invalid business(%s)", param.Business)
		return
	}
	if s.cheeseDao.HasCheese(plat, int(dev.Build), false) && (param.Business == "all" || param.Business == "archive") {
		businesses = append(businesses, _cheeseStr)
	}
	if ((plat == model.PlatIPad && dev.Build > int64(s.c.BuildLimit.IPadCheese)) || plat == model.PlatAndroidHD || (plat == model.PlatIpadHD && dev.Build > int64(s.c.BuildLimit.IPadHDCheese))) &&
		(param.Business == "all" || param.Business == "archive") {
		businesses = append(businesses, _cheeseIPadStr)
	}
	err = s.historyDao.Clear(c, param.Mid, businesses, dev.Buvid)
	if err != nil {
		log.Error("%+v ", err)
		return
	}
	return
}

// CursorGRPC
func (s *Service) CursorGRPC(c context.Context, mid int64, dev device.Device, plat int8, arg *api.CursorReq, net network.Network) (*api.CursorReply, error) {
	res := new(api.CursorReply)
	param := &history.HisParam{
		Ps:  int32(20),
		Mid: mid,
	}
	if arg.Cursor != nil {
		param.Max = arg.Cursor.Max
		param.MaxTP = arg.Cursor.MaxTp
	}
	if arg.Business == "" {
		arg.Business = "all"
	}
	param.Business = arg.Business
	param.Build = dev.Build
	param.Platform = dev.RawPlatform
	param.Buvid = dev.Buvid
	param.MobiApp = dev.RawMobiApp
	liveParam := &history.LiveParam{
		Uid:        mid,
		Build:      param.Build,
		Platform:   param.Platform,
		DeviceName: dev.Model,
		NetWork:    fetchLiveNetType(net),
		ReqBiz:     _hisCursor,
	}
	hisCursor, hasMore, err := s.Cursor(c, param, param.Ps, plat, true, liveParam)
	if err != nil {
		return nil, err
	}
	if hisCursor == nil {
		return res, nil
	}
	if hisCursor.Cursor != nil {
		res.Cursor = &api.Cursor{
			MaxTp: hisCursor.Cursor.MaxTP,
			Max:   hisCursor.Cursor.Max,
		}
	}
	for _, v := range hisCursor.Tab {
		if v == nil {
			continue
		}
		res.Tab = append(res.Tab, &api.CursorTab{
			Business: v.Business,
			Name:     v.Name,
		})
	}
	if len(hisCursor.List) == 0 {
		return res, nil
	}
	//历史服务端未到底+网关测当页有数据 认为还可翻页
	res.HasMore = hasMore
	res.Items = s.buildGRPCRes(c, hisCursor.List, "", s.c.Custom.HisHasShare)
	return res, nil
}

// Search
func (s *Service) Search(c context.Context, mid int64, plat int8, arg *api.SearchReq, dev device.Device, net network.Network) (*api.SearchReply, error) {
	ps := int32(20)
	hisBusTabMap := busTabMap
	if model.IsOverseas(plat) {
		// 新版繁体版出直播
		if needShowLiveContent(c) {
			hisBusTabMap = busTabNewOverseaMap
		} else {
			hisBusTabMap = busTabOverseaMap
		}
	}

	businesses := hisBusTabMap[arg.Business]
	if model.IsPinkAndBlue(plat) && (arg.Business == "all" || arg.Business == "archive") {
		businesses = append(businesses, _cheeseStr)
	}
	if (plat == model.PlatIPad || plat == model.PlatIpadHD) && (arg.Business == "all" || arg.Business == "archive") {
		businesses = append(businesses, _cheeseIPadStr)
	}
	searchRes, total, err := s.historyDao.Search(c, mid, int32(arg.Pn), ps, arg.Keyword, businesses, dev.Buvid)
	if err != nil {
		log.Error("%+v ", err)
		return nil, err
	}
	res := &api.SearchReply{
		Page: &api.Page{
			Pn:    arg.Pn,
			Total: int64(total),
		},
		HasMore: len(searchRes) >= int(ps),
	}
	if len(searchRes) == 0 {
		return res, nil
	}
	liveParam := &history.LiveParam{
		Uid:        mid,
		Platform:   dev.RawPlatform,
		DeviceName: dev.Model,
		NetWork:    fetchLiveNetType(net),
		Build:      dev.Build,
		ReqBiz:     _hisSearch,
	}
	tmpRes := s.TogetherHistory(c, searchRes, mid, dev.Build, plat, dev.MobiApp(), false, liveParam)
	res.Items = s.buildGRPCRes(c, tmpRes, arg.Keyword, s.c.Custom.HisHasShare)
	return res, nil
}

func (s *Service) buildGRPCRes(ctx context.Context, list []*history.ListRes, keyword string, hasShare bool) (res []*api.CursorItem) {
	for _, v := range list {
		if v == nil {
			continue
		}
		item := &api.CursorItem{
			Title:    api.FromTitle(v.Title, keyword),
			Uri:      v.URI,
			ViewAt:   v.ViewAt,
			Kid:      v.History.Kid,
			Oid:      v.History.Oid,
			Business: v.History.Business,
			Tp:       v.History.Tp,
			Dt:       model.HistoryDt(int8(v.History.Dt), s.c.HisIcon),
		}
		switch v.Goto {
		case model.GotoAv:
			item.HasShare = hasShare
			item.CardItem = api.FromCardUGC(v)
		case model.GotoPGC:
			item.HasShare = hasShare
			item.CardItem = api.FromCardOGV(v)
		case model.GotoArticle:
			item.CardItem = api.FromCardArticle(v)
		case model.GotoLive:
			if showLiveShare(ctx) {
				item.HasShare = hasShare
			}
			item.CardItem = api.FromCardLive(v)
		case model.GotoCheese:
			item.CardItem = api.FromCardCheese(v)
		}
		res = append(res, item)
	}
	return
}

// 直播卡改造需求暂时挂起
func showLiveShare(ctx context.Context) bool {
	return false
}

func (s *Service) ClearGRPC(c context.Context, arg *api.ClearReq, dev device.Device, plat int8, mid int64) error {
	if arg == nil {
		return ecode.RequestErr
	}
	param := &history.HisParam{
		Business: arg.Business,
		Mid:      mid,
	}
	return s.Clear(c, param, plat, dev)
}

// LatestHistoryGRPC
func (s *Service) LatestHistoryGRPC(c context.Context, mid int64, dev device.Device, plat int8, arg *api.LatestHistoryReq) (*api.LatestHistoryReply, error) {
	res := new(api.LatestHistoryReply)
	if s.c.Custom.HisLTime == 0 { //关掉续播提醒
		return res, nil
	}
	param := &history.HisParam{
		Ps:  int32(20),
		Mid: mid,
	}
	if arg.Business == "" {
		arg.Business = "all"
	}
	param.Business = arg.Business
	param.Build = dev.Build
	param.Platform = dev.RawPlatform
	param.Buvid = dev.Buvid
	param.MobiApp = dev.RawMobiApp
	hisCursor, _, err := s.Cursor(c, param, param.Ps, plat, true, nil)
	if err != nil {
		return nil, err
	}
	if hisCursor == nil || len(hisCursor.List) == 0 {
		return res, nil
	}
	rtime, ltime, flag := s.fetchTimeExpVal(param.Buvid)
	for _, v := range hisCursor.List {
		if v.Progress == -1 || v.Duration <= 0 ||
			float64(v.Progress)/float64(v.Duration) > s.c.Custom.LatestHistoryPro ||
			(time.Now().Unix()-v.ViewAt) > ltime ||
			v.History.WatchedBuvid == dev.Buvid { // （播完 || duration不符合要求 || 播放进度大于95% || 播放间隔大于配置值 || 最近的历史记录是本机产生 不展示
			continue
		}
		rawRes := []*history.ListRes{v}
		tempIteams := s.buildGRPCRes(c, rawRes, "", false) //不需要分享按钮
		if len(tempIteams) <= 0 {
			continue
		}
		res.Items = tempIteams[0]
		res.Scene = _homePage
		res.Rtime = rtime
		res.Flag = flag
		return res, nil
	}
	return res, nil
}

func (s *Service) fetchTimeExpVal(buvid string) (int64, int64, string) { //rtime, ltime, flag
	var (
		lsalt        = "ltimeHash"
		rsalt        = "rtimeHash"
		group        uint32
		exp          string
		rtime, ltime int64
	)
	//两个实验交错进行
	if len(s.c.HisRTimeMap) != 0 {
		//nolint:gomnd
		group = crc32.ChecksumIEEE([]byte(buvid+rsalt)) % 5
		exp = _hdExpRtime
	}
	if len(s.c.HisLTimeMap) != 0 {
		//nolint:gomnd
		group = crc32.ChecksumIEEE([]byte(buvid+lsalt)) % 5
		exp = _hdExpLtime
	}
	if exp == "" { //两个实验都关闭
		return s.c.Custom.HisRTime, s.c.Custom.HisLTime, ""
	}
	switch exp {
	case _hdExpRtime:
		rtime = s.c.HisRTimeMap[strconv.Itoa(int(group))]
	case _hdExpLtime:
		ltime = s.c.HisLTimeMap[strconv.Itoa(int(group))]
	default:
		log.Warn("exp error")
	}
	if rtime == 0 {
		rtime = s.c.Custom.HisRTime
	}
	if ltime == 0 {
		ltime = s.c.Custom.HisLTime
	}
	return rtime, ltime, strconv.Itoa(int(group))
}

// 判断是否是新版繁体版，是否出直播
func needShowLiveContent(c context.Context) bool {
	return pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroidI().And().Build(">=", int64(3000500))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhoneI().And().Build(">=", int64(64400200))
	}).FinishOr(false)
}
