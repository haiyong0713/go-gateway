package usermodel

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-interface/interface-legacy/model/usermodel"
)

const (
	_addUserModelSQL          = "INSERT INTO teenager_users (mid,password,state,model,operation,quit_time,pwd_type) VALUES (?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),password=VALUES(password),state=VALUES(state),model=VALUES(model),operation=VALUES(operation),quit_time=VALUES(quit_time),pwd_type=VALUES(pwd_type)"
	_getUserModelAllSQL       = "SELECT id,mid,password,state,model,operation,quit_time,pwd_type,manual_force,mf_operator,mf_time FROM teenager_users WHERE mid=?"
	_addDeviceUserModelSQL    = "INSERT INTO device_user_model (mobi_app,device_token,password,state,model,operation,quit_time,pwd_type) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id),password=VALUES(password),state=VALUES(state),model=VALUES(model),operation=VALUES(operation),quit_time=VALUES(quit_time),pwd_type=VALUES(pwd_type)"
	_getDeviceUserModelAllSQL = "SELECT id,mobi_app,device_token,password,state,model,operation,quit_time,pwd_type FROM device_user_model WHERE mobi_app=? AND device_token=?"
	_addSpModeLogSQL          = "INSERT INTO special_mode_log (related_key,operator_uid,operator,content) VALUES (?,?,?,?)"
	_addManualForceLogSQL     = "INSERT INTO teenager_manual_log (mid, operator, content, remark) VALUES (?,?,?,?)"
	_updateManualForceSQL     = "UPDATE teenager_users SET manual_force=?,mf_operator=?,mf_time=? WHERE id=?"
	_updateOperationSQL       = "UPDATE teenager_users SET operation=? WHERE id=?"
	_getTeenagerModelPWDSQL   = "SELECT pwd FROM teenager_model_wsxcde WHERE wsxcde = ?"
)

func (d *dao) RawUserModels(ctx context.Context, mid int64, mobiApp, deviceToken string) ([]*usermodel.User, error) {
	if mid != 0 {
		rows, err := d.db.Query(ctx, _getUserModelAllSQL, mid)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var users []*usermodel.User
		for rows.Next() {
			user := &usermodel.User{}
			if err = rows.Scan(&user.ID, &user.Mid, &user.Password, &user.State, &user.Model, &user.Operation, &user.QuitTime, &user.PwdType, &user.ManualForce, &user.MfOperator, &user.MfTime); err != nil {
				return nil, err
			}
			users = append(users, user)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return users, nil
	}
	rows, err := d.db.Query(ctx, _getDeviceUserModelAllSQL, mobiApp, deviceToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*usermodel.User
	for rows.Next() {
		user := &usermodel.User{}
		if err = rows.Scan(&user.ID, &user.MobiApp, &user.DeviceToken, &user.Password, &user.State, &user.Model, &user.Operation, &user.QuitTime, &user.PwdType); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (d *dao) addUserModel(ctx context.Context, user *usermodel.User) (userID int64, devID int64, err error) {
	if user.Mid != 0 {
		res, err := d.db.Exec(ctx, _addUserModelSQL, user.Mid, user.Password, user.State, user.Model, user.Operation, user.QuitTime, user.PwdType)
		if err != nil {
			log.Error("Fail to create teenager_users, user=%+v error=%+v", user, err)
			return 0, 0, err
		}
		if userID, err = res.LastInsertId(); err != nil {
			log.Error("Fail to get LastInsertId of teenager_users, user=%+v error=%+v", user, err)
			return 0, 0, err
		}
		return userID, 0, nil
	}
	res, err := d.db.Exec(ctx, _addDeviceUserModelSQL, user.MobiApp, user.DeviceToken, user.Password, user.State, user.Model, user.DevOperation, user.QuitTime, user.PwdType)
	if err != nil {
		log.Error("Fail to create device_user_model, user=%+v error=%+v", user, err)
		return 0, 0, err
	}
	if devID, err = res.LastInsertId(); err != nil {
		log.Error("Fail to get LastInsertId of device_user_model, user=%+v error=%+v", user, err)
		return 0, 0, err
	}
	return 0, devID, nil
}

func (d *dao) addSyncUserModel(ctx context.Context, user *usermodel.User) (userID int64, devID int64, err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	userRes, err := tx.Exec(_addUserModelSQL, user.Mid, user.Password, user.State, user.Model, user.Operation, user.QuitTime, user.PwdType)
	if err != nil {
		log.Error("Fail to create teenager_users, user=%+v error=%+v", user, err)
		return 0, 0, err
	}
	if userID, err = userRes.LastInsertId(); err != nil {
		log.Error("Fail to get LastInsertId of teenager_users, user=%+v error=%+v", user, err)
		return 0, 0, err
	}
	devRes, err := tx.Exec(_addDeviceUserModelSQL, user.MobiApp, user.DeviceToken, user.Password, user.State, user.Model, user.DevOperation, user.QuitTime, user.PwdType)
	if err != nil {
		log.Error("Fail to create device_user_model, user=%+v error=%+v", user, err)
		return 0, 0, err
	}
	if devID, err = devRes.LastInsertId(); err != nil {
		log.Error("Fail to get LastInsertId of device_user_model, user=%+v error=%+v", user, err)
		return 0, 0, err
	}
	return userID, devID, nil
}

func (d *dao) AddSpecialModeLog(ctx context.Context, fields *usermodel.SpecialModeLog) error {
	if _, err := d.db.Exec(ctx, _addSpModeLogSQL, fields.RelatedKey, fields.OperatorUid, fields.Operator, fields.Content); err != nil {
		log.Error("Fail to create special_mode_log, fields=%+v error=%+v", fields, err)
		return err
	}
	return nil
}

func (d *dao) AddManualForceLog(ctx context.Context, fields *usermodel.ManualForceLog) error {
	if _, err := d.db.Exec(ctx, _addManualForceLogSQL, fields.Mid, fields.Operator, fields.Content, fields.Remark); err != nil {
		log.Error("Fail to create teenager_manual_log, item=%+v error=%+v", fields, err)
		return err
	}
	return nil
}

func (d *dao) updateManualForce(ctx context.Context, id, mf int64, mfTime xtime.Time, mfOperator string) error {
	if _, err := d.db.Exec(ctx, _updateManualForceSQL, mf, mfOperator, mfTime, id); err != nil {
		log.Error("Fail to updateManualForce, id=%+v error=%+v", id, err)
		return err
	}
	return nil
}

// updateOperationDB .
func (d *dao) updateOperationDB(ctx context.Context, id int64, op int) error {
	if _, err := d.db.Exec(ctx, _updateOperationSQL, op, id); err != nil {
		log.Error("Fail to updateOperation, id=%+v error=%+v", id, err)
		return err
	}
	return nil
}

func (d *dao) GetTeenagerModelPWD(ctx context.Context, wsxcde string) (string, error) {
	var pwd string
	if err := d.db.QueryRow(ctx, _getTeenagerModelPWDSQL, wsxcde).Scan(&pwd); err != nil {
		return "", err
	}
	if pwd == "" {
		return "", errors.Wrap(ecode.ServerErr, "pwd is empty")
	}
	return pwd, nil
}
