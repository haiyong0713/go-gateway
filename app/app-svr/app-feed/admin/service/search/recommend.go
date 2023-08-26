package search

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	inlinegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/game"
	"go-gateway/app/app-svr/app-feed/admin/model/manager"
	searchModel "go-gateway/app/app-svr/app-feed/admin/model/search"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	PLAT_APP = 1
	PLAT_WEB = 2
)

var (
	_emptyPlatVer = make([]*searchModel.PlatVer, 0)
	_emptyExtra   = &searchModel.SpreadConfigExtra{}
)

// pgcValues get pgc values
func (s *Service) pgcValues(c context.Context, ids []int32) (res map[int32]*seasongrpc.CardInfoProto, err error) {
	return s.pgcDao.CardsInfoReply(c, ids)
}

func (s *Service) pgcValues2(c context.Context, ids []int32) (res map[int32]*seasongrpc.CardInfoProto, err error) {
	return s.pgcDao.CardsInfoReply2(c, ids)
}

func (s *Service) specilValues(ids []int64) (map[int64]*manager.SpecialCard, error) {
	return s.managerDao.SpecialCards(ids)
}

func (s *Service) contentValues(ids []int64) (map[int64]*manager.ContentCard, error) {
	return s.managerDao.ContentCards(ids)
}

func (s *Service) navigationValues(ids []int64) (map[int64]*searchModel.NavigationCard, error) {
	return s.dao.NavigationCards(ids)
}

func (s *Service) epValues(c context.Context, ids []int32) (map[int32]*inlinegrpc.EpisodeCard, error) {
	return s.pgcDao.InlineCardEpInfoReply(c, ids)
}

func (s *Service) gameValues(c context.Context, ids []int64) (map[int64]*game.Info, error) {
	return s.gameDao.GameListApp(c, ids)
}

func (s *Service) platValues(v *searchModel.SpreadConfig) {
	var (
		err error
	)
	if v.Platform == searchModel.PlatFromApp {
		var (
			platver []*searchModel.PlatVer
		)
		if err = json.Unmarshal([]byte(v.PlatVerStr), &platver); err != nil {
			log.Error("platValues.Unmarshal(%s) error(%v)", v.PlatVerStr, err)
			v.PlatVer = _emptyPlatVer
			return
		}
		v.PlatVer = platver
	} else {
		v.PlatVer = _emptyPlatVer
	}
}

func (s *Service) extraValues(v *searchModel.SpreadConfig) {
	var (
		err               error
		extraInfo         = &searchModel.SpreadConfigExtra{}
		extraSupportTypes = []int{
			// inline support types
			searchModel.CardInlineUgc,
			searchModel.CardInlineLive,
			searchModel.CardInlineOgv,

			// wiki corner support types
			// 特殊大卡、特殊小卡、视频卡、inline UGC、inline OGV、inline 直播、聚合卡，即只有这几种卡片类型才有百科角标配置项
			// https://info.bilibili.co/pages/viewpage.action?pageId=241699280
			searchModel.CardSpecial,
			searchModel.CardSpecialSmall,
			searchModel.CardVideo,
			searchModel.CardUnion,
			searchModel.CardGameBig,
		}
	)

	if len(v.ExtraStr) == 0 {
		v.Extra = _emptyExtra
		return
	}

	if util.IsIntInArray(extraSupportTypes, int(v.CardType)) {
		if err = json.Unmarshal([]byte(v.ExtraStr), extraInfo); err != nil {
			log.Error("extraValues.Unmarshal(%s) error(%v)", v.ExtraStr, err)
			v.Extra = _emptyExtra
			return
		}
		v.Extra = extraInfo
	} else {
		v.Extra = _emptyExtra
	}
}

func (s *Service) OpenRecommend2(c context.Context, param *searchModel.RecomParam) (res *searchModel.RecomRes, err error) {
	now := time.Now().Unix()
	if len(param.CardType) == 0 && param.Pos == 0 && param.Ts == 0 &&
		(param.EndTs-param.StartTs == 60*30) &&
		//nolint:gomnd
		(math.Abs(float64(param.StartTs-now)) <= 60) {

		// AI访问openRecommend时传入参数：
		// 1. 无card_type，pos，ts等额外参数
		// 2. 开始时间为当前时间
		// 3. 结束时间为当前时间+30分钟
		// 此处缓存适用于以上条件成立时，并且入参开始时间和当前时间的差值不大于1分钟，则使用内存缓存
		return s.OpenRecommendCache(c, param)

	} else if param.Plat == 0 && param.Pos == 0 && param.Ts == 0 && param.StartTs == 0 && param.EndTs == 0 {
		// 搜索干预获取特殊小卡访问openRecommend时传入参数：
		// 1. 只会传入card_type，pn，ps，其他参数都不传
		return s.OpenRecommendCacheNow(c, param)

	} else {
		return s.OpenRecommend(c, param)
	}
}

