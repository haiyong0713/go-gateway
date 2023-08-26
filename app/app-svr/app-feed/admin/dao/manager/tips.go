package manager

import (
	"context"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"

	model "go-gateway/app/app-svr/app-feed/admin/model/tips"
)

// 新增一条 tip
func (d *Dao) InsertTip(c context.Context, tip model.SearchTipDB, queryList []model.SearchTipQueryDB) (err error) {
	tx := d.DBShow.BeginTx(c, nil)
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit().Error
		//nolint:gosimple
		return
	}()

	query := tx.Model(&model.SearchTipDB{})
	if err = query.Create(&tip).Error; err != nil {
		log.Error("tips dao tip Create() error(%v)", err)
		return
	}
	log.Info("tips dao tip Create() success(%v)", tip)

	for _, v := range queryList {
		insertVal := model.SearchTipQueryDB{
			SearchWord:  v.SearchWord,
			SearchTipID: tip.ID,
		}
		query = tx.Model(model.SearchTipQueryDB{})
		if err = query.Create(&insertVal).Error; err != nil {
			log.Error("tips dao queryList Create() error(%v)", err)
			return
		}
	}
	log.Info("tips dao queryList Create() success(%v)", queryList)

	return
}

// 更新一条 tip
func (d *Dao) UpdateTip(c context.Context, newTip model.SearchTipDB, newQueryList []model.SearchTipQueryDB) (err error) {
	id := newTip.ID
	if id == 0 {
		err = ecode.Error(ecode.RequestErr, "lost param id")
		return
	}
	tx := d.DBShow.BeginTx(c, nil)
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit().Error
		//nolint:gosimple
		return
	}()

	// 找出需要变更的配置
	var currentTip model.SearchTipDB
	query := tx.Model(&model.SearchTipDB{}).Where("id = ?", id)
	if err = query.Find(&currentTip).Error; err != nil {
		err = ecode.Error(ecode.RequestErr, "id not exist")
		return
	}

	// 更新配置
	updateMap := map[string]interface{}{
		"title":         newTip.Title,
		"sub_title":     newTip.SubTitle,
		"is_immediate":  newTip.IsImmediate,
		"stime":         newTip.STime,
		"etime":         newTip.ETime,
		"online_status": newTip.Status,
		"has_bg_img":    newTip.HasBgImg,
		"jump_url":      newTip.JumpUrl,
		"plat":          0,
	}
	if err = query.Updates(updateMap).Error; err != nil {
		log.Error("tips dao tip Updates() error(%v) id(%v)", err, id)
		return
	}
	log.Info("tips dao tip Updates() success(%v) id(%v)", updateMap, id)

	// 找出当前query
	var currentQueryList []model.SearchTipQueryDB
	query = tx.Model(&model.SearchTipQueryDB{}).Where("s_t_id = ?", id).Where("deleted = ?", 0)
	if err = query.Find(&currentQueryList).Error; err != nil {
		log.Error("tips dao tip Find() error(%v) id(%v)", err, id)
		return
	}

	var currentQueryMap = map[string]int64{}
	for _, currentQuery := range currentQueryList {
		currentQueryMap[currentQuery.SearchWord] = currentQuery.ID
	}

	// 生成需要删除和插入的query数据
	var newQueryMap = map[string]bool{}
	for _, newQuery := range newQueryList {
		newQueryMap[newQuery.SearchWord] = true
		if _, ok := currentQueryMap[newQuery.SearchWord]; !ok {
			needInsertQuery := model.SearchTipQueryDB{
				SearchTipID: id,
				SearchWord:  newQuery.SearchWord,
			}

			// 插入query
			query = tx.Model(&model.SearchTipQueryDB{})
			if err = query.Create(&needInsertQuery).Error; err != nil {
				log.Error("tips dao tip Create() error(%v) id(%v)", err, id)
				return
			}
			log.Info("tips dao newQueryList Create() success(%v) id(%v)", needInsertQuery, id)
		}
	}

	for _, currentQuery := range currentQueryList {
		if _, ok := newQueryMap[currentQuery.SearchWord]; !ok {
			// 删除query
			query = tx.Model(&model.SearchTipQueryDB{}).Where("id = ?", currentQuery.ID)
			if err = query.Updates(model.SearchTipQueryDB{
				Deleted: 1,
			}).Error; err != nil {
				log.Error("tips dao tip Updates() error(%v) id(%v)", err, id)
				return
			}
			log.Info("tips dao queryList Delete() success(%v) id(%v)", currentQuery, id)
		}
	}

	return
}

// 更新上下线状态
func (d *Dao) UpdateTipOperation(c context.Context, id []int64, operation int) (err error) {
	// 找出需要变更的配置
	var currentTip model.SearchTipDB
	query := d.DBShow.Model(&model.SearchTipDB{}).Where("id in (?)", id)
	if err = query.Find(&currentTip).Error; err != nil {
		err = ecode.Error(ecode.RequestErr, "id not exist")
		return
	}

	updateTip := map[string]interface{}{
		"online_status": operation,
	}

	//nolint:gomnd
	if operation == 2 {
		updateTip["etime"] = xtime.Time(time.Now().Unix())
	}

	if err = query.Updates(updateTip).Error; err != nil {
		log.Error("tips dao tip Updates() error(%v) id(%v)", err, id)
		return
	}
	log.Info("tips dao onlineStatus Updates() success(%v) id(%v)", updateTip, id)

	return
}

