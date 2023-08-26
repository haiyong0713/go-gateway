package pwd_appeal

import (
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-feed/admin/model/pwd_appeal"
)

func (d *Dao) CreatePwdAppealLog(item *model.PwdAppealLog) (int64, error) {
	if err := d.db.Model(&model.PwdAppealLog{}).Create(item).Error; err != nil {
		log.Error("Fail to create pwd_appeal_log, item=%+v error=%+v", item, err)
		return 0, err
	}
	return item.ID, nil
}
