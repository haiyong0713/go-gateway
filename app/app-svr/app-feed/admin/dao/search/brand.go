package search

import (
	"context"
	"encoding/json"
	"fmt"

	model "go-gateway/app/app-svr/app-feed/admin/model/search"
	"go-gateway/app/app-svr/app-feed/ecode"
)

const (
	STATE_ENABLED  = 1
	STATE_DISABLED = 2
)

const (
	ROW_DELETED_TRUE  = 1
	ROW_DELETED_FALSE = 0
)
const (
	ORDER_ASC  = 1
	ORDER_DESC = 2
)

const (
	_blacklistQueryAdd = `INSERT INTO search_brand_blacklist_query(blacklist_id,query,state,c_uname,m_uname) VALUES %s`
)

func (d *Dao) BrandBlacklistAdd(c context.Context, req *model.BrandBlacklistAddReq) (id int64, enabledQuery []string, err error) {
	var (
		batchSql  string
		queryList []string
	)

	// Parse query list and check if already enabled
	if queryList, err = d.parseQueryList(req.QueryList); err != nil {
		return
	}
	if len(queryList) == 0 {
		return 0, nil, ecode.BrandBlacklistQueryInvalid
	}
	if enabledQuery, err = d.GetEnabledQuery(c, 0, queryList); err != nil {
		return
	}
	if len(enabledQuery) > 0 {
		return 0, enabledQuery, ecode.BrandBlacklistQueryExists
	}

	// start transaction
	tx := d.DB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// insert search blacklist
	list := &model.BrandBlacklist{
		Desc:   req.Desc,
		CUname: req.Username,
		MUname: req.Username,
		State:  STATE_ENABLED,
	}
	if err = tx.Create(list).Error; err != nil {
		return
	}

	// batch insert search blacklist query
	if batchSql, err = d.batchAddBlacklistQuery(req.Username, list.Id, queryList, STATE_ENABLED); err != nil {
		return
	}
	if err = tx.Exec(batchSql).Error; err != nil {
		return
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	return list.Id, enabledQuery, nil
}

func (d *Dao) BrandBlacklistEdit(c context.Context, req *model.BrandBlacklistEditReq) (enabledQuery []string, oldBlacklistItem *model.BrandBlacklistItem, err error) {
	var (
		batchSql  string
		queryList []string

		oldBlacklist      = &model.BrandBlacklist{}
		oldBlacklistQuery []*model.BrandBlacklistQuery
	)

	// Parse query list and check if already enabled
	if queryList, err = d.parseQueryList(req.QueryList); err != nil {
		return
	}
	if enabledQuery, err = d.GetEnabledQuery(c, req.BlacklistId, queryList); err != nil {
		return
	}

	// start transaction
	tx := d.DB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// update blacklist record
	db1 := tx.Model(&model.BrandBlacklist{}).
		Where("deleted = ?", ROW_DELETED_FALSE).
		Where("id = ?", req.BlacklistId).
		Scan(oldBlacklist)
	if db1.RecordNotFound() {
		return nil, nil, ecode.BrandBlacklistNotFound
	}
	if oldBlacklist.State == STATE_ENABLED && len(enabledQuery) > 0 {
		return enabledQuery, nil, ecode.BrandBlacklistQueryExists
	}

	err = db1.Update(map[string]interface{}{
		"desc":    req.Desc,
		"m_uname": req.Username,
	}).Error
	if err != nil {
		return nil, nil, err
	}

	// delete old blacklist query records
	err = tx.Model(&model.BrandBlacklistQuery{}).
		Where("deleted = ?", ROW_DELETED_FALSE).
		Where("blacklist_id = ?", req.BlacklistId).
		Scan(&oldBlacklistQuery).
		Update(map[string]interface{}{
			"deleted": ROW_DELETED_TRUE,
			"m_uname": req.Username,
		}).Error
	if err != nil {
		return
	}

	// batch insert blacklist query
	if batchSql, err = d.batchAddBlacklistQuery(req.Username, req.BlacklistId, queryList, oldBlacklist.State); err != nil {
		return
	}
	if err = tx.Exec(batchSql).Error; err != nil {
		return
	}

	// commit transaction
	if err = tx.Commit().Error; err != nil {
		return
	}
	oldBlacklistItem = d.convert2Item(oldBlacklist, oldBlacklistQuery)
	return
}

func (d *Dao) BrandBlacklistOption(c context.Context, req *model.BrandBlacklistOptionReq) (enabledQueryWords []string, affectedQuery []*model.BrandBlacklistQuery, err error) {
	tx := d.DB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// if option is "enable", check if any query is being enabled
	if req.Option == STATE_ENABLED {
		var (
			queryList []string
		)
		if queryList, err = d.GetQueryListById(c, req.BlacklistId); err != nil {
			return
		}
		if enabledQueryWords, err = d.GetEnabledQuery(c, req.BlacklistId, queryList); err != nil {
			return
		}
		if len(enabledQueryWords) > 0 {
			return enabledQueryWords, affectedQuery, ecode.BrandBlacklistQueryExists

		}
	}

	// update blacklist state
	err1 := tx.Model(&model.BrandBlacklist{}).
		Where("deleted = ?", ROW_DELETED_FALSE).
		Where("id = ?", req.BlacklistId).
		Scan(&model.BrandBlacklist{}).
		Update(map[string]interface{}{
			"state":   req.Option,
			"m_uname": req.Username,
		})
	if err1.RecordNotFound() {
		return nil, affectedQuery, ecode.BrandBlacklistNotFound
	}
	if err != nil {
		return
	}

	// update blacklist query state
	err = tx.Model(&model.BrandBlacklistQuery{}).
		Where("deleted = ?", ROW_DELETED_FALSE).
		Where("blacklist_id = ?", req.BlacklistId).
		Scan(&affectedQuery).
		Update(map[string]interface{}{
			"State":   req.Option,
			"m_uname": req.Username,
		}).Error
	if err != nil {
		return
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	return
}

func (d *Dao) BrandBlacklistList(c context.Context, req *model.BrandBlacklistListReq) (total int, ret []*model.BrandBlacklistItem, err error) {
	var (
		offset, limit int
	)

	// paginate
	offset, limit, err = d.paginate(req.Pn, req.Ps)
	if err != nil {
		return
	}

	// assemble sql conditions
	db1 := d.DB.Table("search_brand_blacklist bl").
		Where("bl.deleted = ?", ROW_DELETED_FALSE)
	if len(req.Keyword) > 0 {
		db1 = db1.Joins("left join search_brand_blacklist_query q ON q.blacklist_id = bl.id AND q.deleted = ?", ROW_DELETED_FALSE)
		db1 = db1.Where("q.query like ?", "%"+req.Keyword+"%")
	}
	if req.State > 0 {
		db1 = db1.Where("bl.state = ?", req.State)
	}
	if req.Order == ORDER_ASC {
		db1 = db1.Order("bl.mtime asc")
	} else if req.Order == ORDER_DESC {
		db1 = db1.Order("bl.mtime desc")
	} else {
		db1 = db1.Order("bl.id desc")
	}

	// get blacklist records
	err = db1.Select("count(DISTINCT(bl.id))").Count(&total).
		Offset(offset).Limit(limit).
		Select("DISTINCT(bl.id) as blacklist_id, bl.desc, bl.state, bl.c_uname, bl.m_uname, bl.ctime, bl.mtime").
		Scan(&ret).
		Error
	if err != nil {
		return
	}

	// get blacklist query records
	for _, v := range ret {
		err = d.DB.Model(&model.BrandBlacklistQuery{}).
			Where("deleted = ?", ROW_DELETED_FALSE).
			Where("blacklist_id = ?", v.BlacklistId).
			Pluck("query", &v.QueryList).
			Error
		if err != nil {
			return
		}
	}

	return
}

func (d *Dao) GetEnabledQuery(c context.Context, exceptBlacklistId int64, queryList []string) (ret []string, err error) {
	var obj []*model.BrandBlacklistQuery
	db := d.DB.Model(&model.BrandBlacklistQuery{}).
		Where("deleted = ?", ROW_DELETED_FALSE).
		Where("state = ?", STATE_ENABLED)
	if exceptBlacklistId > 0 {
		db = db.Where("blacklist_id != ?", exceptBlacklistId)
	}
	if len(queryList) > 0 {
		db = db.Where("query in (?)", queryList)
	}
	if err = db.Select("query").Scan(&obj).Error; err != nil {
		return
	}

	ret = make([]string, len(obj))
	for i, v := range obj {
		ret[i] = v.Query
	}
	return
}

func (d *Dao) GetQueryListById(c context.Context, blacklistId int64) (ret []string, err error) {
	var obj []*model.BrandBlacklistQuery
	err = d.DB.Model(&model.BrandBlacklistQuery{}).
		Where("deleted = ?", ROW_DELETED_FALSE).
		Where("blacklist_id = ?", blacklistId).
		Select("query").
		Scan(&obj).Error
	if err != nil {
		return
	}

	ret = make([]string, len(obj))
	for i, v := range obj {
		ret[i] = v.Query
	}
	return
}

func (d *Dao) batchAddBlacklistQuery(username string, blacklistId int64, queryList []string, state int32) (sql string, err error) {
	if len(queryList) == 0 {
		return "", ecode.BrandBlacklistQueryInvalid
	}
	for _, query := range queryList {
		if len(sql) != 0 {
			sql = sql + ","
		}
		sql += fmt.Sprintf(`(%d,"%s",%d,"%s","%s")`, blacklistId, query, state, username, username)
	}
	sql = fmt.Sprintf(_blacklistQueryAdd, sql)
	return
}

func (d *Dao) parseQueryList(src string) (ret []string, err error) {
	err = json.Unmarshal([]byte(src), &ret)
	return
}

func (d *Dao) convert2Item(bl *model.BrandBlacklist, query []*model.BrandBlacklistQuery) (ret *model.BrandBlacklistItem) {
	list := make([]string, len(query))
	for i, v := range query {
		list[i] = v.Query
	}
	ret = &model.BrandBlacklistItem{
		BlacklistId: bl.Id,
		QueryList:   list,
		Desc:        bl.Desc,
		State:       bl.State,
		CUname:      bl.CUname,
		MUname:      bl.MUname,
		Ctime:       bl.Ctime,
		Mtime:       bl.Mtime,
	}
	return
}

func (d *Dao) paginate(pn, ps int) (offset, limit int, err error) {
	limit = ps
	if pn < 1 {
		err = ecode.PageInvalid
		return
	}

	offset = (pn - 1) * ps
	return
}