// 获取列表
func (d *Dao) FindTipList(c context.Context, idList []int64, startTs, endTs xtime.Time, searchWord string, status int, ps, pn int) (tipList []model.SearchTipDB, total int, err error) {
	// status: -1 全部；0 待生效；1 生效中；2 手动下线
	query := d.DBShow.Model(&model.SearchTipDB{})

	//nolint:gosimple
	if searchWord != "" && (idList == nil || len(idList) == 0) {
		var queryRes []model.SearchTipQueryDB
		searchWordQuery := d.DBShow.Model(&model.SearchTipQueryDB{}).
			Where("search_word = ?", searchWord).
			Where("deleted = ?", 0)
		if err = searchWordQuery.Find(&queryRes).Error; err != nil {
			log.Error("tips dao tip Find() error(%v)", err)
			return
		}

		var searchWordIdList []int64
		if len(queryRes) == 0 {
			return
		} else {
			for _, q := range queryRes {
				searchWordIdList = append(searchWordIdList, q.SearchTipID)
			}
		}
		query = query.Where("id in (?)", searchWordIdList)
	}

	//nolint:gosimple
	if idList != nil && len(idList) > 0 {
		query = query.Where("id in (?)", idList).Where("deleted = ?", 0)
		if err = query.Find(&tipList).Error; err != nil {
			log.Error("tips dao tip Find() error(%v)", err)
			return
		}
		return
	}

	if startTs != 0 && endTs != 0 {
		query = query.Where("(online_status = 0 AND stime >= ? AND stime <= ?) OR (online_status = 1 AND stime <= ? AND stime <= ?)", startTs, endTs, startTs, endTs)
	} else if status != -1 {
		query = query.Where("online_status = ?", status)
	}

	query = query.Where("deleted = ?", 0)

	if err = query.Count(&total).Error; err != nil {
		log.Error("tips dao tip Count() error(%v)", err)
		return
	}

	if pn != 0 && ps != 0 {
		query = query.Offset((pn - 1) * ps).Limit(ps)
	}

	if err = query.Order("ctime desc").Find(&tipList).Error; err != nil {
		log.Error("tips dao tip Find() error(%v)", err)
		return
	}

	return
}

// 获取query的map
func (d *Dao) FindQueryMap(c context.Context, ids []int64) (queryMap map[int64][]model.SearchTipQueryDB, err error) {
	var queryList []model.SearchTipQueryDB
	query := d.DBShow.Model(&model.SearchTipQueryDB{}).Where("deleted = ?", 0)
	if err = query.Find(&queryList).Error; err != nil {
		log.Error("tips dao tip Find() error(%v)", err)
		return
	}
	queryMap = map[int64][]model.SearchTipQueryDB{}
	for _, queryItem := range queryList {
		queryMap[queryItem.SearchTipID] = append(queryMap[queryItem.SearchTipID], queryItem)
	}

	return
}

// 检查冲突
func (d *Dao) CheckConflict(c context.Context, id int64, stime xtime.Time, _ int, searchWordList []model.SearchTipQueryDB) (pass bool, err error) {
	// id 不相等，时间有交集，status = 0、1，plat 相同 或者 3
	var tipIds []int64
	var findTip []model.SearchTipDB
	timeStr := time.Unix(int64(stime), 0).Format("2006-01-02 15:04:05")
	tipQuery := d.DBShow.Model(&model.SearchTipDB{}).
		Where("id != ?", id).
		Where("online_status = 1 OR (online_status = 0 AND stime <= ?)", timeStr).
		//Where("plat = ? OR plat = 3", plat).
		Where("deleted = 0")
	if err = tipQuery.Find(&findTip).Error; err != nil {
		log.Error("tip dao CheckConflict tipIds Find() error(%v)", err)
		return
	}
	if len(findTip) == 0 {
		pass = true
		return
	} else {
		for _, t := range findTip {
			tipIds = append(tipIds, t.ID)
		}
	}

	// query 需要有交集
	var dupSearchWords []string
	var searchWordsInDB []model.SearchTipQueryDB
	var searchWords []string
	for _, v := range searchWordList {
		searchWords = append(searchWords, v.SearchWord)
	}
	queryQuery := d.DBShow.Model(&model.SearchTipQueryDB{}).
		Select("distinct(search_word)").
		Where("s_t_id in (?)", tipIds).
		Where("search_word in (?)", searchWords).
		Where("deleted = 0")
	if err = queryQuery.Find(&searchWordsInDB).Error; err != nil {
		log.Error("tip dao CheckConflict searchWords Find() error(%v)", err)
		return
	}

	for _, q := range searchWordsInDB {
		dupSearchWords = append(dupSearchWords, q.SearchWord)
	}

	if len(dupSearchWords) > 0 {
		pass = false
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("检索词重复%v", dupSearchWords))
	} else {
		pass = true
	}

	return
}
