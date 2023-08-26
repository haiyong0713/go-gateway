package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	gameModel "go-gateway/app/app-svr/app-feed/admin/model/game"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"

	eg "go-common/library/sync/errgroup.v2"

	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

const (
	//moreShowAv 稿件
	moreShowAv = 2
	//moreShowPgc pgc
	moreShowPgc = 3
	//moreShowArticle 专栏
	moreShowArticle = 4
	//pgcMoreUrlH5 pgc 聚合卡 更多跳转H5
	pgcMoreUrlH5 = 1
	//pgcMoreUrlH5 pgc 聚合卡 番剧垂搜
	//pgcMoreUrlOgv = 2
	//pgcMoreUrlMovies pgc 聚合卡 影视垂搜
	pgcMoreUrlMovies = 3
	//OgvStatusShow .
	OgvStatusShow   = 1
	_maxOgvQuery    = 15
	_maxOgvTitle    = 15
	_maxOgvSubTitle = 30
	_maxOgvMoreshow = 10
	_colorSelfDef   = 11
)

// OgvList ogv list
//
//nolint:gocognit
func (s *Service) OgvList(c context.Context, lp *show.SearchOgvLP) (pager *show.SearchOgvPager, err error) {
	var (
		ids    []int64
		resOgv = make([]*show.SearchOgv, 0)
	)
	//1-粉 2-红 3-橙 4-黄 5-绿 6-青 7-湖蓝 8-蓝 9-紫 10-灰黑
	colorMap := map[int64]string{
		1:  "#D85D84",
		2:  "#BA4B45",
		3:  "#D48735",
		4:  "#CC9B14",
		5:  "#50A164",
		6:  "#3F9B97",
		7:  "#1389BF",
		8:  "#3A69B7",
		9:  "#5A62C7",
		10: "#363E53",
	}
	pager = &show.SearchOgvPager{
		Page: common.Page{
			Num:  lp.Pn,
			Size: lp.Ps,
		},
	}
	w := map[string]interface{}{}
	query := s.showDao.DB.Model(&show.SearchOgv{})
	if lp.ID > 0 {
		w["id"] = lp.ID
	}
	if lp.Person != "" {
		query = query.Where("person like ?", "%"+lp.Person+"%")
	}
	if lp.Title != "" {
		query = query.Where("hd_title like ?", "%"+lp.Title+"%")
	}
	if lp.Query != "" {
		var (
			ogvQuery []*show.SearchOgvQuery
			sids     []int64
		)
		where := map[string]interface{}{
			"deleted": common.NotDeleted,
		}
		if err = s.showDao.DB.Model(&show.SearchOgvQuery{}).Where(where).Where("value like ?", "%"+lp.Query+"%").Find(&ogvQuery).Error; err != nil {
			log.Error("OgvList Find error(%v)", err)
			return
		}
		if len(ogvQuery) == 0 {
			return
		}
		for _, v := range ogvQuery {
			sids = append(sids, v.SID)
		}
		if len(sids) > 0 {
			query = query.Where("id in (?)", sids)
		}
	}
	if lp.Check != 0 {
		if lp.Check == common.Valid {
			cTimeStr := util.CTimeStr()
			//已通过 已生效
			query = query.Where("`check` = ?", common.Pass)
			if lp.Ts != 0 {
				//搜索 传时间戳
				cTimeStr = time.Unix(lp.Ts, 0).Format("2006-01-02 15:04:05")
			}
			query = query.Where("stime <= ?", cTimeStr)
		} else {
			query = query.Where("`check` = ? ", lp.Check)
		}
	}
	if lp.Stime != 0 {
		//后台查询
		query = query.Where("stime >= ?", time.Unix(lp.Stime, 0).Format("2006-01-02 15:04:05"))
	}
	if lp.Etime != 0 {
		t, e := time.ParseInLocation("2006-01-02 15:04:05", time.Unix(lp.Etime, 0).Format("2006-01-02")+" 23:59:59", time.Local)
		if e != nil {
			err = e
			return
		}
		query = query.Where("stime <= ?", time.Unix(t.Unix(), 0).Format("2006-01-02 15:04:05"))
	}
	if err = query.Where(w).Count(&pager.Page.Total).Error; err != nil {
		log.Error("OgvList count error(%v)", err)
		return
	}
	ogv := make([]*show.SearchOgv, 0)
	if err = query.Where(w).Order("`mtime` DESC, `id` DESC").Offset((lp.Pn - 1) * lp.Ps).Limit(lp.Ps).Find(&ogv).Error; err != nil {
		log.Error("OgvList Find error(%v)", err)
		return
	}
	if len(ogv) > 0 {
		for _, v := range ogv {
			//已生效
			if v.Check == common.Pass {
				if time.Now().Unix() >= v.Stime.Time().Unix() {
					v.Check = common.Valid
				}
			}
			ids = append(ids, v.ID)
		}
	}
	if len(ids) > 0 {
		var (
			query    []*show.SearchOgvQuery
			moreShow []*show.SearchOgvMoreshow
		)
		mapQuery := make(map[int64][]*show.SearchOgvQuery, len(ids))
		mapMoreShow := make(map[int64][]*show.SearchOgvMoreshow, len(ids))
		wQuery := map[string]interface{}{
			"deleted": common.NotDeleted,
		}
		if err = s.showDao.DB.Model(&show.SearchOgvQuery{}).Where(wQuery).Where("sid in (?)", ids).Find(&query).Error; err != nil {
			log.Error("OgvList Find SearchOgvQuery error(%v)", err)
			return
		}
		for _, v := range query {
			mapQuery[v.SID] = append(mapQuery[v.SID], v)
		}
		if err = s.showDao.DB.Model(&show.SearchOgvMoreshow{}).Where(wQuery).Where("sid in (?)", ids).Find(&moreShow).Error; err != nil {
			log.Error("OgvList Find SearchOgvMoreshow error(%v)", err)
			return
		}
		for _, v := range moreShow {
			mapMoreShow[v.Sid] = append(mapMoreShow[v.Sid], v)
		}
		for _, v := range ogv {
			if query, ok := mapQuery[v.ID]; ok {
				v.Query = query
			} else {
				v.Query = []struct{}{}
			}
			if show, ok := mapMoreShow[v.ID]; ok {
				v.MoreshowValue = show
			} else {
				v.MoreshowValue = []struct{}{}
			}
			if v.Color == _colorSelfDef {
				//自定义色值
				v.ColorStr = v.ColorValue
			} else {
				if color, ok := colorMap[v.Color]; ok {
					v.ColorStr = color
				}
			}
			pgcIDsStr := strings.Split(v.PgcIds, ",")
			var (
				pgcIDs     []int32
				pgcID      int64
				seasonInfo map[int32]*seasongrpc.CardInfoProto
			)
			for _, pgcIDStr := range pgcIDsStr {
				if pgcID, err = strconv.ParseInt(pgcIDStr, 10, 64); err != nil {
					log.Error("OgvList strconv.ParseInt(%s) error(%v)", pgcIDStr, err)
					err = nil
					continue
				}
				pgcIDs = append(pgcIDs, int32(pgcID))
			}
			if seasonInfo, err = s.pgcDao.CardsInfoReply(c, pgcIDs); err != nil {
				log.Error("OgvList CardsInfoReply(%v) error(%v)", pgcIDs, err)
				err = nil
				continue
			}
			// 只返回存在有效pgc的ogv card数据
			if len(seasonInfo) > 0 {
				effectivePgcIDMap := make(map[string]bool)
				for _, season := range seasonInfo {
					effectivePgcIDMap[strconv.FormatInt(int64(season.SeasonId), 10)] = true
					v.PgcMediaID = append(v.PgcMediaID, season.MediaId)
				}
				// 同时过滤掉原数据中已失效的pgc id
				effectivePgcIDStrs := make([]string, 0, len(seasonInfo))
				for _, pgcIDStr := range pgcIDsStr {
					if effectivePgcIDMap[strings.TrimSpace(pgcIDStr)] {
						effectivePgcIDStrs = append(effectivePgcIDStrs, strings.TrimSpace(pgcIDStr))
					}
				}

				v.PgcIds = strings.Join(effectivePgcIDStrs, ",")
				resOgv = append(resOgv, v)
			}
		}
	}
	pager.Item = resOgv
	return
}

