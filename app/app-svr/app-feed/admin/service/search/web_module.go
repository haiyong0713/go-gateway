package search

import (
	"context"
	"encoding/json"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// SearchWebModuleQuerys .
func (s *Service) SearchWebModuleQuerys(ids []int64) (res map[int64][]*show.SearchWebModuleQuery, err error) {
	var (
		querys []*show.SearchWebModuleQuery
	)
	res = make(map[int64][]*show.SearchWebModuleQuery)
	where := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	dbQuery := s.showDao.DB.Model(&show.SearchWebModuleQuery{}).Where(where).Where("sid in (?)", ids)
	if err = dbQuery.Find(&querys).Error; err != nil {
		log.Error("WebModuleQuerys Find param(%v) error(%v)", ids, err)
		return
	}
	if len(querys) > 0 {
		for _, v := range querys {
			res[v.Sid] = append(res[v.Sid], v)
		}
	}
	return
}

// SearchWebModuleModules .
func (s *Service) SearchWebModuleModules(ids []int64) (res map[int64][]*show.SearchWebModuleModule, err error) {
	var (
		moduless []*show.SearchWebModuleModule
	)
	res = make(map[int64][]*show.SearchWebModuleModule)
	where := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	dbQuery := s.showDao.DB.Model(&show.SearchWebModuleModule{}).Where(where).Where("sid in (?)", ids)
	if err = dbQuery.Order("`order` ASC").Find(&moduless).Error; err != nil {
		log.Error("WebModuleModules Find param(%v) error(%v)", ids, err)
		return
	}
	if len(moduless) > 0 {
		for _, v := range moduless {
			res[v.Sid] = append(res[v.Sid], v)
		}
	}
	return
}

// UpdateSearchWebModule search special list
func (s *Service) UpdateSearchWebModule(c context.Context, param *show.SearchWebModuleUP) (err error) {
	if err = s.showDao.WebModuleUpdate(param); err != nil {
		return
	}
	var (
		module string
		querys []string
	)
	if module, err = s.getModuleName(param.Module); err != nil {
		return
	}
	if querys, err = s.getQuery(param.Query); err != nil {
		return
	}
	query := strings.Join(querys, "")
	if err = util.AddWebModuleLogs(common.LogWebSerModule, param.UserName, 0, param.ID, common.ActionUpdate, query, module); err != nil {
		log.Error("WebModuleUpdate AddLog error(%v)", err)
		return
	}
	return nil
}

// AddSearchWebModule search special add
func (s *Service) AddSearchWebModule(c context.Context, param *show.SearchWebModuleAP) (err error) {
	param.Check = common.StatusOnline
	if err = s.showDao.WebModuleAdd(param); err != nil {
		return
	}
	var (
		module string
		querys []string
	)
	if module, err = s.getModuleName(param.Module); err != nil {
		return
	}
	if querys, err = s.getQuery(param.Query); err != nil {
		return
	}
	query := strings.Join(querys, "")
	if err = util.AddWebModuleLogs(common.LogWebSerModule, param.UserName, 0, param.ID, common.ActionAdd, query, module); err != nil {
		log.Error("WebModuleAdd AddLog error(%v)", err)
		return
	}
	return
}

// getModuleName.
func (s *Service) getModuleName(query string) (module string, err error) {
	var querys []*show.SearchWebModuleQuery
	if err := json.Unmarshal([]byte(query), &querys); err != nil {
		log.Error("getModuleName json.Unmarshal(%s) error(%v)", query, err)
		return "", err
	}
	//	//模块 1-游戏小卡 2-特殊小卡 3-视频聚合卡 4-影视综合卡 5-番剧卡 6-番剧卡
	for _, v := range querys {
		switch v.Value {
		case "1":
			module += " 游戏小卡 "
		case "2":
			module += " 特殊小卡 "
		case "3":
			module += " 视频聚合卡 "
		case "4":
			module += " 影视综合卡 "
		case "5":
			module += " 番剧卡 "
		case "6":
			module += " 用户卡 "
		}
	}
	return
}

// getModuleName.
func (s *Service) getQuery(query string) (querys []string, err error) {
	var (
		newQuerys []*show.SearchWebModuleQuery
	)
	if err = json.Unmarshal([]byte(query), &newQuerys); err != nil {
		return
	}
	for _, v := range newQuerys {
		querys = append(querys, v.Value)
	}
	return
}

// OptionSearchWebModule option search web
func (s *Service) OptionSearchWebModule(up *show.SearchWebModuleOption) (err error) {
	if err = s.showDao.DB.Model(&show.SearchWebModuleOption{}).Update(up).Error; err != nil {
		log.Error("dao.SearchShieldOption Updates(%+v) error(%v)", up, err)
	}
	var (
		action string
	)
	if up.Check == common.StatusOnline {
		action = common.ActionOnline
	} else {
		action = common.ActionOffline
	}
	if err = util.AddLogs(common.LogWebSerModule, up.UserName, up.UID, up.ID, action, up); err != nil {
		log.Error("searchSvc.OptionSearchShield AddLog error(%v)", err)
		return
	}
	return
}

// SearchWebModuleList search special list
func (s *Service) OpenSearchWebModule(c context.Context, param *show.SearchWebModuleLP) (pager *show.SearchWebModulePager, err error) {
	var (
		mapQuery  map[int64][]*show.SearchWebModuleQuery
		mapModule map[int64][]*show.SearchWebModuleModule
	)
	pager = &show.SearchWebModulePager{
		Item: make([]*show.SearchWebModule, 0),
		Page: common.Page{
			Num:  param.Pn,
			Size: param.Ps,
		},
	}
	where := map[string]interface{}{
		"check": common.StatusOnline,
	}
	query := s.showDao.DB.Model(&show.SearchWebModule{}).Where(where)
	if err = query.Count(&pager.Page.Total).Error; err != nil {
		log.Error("OpenSearchWebModule count error(%v)", err)
		return
	}
	WebModules := make([]*show.SearchWebModule, 0)
	if err = query.Order("`id` DESC").Offset((param.Pn - 1) * param.Ps).Limit(param.Ps).Find(&WebModules).Error; err != nil {
		log.Error("OpenSearchWebModule Find error(%v)", err)
		return
	}
	var (
		ids []int64
	)
	for _, v := range WebModules {
		ids = append(ids, v.ID)
	}
	if len(ids) > 0 {
		if mapQuery, err = s.SearchWebModuleQuerys(ids); err != nil {
			log.Error("OpenSearchWebModule WebModuleQuerys param(%v) error(%v)", ids, err)
		}
	}
	if len(ids) > 0 {
		if mapModule, err = s.SearchWebModuleModules(ids); err != nil {
			log.Error("OpenSearchWebModule WebModuleModules param(%v) error(%v)", ids, err)
		}
	}
	for _, special := range WebModules {
		id := special.ID
		if v, ok := mapQuery[id]; ok {
			special.Querys = v
		}
		if v, ok := mapModule[id]; ok {
			special.Modules = v
		}
	}
	pager.Item = WebModules
	return
}

// SearchWebModuleList search special list
func (s *Service) SearchWebModuleList(c context.Context, param *show.SearchWebModuleLP) (pager *show.SearchWebModulePager, err error) {
	var (
		sids      []int64
		mapQuery  map[int64][]*show.SearchWebModuleQuery
		mapModule map[int64][]*show.SearchWebModuleModule
	)
	pager = &show.SearchWebModulePager{
		Item: make([]*show.SearchWebModule, 0),
		Page: common.Page{
			Num:  param.Pn,
			Size: param.Ps,
		},
	}
	query := s.showDao.DB.Model(&show.SearchWebModule{})
	if param.Check != 0 {
		query = query.Where("`check` = ?", param.Check)
	}
	if param.Query != "" {
		WebModuleQuery := make([]*show.SearchWebModuleQuery, 0)
		where := map[string]interface{}{
			"deleted": common.NotDeleted,
		}
		if err = s.showDao.DB.Model(&show.SearchWebModuleQuery{}).Where(where).Where("value like ?", "%"+param.Query+"%").Find(&WebModuleQuery).Error; err != nil {
			log.Error("WebModuleList Find error(%v)", err)
			return
		}
		for _, v := range WebModuleQuery {
			sids = append(sids, v.Sid)
		}
		if len(sids) == 0 {
			return
		}
		query = query.Where("id in (?)", sids)
	}
	if err = query.Count(&pager.Page.Total).Error; err != nil {
		log.Error("WebModuleList count error(%v)", err)
		return
	}
	WebModules := make([]*show.SearchWebModule, 0)
	if err = query.Order("`id` DESC").Offset((param.Pn - 1) * param.Ps).Limit(param.Ps).Find(&WebModules).Error; err != nil {
		log.Error("WebModuleList Find error(%v)", err)
		return
	}
	var (
		ids []int64
	)
	for _, v := range WebModules {
		ids = append(ids, v.ID)
	}
	if len(ids) > 0 {
		if mapQuery, err = s.SearchWebModuleQuerys(ids); err != nil {
			log.Error("WebModuleList WebModuleQuerys param(%v) error(%v)", ids, err)
		}
	}
	if len(ids) > 0 {
		if mapModule, err = s.SearchWebModuleModules(ids); err != nil {
			log.Error("WebModuleList WebModuleModules param(%v) error(%v)", ids, err)
		}
	}
	for _, special := range WebModules {
		id := special.ID
		if v, ok := mapQuery[id]; ok {
			special.Querys = v
		}
		if v, ok := mapModule[id]; ok {
			special.Modules = v
		}
	}
	pager.Item = WebModules
	return
}
