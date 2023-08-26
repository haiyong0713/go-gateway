package spmode

import (
	"time"

	"github.com/jinzhu/gorm"
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-feed/admin/model/spmode"
)

func (d *Dao) TeenagerUsersByMid(mid int64) ([]*model.TeenagerUsers, error) {
	var items []*model.TeenagerUsers
	if err := d.db.Model(&model.TeenagerUsers{}).Where("mid=?", mid).Find(&items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*model.TeenagerUsers{}, nil
		}
		log.Error("Fail to query teenager_users, mid=%d error=%+v", mid, err)
		return nil, err
	}
	return items, nil
}

func (d *Dao) RelieveTeenagerUsers(id, operation int64) (int64, error) {
	fields := map[string]interface{}{
		"state":     model.StateQuit,
		"operation": operation,
		"password":  "",
		"quit_time": time.Now().Unix(),
	}
	db := d.db.Model(&model.TeenagerUsers{}).Where("id=?", id).Where("state=?", model.StateOpen).Update(fields)
	if err := db.Error; err != nil {
		log.Error("Fail to relieve teenager_users, id=%d error=%+v", id, err)
		return 0, err
	}
	return db.RowsAffected, nil
}

func (d *Dao) TeenagerUsersByID(id int64) (*model.TeenagerUsers, error) {
	item := &model.TeenagerUsers{}
	if err := d.db.Model(&model.TeenagerUsers{}).Where("id=?", id).First(item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Error("Fail to query teenager_users, id=%d error=%+v", id, err)
		return nil, err
	}
	return item, nil
}

func (d *Dao) PagingTeenManual(mid int64, mfOperator string, pn, ps int64) (int64, []*model.TeenagerUsers, error) {
	var offset int64
	if pn > 0 {
		offset = (pn - 1) * ps
	}
	db := d.db.Model(&model.TeenagerUsers{}).Where("model=?", model.ModelTeenager)
	if mid > 0 {
		db = db.Where("mid=?", mid)
	}
	if mfOperator != "" {
		db = db.Where("mf_operator=?", mfOperator)
	}
	if mid <= 0 && mfOperator == "" {
		db = db.Where("manual_force=?", model.ManualForceOpen)
	}
	total := func() int64 {
		var total int64
		if err := db.Count(&total).Error; err != nil {
			log.Error("Fail to count PagingTeenManual, mid=%+v op=%+v pn=%+v ps=%+v error=%+v", mid, mfOperator, pn, ps, err)
			return 0
		}
		return total
	}()
	var items []*model.TeenagerUsers
	if err := db.Order("mf_time DESC").Offset(offset).Limit(ps).Find(&items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return total, []*model.TeenagerUsers{}, nil
		}
		log.Error("Fail to query PagingTeenManual, mid=%+v op=%+v pn=%+v ps=%+v error=%+v", mid, mfOperator, pn, ps, err)
		return 0, nil, err
	}
	return total, items, nil
}

func (d *Dao) TeenagerUserByMidModel(mid, mod int64) (*model.TeenagerUsers, error) {
	item := &model.TeenagerUsers{}
	if err := d.db.Model(&model.TeenagerUsers{}).Where("mid=?", mid).Where("model=?", mod).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Error("Fail to query teenager_users, mid=%+v model=%+v error=%+v", mid, mod, err)
		return nil, err
	}
	return item, nil
}

func (d *Dao) OpenManualForce(id int64, operator string) (int64, error) {
	fields := map[string]interface{}{
		"manual_force": model.ManualForceOpen,
		"mf_operator":  operator,
		"mf_time":      time.Now().Unix(),
	}
	db := d.db.Model(&model.TeenagerUsers{}).Where("id=?", id).Where("manual_force=?", model.ManualForceQuit).Update(fields)
	if err := db.Error; err != nil {
		log.Error("Fail to OpenManualForce, id=%+v op=%+v error=%+v", id, operator, err)
		return 0, err
	}
	return db.RowsAffected, nil
}

func (d *Dao) QuitManualForce(id int64, operator string) (int64, error) {
	fields := map[string]interface{}{
		"manual_force": model.ManualForceQuit,
		"mf_operator":  operator,
		"mf_time":      time.Now().Unix(),
	}
	db := d.db.Model(&model.TeenagerUsers{}).Where("id=?", id).Where("manual_force=?", model.ManualForceOpen).Update(fields)
	if err := db.Error; err != nil {
		log.Error("Fail to QuitManualForce, id=%+v op=%+v error=%+v", id, operator, err)
		return 0, err
	}
	return db.RowsAffected, nil
}

func (d *Dao) AddTeenUser(item *model.TeenagerUsers) error {
	if err := d.db.Model(&model.TeenagerUsers{}).Create(item).Error; err != nil {
		log.Error("Fail to create teenager_users, item=%+v error=%+v", item, err)
		return err
	}
	return nil
}
