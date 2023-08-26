package search

import (
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/search"
)

// DySeachAdd add
func (d *Dao) DySeachAdd(param *search.DySeachAP) (err error) {
	if err = d.DB.Create(param).Error; err != nil {
		log.Error("DynamicDao.DySeachAdd error(%v)", err)
		return
	}
	return
}

// DySeachUpdate update
func (d *Dao) DySeachUpdate(param *search.DySeachUP) (err error) {
	if err = d.DB.Model(&search.DySeachUP{}).Update(param).Error; err != nil {
		log.Error("DynamicDao.DySeachUpdate error(%v)", err)
		return
	}
	return
}

// DySeachDelete delete
func (d *Dao) DySeachDelete(id int64) (err error) {
	up := map[string]interface{}{
		"is_deleted": common.Deleted,
	}
	if err = d.DB.Model(&search.DySeach{}).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("DynamicDao.DySeachDelete error(%v)", err)
		return
	}
	return
}

// DySeachValidat is position exists
func (d *Dao) DySeachValidat(position, id int64, word string) (err error) {
	var (
		count int
	)
	w := map[string]interface{}{
		"is_deleted": common.NotDeleted,
	}
	if position != 0 {
		w["position"] = position
	}
	if word != "" {
		w["word"] = word
	}
	query := d.DB.Model(&search.DySeach{}).Where(w)
	if id != 0 {
		query = query.Where("id != ?", id)
	}
	if err = query.Count(&count).Error; err != nil {
		log.Error("DynamicDao.DySeachValidat error(%v)", err)
		return
	}
	if count > 0 {
		if position != 0 {
			return fmt.Errorf("已设置相同顺位")
		}
		if word != "" {
			return fmt.Errorf("已添加相同热搜词")
		}
	}
	return
}
