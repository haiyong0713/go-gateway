package show

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

// SearchWebAdd add search web
func (d *Dao) SearchWebAdd(param *show.SearchWebAP) (err error) {
	var (
		querys []*show.SearchWebQuery
	)
	if param.Query != "" {
		if err = json.Unmarshal([]byte(param.Query), &querys); err != nil {
			return
		}
	}
	tx := d.DB.Begin()
	if err = tx.Model(&show.SearchWeb{}).Create(param).Error; err != nil {
		log.Error("SearchWebAdd tx.Model Create(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}

	platSql, platParam := show.BatchAddPlatSQL(param.ID, param.PlatVer)
	if err = tx.Model(&show.SearchWebPlat{}).Exec(platSql, platParam...).Error; err != nil {
		log.Error("SearchWebAdd tx.Model Exec sql(%v) param(%+v) error(%v)", platSql, platParam, err)
		err = tx.Rollback().Error
		return
	}

	if len(querys) > 0 {
		sql, sqlParam := show.BatchAddQuerySQL(param.ID, querys)
		if err = tx.Model(&show.SearchWeb{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("SearchWebAdd tx.Model Exec(%+v) error(%v)", param, err)
			err = tx.Rollback().Error
			return
		}
	}
	err = tx.Commit().Error
	return
}

// SearchWebUpdate update
func (d *Dao) SearchWebUpdate(param *show.SearchWebUP) (err error) {
	var (
		newQuerys []*show.SearchWebQuery
	)
	if param.Query != "" {
		if err = json.Unmarshal([]byte(param.Query), &newQuerys); err != nil {
			return
		}
	}
	tx := d.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("dao.SearchWebUpdate.DB.Begin error(%v)", err)
		return
	}
	//TODO: 此处容易造成空值不被更新，建议使用map指定字段，如rec_reason
	if err = tx.Model(&show.SearchWeb{}).Update(param).Error; err != nil {
		log.Error("dao.SearchWebUpdate (%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	var (
		mapOldCData, mapNewQData    map[int64]*show.SearchWebQuery
		upQData, addQData, oldQData []*show.SearchWebQuery
		delQData                    []int64
	)

	// update search web query info
	if len(newQuerys) > 0 {
		// todo: Model 是不是查错表了？
		if err = d.DB.Model(&show.SearchWeb{}).Where("sid=?", param.ID).Where("deleted=?", common.NotDeleted).Find(&oldQData).Error; err != nil {
			log.Error("dao.SearchWebUpdate Find Old data (%+v) error(%v)", param.ID, err)
			return
		}
		mapOldCData = make(map[int64]*show.SearchWebQuery, len(oldQData))
		for _, v := range oldQData {
			mapOldCData[v.ID] = v
		}
		// 新数据在老数据中 更新老数据。新的数据不在老数据 添加新数据
		for _, qData := range newQuerys {
			if _, ok := mapOldCData[qData.ID]; ok {
				upQData = append(upQData, qData)
			} else {
				addQData = append(addQData, qData)
			}
		}
		mapNewQData = make(map[int64]*show.SearchWebQuery, len(newQuerys))
		for _, v := range newQuerys {
			mapNewQData[v.ID] = v
		}
		// 老数据在新数据中 上面已经处理。老数据不在新数据中 删除老数据
		for _, qData := range oldQData {
			if _, ok := mapNewQData[qData.ID]; !ok {
				delQData = append(delQData, qData.ID)
			}
		}
		if len(upQData) > 0 {
			sql, param := show.BatchEditQuerySQL(upQData)
			if err = tx.Model(&show.SearchWebQuery{}).Exec(sql, param...).Error; err != nil {
				log.Error("dao.SearchWebUpdate tx.Model Exec(%+v) error(%v)", upQData, err)
				err = tx.Rollback().Error
				return
			}
		}
		if len(delQData) > 0 {
			if err = tx.Model(&show.SearchWebQuery{}).Where("id IN (?)", delQData).Updates(map[string]interface{}{"deleted": common.Deleted}).Error; err != nil {
				log.Error("dao.SearchWebUpdate Updates(%+v) error(%v)", delQData, err)
				err = tx.Rollback().Error
				return
			}
		}
		if len(addQData) > 0 {
			sql, sqlParam := show.BatchAddQuerySQL(param.ID, addQData)
			if err = tx.Model(&show.SearchWebQuery{}).Exec(sql, sqlParam...).Error; err != nil {
				log.Error("EditContest s.dao.DB.Model Create(%+v) error(%v)", addQData, err)
				err = tx.Rollback().Error
				return
			}
		}
	} else {
		if err = tx.Model(&show.SearchWebQuery{}).Where("sid IN (?)", param.ID).Updates(map[string]interface{}{"deleted": common.Deleted}).Error; err != nil {
			log.Error("dao.SearchWebUpdate Updates(%+v) error(%v)", param.ID, err)
			err = tx.Rollback().Error
			return
		}
	}

	// update search web plat info
	if err = tx.Model(&show.SearchWebPlat{}).
		Where("deleted = ? AND sid = ?", common.NotDeleted, param.ID).
		Update(map[string]interface{}{
			"deleted": common.Deleted,
		}).Error; err != nil {

		log.Error("dao.SearchWebUpdate find old plat data sid(%v) error(%v)", param.ID, err)
		err = tx.Rollback().Error
		return
	}
	platSql, platParam := show.BatchAddPlatSQL(param.ID, param.PlatVer)
	if err = tx.Model(&show.SearchWebPlat{}).Exec(platSql, platParam...).Error; err != nil {
		log.Error("SearchWebAdd tx.Model Exec sql(%v) param(%+v) error(%v)", platSql, platParam, err)
		err = tx.Rollback().Error
		return
	}

	if err = tx.Commit().Error; err != nil {
		log.Error("search_web commit err %v", err)
	}
	return
}

// SearchWebDelete delete search web
func (d *Dao) SearchWebDelete(id int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.Deleted,
	}
	tx := d.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("dao.SearchWebDelete.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&show.SearchWeb{}).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("dao.show.SearchWebDelete(%+v) error(%v)", id, err)
		err = tx.Rollback().Error
		return
	}
	if err = tx.Model(&show.SearchWebQuery{}).Where("sid = ?", id).Updates(map[string]interface{}{"deleted": common.Deleted}).Error; err != nil {
		log.Error("dao.SearchWebDelete Updates(%+v) error(%v)", id, err)
		err = tx.Rollback().Error
		return
	}
	if err = tx.Model(&show.SearchWebPlat{}).Where("sid = ?", id).Update(map[string]interface{}{"deleted": common.Deleted}).Error; err != nil {
		log.Error("dao.SearchWebDelete plat sid(%v) error(%v)", id, err)
		err = tx.Rollback().Error
		return
	}
	err = tx.Commit().Error
	return
}

// SearchWebOption option search web
func (d *Dao) SearchWebOption(up *show.SearchWebOption) (err error) {
	if err = d.DB.Model(&show.SearchWebOption{}).Update(up).Error; err != nil {
		log.Error("dao.SearchWebOption Updates(%+v) error(%v)", up, err)
	}
	return
}

// SearchOptWeb search batch option web
func (d *Dao) SearchOptWeb(c context.Context, ids []int64, option string) (err error) {
	update := map[string]interface{}{}
	switch option {
	case common.OptionBatchPass:
		update["status"] = common.StatusOnline
		update["check"] = common.Pass
	case common.OptionBatchReject:
		update["status"] = common.StatusDownline
		update["check"] = common.Rejecte
	case common.OptionBatchHidden:
		update["status"] = common.StatusDownline
		update["check"] = common.InValid
	}
	err = d.DB.Model(&show.SearchWebOption{}).
		Where("id in (?)", ids).
		Update(update).Error
	return
}

func (d *Dao) SearchWebOptionQueryById(c context.Context, ids []int64) (ret map[int64]*show.SearchWebOption, err error) {
	var obj []*show.SearchWebOption
	if err = d.DB.Model(&show.SearchWebOption{}).
		Where("id in (?)", ids).
		Scan(&obj).
		Error; err != nil {
		return
	}

	ret = make(map[int64]*show.SearchWebOption)
	for _, conf := range obj {
		ret[conf.ID] = conf
	}
	return
}

// SWTimeValid search web time validate
func (d *Dao) SWTimeValid(param *show.SWTimeValid) (count int, err error) {
	query := d.DB.Table("search_web_query").
		Select("search_web_query.id").
		Joins("left join search_web ON search_web.id = search_web_query.sid").
		Joins("left join search_web_plat ON search_web.id = search_web_plat.sid").
		Where("card_type = ?", param.CardType).
		Where("value = ?", param.Query).
		Where("priority = ?", param.Priority).
		Where("`check` in (?)", []int{common.Verify, common.Pass, common.Valid}).
		Where("stime < ?", param.ETime).
		Where("etime > ?", param.STime).
		Where("plat = ?", param.Plat).
		Where("search_web_query.deleted = 0").
		Where("search_web_plat.deleted = 0").
		Where("search_web.deleted = 0")
	if param.ID != 0 {
		query = query.Where("search_web.id != ?", param.ID)
	}
	if err = query.Count(&count).Error; err != nil {
		log.Error("dao.SWTimeValid Count error(%v)", err)
	}
	return
}

// SWFindByID search web table value find by id
func (d *Dao) SWFindByID(id int64) (value *show.SearchWeb, err error) {
	value = &show.SearchWeb{}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
		"id":      id,
	}
	if err = d.DB.Model(&show.SearchWeb{}).Where(w).Find(value).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = fmt.Errorf("ID为%d的数据不存在", id)
			return
		}
		return
	}
	return
}

func (d *Dao) ReleaseSearchWeb(c context.Context) (err error) {
	var oldWebs []*show.SearchWeb
	if err = d.DB.Model(&show.SearchWeb{}).
		Where("search_web.deleted = ?", common.NotDeleted).
		Joins("LEFT JOIN search_web_plat p ON p.sid = search_web.id AND p.deleted != ?", common.Deleted).
		Where("p.id IS NULL").
		Scan(&oldWebs).Error; err != nil {

		log.Error("ReleaseSearchWeb Get id error(%v)", err)
		return
	}

	if len(oldWebs) == 0 {
		return
	}

	var rowStrings []string
	var param []interface{}
	for _, v := range oldWebs {
		rowStrings = append(rowStrings, "(?,?)")
		param = append(param, v.ID, common.PlatWeb)
	}
	sql := fmt.Sprintf("INSERT INTO search_web_plat(sid,plat) VALUES %s", strings.Join(rowStrings, ","))

	if err = d.DB.Model(&show.SearchWebPlat{}).Exec(sql, param...).Error; err != nil {
		log.Error("ReleaseSearchWeb Exec sql(%v) param(%+v) error(%v)", sql, param, err)
		return
	}

	return
}
