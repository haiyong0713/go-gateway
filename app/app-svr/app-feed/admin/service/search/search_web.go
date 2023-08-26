package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	pb "go-gateway/app/app-svr/app-feed/admin/api/search"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/game"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	showModel "go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/jinzhu/gorm"
)

const (
	// _ActAddSearchWebCard log action
	_ActAddSearchWebCard = "ActAddSearchWebCard"
	// _ActUpSearchWebCard log action
	_ActUpSearchWebCard = "ActUpSearchWebCard"
	// _ActDelSearchWebCard log action
	_ActDelSearchWebCard = "ActDelSearchWebCard"
	// _ActAddSearchWeb log action
	_ActAddSearchWeb = "ActAddSearchWeb"
	// _ActUpSearchWeb log action
	_ActUpSearchWeb = "ActUpSearchWeb"
	// _ActDelSearchWeb log action
	_ActDelSearchWeb = "ActDelSearchWeb"
	// _ActOptSearchWeb log action
	_ActOptSearchWeb = "ActOptSearchWeb"
)

var (
	_emptyWebQuery = make([]*show.SearchWebQuery, 0)
	_emptyWebPlat  = make([]*show.SearchWebPlat, 0)
)

// SearchWebCardList channel SearchWebCard list
func (s *Service) SearchWebCardList(lp *show.SearchWebCardLP) (pager *show.SearchWebCardPager, err error) {
	pager = &show.SearchWebCardPager{
		Page: common.Page{
			Num:  lp.Pn,
			Size: lp.Ps,
		},
	}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	query := s.showDao.DB.Model(&show.SearchWebCard{})
	if lp.ID > 0 {
		w["id"] = lp.ID
	}
	if lp.Person != "" {
		query = query.Where("person like ?", "%"+lp.Person+"%")
	}
	if lp.Title != "" {
		query = query.Where("title like ?", "%"+lp.Title+"%")
	}
	if lp.STime != "" {
		query = query.Where("ctime >= ?", lp.STime)
	}
	if lp.ETime != "" {
		query = query.Where("ctime <= ?", lp.ETime)
	}
	if err = query.Where(w).Count(&pager.Page.Total).Error; err != nil {
		log.Error("searchWebSvc.SearchWebCardList count error(%v)", err)
		return
	}
	SearchWebCards := make([]*show.SearchWebCard, 0)
	if err = query.Where(w).Order("`id` DESC").Offset((lp.Pn - 1) * lp.Ps).Limit(lp.Ps).Find(&SearchWebCards).Error; err != nil {
		log.Error("searchWebSvc.SearchWebCardList Find error(%v)", err)
		return
	}
	pager.Item = SearchWebCards
	return
}

// AddSearchWebCard add channel SearchWebCard
func (s *Service) AddSearchWebCard(c context.Context, param *show.SearchWebCardAP, name string, uid int64) (err error) {
	if err = s.showDao.SearchWebCardAdd(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogSWEBCard, name, uid, 0, _ActAddSearchWebCard, param); err != nil {
		log.Error("searchWebSvc.AddSearchWebCard AddLog error(%v)", err)
		return
	}
	return
}

// UpdateSearchWebCard update channel SearchWebCard
func (s *Service) UpdateSearchWebCard(c context.Context, param *show.SearchWebCardUP, name string, uid int64) (err error) {
	if err = s.showDao.SearchWebCardUpdate(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogSWEBCard, name, uid, param.ID, _ActUpSearchWebCard, param); err != nil {
		log.Error("searchWebSvc.UpdateSearchWebCard AddLog error(%v)", err)
		return
	}
	return
}

// DeleteSearchWebCard delete channel SearchWebCard
func (s *Service) DeleteSearchWebCard(id int64, name string, uid int64) (err error) {
	if err = s.showDao.SearchWebCardDelete(id); err != nil {
		return
	}
	if err = util.AddLogs(common.LogSWEBCard, name, uid, id, _ActDelSearchWebCard, id); err != nil {
		log.Error("searchWebSvc.DeleteSearchWebCard AddLog error(%v)", err)
		return
	}
	return
}

