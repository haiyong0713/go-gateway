package spmode

import (
	"github.com/jinzhu/gorm"
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-feed/admin/model/family"
)

func (d *Dao) FamilyRelsOfParent(pmid int64) ([]*model.FamilyRelation, error) {
	var items []*model.FamilyRelation
	if err := d.db.Model(&model.FamilyRelation{}).Where("parent_mid=?", pmid).Where("state=?", model.StateBind).Find(&items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Error("Fail to query family_relation of parent, pmid=%+v error=%+v", pmid, err)
		return nil, err
	}
	return items, nil
}

func (d *Dao) FamilyRelOfChild(cmid int64) (*model.FamilyRelation, error) {
	item := new(model.FamilyRelation)
	if err := d.db.Model(&model.FamilyRelation{}).Where("child_mid=?", cmid).Where("state=?", model.StateBind).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Error("Fail to query family_relation of child, cmid=%+v error=%+v", cmid, err)
		return nil, err
	}
	return item, nil
}

func (d *Dao) FamilyRelById(id int64) (*model.FamilyRelation, error) {
	item := new(model.FamilyRelation)
	if err := d.db.Model(&model.FamilyRelation{}).Where("id=?", id).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Error("Fail to query family_relation by id, id=%+v error=%+v", id, err)
		return nil, err
	}
	return item, nil
}

func (d *Dao) UnbindFamily(id int64) error {
	fields := map[string]interface{}{
		"state": model.StateUnbind,
	}
	if err := d.db.Model(&model.FamilyRelation{}).Where("id=?", id).Update(fields).Error; err != nil {
		log.Error("Fail to unbind family_relation, id=%+v error=%+v", id, err)
		return err
	}
	return nil
}
