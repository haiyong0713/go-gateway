package entry

import (
	"context"
	"time"

	"go-gateway/app/app-svr/resource/service/model"
)

// -------------------------- 业务dao --------------------------
const (
	dateFormat = "2006-01-02 15:04:05"
	//notPushedTime    = "2009-12-31 23:59:59"
	getEffectiveEntrySQL              = "select id, entry_name, platforms, etime from app_entries where online_status = 1 and is_deprecated = 0 and stime <= ? and etime > ?"
	GetStateByIDSQL                   = "select id, state_name, url, static_icon, dynamic_icon, loop_count, entry_id from app_entry_states where id = ? and is_deprecated = 0"
	GetPushedTimeSettingsByEntryIDSQL = "select id, entry_id, state_id, stime, push_time, sent_loop from app_entry_time_settings where entry_id = ? AND is_deprecated = 0 and push_time != '2009-12-31 23:59:59' order by push_time desc"
)

// APP_ENTRY
// 获取所有在线Entry
func (d *Dao) GetEffectiveEntries(ctx context.Context) (result []*model.AppEntry, err error) {
	now := time.Now().Format(dateFormat)
	rows, err := d.db.Query(ctx, getEffectiveEntrySQL, now, now)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		temp := &model.AppEntry{}
		if err = rows.Scan(&temp.ID, &temp.EntryName, &temp.Platforms, &temp.ETime); err != nil {
			return
		}
		result = append(result, temp)
	}

	err = rows.Err()
	return result, err
}

// APP_ENTRY_STATE
// 根据state_id获取所有state
func (d *Dao) GetStateByID(ctx context.Context, stateID int32) (result *model.AppEntryState, err error) {
	rows, err := d.db.Query(ctx, GetStateByIDSQL, stateID)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		result = &model.AppEntryState{}
		if err = rows.Scan(&result.ID, &result.StateName, &result.Url, &result.StaticIcon, &result.DynamicIcon, &result.LoopCount, &result.EntryID); err != nil {
			return
		}
	}
	err = rows.Err()
	return
}

// APP_ENTRY_TIME_SETTINGS
// 获取已推送的时间配置
func (d *Dao) GetPushedTimeSettingsByEntryID(ctx context.Context, entryID int32) (result *model.AppEntryTimeSetting, err error) {
	rows, err := d.db.Query(ctx, GetPushedTimeSettingsByEntryIDSQL, entryID)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		result = &model.AppEntryTimeSetting{}
		if err = rows.Scan(&result.ID, &result.EntryID, &result.StateID, &result.STime, &result.PushTime, &result.SentLoop); err != nil {
			return
		}
	}
	err = rows.Err()
	return result, err
}
