package dao

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	xecode "go-common/library/ecode"

	"go-gateway/app/web-svr/space/admin/model"
)

// GetByTID 根据头图主键ID获取日志
func (d *Dao) GetByTID(tid int64) (vipAuditLog *model.VipAuditLog, err error) {
	vipAuditLog = &model.VipAuditLog{}

	query := d.DB.Model(model.VipAuditLog{}).Where("tid = ?", tid)

	if err = query.Find(vipAuditLog).Error; err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			err = nil
		} else {
			err = errors.Wrapf(err, "Dao: GetByTID %d", tid)
		}
		return nil, err
	}

	return vipAuditLog, nil
}

// AddLog 添加日志
func (d *Dao) AddLog(vipAuditLog *model.VipAuditLog, tx *gorm.DB) (err error) {
	if vipAuditLog == nil {
		return
	}
	if tx == nil {
		tx = d.DB.Begin()
		defer func() {
			if err != nil {
				err = tx.Rollback().Error
			} else {
				err = tx.Commit().Error
			}
		}()
	}
	if err = tx.Model(model.VipAuditLog{}).Create(vipAuditLog).Error; err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			err = nil
		} else {
			err = errors.Wrapf(err, "Dao: AddLog %+v", vipAuditLog)
		}
		return err
	}

	return

}

// EditByTID 根据头图ID修改日志
func (d *Dao) EditByTID(toUpdate *model.VipAuditLog, tx *gorm.DB) (err error) {
	if toUpdate == nil {
		return xecode.RequestErr
	}
	if tx == nil {
		tx = d.DB.Begin()
		defer func() {
			if err != nil {
				err = tx.Rollback().Error
			} else {
				err = tx.Commit().Error
			}
		}()
	}

	up := map[string]interface{}{
		"mid":            toUpdate.MID,
		"reason":         toUpdate.Reason,
		"reason_default": toUpdate.ReasonDefault,
		"operator":       toUpdate.Operator,
	}

	if err = tx.Model(model.VipAuditLog{}).Where("tid = ?", toUpdate.TID).Update(up).Error; err != nil {
		return errors.Wrapf(err, "Dao: EditByTID %d", toUpdate.TID)
	}
	return
}

// GetAuditInfosByPage 分页获取日志列表
func (d *Dao) GetAuditInfosByPage(params *model.VipAuditLogSearch, pn int, ps int) (vipAuditLogList []*model.VipAuditLog, total int, err error) {
	vipAuditLogList = make([]*model.VipAuditLog, 0)

	//获取日志
	query := d.DB.Model(model.VipAuditLog{})
	if params.AuditTimeStart != "" {
		query = query.Where("ctime >= ?", params.AuditTimeStart)
	}
	if params.AuditTimeEnd != "" {
		query = query.Where("ctime <= ?", params.AuditTimeEnd)
	}
	if len(params.MIDs) > 0 {
		query = query.Where("mid IN (?)", params.MIDs)
	}
	if params.Operator != "" {
		query = query.Where("operator = ?", params.Operator)
	}
	if err = query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "Dao: GetAuditInfosByPage %+v Count", params)
	}
	if err = query.Order("id DESC").Limit(ps).Offset((pn - 1) * ps).Find(&vipAuditLogList).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "Dao: GetAuditInfosByPage %+v Find", params)
	}

	return
}

// GetTopPhotoInfosByTIDs 获取日志对应的审核信息
func (d *Dao) GetTopPhotoInfosByTIDs(params *model.VipAuditLogSearch, tids []int64) (memberUploadTopPhotoList []*model.MemberUploadTopPhoto, err error) {
	memberUploadTopPhotoList = make([]*model.MemberUploadTopPhoto, 0)

	photoQuery := d.DB.Model(model.MemberUploadTopPhoto{}).Where("id IN (?)", tids)
	if params.UploadTimeStart != "" {
		photoQuery = photoQuery.Where("upload_date >= ?", params.UploadTimeStart)
	}
	if params.UploadTimeEnd != "" {
		photoQuery = photoQuery.Where("upload_date <= ?", params.UploadTimeEnd)
	}
	if len(params.Status) > 0 {
		photoQuery = photoQuery.Where("status IN (?)", params.Status)
	}
	if len(params.Platfrom) > 0 {
		photoQuery = photoQuery.Where("platfrom IN (?)", params.Platfrom)
	}
	if err = photoQuery.Find(&memberUploadTopPhotoList).Error; err != nil {
		return nil, errors.Wrapf(err, "Dao: GetTopPhotoInfosByTIDs (%+v, %+v)", params, tids)
	}

	return
}
