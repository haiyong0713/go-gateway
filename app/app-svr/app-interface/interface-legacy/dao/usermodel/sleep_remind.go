package usermodel

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-interface/interface-legacy/model/anti_addiction"
)

const (
	_sleepRemindSQL       = "SELECT id,mid,switch,stime,etime FROM `sleep_remind` WHERE mid=?"
	_addSleepRemindSQL    = "INSERT INTO `sleep_remind` (mid,switch,stime,etime) VALUES (?,?,?,?)"
	_updateSleepRemindSQL = "UPDATE `sleep_remind` SET switch=?,stime=?,etime=? WHERE id=?"
)

func (d *dao) RawSleepRemind(ctx context.Context, mid int64) (*model.SleepRemind, error) {
	row := d.db.QueryRow(ctx, _sleepRemindSQL, mid)
	val := &model.SleepRemind{}
	if err := row.Scan(&val.ID, &val.Mid, &val.Switch, &val.Stime, &val.Etime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("Fail to scan sleep_remind, mid=%+v error=%+v", mid, err)
		return nil, err
	}
	return val, nil
}

func (d *dao) addSleepRemind(ctx context.Context, val *model.SleepRemind) error {
	if _, err := d.db.Exec(ctx, _addSleepRemindSQL, val.Mid, val.Switch, val.Stime, val.Etime); err != nil {
		log.Error("Fail to create sleep_remind, val=%+v error=%+v", val, err)
		return err
	}
	return nil
}

func (d *dao) updateSleepRemind(ctx context.Context, val *model.SleepRemind) error {
	if _, err := d.db.Exec(ctx, _updateSleepRemindSQL, val.Switch, val.Stime, val.Etime, val.ID); err != nil {
		log.Error("Fail to update sleep_remind, val=%+v error=%+v", val, err)
		return err
	}
	return nil
}
