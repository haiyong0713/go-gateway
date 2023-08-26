package spmode

import (
	"github.com/jinzhu/gorm"
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-feed/admin/model/spmode"
)

func (d *Dao) AddSpecialModeLog(item *model.SpecialModeLog) error {
	if err := d.db.Model(&model.SpecialModeLog{}).Create(item).Error; err != nil {
		log.Error("Fail to create special_mode_log, item=%+v error=%+v", item, err)
		return err
	}
	return nil
}

func (d *Dao) SpecialModeLogsByKey(key string) ([]*model.SpecialModeLog, error) {
	var items []*model.SpecialModeLog
	if err := d.db.Model(&model.TeenagerUsers{}).Where("related_key=?", key).Order("id DESC").Find(&items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*model.SpecialModeLog{}, nil
		}
		log.Error("Fail to query special_mode_log, related_key=%s error=%+v", key, err)
		return nil, err
	}
	return items, nil
}
