package laser

import (
	"context"
	"errors"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// AppLaserCmdList get app laser list.
func (s *Service) AppLaserCmdList(c context.Context, appKey, platform, buvid, action, operator string, mid, taskID int64,
	status, pn, ps int) (res *model.LaserCmdResult, err error) {
	var total int
	if total, err = s.fkDao.LaserCmdCount(c, appKey, platform, buvid, action, operator, mid, taskID, status); err != nil {
		log.Error("%v", err)
		return
	}
	if total == 0 {
		return
	}
	var ls []*appmdl.LaserCmd
	if ls, err = s.fkDao.LaserCmdList(c, appKey, platform, buvid, action, operator, taskID, mid, status, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	res = &model.LaserCmdResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
		Items: ls,
	}
	return
}

// AppLaserCmdAdd add laser command
func (s *Service) AppLaserCmdAdd(c context.Context, appKey, buvid, action, description, paramsStr, operator string, mid int64) (err error) {
	var app *appmdl.APP
	if app, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if app == nil {
		return errors.New("该app没有在fawkes平台注册")
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
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
	var rowId int64
	if rowId, err = s.fkDao.TxAddLaserCmd(tx, appKey, app.MobiApp, app.Platform, buvid, action, description, paramsStr, operator, mid); err != nil {
		log.Error("%v", err)
		return
	}
	// 推送Broadcast消息
	pushError := s.LaserPushCommand(context.Background(), appKey, &appmdl.LaserCmd{
		Action:  action,
		Params:  paramsStr,
		ID:      rowId,
		Buvid:   buvid,
		MID:     mid,
		MobiApp: app.MobiApp,
	})
	var status int
	if pushError != nil {
		status = model.LaserParseStatusFailed
	} else {
		status = model.LaserParseStatusRunning
	}
	_, _ = s.fkDao.TxUpLaserCmdStatus(tx, rowId, status, "", "")
	return
}

// AppLaserCmdDel delete laser cmd
func (s *Service) AppLaserCmdDel(c context.Context, taskID int64) (err error) {
	if err = s.fkDao.DelLaserCmd(c, taskID); err != nil {
		log.Error("AppLaserCmdDel error: %v", err)
	}
	return
}

// AppLaserCmdActionAdd add laser cmd action
func (s *Service) AppLaserCmdActionAdd(c context.Context, name, platform, paramsJSON, operator, description string) (err error) {
	if err = s.fkDao.AddLaserCmdAction(c, name, platform, paramsJSON, operator, description); err != nil {
		log.Error("AddLaserCmdAction error: %v", err)
	}
	return
}

// AppLaserCmdActionUpdate update laser cmd action
func (s *Service) AppLaserCmdActionUpdate(c context.Context, id int64, name, platform, paramsJSON, operator, description string) (err error) {
	if err = s.fkDao.UpdateLaserCmdAction(c, id, name, platform, paramsJSON, operator, description); err != nil {
		log.Error("UpdateLaserCmdAction error: %v", err)
	}
	return
}

// AppLaserCmdActionDel del laer cmd action
func (s *Service) AppLaserCmdActionDel(c context.Context, id int64) (err error) {
	if err = s.fkDao.DelLaserCmdAction(c, id); err != nil {
		log.Error("delLaserCmdAction error: %v", err)
	}
	return
}

// LaserCmdActionList get laser cmd action list
func (s *Service) LaserCmdActionList(c context.Context, name, platform string) (res []*appmdl.LaserCmdAction, err error) {
	if res, err = s.fkDao.LaserCmdActionList(c, name, platform); err != nil {
		log.Error("s.fkDao.LaserCmdActionList err: %v", err)
	}
	return
}
