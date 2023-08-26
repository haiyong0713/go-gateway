package show

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unicode/utf8"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
)

const (
	_moduleCount = 6
	_sql         = "SELECT search_web_special_query.id FROM search_web_special_query LEFT JOIN search_web_special ON search_web_special.id = search_web_special_query.sid WHERE search_web_special_query.`value` = ? AND search_web_special_query.deleted = 0 AND search_web_special.`check` = 1"
)

func (d *Dao) ValQuery(id int64, values []*show.SearchWebModuleQuery) (err error) {
	var (
		querys []string
	)
	for _, v := range values {
		if v.Value == "" {
			return fmt.Errorf("query不能为空")
		}
		//nolint:gomnd
		if utf8.RuneCountInString(v.Value) > 20 {
			return fmt.Errorf("query最大不能超过20个字符")
		}
		querys = append(querys, v.Value)
	}
	if len(querys) == 0 {
		return fmt.Errorf("query不能为空")
	}
	sql := _sql
	if id != 0 {
		sql += fmt.Sprintf(" AND search_web_special_query.sid != %d", id)
	}
	for _, v := range querys {
		var res []*show.SearchWebModuleQuery
		if err = d.DB.Raw(sql, v).Scan(&res).Error; err != nil {
			return
		}
		if len(res) > 0 {
			return fmt.Errorf("query %s 已存在 不能重复添加", v)
		}
	}
	return nil
}

func (d *Dao) valModule(values []*show.SearchWebModuleModule) error {
	m := make(map[string]bool)
	for _, value := range values {
		if _, ok := m[value.Value]; ok {
			return fmt.Errorf("模块不能重复")
		}
		m[value.Value] = true
	}
	moduleMap := make(map[string]bool)
	if len(values) != _moduleCount {
		return fmt.Errorf("模块参数错误")
	}
	for _, v := range values {
		moduleMap[v.Value] = true
	}
	for i := 1; i <= 6; i++ {
		if _, ok := moduleMap[strconv.Itoa(i)]; !ok {
			return fmt.Errorf("模块参数错误")
		}
	}
	return nil
}