// SearchWebList WebSearch list
//
//nolint:gocognit
func (s *Service) SearchWebList(c context.Context, lp *show.SearchWebLP) (pager *show.SearchWebPager, err error) {
	pager = &show.SearchWebPager{
		Page: common.Page{
			Num:  lp.Pn,
			Size: lp.Ps,
		},
	}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	query := s.showDao.DB.Model(&show.SearchWeb{})
	if lp.ID > 0 {
		query = query.Where("card_value = ?", lp.ID)
	}
	if lp.Person != "" {
		query = query.Where("person like ?", "%"+lp.Person+"%")
	}
	if lp.STime != "" {
		query = query.Where("stime >= ?", lp.STime)
	}
	if lp.ETime != "" {
		query = query.Where("etime <= ?", lp.ETime)
	}
	if lp.CardType != 0 {
		query = query.Where("card_type = ?", lp.CardType)
	}
	if lp.Keyword != "" {
		query = query.Joins("LEFT JOIN search_web_query q ON q.sid = search_web.id").
			Where("q.value = ?", lp.Keyword)
	}
	cTimeStr := util.CTimeStr()
	if lp.Check != 0 {
		if lp.Check == common.Pass {
			// 已通过 未生效
			query = query.Where("`check` = ?", common.Pass)
			query = query.Where("stime > ?", cTimeStr)
		} else if lp.Check == common.Valid {
			// 已通过 已生效
			query = query.Where("`check` = ?", common.Pass)
			query = query.Where("stime <= ?", cTimeStr).Where("etime >= ?", cTimeStr)
		} else if lp.Check == common.InValid {
			// 已通过 已失效
			query = query.Where("(`check` = ? AND etime <= ?) OR (`check` = ?)", common.Pass, cTimeStr, common.InValid)
		} else {
			query = query.Where("`check` = ? ", lp.Check)
		}
	}
	if err = query.Where(w).Count(&pager.Page.Total).Error; err != nil {
		log.Error("searchSvc.SearchWebList count error(%v)", err)
		return
	}
	SearchWebs := make([]*show.SearchWeb, 0)
	if err = query.Where(w).Order("`id` DESC").Offset((lp.Pn - 1) * lp.Ps).Limit(lp.Ps).Find(&SearchWebs).Error; err != nil {
		log.Error("searchSvc.SearchWebList Find error(%v)", err)
		return
	}
	if len(SearchWebs) > 0 {
		var (
			ids      []int64
			queryMap map[int64][]*show.SearchWebQuery
		)
		for _, v := range SearchWebs {
			// 修改卡片状态
			// todo: 不应该和数据库状态有差异
			if v.Check == common.Pass {
				v.Check, v.Status = s.getCheckStatus(v)
			}

			// 根据类型获取卡片内容
			webCard := &show.SearchWebCard{}
			if v.CardType == common.WebSearchSpecialSmall || v.CardType == common.WebSearchVideoSpecialSmall {
				cardWhere := map[string]interface{}{
					"deleted": common.NotDeleted,
					"id":      v.CardValue,
				}
				// 特殊小卡从库里取
				if err = s.showDao.DB.Model(&show.SearchWebCard{}).Where(cardWhere).First(webCard).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						err = nil
					} else {
						log.Error("searchSvc.SearchWebCard Find error(%v)", err)
					}
				}
				v.Card = webCard
			} else if v.CardType == common.WebSearchUpUser {
				var (
					id     int64
					upCard *account.ProfileStatReply
				)
				if id, err = strconv.ParseInt(v.CardValue, 10, 64); err != nil {
					log.Error("searchSvc.SearchWebCard ParseInt(%s) error(%v)", v.CardValue, err)
				} else {
					if upCard, err = s.accDao.ProfileWithStat3(c, id); err != nil {
						log.Error("searchSvc.ProfileWithStat3 id(%v) error(%v)", v.CardValue, err)
					}
				}
				v.Card = upCard
			} else {
				if gameInfo, ok := s.EntryGameCache[v.CardValue]; !ok {
					log.Error("searchSvc.SearchWebCard GameEntryInfo(%v) record not found", v.CardValue)
				} else {
					webCard.Title = gameInfo.Name
				}
				v.Card = webCard
			}
			ids = append(ids, v.ID)
		}
		where := map[string]interface{}{
			"deleted": common.NotDeleted,
		}
		SearchWebQuery := make([]*show.SearchWebQuery, 0)
		if err = s.showDao.DB.Model(&show.SearchWebQuery{}).Where(where).Where("sid in (?)", ids).Find(&SearchWebQuery).Error; err != nil {
			log.Error("searchSvc.SearchWebList Find error(%v)", err)
			return
		}
		queryMap = make(map[int64][]*show.SearchWebQuery, len(SearchWebQuery))
		for _, v := range SearchWebQuery {
			queryMap[v.SID] = append(queryMap[v.SID], v)
		}

		searchWebPlat := make([]*show.SearchWebPlat, 0)
		if err = s.showDao.DB.Model(&show.SearchWebPlat{}).Where(where).Where("sid in (?)", ids).Find(&searchWebPlat).Error; err != nil {
			log.Error("searchSvc.SearchWebList find plat error(%v)", err)
			return
		}
		platMap := make(map[int64][]*show.SearchWebPlat)
		for _, v := range searchWebPlat {
			platMap[v.SId] = append(platMap[v.SId], v)
		}

		for _, v := range SearchWebs {
			if value, ok := queryMap[v.ID]; ok {
				v.Query = value
			} else {
				v.Query = _emptyWebQuery
			}
			if value, ok := platMap[v.ID]; ok {
				v.PlatVer = value
			} else {
				v.PlatVer = _emptyWebPlat
			}
		}
	}
	pager.Item = SearchWebs
	return
}