func (s *Service) getOgvQuery(query string) (res string, err error) {
	var (
		querys   []*show.SearchOgvQuery
		tmpQuery []string
	)
	if err = json.Unmarshal([]byte(query), &querys); err != nil {
		return
	}
	for _, v := range querys {
		if len([]rune(v.Value)) > _maxOgvQuery {
			return "", fmt.Errorf("query最大长度不能超过%d个字符", _maxOgvQuery)
		}
		tmpQuery = append(tmpQuery, v.Value)
	}
	res = strings.Join(tmpQuery, ",")
	return
}

func (s *Service) valOgvTitle(title, _ string) error {
	if len([]rune(title)) > _maxOgvTitle {
		return fmt.Errorf("标题最大长度不能超过%d个字符", _maxOgvTitle)
	}
	if len([]rune(title)) > _maxOgvTitle {
		return fmt.Errorf("副标题最大长度不能超过%d个字符", _maxOgvSubTitle)
	}
	return nil
}

// OgvAdd add ogv
func (s *Service) OgvAdd(c context.Context, param *show.SearchOgvAP) (err error) {
	var (
		query string
	)
	if err = s.ValidatOgv(c, param.MoreshowStatus, param.GameStatus, param.MoreshowValue, param.Plat, param.GameValue, param.PgcIds, param.PgcMoreURL, param.PgcMoreType); err != nil {
		return
	}
	if err = s.valOgvTitle(param.HdTitle, param.HdSubtitle); err != nil {
		return
	}
	if err = s.showDao.SearchOgvAdd(param); err != nil {
		return
	}
	if query, err = s.getOgvQuery(param.Query); err != nil {
		return
	}
	if err = util.AddOgvLogs(param.Person, 0, param.ID, common.ActionAdd, query, param.HdTitle); err != nil {
		log.Error("OgvAdd AddLog error(%v)", err)
		return
	}
	return nil
}

