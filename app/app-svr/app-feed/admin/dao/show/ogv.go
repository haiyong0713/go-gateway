package show

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
)

const (
	statusShow = 1
	//OgvStatusShow .
	OgvStatusShow = 1
)

// ValOgvQuery validate query 重复query检测逻辑
func (d *Dao) ValOgvQuery(param string) (querys []*show.SearchOgvQuery, err error) {
	var (
		query     []string
		ogvQuerys []*show.SearchOgvQuery
	)
	if err = json.Unmarshal([]byte(param), &querys); err != nil {
		return
	}
	if len(querys) == 0 {
		return nil, fmt.Errorf("query不能为空")
	}
	mapQuery := make(map[string]bool, len(querys))
	queryIDMap := make(map[string]int64, len(querys))
	for _, v := range querys {
		if mapQuery[v.Value] {
			return nil, fmt.Errorf("query不能重复")
		}
		mapQuery[v.Value] = true
		queryIDMap[v.Value] = v.ID
		query = append(query, v.Value)
	}
	where := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.SearchOgvQuery{}).Where(where).Where("value in (?)", query).Find(&ogvQuerys).Error; err != nil {
		return
	}
	for _, v := range ogvQuerys {
		if tmp, ok := queryIDMap[v.Value]; ok {
			if tmp != v.ID {
				return nil, fmt.Errorf("query(%s)已存在", v.Value)
			}
		}
	}
	return
}

// ValPos validate position
func (d *Dao) ValPos(gameStatus, gamePos, moreshowStatus, moreShowPos, pgcPos int64) (err error) {
	var sortPos []int
	sortPos = append(sortPos, int(pgcPos))
	if gameStatus == statusShow && moreshowStatus == statusShow {
		if gamePos == 0 || pgcPos == 0 {
			return fmt.Errorf("展示的模块 位置不能为0")
		}
		if gamePos == moreShowPos {
			return fmt.Errorf("游戏卡片和发现更多精彩位置不能冲突")
		}
		if gamePos == pgcPos {
			return fmt.Errorf("游戏卡片和PGC聚合卡片位置不能冲突")
		}
		if moreShowPos == pgcPos {
			return fmt.Errorf("发现更多精彩和PGC聚合卡片位置不能冲突")
		}
		sortPos = append(sortPos, int(gamePos))
		sortPos = append(sortPos, int(moreShowPos))
	} else if gameStatus == statusShow {
		if gamePos == 0 {
			return fmt.Errorf("展示的模块 位置不能为0")
		}
		if gamePos == pgcPos {
			return fmt.Errorf("游戏卡片和PGC聚合卡片位置不能冲突")
		}
		sortPos = append(sortPos, int(gamePos))
	} else if moreshowStatus == statusShow {
		if moreShowPos == 0 {
			return fmt.Errorf("展示的模块 位置不能为0")
		}
		if moreShowPos == pgcPos {
			return fmt.Errorf("发现更多精彩和PGC聚合卡片位置不能冲突")
		}
		sortPos = append(sortPos, int(moreShowPos))
	}
	sort.Ints(sortPos)
	for k, v := range sortPos {
		if k+1 != v {
			return fmt.Errorf("卡片位置参数不正确")
		}
	}
	return nil
}