// WebModuleAdd add search special
func (d *Dao) WebModuleAdd(param *show.SearchWebModuleAP) (err error) {
	var (
		querys       []*show.SearchWebModuleQuery
		paramModules []*show.SearchWebModuleModule
	)
	if param.Query != "" {
		if err = json.Unmarshal([]byte(param.Query), &querys); err != nil {
			log.Error("WebModuleAdd json.Unmarshal(%s) error(%v)", param.Query, err)
			return err
		}
		if err = d.ValQuery(0, querys); err != nil {
			return
		}
	}
	if param.Module != "" {
		if err = json.Unmarshal([]byte(param.Module), &paramModules); err != nil {
			log.Error("WebModuleAdd json.Unmarshal(%s) error(%v)", param.Module, err)
			return
		}
		if err = d.valModule(paramModules); err != nil {
			return
		}
	}
	tx := d.DB.Begin()
	if err = tx.Create(param).Error; err != nil {
		log.Error("WebModuleAdd tx.Model Create error(%v)", err)
		err = tx.Rollback().Error
		return
	}
	if len(querys) > 0 {
		sql, sqlParam := show.SpecialQuerySQL(param.ID, querys)
		if err = tx.Exec(sql, sqlParam...).Error; err != nil {
			log.Error("WebModuleAdd tx.Model Exec error(%v)", err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(paramModules) > 0 {
		var modules []*show.SearchWebModuleModule
		for k, v := range paramModules {
			module := &show.SearchWebModuleModule{
				Sid:   param.ID,
				Order: k + 1,
				Value: v.Value,
			}
			modules = append(modules, module)
		}
		sql, sqlParam := show.SpecialModuleSQL(param.ID, modules)
		if err = tx.Exec(sql, sqlParam...).Error; err != nil {
			log.Error("WebModuleAdd tx.Model Exec error(%v)", err)
			err = tx.Rollback().Error
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	return
}

// WebModuleUpdate update
//
//nolint:gocognit
func (d *Dao) WebModuleUpdate(param *show.SearchWebModuleUP) (err error) {
	var (
		newQuerys  []*show.SearchWebModuleQuery
		newModules []*show.SearchWebModuleModule
	)
	if param.Query != "" {
		if err = json.Unmarshal([]byte(param.Query), &newQuerys); err != nil {
			return
		}
		if err = d.ValQuery(param.ID, newQuerys); err != nil {
			return
		}
	}
	if param.Module != "" {
		if err = json.Unmarshal([]byte(param.Module), &newModules); err != nil {
			return
		}
		if err = d.valModule(newModules); err != nil {
			return
		}
	}
	tx := d.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("dao.WebModuleUpdate.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&show.SearchWebModule{}).Update(param).Error; err != nil {
		log.Error("dao.WebModuleUpdate  error(%v)", err)
		err = tx.Rollback().Error
		return
	}
	//更新query
	if len(newQuerys) > 0 {
		var (
			mapOldCData, mapNewQData    map[int64]*show.SearchWebModuleQuery
			upQData, addQData, oldQData []*show.SearchWebModuleQuery
			delQData                    []int64
		)
		if err = d.DB.Model(&show.SearchWebModuleQuery{}).Where("sid=?", param.ID).Where("deleted=?", common.NotDeleted).Find(&oldQData).Error; err != nil {
			log.Error("dao.WebModuleUpdate Find Old data (%+v) error(%v)", param.ID, err)
			return
		}
		mapOldCData = make(map[int64]*show.SearchWebModuleQuery, len(oldQData))
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
		mapNewQData = make(map[int64]*show.SearchWebModuleQuery, len(newQuerys))
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
			sql, param := show.SpecialQueryUpSQL(upQData)
			if err = tx.Exec(sql, param...).Error; err != nil {
				log.Error("dao.WebModuleUpdate tx.Model Exec(%+v) error(%v)", upQData, err)
				err = tx.Rollback().Error
				return
			}
		}
		if len(delQData) > 0 {
			if err = tx.Model(&show.SearchWebModuleQuery{}).Where("id IN (?)", delQData).Updates(map[string]interface{}{"deleted": common.Deleted}).Error; err != nil {
				log.Error("dao.WebModuleUpdate Updates(%+v) error(%v)", delQData, err)
				err = tx.Rollback().Error
				return
			}
		}
		if len(addQData) > 0 {
			sql, sqlParam := show.SpecialQuerySQL(param.ID, addQData)
			if err = tx.Exec(sql, sqlParam...).Error; err != nil {
				log.Error("dao.WebModuleUpdate tx.Exec error(%v)", err)
				err = tx.Rollback().Error
				return
			}
		}
	} else {
		if err = tx.Model(&show.SearchWebModuleQuery{}).Where("sid IN (?)", param.ID).Updates(map[string]interface{}{"deleted": common.Deleted}).Error; err != nil {
			log.Error("dao.SearchWebUpdate Updates(%+v) error(%v)", param.ID, err)
			err = tx.Rollback().Error
			return
		}
	}
	//更新模块排序
	if len(newModules) > 0 {
		var (
			mapOldMData, mapNewMData    map[int64]*show.SearchWebModuleModule
			upMData, addMData, oldMData []*show.SearchWebModuleModule
			delMData                    []int64
		)
		if err = d.DB.Model(&show.SearchWebModuleModule{}).Where("sid=?", param.ID).Where("deleted=?", common.NotDeleted).Find(&oldMData).Error; err != nil {
			log.Error("dao.WebModuleUpdate Find Old data (%+v) error(%v)", param.ID, err)
			return
		}
		mapOldMData = make(map[int64]*show.SearchWebModuleModule, len(oldMData))
		for _, v := range oldMData {
			mapOldMData[v.ID] = v
		}
		//新数据在老数据中 更新老数据。新的数据不在老数据 添加新数据
		for k, mData := range newModules {
			mData.Order = k + 1
			if _, ok := mapOldMData[mData.ID]; ok {
				upMData = append(upMData, mData)
			} else {
				addMData = append(addMData, mData)
			}
		}
		mapNewMData = make(map[int64]*show.SearchWebModuleModule, len(newModules))
		for _, v := range newModules {
			mapNewMData[v.ID] = v
		}
		//老数据在新数据中 上面已经处理。老数据不在新数据中 删除老数据
		for _, qData := range oldMData {
			if _, ok := mapNewMData[qData.ID]; !ok {
				delMData = append(delMData, qData.ID)
			}
		}
		if len(upMData) > 0 {
			sql, param := show.SpecialModuleUpSQL(upMData)
			if err = tx.Exec(sql, param...).Error; err != nil {
				log.Error("dao.WebModuleUpdate tx.Model Exec(%+v) error(%v)", upMData, err)
				err = tx.Rollback().Error
				return
			}
		}
		if len(delMData) > 0 {
			if err = tx.Model(&show.SearchWebModuleModule{}).Where("id IN (?)", delMData).Updates(map[string]interface{}{"deleted": common.Deleted}).Error; err != nil {
				log.Error("dao.WebModuleUpdate Updates(%+v) error(%v)", delMData, err)
				err = tx.Rollback().Error
				return
			}
		}
		if len(addMData) > 0 {
			sql, sqlParam := show.SpecialModuleSQL(param.ID, addMData)
			if err = tx.Exec(sql, sqlParam...).Error; err != nil {
				log.Error("dao.WebModuleUpdate tx.Exec error(%v)", err)
				err = tx.Rollback().Error
				return
			}
		}
	} else {
		if err = tx.Model(&show.SearchWebModuleModule{}).Where("sid IN (?)", param.ID).Updates(map[string]interface{}{"deleted": common.Deleted}).Error; err != nil {
			log.Error("dao.WebModuleUpdate Updates(%+v) error(%v)", param.ID, err)
			err = tx.Rollback().Error
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		log.Error("dao.WebModuleUpdate commit err %v", err)
	}
	return
}