func (s *Service) OpenRecommendCache(c context.Context, param *searchModel.RecomParam) (res *searchModel.RecomRes, err error) {
	res = &searchModel.RecomRes{
		Page: common.Page{
			Num:  param.Pn,
			Size: param.Ps,
		},
	}

	var list []*searchModel.SpreadConfig
	if param.Plat == PLAT_APP {
		if len(param.SearchGroup) == 0 {
			list = s.RcmdAppCache
		} else {
			for _, rcmd := range s.RcmdAppCache {
				if util.IsStringInArray(param.SearchGroup, rcmd.SearchGroup) {
					list = append(list, rcmd)
				}
			}
		}
	} else if param.Plat == PLAT_WEB {
		list = s.RcmdWebCache
	}

	sliceStart, sliceEnd := util.PaginateSlice(param.Pn, param.Ps, len(list))
	res.Item = list[sliceStart:sliceEnd]
	res.Page.Total = len(list)
	return
}

func (s *Service) OpenRecommendCacheNow(c context.Context, param *searchModel.RecomParam) (res *searchModel.RecomRes, err error) {
	res = &searchModel.RecomRes{
		Page: common.Page{
			Num:  param.Pn,
			Size: param.Ps,
		},
	}

	var list []*searchModel.SpreadConfig
	if len(param.CardType) == 0 {
		list = s.RcmdCache
	} else {
		list = make([]*searchModel.SpreadConfig, 0, len(s.RcmdCache))
		for _, rcmd := range s.RcmdCache {
			if util.IsIntInArray(param.CardType, int(rcmd.CardType)) {
				list = append(list, rcmd)
			}
		}
	}

	sliceStart, sliceEnd := util.PaginateSlice(param.Pn, param.Ps, len(list))
	res.Item = list[sliceStart:sliceEnd]
	res.Page.Total = len(list)
	return
}

func (s *Service) OpenRecommendCount(c context.Context, param *searchModel.RecomParam) (count int, err error) {
	count, err = s.dao.GetSearchSpreadCount(c, param)
	if err == nil && count > 0 {
		return
	}

	var (
		sTimeStr, eTimeStr string
	)
	w := map[string]interface{}{
		"valid_status": common.StatusOnline,
	}
	if param.Ts == 0 {
		if param.StartTs != 0 && param.EndTs != 0 {
			sTimeStr = time.Unix(param.StartTs, 0).Format("2006-01-02 15:04:05")
			eTimeStr = time.Unix(param.EndTs, 0).Format("2006-01-02 15:04:05")
		} else {
			sTimeStr = util.CTimeStr()
			eTimeStr = sTimeStr
		}
	} else {
		sTimeStr = time.Unix(param.Ts, 0).Format("2006-01-02 15:04:05")
		eTimeStr = sTimeStr
	}
	query := s.dao.DB.Model(searchModel.SpreadConfig{})
	query = query.Where(w).Where("start_time <= ?", eTimeStr).Where("end_time >= ?", sTimeStr)
	if param.Plat != 0 {
		query = query.Where("platform = ?", param.Plat)
	}
	if param.Pos != 0 {
		query = query.Where("position = ?", param.Pos)
	}
	if len(param.CardType) > 0 {
		query = query.Where("card_type in (?)", param.CardType)
	}
	if err = query.Count(&count).Error; err != nil {
		log.Errorc(c, "service.OpenRecommend query count error(%v)", err)
		return
	}
	if err = s.dao.SetSearchSpreadCount(c, param, count); err != nil {
		log.Errorc(c, "service.OpenRecommend set search spread count(%v) error(%v)", count, err)
	}
	return
}