// SearchOgvAdd add search web
func (d *Dao) SearchOgvAdd(param *show.SearchOgvAP) (err error) {
	var (
		querys   []*show.SearchOgvQuery
		moreShow []*show.SearchOgvMoreshow
	)
	if err = d.ValPos(param.GameStatus, param.GamePos, param.MoreshowStatus, param.MoreshowPos, param.PgcPos); err != nil {
		return
	}
	if param.Query != "" {
		if querys, err = d.ValOgvQuery(param.Query); err != nil {
			return
		}
	}
	if param.MoreshowStatus == OgvStatusShow {
		if err = json.Unmarshal([]byte(param.MoreshowValue), &moreShow); err != nil {
			return
		}
		if len(moreShow) == 0 {
			return fmt.Errorf("发现更多精彩不能为空")
		}
	}
	tx := d.DB.Begin()
	if err = tx.Model(&show.SearchOgv{}).Create(param).Error; err != nil {
		log.Error("SearchOgvAdd tx.Model Create error(%v)", err)
		err = tx.Rollback().Error
		return
	}
	if len(querys) > 0 {
		sql, sqlParam := show.BatchAddOgvQuerySQL(param.ID, querys)
		if err = tx.Model(&show.SearchOgvQuery{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("SearchOgvAdd tx.Model Exec(%+v) error(%v)", sqlParam, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(moreShow) > 0 {
		sql, sqlParam := show.BatchAddOgvMoreShowSQL(param.ID, moreShow)
		if err = tx.Model(&show.SearchOgvMoreshow{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("SearchOgvAdd tx.Model Exec(%+v) error(%v)", sqlParam, err)
			err = tx.Rollback().Error
			return
		}
	}
	err = tx.Commit().Error
	return
}

// SearchOgvUpdate update
func (d *Dao) SearchOgvUpdate(param *show.SearchOgvUP) (err error) {
	var (
		querys   []*show.SearchOgvQuery
		moreShow []*show.SearchOgvMoreshow
	)
	if err = d.ValPos(param.GameStatus, param.GamePos, param.MoreshowStatus, param.MoreshowPos, param.PgcPos); err != nil {
		return
	}
	if param.Query != "" {
		if param.Query != "" {
			if querys, err = d.ValOgvQuery(param.Query); err != nil {
				return
			}
		}
	}
	if param.MoreshowStatus == OgvStatusShow {
		if err = json.Unmarshal([]byte(param.MoreshowValue), &moreShow); err != nil {
			return
		}
		if len(moreShow) == 0 {
			return fmt.Errorf("发现更多精彩不能为空")
		}
	}
	tx := d.DB.Begin()
	if err = tx.Model(&show.SearchOgv{}).Save(param).Error; err != nil {
		log.Error("SearchOgvUpdate tx.Model Create error(%v)", err)
		err = tx.Rollback().Error
		return
	}
	if len(querys) > 0 {
		if err = tx.Model(&show.SearchOgvQuery{}).Where("sid = ?", param.ID).UpdateColumn("deleted", common.Deleted).Error; err != nil {
			log.Error("SearchOgvUpdate SearchOgvQuery UpdateColumn error(%v)", err)
			err = tx.Rollback().Error
			return
		}
		sql, sqlParam := show.BatchAddOgvQuerySQL(param.ID, querys)
		if err = tx.Model(&show.SearchOgvQuery{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("SearchOgvUpdate SearchOgvQuery Exec(%+v) error(%v)", sqlParam, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(moreShow) > 0 {
		if err = tx.Model(&show.SearchOgvMoreshow{}).Where("sid = ?", param.ID).UpdateColumn("deleted", common.Deleted).Error; err != nil {
			log.Error("SearchOgvUpdate SearchOgvMoreshow UpdateColumn error(%v)", err)
			err = tx.Rollback().Error
			return
		}
		sql, sqlParam := show.BatchAddOgvMoreShowSQL(param.ID, moreShow)
		if err = tx.Model(&show.SearchOgvMoreshow{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("SearchOgvUpdate tx.Model Exec(%+v) error(%v)", sqlParam, err)
			err = tx.Rollback().Error
			return
		}
	}
	err = tx.Commit().Error
	return
}

// SearchOgvOption option search web
func (d *Dao) SearchOgvOption(up *show.SearchOgvOption) (err error) {
	if err = d.DB.Model(&show.SearchOgvOption{}).Where("id = ?", up.ID).UpdateColumn("check", up.Check).Error; err != nil {
		log.Error("dao.SearchOgvOption Updates(%+v) error(%v)", up, err)
	}
	return
}

func (d *Dao) SearchOgvFind(id int64) (res *show.SearchOgv, err error) {
	var (
		ogvQuery []*show.SearchOgvQuery
		query    []string
	)
	res = &show.SearchOgv{}
	if err = d.DB.Model(&show.SearchOgv{}).Where("id = ?", id).First(res).Error; err != nil {
		log.Error("SearchOgvFind Find error(%v)", err)
		return
	}
	where := map[string]interface{}{
		"sid":     id,
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.SearchOgvQuery{}).Where(where).Find(&ogvQuery).Error; err != nil {
		log.Error("SearchShieldList Find error(%v)", err)
		return
	}
	for _, v := range ogvQuery {
		query = append(query, v.Value)
	}
	res.QueryStr = strings.Join(query, ",")
	return
}