//nolint:gocognit
func (s *Service) ValidatOgv(c context.Context, moreShowStatus, gameStatus int64, moreShowStr, plat, game, pgcIDStr, pgcMoreUrl string, pgcMoreType int64) (err error) {
	var (
		moreShow   []*show.SearchOgvMoreshow
		pgcIDs     []int32
		avIDs      []int64
		articleIDs []int64
	)
	if pgcMoreType > pgcMoreUrlMovies || pgcMoreType < pgcMoreUrlH5 {
		return fmt.Errorf("pgc聚合卡参数错误")
	}
	if pgcMoreType == pgcMoreUrlH5 && pgcMoreUrl == "" {
		return fmt.Errorf("pgc聚合卡 h5跳转地址不能为空")
	}
	if moreShowStatus == OgvStatusShow {
		if err = json.Unmarshal([]byte(moreShowStr), &moreShow); err != nil {
			return
		}
		if len(moreShow) == 0 {
			return fmt.Errorf("发现更多精彩不能为空")
		}
		for _, v := range moreShow {
			if len([]rune(v.Word)) > _maxOgvMoreshow {
				return fmt.Errorf("展示词最多不能超过%d", _maxOgvMoreshow)
			}
		}
	}
	for _, v := range moreShow {
		if v.Type == moreShowArticle || v.Type == moreShowAv || v.Type == moreShowPgc {
			var id int64
			id, err = strconv.ParseInt(v.Value, 10, 64)
			if err != nil {
				return
			}
			if v.Type == moreShowArticle {
				articleIDs = append(articleIDs, id)
			} else if v.Type == moreShowAv {
				avIDs = append(avIDs, id)
			} else if v.Type == moreShowPgc {
				pgcIDs = append(pgcIDs, int32(id))
			}
		}
	}
	eg := eg.Group{}
	if gameStatus == OgvStatusShow {
		eg.Go(func(ctx context.Context) (gameError error) {
			var (
				vs       []*common.Version
				gameID   int64
				gameInfo *gameModel.Info
			)
			if gameError = json.Unmarshal([]byte(plat), &vs); gameError != nil {
				return
			}
			for _, v := range vs {
				var plat int
				if v.Conditions == common.Android {
					plat = 1
				} else {
					plat = 2
				}
				if gameID, gameError = strconv.ParseInt(game, 10, 64); gameError != nil {
					return fmt.Errorf("游戏ID(%d)错误(%v)", gameID, gameError)
				}
				if gameInfo, gameError = s.gameDao.GameInfo(ctx, gameID, plat); gameError != nil {
					return fmt.Errorf("游戏ID(%d)错误(%v)", gameID, gameError)
				}
				if gameInfo == nil {
					return fmt.Errorf("游戏ID(%d)不存在", gameID)
				}
			}
			return
		})
	}
	if pgcIDStr != "" {
		eg.Go(func(ctx context.Context) (pgcError error) {
			var (
				ids    []int32
				mapPgc map[int32]*seasongrpc.CardInfoProto
			)
			dupMap := make(map[int32]bool)
			pgcIDStr := strings.Split(pgcIDStr, ",")
			for _, v := range pgcIDStr {
				var id int64
				id, pgcError = strconv.ParseInt(v, 10, 64)
				if pgcError != nil {
					return
				}
				ids = append(ids, int32(id))
				if _, ok := dupMap[int32(id)]; ok {
					return fmt.Errorf("id 为%d的pgc数据重复", int32(id))
				}
				dupMap[int32(id)] = true
			}
			if len(ids) == 0 || len(ids) > 50 {
				return fmt.Errorf("pgc ID错误")
			}
			if mapPgc, pgcError = s.pgcValues(ctx, ids); pgcError != nil {
				return
			}
			for _, v := range ids {
				if _, ok := mapPgc[v]; !ok {
					return fmt.Errorf("id 为%d的pgc数据不存在", v)
				}
			}
			return
		})
	}
	if len(pgcIDs) > 0 {
		eg.Go(func(ctx context.Context) (pgcError error) {
			var (
				mapPgc map[int32]*seasongrpc.CardInfoProto
			)
			dupMap := make(map[int32]bool)
			for _, v := range pgcIDs {
				if _, ok := dupMap[v]; ok {
					return fmt.Errorf("id 为%d的pgc数据重复", v)
				}
				dupMap[v] = true
			}
			if mapPgc, pgcError = s.pgcValues(ctx, pgcIDs); pgcError != nil {
				return
			}
			for _, v := range pgcIDs {
				if _, ok := mapPgc[v]; !ok {
					return fmt.Errorf("id 为%d的pgc数据不存在", v)
				}
			}
			return
		})
	}
	if len(articleIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			articleValues, arcInfoErr := s.articleDao.ArticlesInfo(ctx, articleIDs)
			if arcInfoErr != nil {
				return arcInfoErr
			}
			for _, v := range articleIDs {
				if _, ok := articleValues[v]; !ok {
					return fmt.Errorf("id 为%d的专栏数据不存在", v)
				}
			}
			return nil
		})
	}
	if len(avIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			arcValues, arcErr := s.arcDao.Arcs(ctx, avIDs)
			if arcErr != nil {
				return arcErr
			}
			for _, v := range avIDs {
				if _, ok := arcValues[v]; !ok {
					return fmt.Errorf("id 为%d的稿件数据不存在", v)
				}
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	return
}

// OgvUpdate update ogv
func (s *Service) OgvUpdate(c context.Context, param *show.SearchOgvUP) (err error) {
	var (
		query string
	)
	if err = s.ValidatOgv(c, param.MoreshowStatus, param.GameStatus, param.MoreshowValue, param.Plat, param.GameValue, param.PgcIds, param.PgcMoreURL, param.PgcMoreType); err != nil {
		return
	}
	if err = s.valOgvTitle(param.HdTitle, param.HdSubtitle); err != nil {
		return
	}
	oldOGv, err := s.showDao.SearchOgvFind(param.ID)
	if err != nil {
		return
	}
	//编辑之后 已上线的 保持原状态
	if oldOGv.Check == common.Pass {
		param.Check = common.Pass
	} else {
		//其它卡片状态变为已提交
		param.Check = common.Verify
	}
	if err = s.showDao.SearchOgvUpdate(param); err != nil {
		return
	}
	if query, err = s.getOgvQuery(param.Query); err != nil {
		log.Error("OgvUpdate getOgvQuery error(%v)", err)
		return
	}
	if err = util.AddOgvLogs(param.Person, 0, param.ID, common.ActionUpdate, query, param.HdTitle); err != nil {
		log.Error("OgvUpdate AddOgvLogs error(%v)", err)
		return
	}
	return nil
}

// OgvOpt opt ogv
func (s *Service) OgvOpt(c context.Context, param *show.SearchOgvOption) (err error) {
	var (
		ogvValue *show.SearchOgv
	)
	if ogvValue, err = s.showDao.SearchOgvFind(param.ID); err != nil {
		return
	}
	if err = s.showDao.SearchOgvOption(param); err != nil {
		return
	}
	var (
		action string
	)
	if param.Check == common.Pass {
		action = common.ActionOnline
	} else {
		action = common.ActionOffline
	}
	if err = util.AddOgvLogs(param.Person, 0, param.ID, action, ogvValue.QueryStr, ogvValue.HdTitle); err != nil {
		log.Error("OgvUpdate AddLog error(%v)", err)
		return
	}
	return nil
}