// OpenRecommend get validate recommend data
//
//nolint:gocognit
func (s *Service) OpenRecommend(c context.Context, param *searchModel.RecomParam) (res *searchModel.RecomRes, err error) {
	var (
		ids           []int64
		seasonIDs     []int32
		specialIDs    []int64
		contentIDs    []int64
		navigationIDs []int64
		epIDs         []int32
		gameIDs       []int64
		pgcInfos      map[int32]*seasongrpc.CardInfoProto
		speInfos      map[int64]*manager.SpecialCard
		contInfos     map[int64]*manager.ContentCard
		navInfos      map[int64]*searchModel.NavigationCard
		epInfos       map[int32]*inlinegrpc.EpisodeCard
		gameInfos     map[int64]*game.Info
		sTimeStr      string
		eTimeStr      string
		querys        map[int64][]*searchModel.SpreadQuery
		empty         []*searchModel.SpreadQuery
	)
	res = &searchModel.RecomRes{
		Page: common.Page{
			Num:  param.Pn,
			Size: param.Ps,
		},
	}
	w := map[string]interface{}{
		"valid_status": common.StatusOnline,
	}
	if param.Ts == 0 {
		if param.StartTs != 0 && param.EndTs != 0 {
			sTimeStr = time.Unix(param.StartTs, 0).Format("2006-01-02 15:04:05")
			eTimeStr = time.Unix(param.EndTs, 0).Format("2006-01-02 15:04:05")
		} else {
			sTimeStr = util.CTimeStr()
			eTimeStr = sTimeStr
		}
	} else {
		sTimeStr = time.Unix(param.Ts, 0).Format("2006-01-02 15:04:05")
		eTimeStr = sTimeStr
	}
	query := s.dao.DB.Model(searchModel.SpreadConfig{})
	query = query.Where(w).Where("start_time <= ?", eTimeStr).Where("end_time >= ?", sTimeStr)
	if param.Plat != 0 {
		query = query.Where("platform = ?", param.Plat)
	}
	if param.Pos != 0 {
		query = query.Where("position = ?", param.Pos)
	}
	if len(param.CardType) > 0 {
		query = query.Where("card_type in (?)", param.CardType)
	}

	count, err := s.dao.GetSearchSpreadCount(c, param)
	if err != nil || count == 0 {
		log.Errorc(c, "OpenRecommend get spread count from mc count(%v) error(%v)", count, err)

		if err = query.Count(&count).Error; err != nil {
			log.Errorc(c, "service.OpenRecommend query count error(%v)", err)
			return
		}

		if err = s.dao.SetSearchSpreadCount(c, param, count); err != nil {
			log.Errorc(c, "service.OpenRecommend set search spread count(%v) error(%v)", res.Page.Total, err)
		}
	}
	res.Page.Total = count

	if err = query.Order("`id` DESC").Offset((param.Pn - 1) * param.Ps).Limit(param.Ps).Find(&res.Item).Error; err != nil {
		log.Error("OpenRecommend.Find error(%v)", err)
		return
	}
	for _, v := range res.Item {
		ids = append(ids, v.ID)
		s.platValues(v)
		s.extraValues(v)
		if v.CardType == searchModel.CardPGC || v.CardType == searchModel.CardPGCFan || v.CardType == searchModel.CardPGCMove {
			seasonIDs = append(seasonIDs, int32(v.ArticleId))
		} else if v.CardType == searchModel.CardSpecial || v.CardType == searchModel.CardSpecialSmall {
			specialIDs = append(specialIDs, v.ArticleId)
		} else if v.CardType == searchModel.CardUnion {
			contentIDs = append(contentIDs, v.ArticleId)
		} else if v.CardType == searchModel.CardNavigation {
			navigationIDs = append(navigationIDs, v.ArticleId)
		} else if v.CardType == searchModel.CardInlineOgv {
			epIDs = append(epIDs, int32(v.ArticleId))
		} else if v.CardType == searchModel.CardGame {
			gameIDs = append(gameIDs, v.ArticleId)
		}
	}
	eg := errgroup.WithCancel(c)
	if len(seasonIDs) > 0 {
		eg.Go(func(ctx context.Context) (pgcError error) {
			if pgcInfos, pgcError = s.pgcValues2(ctx, seasonIDs); pgcError != nil {
				log.Error("OpenRecommend.pgcValues param(%+v) error(%v)", seasonIDs, pgcError)
			}
			return nil
		})
	}
	if len(specialIDs) > 0 {
		eg.Go(func(ctx context.Context) (speError error) {
			if speInfos, speError = s.specilValues(specialIDs); speError != nil {
				log.Error("OpenRecommend.specilValues param(%+v) error(%v)", specialIDs, speError)
			}
			return nil
		})
	}
	if len(contentIDs) > 0 {
		eg.Go(func(ctx context.Context) (contError error) {
			if contInfos, contError = s.contentValues(contentIDs); contError != nil {
				log.Error("OpenRecommend.contentValues param(%+v) error(%v)", contentIDs, contError)
			}
			return nil
		})
	}
	if len(navigationIDs) > 0 {
		eg.Go(func(ctx context.Context) (navError error) {
			if navInfos, navError = s.navigationValues(navigationIDs); navError != nil {
				log.Error("OpenRecommend.navigationValues param(%+v) error(%v)", navigationIDs, navError)
			}
			return nil
		})
	}
	if len(epIDs) > 0 {
		eg.Go(func(ctx context.Context) (epError error) {
			if epInfos, epError = s.epValues(ctx, epIDs); epError != nil {
				log.Error("OpenRecommend.epValues param(%+v) error(%v)", epIDs, epError)
			}
			return nil
		})
	}
	if len(gameIDs) > 0 {
		eg.Go(func(ctx context.Context) (gameError error) {
			if gameInfos, gameError = s.gameValues(c, gameIDs); gameError != nil {
				log.Error("OpenRecommend.gameValues param(%+v) error(%v)", gameIDs, gameError)
			}
			return nil
		})
	}
	if len(ids) > 0 {
		eg.Go(func(ctx context.Context) (queryError error) {
			if querys, queryError = s.getSpreadQuery(ids); queryError != nil {
				log.Error("OpenRecommend.getSpreadQuery param(%+v) error(%v)", ids, queryError)
			}
			return nil
		})
	}
	//nolint:errcheck
	eg.Wait()
	item := make([]*searchModel.SpreadConfig, 0, len(res.Item))
	for _, spread := range res.Item {
		if v, ok := querys[spread.ID]; ok {
			spread.Query = v
		} else {
			spread.Query = empty
		}
		if spread.CardType == searchModel.CardPGC || spread.CardType == searchModel.CardPGCFan || spread.CardType == searchModel.CardPGCMove {
			pgcID := int32(spread.ArticleId)
			if pgc, ok := pgcInfos[pgcID]; ok {
				spread.PgcSeason = pgc
			} else {
				log.Error("OpenRecommend pgc card no value ID(%+v)", pgcID)
			}
		} else if spread.CardType == searchModel.CardSpecial || spread.CardType == searchModel.CardSpecialSmall {
			if special, ok := speInfos[spread.ArticleId]; ok {
				spread.Special = special
			} else {
				log.Error("OpenRecommend special card no value ID(%+v)", spread.ArticleId)
			}
		} else if spread.CardType == searchModel.CardUnion {
			if union, ok := contInfos[spread.ArticleId]; ok {
				spread.Union = union
			} else {
				log.Error("OpenRecommend union card no value ID(%+v)", spread.ArticleId)
			}
		} else if spread.CardType == searchModel.CardNavigation {
			if nav, ok := navInfos[spread.ArticleId]; ok {
				spread.Navigation = nav
			} else {
				log.Error("OpenRecommend navigation card no value ID(%+v)", spread.ArticleId)
			}
		} else if spread.CardType == searchModel.CardInlineOgv {
			if ogv, ok := epInfos[int32(spread.ArticleId)]; ok {
				spread.Ogv = ogv
			} else {
				log.Error("OpenRecommend inline ogv card no value ID(%+v)", spread.ArticleId)
			}
		} else if spread.CardType == searchModel.CardGame {
			// 2021M7W4: 未上线的游戏不下发给AI
			if _, ok := gameInfos[spread.ArticleId]; !ok {
				log.Error("OpenRecommend game card no value ID(%+v)", spread.ArticleId)
				continue
			}
		}
		item = append(item, spread)
	}
	res.Item = item
	return
}

