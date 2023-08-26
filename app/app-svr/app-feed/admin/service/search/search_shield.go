package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"

	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

const (
	_ActUpSearchShield  = "ActUpSearchShield"
	_ActAddSearchShield = "ActAddSearchShield"
)

var (
	_emptySearchShield = make([]*show.SearchShieldQuery, 0)
)

// SearchShieldList search shield list
func (s *Service) SearchShieldList(c context.Context, param *show.SearchShieldLP) (pager *show.SearchShieldPager, err error) {
	var (
		sids []int64
	)
	pager = &show.SearchShieldPager{
		Page: common.Page{
			Num:  param.Pn,
			Size: param.Ps,
		},
	}
	query := s.showDao.DB.Model(&show.SearchShield{})
	if param.ID != 0 {
		query = query.Where("card_value = ?", param.ID)
	}
	if param.CardType != 0 {
		query = query.Where("card_type = ?", param.CardType)
	}
	if param.Check != 0 {
		query = query.Where("`check` = ?", param.Check)
	}
	if param.Query != "" {
		SearchShieldQuery := make([]*show.SearchShieldQuery, 0)
		where := map[string]interface{}{
			"deleted": common.NotDeleted,
		}
		if err = s.showDao.DB.Model(&show.SearchShieldQuery{}).Where(where).Where("value like ?", "%"+param.Query+"%").Find(&SearchShieldQuery).Error; err != nil {
			log.Error("SearchShieldList Find error(%v)", err)
			return
		}
		if len(SearchShieldQuery) == 0 {
			return
		}
		for _, v := range SearchShieldQuery {
			sids = append(sids, v.SID)
		}
		if len(sids) > 0 {
			query = query.Where("id in (?)", sids)
		}
	}
	if err = query.Count(&pager.Page.Total).Error; err != nil {
		log.Error("searchSvc.SearchShieldList count error(%v)", err)
		return
	}
	SearchShields := make([]*show.SearchShield, 0)
	if err = query.Order("`mtime` DESC").Offset((param.Pn - 1) * param.Ps).Limit(param.Ps).Find(&SearchShields).Error; err != nil {
		log.Error("searchSvc.SearchShieldList Find error(%v)", err)
		return
	}
	if len(SearchShields) > 0 {
		var (
			ids      []int64
			queryMap map[int64][]*show.SearchShieldQuery
		)
		for _, v := range SearchShields {
			if v == nil {
				continue
			}
			ids = append(ids, v.ID)
		}
		where := map[string]interface{}{
			"deleted": common.NotDeleted,
		}
		SearchShieldQuery := make([]*show.SearchShieldQuery, 0)
		if err = s.showDao.DB.Model(&show.SearchShieldQuery{}).Where(where).Where("sid in (?)", ids).Find(&SearchShieldQuery).Error; err != nil {
			log.Error("searchSvc.SearchShieldList Find error(%v)", err)
			return
		}
		queryMap = make(map[int64][]*show.SearchShieldQuery, len(SearchShieldQuery))
		for _, v := range SearchShieldQuery {
			queryMap[v.SID] = append(queryMap[v.SID], v)
		}
		for _, v := range SearchShields {
			if value, ok := queryMap[v.ID]; ok {
				v.Query = value
			} else {
				v.Query = _emptySearchShield
			}
		}
	}
	pager.Item = SearchShields
	return
}

