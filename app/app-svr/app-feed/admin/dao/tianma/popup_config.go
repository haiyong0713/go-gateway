package tianma

import (
	xecode "go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	model "go-gateway/app/app-svr/app-feed/admin/model/tianma"
	"time"

	"github.com/jinzhu/gorm"
)

// GetPopupConfigWithConflictDuration 查询时间冲突的配置。若为编辑需要排除掉自身
func (d *Dao) GetPopupConfigWithConflictDuration(popupConfig *model.PopupConfig) (configList []*model.PopupConfig, err error) {
	if popupConfig == nil {
		err = xecode.NothingFound
		return nil, err
	}
	query := d.DB.Model(&model.PopupConfig{}).
		Where("deleted_flag = ?", model.PopupNotDeleted).
		Where("stime <= ? AND etime >= ? AND audit_state != ? AND audit_state != ?", popupConfig.ETime, popupConfig.STime, model.PopupAuditStateOffline, model.PopupAuditStateCanceled)

	if popupConfig.ID != 0 {
		query = query.Where("id != ?", popupConfig.ID)
	}

	if err = query.Find(&configList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}
	return
}

// GetPopupConfigByID 根据ID查询弹窗配置数据
func (d *Dao) GetPopupConfigByID(id int64) (configDetail *model.PopupConfig, err error) {
	configDetail = &model.PopupConfig{}
	if err = d.DB.Model(&model.PopupConfig{}).
		Where("deleted_flag = ?", model.PopupNotDeleted).
		First(&configDetail, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}
	return
}

// GetPopupConfigAll 查询弹窗配置数据
func (d *Dao) GetPopupConfigAll(ps int, pn int, order string) (configList []*model.PopupConfig, total int64, err error) {
	query := d.DB.Model(&model.PopupConfig{}).Where("deleted_flag = ?", model.PopupNotDeleted).Order(order)
	if err = query.Count(&total).Error; err != nil {
		return
	}
	if ps > 0 && pn >= 1 {
		query = query.Offset(ps * (pn - 1)).Limit(ps)
	}
	if err = query.Find(&configList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			configList = make([]*model.PopupConfig, 0)
			total = 0
		}
		return
	}
	return
}

// GetPopupConfigOnline 查询生效中弹窗配置数据
func (d *Dao) GetPopupConfigOnline() (configDetail *model.PopupConfig, err error) {
	configDetail = &model.PopupConfig{}
	now := time.Now()
	query := d.DB.Model(&model.PopupConfig{}).Where("deleted_flag = ? AND stime <= ? AND etime >= ? AND audit_state != ?", model.PopupNotDeleted, now, now, model.PopupAuditStateOffline)
	if err = query.First(configDetail).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}
	return
}

// AddPopupConfig 添加一个弹窗配置
func (d *Dao) AddPopupConfig(toAddConfig *model.PopupConfig, username string, uid int64) (id int64, err error) {
	insertItem := toAddConfig
	// 默认值处理
	insertItem.CUser = username
	insertItem.MUser = username
	now := xtime.Time(time.Now().Unix())
	insertItem.CTime = now
	insertItem.MTime = now
	if insertItem.ReType == 0 {
		// 默认跳转为不跳转
		insertItem.ReType = model.PopupReTypeNone
	}
	if insertItem.PopupType == 0 {
		// 默认为天马业务弹窗
		insertItem.PopupType = model.PopupTypeBusiness
	}
	if insertItem.TeenagePushFlag == 0 {
		// 默认青少年模式不推送
		insertItem.TeenagePushFlag = model.PopupTeenageNotPush
	}
	if insertItem.AutoHideStatus == 0 {
		// 默认自动关闭
		insertItem.AutoHideStatus = model.PopupAutoHideStatusHide
	}
	if insertItem.AutoHideCountdown == 0 {
		// 默认倒数x秒关闭
		insertItem.AutoHideCountdown = conf.Conf.Popup.AutoHideCountdown
	}
	if insertItem.CrowdType == 0 {
		// 默认不定向
		insertItem.CrowdType = model.PopupCrowdTypeNone
	}
	if insertItem.CrowdBase == 0 {
		// 默认是MID定向
		insertItem.CrowdBase = model.PopupCrowdBaseBGroupMID
	}
	if insertItem.AuditState == 0 {
		// 默认通过审核
		insertItem.AuditState = model.PopupAuditStatePass
	}
	if err = d.DB.Model(&model.PopupConfig{}).Create(insertItem).Error; err != nil {
		log.Error("popupConfig.dao.AddPopupConfig error(%v)", err)
	}
	id = insertItem.ID
	return
}

// UpdatePopupConfig 更新弹窗配置
func (d *Dao) UpdatePopupConfig(updateConfig *model.PopupConfig, username string, uid int64) (err error) {
	toUpdateConfig := updateConfig
	toUpdateConfig.MUser = username
	toUpdateConfig.MTime = xtime.Time(time.Now().Unix())
	updateQuery := d.DB.Model(&model.PopupConfig{}).
		Where("id = ?", updateConfig.ID)
	if err = updateQuery.Updates(toUpdateConfig).Error; err != nil {
		log.Error("popupConfig.dao.UpdatePopupConfig error(%v)", err)
	}
	return
}

// DeletePopupConfig 删除弹窗配置
func (d *Dao) DeletePopupConfig(deleteConfig *model.PopupConfig, username string, uid int64) (err error) {
	toDeleteConfig := &model.PopupConfig{
		DeletedFlag: model.PopupIsDeleted,
		MUser:       username,
		MTime:       xtime.Time(time.Now().Unix()),
	}
	deleteQuery := d.DB.Model(&model.PopupConfig{}).
		Where("id = ?", deleteConfig.ID)
	if err = deleteQuery.Updates(toDeleteConfig).Error; err != nil {
		log.Error("popupConfig.dao.DeletePopupConfig error(%v)", err)
	}
	return
}