// getSpreadQuery .
func (s *Service) getSpreadQuery(ids []int64) (res map[int64][]*searchModel.SpreadQuery, err error) {
	where := map[string]interface{}{
		"del_status": common.NotDeleted,
	}
	SearchWebQuery := make([]*searchModel.SpreadQuery, 0)
	if err = s.dao.DB.Model(&searchModel.SpreadQuery{}).Where(where).Where("spread_id in (?)", ids).Find(&SearchWebQuery).Error; err != nil {
		log.Error("getSpreadQuery Find error(%v)", err)
		return
	}
	res = make(map[int64][]*searchModel.SpreadQuery, len(SearchWebQuery))
	for _, v := range SearchWebQuery {
		res[v.SpreadId] = append(res[v.SpreadId], v)
	}
	return
}

func (s *Service) OpenChannelIdsCache(ps, pn int) (ids []int64, total int, err error) {
	list := s.ChannelIdCache
	sliceStart, sliceEnd := util.PaginateSlice(pn, ps, len(list))
	ids = list[sliceStart:sliceEnd]
	total = len(list)
	return
}

// 给频道服务端用，返回管理后台所有配置过的频道id
func (s *Service) OpenChannelIds(ps, pn int) (ids []int64, total int, err error) {
	var (
		configList []*searchModel.SpreadConfig
	)
	cTimeStr := util.CTimeStr()
	query := s.dao.DB.Model(&searchModel.SpreadConfig{}).
		Where(map[string]interface{}{
			"valid_status": common.StatusOnline,
			"card_type":    15,
		}).Where("start_time <= ?", cTimeStr).
		Where("end_time >= ?", cTimeStr)
	if err = query.Select("count(distinct(article_id))").
		Count(&total).Error; err != nil {
		log.Error("OpenChannelIds Count error(%v)", err)
		return
	}
	db := query.Select("distinct article_id").Order("article_id")
	if ps > 0 && pn > 0 {
		db = db.Offset((pn - 1) * ps).Limit(ps)
	}
	if err = db.Find(&configList).Error; err != nil {
		log.Error("OpenChannelIds Find error(%v)", err)
		return
	}
	for _, config := range configList {
		ids = append(ids, config.ArticleId)
	}
	return
}
