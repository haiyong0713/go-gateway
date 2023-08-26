package teen_manual

import (
	"github.com/jinzhu/gorm"
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-feed/admin/model/teen_manual"
)

func (d *Dao) LogsByMid(mid int64) ([]*model.TeenagerManualLog, error) {
	var items []*model.TeenagerManualLog
	if err := d.db.Model(&model.TeenagerManualLog{}).Where("mid=?", mid).Order("id DESC").Find(&items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*model.TeenagerManualLog{}, nil
		}
		log.Error("Fail to query teenager_manual_log, mid=%+v error=%+v", mid, err)
		return nil, err
	}
	return items, nil
}

func (d *Dao) AddManualLog(item *model.TeenagerManualLog) error {
	if err := d.db.Model(&model.TeenagerManualLog{}).Create(item).Error; err != nil {
		log.Error("Fail to create teenager_manual_log, item=%+v error=%+v", item, err)
		return err
	}
	return nil
}
