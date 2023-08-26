package show

import (
	"encoding/json"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
)

func (d *Dao) ParseQuery(value string) ([]*show.SearchShieldQuery, error) {
	var res []*show.SearchShieldQuery
	if err := json.Unmarshal([]byte(value), &res); err != nil {
		return res, err
	}
	queryMap := make(map[string]bool, len(res))
	for _, v := range res {
		if queryMap[v.Value] {
			return res, fmt.Errorf("query 不能重复")
		}
		queryMap[v.Value] = true
	}
	return res, nil
}

// SearchShieldAdd add search shield
func (d *Dao) SearchShieldAdd(param *show.SearchShieldAP) (err error) {
	var (
		querys []*show.SearchShieldQuery
	)
	if param.Query != "" {
		if querys, err = d.ParseQuery(param.Query); err != nil {
			return
		}
	}
	tx := d.DB.Begin()
	if err = tx.Model(&show.SearchShield{}).Create(param).Error; err != nil {
		log.Error("SearchShieldAdd tx.Model Create(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	if len(querys) > 0 {
		sql, sqlParam := show.BatchAddShieldSQL(param.ID, querys)
		if err = tx.Model(&show.SearchShield{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("SearchShieldAdd tx.Model Exec(%+v) error(%v)", param, err)
			err = tx.Rollback().Error
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	return
}

// SearchShieldUpdate update
func (d *Dao) SearchShieldUpdate(param *show.SearchShieldUP) (err error) {
	var (
		newQuerys []*show.SearchShieldQuery
	)
	if param.Query != "" {
		if newQuerys, err = d.ParseQuery(param.Query); err != nil {
			return
		}
	}
	tx := d.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("dao.SearchShieldUpdate.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&show.SearchShield{}).Update(param).Error; err != nil {
		log.Error("dao.SearchShieldUpdate (%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	var (
		mapOldCData, mapNewQData    map[int64]*show.SearchShieldQuery
		upQData, addQData, oldQData []*show.SearchShieldQuery
		delQData                    []int64
	)
	if len(newQuerys) > 0 {
		if err = d.DB.Model(&show.SearchShield{}).Where("sid=?", param.ID).Where("deleted=?", common.NotDeleted).Find(&oldQData).Error; err != nil {
			log.Error("dao.SearchShieldUpdate Find Old data (%+v) error(%v)", param.ID, err)
			return
		}
		mapOldCData = make(map[int64]*show.SearchShieldQuery, len(oldQData))
		for _, v := range oldQData {
			mapOldCData[v.ID] = v
		}
		//新数据在老数据中 更新老数据。新的数据不在老数据 添加新数据
		for _, qData := range newQuerys {
			if _, ok := mapOldCData[qData.ID]; ok {
				upQData = append(upQData, qData)
			} else {
				addQData = append(addQData, qData)
			}
		}
		mapNewQData = make(map[int64]*show.SearchShieldQuery, len(newQuerys))
		for _, v := range newQuerys {
			mapNewQData[v.ID] = v
		}
		//老数据在新数据中 上面已经处理。老数据不在新数据中 删除老数据
		for _, qData := range oldQData {
			if _, ok := mapNewQData[qData.ID]; !ok {
				delQData = append(delQData, qData.ID)
			}
		}
		if len(upQData) > 0 {
			sql, param := show.BatchEditShieldSQL(upQData)
			if err = tx.Model(&show.SearchShieldQuery{}).Exec(sql, param...).Error; err != nil {
				log.Error("dao.SearchShieldUpdate tx.Model Exec(%+v) error(%v)", upQData, err)
				err = tx.Rollback().Error
				return
			}
		}
		if len(delQData) > 0 {
			if err = tx.Model(&show.SearchShieldQuery{}).Where("id IN (?)", delQData).Updates(map[string]interface{}{"deleted": common.Deleted}).Error; err != nil {
				log.Error("dao.SearchShieldUpdate Updates(%+v) error(%v)", delQData, err)
				err = tx.Rollback().Error
				return
			}
		}
		if len(addQData) > 0 {
			sql, sqlParam := show.BatchAddShieldSQL(param.ID, addQData)
			if err = tx.Model(&show.SearchShieldQuery{}).Exec(sql, sqlParam...).Error; err != nil {
				log.Error("EditContest s.dao.DB.Model Create(%+v) error(%v)", addQData, err)
				err = tx.Rollback().Error
				return
			}
		}
	} else {
		if err = tx.Model(&show.SearchShieldQuery{}).Where("sid IN (?)", param.ID).Updates(map[string]interface{}{"deleted": common.Deleted}).Error; err != nil {
			log.Error("dao.SearchShieldUpdate Updates(%+v) error(%v)", param.ID, err)
			err = tx.Rollback().Error
			return
		}
	}
	err = tx.Commit().Error
	return
}

// SearchShieldOption option search web
func (d *Dao) SearchShieldOption(up *show.SearchShieldOption) (err error) {
	if err = d.DB.Model(&show.SearchShieldOption{}).Update(up).Error; err != nil {
		log.Error("dao.SearchShieldOption Updates(%+v) error(%v)", up, err)
	}
	return
}

// SearchShieldValid search shield validate
func (d *Dao) SearchShieldValid(param *show.SearchShieldValid) (count int, err error) {
	query := d.DB.Table("search_shield").
		Select("search_shield_query.id").
		Joins("left join search_shield_query ON search_shield.id = search_shield_query.sid").
		Where("value = ?", param.Query).
		Where("search_shield.card_type = ?", param.CardType).
		Where("search_shield.card_value = ?", param.CardValue).
		Where("search_shield_query.deleted = 0")
	if param.ID != 0 {
		query = query.Where("search_shield.id != ?", param.ID)
	}
	if err = query.Count(&count).Error; err != nil {
		log.Error("SearchShieldValid Count error(%v)", err)
	}
	return
}
