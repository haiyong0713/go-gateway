package spmode

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-feed/admin/model/family"
)

const (
	_batchAddLogsSQL = "INSERT INTO family_log (mid,operator,content) VALUES %s"
)

func (d *Dao) PagingFamilyLog(mid, pn, ps int64) (int64, []*model.FamilyLog, error) {
	var offset int64
	if pn > 0 {
		offset = (pn - 1) * ps
	}
	db := d.db.Model(&model.FamilyLog{}).Where("mid=?", mid)
	total := func() int64 {
		var total int64
		if err := db.Count(&total).Error; err != nil {
			log.Error("Fail to count PagingFamilyLog, mid=%+v pn=%+v ps=%+v error=%+v", mid, pn, ps, err)
			return 0
		}
		return total
	}()
	var items []*model.FamilyLog
	if err := db.Order("id DESC").Offset(offset).Limit(ps).Find(&items).Error; err != nil {
		log.Error("Fail to query PagingFamilyLog, mid=%+v pn=%+v ps=%+v error=%+v", mid, pn, ps, err)
		if err == gorm.ErrRecordNotFound {
			return total, nil, nil
		}
		return 0, nil, err
	}
	return total, items, nil
}

func (d *Dao) BatchAddFamilyLogs(items []*model.FamilyLog) error {
	if len(items) == 0 {
		return nil
	}
	parts := make([]string, 0, len(items))
	args := make([]interface{}, 0, len(items)*3)
	for _, item := range items {
		if item == nil {
			continue
		}
		parts = append(parts, "(?,?,?)")
		args = append(args, item.Mid, item.Operator, item.Content)
	}
	batchSQL := fmt.Sprintf(_batchAddLogsSQL, strings.Join(parts, ","))
	if err := d.db.Model(&model.FamilyLog{}).Exec(batchSQL, args...).Error; err != nil {
		log.Error("Fail to batch create family_log, items=%+v error=%+v", items, err)
		return err
	}
	return nil
}