func (s *Service) getCheckStatus(v *showModel.SearchWeb) (check int, status int) {
	check = v.Check
	status = v.Status

	cur := time.Now().Unix()
	if (cur >= v.Stime.Time().Unix()) && (cur <= v.Etime.Time().Unix()) {
		// 2021M7W4: 未上线的游戏，不变更状态为已生效
		if v.CardType != common.WebSearchGame {
			return common.Valid, status
		}
		if _, ok := s.GameCache[v.CardValue]; !ok {
			log.Warn("getCheckStatus game card SearchGame id(%v) unreleased", v.CardValue)
			return
		}
		check = common.Valid
	} else if cur > v.Etime.Time().Unix() && v.Check != common.InValid {
		return common.InValid, common.StatusDownline
	}
	return
}

// OpenSearchWebList WebSearch list
func (s *Service) OpenSearchWebList(c context.Context) (ret []*show.OpenSearchWeb, err error) {
	cTimeStr := util.CTimeStr()
	SearchWebs := make([]*show.SearchWeb, 0)
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
		"check":   common.Pass,
	}
	query := s.showDao.DB.Model(&show.SearchWeb{})
	// 已通过 已生效
	query = query.Where("stime <= ?", cTimeStr).Where("etime >= ?", cTimeStr)
	if err = query.Where(w).Order("`id` DESC").Find(&SearchWebs).Error; err != nil {
		log.Error("searchSvc.OpenSearchWebList Find error(%v)", err)
		return
	}
	if len(SearchWebs) > 0 {
		var (
			ids      []int64
			queryMap map[int64][]*show.SearchWebQuery
		)
		for _, v := range SearchWebs {
			webCard := &show.SearchWebCard{}
			cardWhere := map[string]interface{}{
				"deleted": common.NotDeleted,
				"id":      v.CardValue,
			}
			if err = s.showDao.DB.Model(&show.SearchWebCard{}).Where(cardWhere).First(webCard).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					err = nil
					webCard = nil
				} else {
					log.Error("searchSvc.OpenSearchWebList Find error(%v)", err)
				}
			}
			if webCard != nil {
				v.Card = webCard
			} else {
				v.Card = struct{}{}
			}
			ids = append(ids, v.ID)
		}
		where := map[string]interface{}{
			"deleted": common.NotDeleted,
		}
		SearchWebQuery := make([]*show.SearchWebQuery, 0)
		if err = s.showDao.DB.Model(&show.SearchWebQuery{}).Where(where).Where("sid in (?)", ids).Find(&SearchWebQuery).Error; err != nil {
			log.Error("searchSvc.OpenSearchWebList Find error(%v)", err)
			return
		}
		queryMap = make(map[int64][]*show.SearchWebQuery, len(SearchWebQuery))
		for _, v := range SearchWebQuery {
			queryMap[v.SID] = append(queryMap[v.SID], v)
		}

		var searchWebPlat []*show.SearchWebPlat
		if err = s.showDao.DB.Model(&show.SearchWebPlat{}).Where(where).Where("sid IN (?)", ids).Scan(&searchWebPlat).Error; err != nil {
			log.Error("searchSvc.OpenSearchWebList plat Scan error(%v)", err)
			return
		}
		platMap := make(map[int64][]*show.SearchWebPlat)
		for _, v := range searchWebPlat {
			platMap[v.SId] = append(platMap[v.SId], v)
		}

		for _, v := range SearchWebs {
			if value, ok := queryMap[v.ID]; ok {
				v.Query = value
			} else {
				v.Query = _emptyWebQuery
			}
			if value, ok := platMap[v.ID]; ok {
				v.PlatVer = value
			} else {
				v.PlatVer = _emptyWebPlat
			}
		}
		// 2021M7W4: 游戏卡未上线的不下发给AI
		ret = make([]*show.OpenSearchWeb, 0, len(SearchWebs))
		for _, v := range SearchWebs {
			if v.CardType == common.WebSearchGame && v.Check == common.Pass {
				if _, ok := s.GameCache[v.CardValue]; !ok {
					log.Warn("searchSvc.OpenSearchWebList game(%v) unreleased", v.CardValue)
					continue
				}
			}
			ret = append(ret, v.Convert())
		}
	}
	return
}

