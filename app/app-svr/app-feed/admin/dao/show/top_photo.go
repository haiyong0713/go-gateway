package show

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	_updateTopPhotoSQL = "UPDATE popular_top_photo SET top_photo=? WHERE id=?"
)

// PopTopPhotoAdd add event topic
func (d *Dao) PopTopPhotoAdd(ctx context.Context, param *show.PopTopPhotoAD) (err error) {
	if err = d.DB.Create(param).Error; err != nil {
		log.Error("dao.show.PopTopPhotoAdd error(%v)", err)
		return
	}
	return
}

// PopTopPhotoState delete event topic
func (d *Dao) PopTopPhotoDeleted(ctx context.Context, id int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.Deleted,
	}
	if err = d.DB.Model(&show.PopTopPhoto{}).Where("id=?", id).Update(up).Error; err != nil {
		log.Error("dao.show.PopTopPhotoState error(%v)", err)
		return
	}
	return
}

// PopTopPhotoNotDeleted not delete event topic
func (d *Dao) PopTopPhotoNotDeleted(ctx context.Context, id int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.PopTopPhoto{}).Where("id=?", id).Update(up).Error; err != nil {
		log.Error("dao.show.PopTopPhotoNotDeleted error(%v)", err)
		return
	}
	return
}

// PopTPFind search top photo
func (d *Dao) PopTPFind(ctx context.Context, pn, ps int) (res []*show.PopTopPhoto, err error) {
	res = []*show.PopTopPhoto{}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.PopTopPhoto{}).Offset((pn - 1) * ps).Limit(ps).Where(w).Find(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			res = nil
			err = nil
		} else {
			log.Error("dao.PopTPFind error(%v)", err)
		}
	}
	return
}

func (d *Dao) PopTopPhotoUpdate(ctx context.Context, id int64, topPhoto string) (err error) {
	if err = d.DB.Exec(_updateTopPhotoSQL, topPhoto, id).Error; err != nil {
		err = errors.Wrapf(err, "id(%d)", id)
	}
	return
}
