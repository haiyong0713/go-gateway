package show

import (
	"context"

	"go-gateway/app/app-svr/app-feed/admin/model/push"
)

func (d *Dao) PushList(_ context.Context) ([]*push.PushDetail, error) {
	pushList := make([]*push.PushDetail, 0)
	if err := d.DB.Model(&push.PushDetail{}).Where("is_deleted=?", 0).Order("id desc").Find(&pushList).Error; err != nil {
		return nil, err
	}
	return pushList, nil
}

func (d *Dao) PushDetail(_ context.Context, id int64) (*push.PushDetail, error) {
	pushDetail := &push.PushDetail{}
	if err := d.DB.Model(pushDetail).Where("id=? AND is_deleted=?", id, 0).Find(pushDetail).Error; err != nil {
		return nil, err
	}
	return pushDetail, nil
}

func (d *Dao) PushCreate(_ context.Context, detail *push.PushDetail) error {
	return d.DB.Model(&push.PushDetail{}).Create(detail).Error
}

func (d *Dao) PushUpdate(_ context.Context, detail *push.PushDetail) error {
	return d.DB.Model(&push.PushDetail{}).Where("id=?", detail.ID).Save(detail).Error
}

func (d *Dao) PushDelete(_ context.Context, id int64) error {
	return d.DB.Table("package_push").Where("id=?", id).Update("is_deleted", 1).Error
}