// Validate validate search web card
func (s *Service) Validate(c context.Context, p *show.SWTimeValid) (err error) {
	var (
		querys   []*show.SearchWebQuery
		webCard  *showModel.SearchWebCard
		upCard   *account.Card
		gameCard *game.EntryInfo
		id       int64
		ok       bool
	)
	if id, err = strconv.ParseInt(p.CardValue, 10, 64); err != nil {
		return
	}
	if p.CardType == common.WebSearchSpecialSmall || p.CardType == common.WebSearchVideoSpecialSmall {
		if webCard, err = s.showDao.SWBFindByID(id); err != nil {
			return err
		}
		if webCard == nil {
			return fmt.Errorf("无效web卡片ID(%d)", id)
		}
	} else if p.CardType == common.WebSearchGame {
		if gameCard, ok = s.EntryGameCache[p.CardValue]; !ok {
			return
		}
		if gameCard == nil || gameCard.Name == "" {
			return fmt.Errorf("无效游戏卡片ID(%d)", id)
		}
	} else if p.CardType == common.WebSearchUpUser {
		if upCard, err = s.accDao.Card3(c, id); err != nil {
			return err
		}
		if upCard == nil {
			return fmt.Errorf("无效UP主卡片ID(%d)", id)
		}
	} else {
		return fmt.Errorf("参数错误ID(%d)", id)
	}
	if err = json.Unmarshal([]byte(p.Query), &querys); err != nil {
		log.Error("searchSvc.Validate json.Unmarshal(%v) error(%v)", p, err)
		return
	}
	if len(querys) == 0 {
		err = fmt.Errorf("query不能为空")
		return
	}
	for _, v := range querys {
		for _, plat := range p.PlatVer {
			count := 0
			p.Query = v.Value
			p.Plat = plat.Plat
			if count, err = s.showDao.SWTimeValid(p); err != nil {
				return
			}
			if count > 0 {
				err = fmt.Errorf("平台(%s)相同query(%s)该位置已有运营卡片", common.PlatDict[p.Plat], v.Value)
			}
			if err != nil {
				return
			}
		}
	}
	return
}

// AddSearchWeb add WebSearch
func (s *Service) AddSearchWeb(c context.Context, param *show.SearchWebAP, name string, uid int64) (err error) {
	p := &show.SWTimeValid{
		Priority:  param.Priority,
		STime:     param.Stime,
		ETime:     param.Etime,
		Query:     param.Query,
		CardValue: param.CardValue,
		CardType:  param.CardType,
		PlatVer:   param.PlatVer,
	}
	if err = s.Validate(c, p); err != nil {
		return
	}
	if err = s.showDao.SearchWebAdd(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogSWEB, name, uid, 0, _ActAddSearchWeb, param); err != nil {
		log.Error("searchSvc.AddSearchWeb AddLog error(%v)", err)
		return
	}
	return
}

// UpdateSearchWeb update WebSearch
func (s *Service) UpdateSearchWeb(c context.Context, param *show.SearchWebUP, name string, uid int64) (err error) {
	var (
		swValue *show.SearchWeb
	)
	p := &show.SWTimeValid{
		ID:        param.ID,
		Priority:  param.Priority,
		STime:     param.Stime,
		ETime:     param.Etime,
		Query:     param.Query,
		CardValue: param.CardValue,
		CardType:  param.CardType,
		PlatVer:   param.PlatVer,
	}
	if err = s.Validate(c, p); err != nil {
		return
	}
	if swValue, err = s.showDao.SWFindByID(param.ID); err != nil {
		log.Error("searchSvc.UpdateSearchWeb AddLog error(%v)", err)
		return
	}
	log.Info("searchSvc.UpdateSearchWeb current web config(%+v)", swValue)

	// [MGR] Y21M4W2逻辑：编辑后所有状态都会修改为【待审核】
	param.Check = common.Verify
	param.Status = common.StatusDownline

	if err = s.showDao.SearchWebUpdate(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogSWEB, name, uid, param.ID, _ActUpSearchWeb, param); err != nil {
		log.Error("searchSvc.UpdateSearchWeb AddLog error(%v)", err)
		return
	}
	return
}

