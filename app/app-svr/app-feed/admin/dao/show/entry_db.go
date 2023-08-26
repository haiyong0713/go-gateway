package show

import (
	"context"
	"time"

	"go-common/library/log"
	model "go-gateway/app/app-svr/app-feed/admin/model/entry"
)

// -------------------------- 业务dao --------------------------
var (
	_dateFormat    = "2006-01-02 15:04:05"
	_notPushedTime = "2009-12-31 23:59:59"
	//_onlineEntriesKey = "online-entries"
)

// APP_ENTRY
// 新增入口
func (d *Dao) CreateEntry(entry *model.AppEntry) (err error) {
	return d.DB.Model(model.AppEntry{}).Create(entry).Error
}

// 删除入口
func (d *Dao) DeleteEntry(entryID int32) (err error) {
	return d.DB.Model(model.AppEntry{}).Where("id = ?", entryID).Update("is_deprecated", 1).Error
}

// 编辑入口
func (d *Dao) EditEntry(_ context.Context, entry *model.AppEntry) (err error) {
	//if mcErr := d.MC.Delete(ctx, _onlineEntriesKey); mcErr != nil {
	//	log.Error("EditEntry from MC error")
	//}
	attrToUpdate := map[string]interface{}{
		"entry_name": entry.EntryName,
		"stime":      entry.STime,
		"etime":      entry.ETime,
		"platforms":  entry.Platforms,
		"created_by": entry.CreatedBy,
		"total_loop": entry.TotalLoop,
	}
	return d.DB.Model(entry).Where("id = ? AND is_deprecated = 0", entry.ID).Updates(attrToUpdate).Error
}

// 入口上线/下线状态切换
func (d *Dao) ToggleEntry(_ context.Context, entryID int32, onlineStatus int32) (err error) {
	//if mcErr := d.MC.Delete(ctx, _onlineEntriesKey); mcErr != nil {
	//	log.Error("ToggleEntry from MC error")
	//}
	record := &model.AppEntry{}
	if err = d.DB.Model(record).Where("id = ? AND is_deprecated = 0", entryID).First(record).Error; err != nil {
		return
	}
	if onlineStatus != 0 && onlineStatus != 1 {
		onlineStatus = record.OnlineStatus*-1 + 1
	}
	return d.DB.Model(record).
		Where("id = ? AND is_deprecated = 0", entryID).
		Update("online_status", onlineStatus).Error
}

// 检查是否有已经在线的Entry
func (d *Dao) CheckEffectiveEntryWithTime(entryID int32, stime time.Time, etime time.Time) (result []*model.AppEntry, err error) {
	err = d.DB.Model(&model.AppEntry{}).
		Not("id", entryID).
		Where("online_status = 1 AND is_deprecated = 0").
		Where(
			"(stime <= ? AND etime > ?) or (stime < ? AND etime >= ?) or (stime >= ? AND etime <= ?)",
			stime, stime, etime, etime, stime, etime,
		).Find(&result).Error

	if err != nil {
		return nil, err
	}
	return result, err
}

// 根据id查entry
func (d *Dao) GetEntryById(entryID int32, record *model.AppEntry) (err error) {
	return d.DB.Model(&model.AppEntry{}).Where("id = ? AND is_deprecated = 0", entryID).First(record).Error
}

// 获取Entry的total
func (d *Dao) GetEntryCount() (count int32, err error) {
	err = d.DB.Model(model.AppEntry{}).Where("is_deprecated = 0").Count(&count).Error
	return
}

// 获取所有Entry
func (d *Dao) GetEntryList(ps int32, pn int32) (result []*model.AppEntry, err error) {
	err = d.DB.Model(model.AppEntry{}).Where("is_deprecated = 0").
		Order("id desc").
		Limit(ps).
		Offset(ps * (pn - 1)).
		Scan(&result).Error
	return
}

// 获取所有线上的Entry
func (d *Dao) GetEffectiveEntry(_ context.Context) (result []*model.AppEntry, err error) {
	now := time.Now().Format(_dateFormat)
	err = d.DB.Model(&model.AppEntry{}).
		Where("online_status = 1 AND is_deprecated = 0 AND stime <= ? AND etime > ?", now, now).
		Find(&result).Error

	if err != nil {
		if err.Error() == "-404" {
			log.Error("GetOnlineEntryList from db with nil")
		} else {
			log.Error("GetOnlineEntryList from db error")
		}
	}
	return result, err
}

// APP_ENTRY_CONFIG
// 新增入口state
func (d *Dao) CreateEntryState(state *model.AppEntryState) (err error) {
	return d.DB.Model(state).Create(state).Error
}

// 删除入口state
func (d *Dao) DeleteEntryState(entryID int32) (err error) {
	return d.DB.Model(&model.AppEntryState{}).Where("entry_id = ?", entryID).Update("is_deprecated", 1).Error
}

// 编辑入口state
func (d *Dao) EditState(state *model.AppEntryState) (err error) {
	attrToUpdate := map[string]interface{}{
		"state_name":    state.StateName,
		"static_icon":   state.StaticIcon,
		"dynamic_icon":  state.DynamicIcon,
		"url":           state.Url,
		"is_deprecated": state.IsDeprecated,
		"loop_count":    state.LoopCount,
	}
	return d.DB.Model(state).Where("id = ?", state.ID).Updates(attrToUpdate).Error
}