// OpenSearchShieldList search shield word list for seach
func (s *Service) OpenSearchShieldList(ctx context.Context) (SearchShields []*show.SearchShield, err error) {
	SearchShields = make([]*show.SearchShield, 0)
	w := map[string]interface{}{
		"check": common.StatusOnline,
	}
	query := s.showDao.DB.Model(&show.SearchShield{})
	if err = query.Where(w).Order("`id` DESC").Find(&SearchShields).Error; err != nil {
		log.Error("OpenSearchShieldList Find error(%v)", err)
		return
	}
	if len(SearchShields) > 0 {
		var (
			ids      []int64
			queryMap map[int64][]*show.SearchShieldQuery
		)
		for _, v := range SearchShields {
			ids = append(ids, v.ID)
		}
		where := map[string]interface{}{
			"deleted": common.NotDeleted,
		}
		SearchShieldQuery := make([]*show.SearchShieldQuery, 0)
		if err = s.showDao.DB.Model(&show.SearchShieldQuery{}).Where(where).Where("sid in (?)", ids).Find(&SearchShieldQuery).Error; err != nil {
			log.Error("searchSvc.OpenSearchShieldList Find error(%v)", err)
			return
		}
		queryMap = make(map[int64][]*show.SearchShieldQuery, len(SearchShieldQuery))
		for _, v := range SearchShieldQuery {
			queryMap[v.SID] = append(queryMap[v.SID], v)
		}
		for _, v := range SearchShields {
			if value, ok := queryMap[v.ID]; ok {
				v.Query = value
			}
			if v.CardType == common.SeaShieldPgc {
				var (
					seasonID    int
					seasonCards map[int32]*seasongrpc.CardInfoProto
				)
				//nolint:gosec
				if seasonID, err = strconv.Atoi(v.CardValue); err != nil {
					log.Error("searchSvc.OpenSearchShieldList strconv Param(%+v) error(%v)", v, err)
					err = nil
					continue
				}
				if seasonCards, err = s.pgcDao.CardsInfoReply(ctx, []int32{int32(seasonID)}); err != nil {
					log.Error("searchSvc.OpenSearchShieldList CardsInfoReply Param(%+v) error(%v)", v, err)
					err = nil
					continue
				}
				if season, ok := seasonCards[int32(seasonID)]; ok {
					v.Season = season
				}
			} else {
				v.Season = struct{}{}
			}
		}
	}
	return
}

// Validate validate search web card
func (s *Service) ValidateShield(c context.Context, p *show.SearchShieldValid) (err error) {
	var (
		querys []*show.SearchShieldQuery
	)
	if err = json.Unmarshal([]byte(p.Query), &querys); err != nil {
		log.Error("ValidateShield json.Unmarshal(%v) error(%v)", p, err)
		return
	}
	if len(querys) == 0 {
		err = fmt.Errorf("query不能为空")
		return
	}
	for _, v := range querys {
		count := 0
		p.Query = v.Value
		if count, err = s.showDao.SearchShieldValid(p); err != nil {
			return
		}
		if count > 0 {
			err = fmt.Errorf("相同卡片类型卡片id(%s) query(%s) 已有运营卡片", p.CardValue, v.Value)
		}
	}
	return
}

// AddSearchShield add search shield
func (s *Service) AddSearchShield(c context.Context, param *show.SearchShieldAP, uid int64) (err error) {
	p := &show.SearchShieldValid{
		Query:     param.Query,
		CardValue: param.CardValue,
		CardType:  param.CardType,
	}
	if err = s.ValidateShield(c, p); err != nil {
		return
	}
	if err = s.showDao.SearchShieldAdd(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogSeashShield, param.Person, uid, 0, _ActAddSearchShield, param); err != nil {
		log.Error("searchSvc.AddSearchShield AddLog error(%v)", err)
		return
	}
	return
}

// UpdateSearchShield update WebSearch
func (s *Service) UpdateSearchShield(c context.Context, param *show.SearchShieldUP, uid int64) (err error) {
	p := &show.SearchShieldValid{
		ID:        param.ID,
		Query:     param.Query,
		CardValue: param.CardValue,
		CardType:  param.CardType,
	}
	if err = s.ValidateShield(c, p); err != nil {
		return
	}
	if err = s.showDao.SearchShieldUpdate(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogSeashShield, param.Person, uid, param.ID, _ActUpSearchShield, param); err != nil {
		log.Error("searchSvc.UpdateSearchShield AddLog error(%v)", err)
		return
	}
	return
}

// OptionSearchShield option WebSearch
func (s *Service) OptionSearchShield(up *show.SearchShieldOption) (err error) {
	if err = s.showDao.SearchShieldOption(up); err != nil {
		return
	}
	var act string
	if up.Check == common.StatusOnline {
		act = common.OptionOnline
	} else {
		act = common.OptionHidden
	}
	if err = util.AddLogs(common.LogSeashShield, up.Name, up.UID, up.ID, act, up); err != nil {
		log.Error("searchSvc.OptionSearchShield AddLog error(%v)", err)
		return
	}
	return
}