// DeleteSearchWeb delete WebSearch
func (s *Service) DeleteSearchWeb(id int64, name string, uid int64) (err error) {
	if err = s.showDao.SearchWebDelete(id); err != nil {
		return
	}
	if err = util.AddLogs(common.LogSWEB, name, uid, id, _ActDelSearchWeb, id); err != nil {
		log.Error("searchSvc.DeleteSearchWeb AddLog error(%v)", err)
		return
	}
	return
}

// OptionSearchWeb option WebSearch
func (s *Service) OptionSearchWeb(id int64, opt string, name string, uid int64) (err error) {
	up := &show.SearchWebOption{}
	if opt == common.OptionOnline {
		up.Status = common.StatusOnline
		up.Check = common.Pass
	} else if opt == common.OptionHidden {
		up.Status = common.StatusDownline
		up.Check = common.InValid
	} else if opt == common.OptionPass {
		up.Status = common.StatusOnline
		up.Check = common.Pass
	} else if opt == common.OptionReject {
		up.Status = common.StatusDownline
		up.Check = common.Rejecte
	} else {
		err = fmt.Errorf("参数不合法")
		return
	}
	up.ID = id
	if err = s.showDao.SearchWebOption(up); err != nil {
		return
	}
	logParam := map[string]interface{}{
		"id":  id,
		"opt": opt,
		"up":  up,
	}
	if err = util.AddLogs(common.LogSWEB, name, uid, id, _ActOptSearchWeb, logParam); err != nil {
		log.Error("searchSvc.OptionSearchWeb AddLog error(%v)", err)
		return
	}
	return
}

// BatchOptSearchWeb 批量审核web端配置
func (s *Service) BatchOptWeb(c *bm.Context, req *pb.BatchOptWebReq) (resp *pb.BatchOptWebResp, err error) {
	resp = &pb.BatchOptWebResp{}
	if resp.InvalidIds, err = s.checkBatchOptWebIds(c, req); err != nil {
		log.Errorc(c, "s.BatchOptWeb checkBatchOptIds req (%+v) err(%v)", req, err)
		return
	}
	if len(resp.InvalidIds) > 0 {
		log.Errorc(c, "s.BatchOptWeb found invalid ids: %+v", resp.InvalidIds)
		return
	}
	if err = s.showDao.SearchOptWeb(c, req.Ids, req.Option); err != nil {
		log.Errorc(c, "s.BatchOptWeb call dao.SearchOptWeb ids(%+v) err(%v)", req.Ids, err)
		return
	}

	//TODO: add action logs
	return
}

func (s *Service) ReleaseSearchWeb(c *bm.Context) (resp *empty.Empty, err error) {
	if err = s.showDao.ReleaseSearchWeb(c); err != nil {
		log.Error("s.ReleaseSearchWeb call dao.SearchWebAll err(%v)", err)
		return nil, err
	}
	return
}

func (s *Service) checkBatchOptWebIds(c *bm.Context, req *pb.BatchOptWebReq) (invalidIds []*pb.BatchInvalidItem, err error) {
	var (
		options map[int64]*showModel.SearchWebOption
	)
	if options, err = s.showDao.SearchWebOptionQueryById(c, req.Ids); err != nil {
		return
	}

	invalidIds = make([]*pb.BatchInvalidItem, 0, len(req.Ids))
	for _, id := range req.Ids {
		var ok bool
		var option *show.SearchWebOption
		var invalidItem = &pb.BatchInvalidItem{Id: id}

		if option, ok = options[id]; !ok {
			invalidItem.Msg = "web端配置ID不存在"
			invalidIds = append(invalidIds, invalidItem)
		} else {
			switch req.Option {
			case common.OptionBatchPass, common.OptionBatchReject:
				if option.Check != common.Verify {
					invalidItem.Msg = "web端配置不是待审核状态"
					invalidIds = append(invalidIds, invalidItem)
				}
			case common.OptionBatchHidden:
				if option.Check != common.Pass && option.Check != common.Valid {
					invalidItem.Msg = "web端配置不是已通过/已生效状态"
					invalidIds = append(invalidIds, invalidItem)
					continue
				}
			}
			//TODO: permit check
		}
	}
	return
}
