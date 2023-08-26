package spmode

import (
	"time"

	"github.com/jinzhu/gorm"
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-feed/admin/model/spmode"
)

func (d *Dao) DeviceUserModelByToken(devToken string) ([]*model.DeviceUserModel, error) {
	var items []*model.DeviceUserModel
	if err := d.db.Model(&model.DeviceUserModel{}).Where("device_token=?", devToken).Find(&items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*model.DeviceUserModel{}, nil
		}
		log.Error("Fail to query device_user_model, deviceToken=%s error=%+v", devToken, err)
		return nil, err
	}
	return items, nil
}

func (d *Dao) RelieveDeviceUserModel(id int64) (int64, error) {
	fields := map[string]interface{}{
		"state":     model.StateQuit,
		"operation": model.OperationQuitManager,
		"password":  "",
		"quit_time": time.Now().Unix(),
	}
	db := d.db.Model(&model.DeviceUserModel{}).Where("id=?", id).Where("state=?", model.StateOpen).Update(fields)
	if err := db.Error; err != nil {
		log.Error("Fail to relieve device_user_model, id=%d error=%+v", id, err)
		return 0, err
	}
	return db.RowsAffected, nil
}

func (d *Dao) DeviceUserModelByID(id int64) (*model.DeviceUserModel, error) {
	item := &model.DeviceUserModel{}
	if err := d.db.Model(&model.DeviceUserModel{}).Where("id=?", id).First(item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Error("Fail to query device_user_model, id=%d error=%+v", id, err)
		return nil, err
	}
	return item, nil
}