// 根据entry_id获取所有state
func (d *Dao) GetStatesByEntryID(entryID int32) (result []*model.AppEntryState, err error) {
	err = d.DB.Model(&model.AppEntryState{}).
		Where("entry_id = ? AND is_deprecated = 0", entryID).
		Find(&result).Error
	return
}

// 根据state_id获取所有state
func (d *Dao) GetStateByID(stateID int32) (result *model.AppEntryState, err error) {
	result = &model.AppEntryState{}
	err = d.DB.Model(&model.AppEntryState{}).
		Where("id = ? AND is_deprecated = 0", stateID).
		First(&result).Error
	return
}

// 根据state id和entry id检查匹配关系
func (d *Dao) CheckEntryStatePair(entryID int32, stateID int32) (isPair bool, err error) {
	var count int
	if err = d.DB.Model(&model.AppEntryState{}).Where("id = ? AND entry_id = ?", stateID, entryID).Count(&count).Error; err != nil {
		return
	}
	isPair = count == 1
	return
}

// APP_ENTRY_TIME_SETTINGS
// 新增入口state的时间配置
func (d *Dao) CreateTimeSetting(setting *model.AppEntryTimeSetting) (err error) {
	return d.DB.Model(setting).Create(setting).Error
}

// 新增入口state的时间配置
func (d *Dao) GetTimeSettingByEntryID(entryID int32) (result []model.AppEntryTimeSetting, err error) {
	err = d.DB.Model(&model.AppEntryTimeSetting{}).
		Where("entry_id = ?", entryID).
		Order("ctime asc").
		Scan(&result).Error
	return
}

// 推送数据后修改时间配置信息
func (d *Dao) ToggleEntryTimeSetting(settingID int32, sentLoop int32) (err error) {
	var now time.Time
	if now, err = time.ParseInLocation(_dateFormat, time.Now().Format(_dateFormat), time.Local); err != nil {
		return err
	}
	// 有时间需要注意
	attrToUpdate := map[string]interface{}{
		"push_time": now,
		"sent_loop": sentLoop,
	}

	return d.DB.Model(&model.AppEntryTimeSetting{}).
		Where("id = ? AND is_deprecated = 0", settingID).
		Updates(attrToUpdate).Error
}

// 获取已推送的时间配置
func (d *Dao) GetPushedTimeSettingsByEntryID(entryID int32) (result *model.AppEntryTimeSetting, err error) {
	result = &model.AppEntryTimeSetting{}
	err = d.DB.Model(result).
		Not("push_time = ?", _notPushedTime).
		Where("entry_id = ? AND is_deprecated = 0", entryID).
		Order("push_time desc").
		First(&result).Error
	return result, err
}

// 获取未推送的最新时间配置
func (d *Dao) GetNotPushedTimeSettingByEntryID(entryID int32) (result *model.AppEntryTimeSetting, err error) {
	result = &model.AppEntryTimeSetting{}
	err = d.DB.Model(&model.AppEntryTimeSetting{}).
		Where("entry_id = ? AND is_deprecated = 0 AND push_time = ?", entryID, _notPushedTime).
		First(&result).Error
	return result, err
}

// 删除所有未推送的时间配置
func (d *Dao) DisableNotPushedTimeSettingByEntryID(entryID int32) (err error) {
	return d.DB.Model(&model.AppEntryTimeSetting{}).
		Where("entry_id = ? AND is_deprecated = 0 AND push_time = ?", entryID, _notPushedTime).
		Update("is_deprecated", 1).Error
}

// 根据id删除所有推送时间配置
func (d *Dao) DisableTimeSettingByEntryID(entryID int32) (err error) {
	return d.DB.Model(&model.AppEntryTimeSetting{}).
		Where("entry_id = ? AND is_deprecated = 0", entryID).
		Update("is_deprecated", 1).Error
}

// 根据id删除所有生效时间大于当前结束时间的推送时间配置
func (d *Dao) DisableTimeSettingByEntryIDWithEtime(entryID int32, etime time.Time) (err error) {
	return d.DB.Model(&model.AppEntryTimeSetting{}).
		Where("entry_id = ? AND is_deprecated = 0 AND stime >= ?", entryID, etime).
		Update("is_deprecated", 1).Error
}

// 根据id删除所有生效时间小于当前开始时间的推送时间配置
func (d *Dao) DisableTimeSettingByEntryIDWithStime(entryID int32, stime time.Time) (err error) {
	return d.DB.Model(&model.AppEntryTimeSetting{}).
		Where("entry_id = ? AND is_deprecated = 0 AND stime <= ?", entryID, stime).
		Update("is_deprecated", 1).Error
}

// 根据id删除所有生效时间小于当前开始时间的推送时间配置
func (d *Dao) GetSentLoopByEntryID(entryID int32) (sum int32, err error) {
	err = d.DB.Raw("select IFNULL(SUM(sent_loop), 0) from app_entry_time_settings where entry_id = ?", entryID).Row().Scan(&sum)
	return sum, err
}
