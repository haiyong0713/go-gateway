package laser

import (
	"context"
	"fmt"
	"time"

	"go-common/library/database/sql"
	"go-common/library/ecode"

	"go-gateway/app/app-svr/fawkes/service/model/app"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	lasermdl "go-gateway/app/app-svr/fawkes/service/model/laser"
	"go-gateway/app/app-svr/fawkes/service/model/template"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

// AppLaserList get app laser list.
func (s *Service) AppLaserList(c context.Context, appKey, platform, buvid, logDate, operator string, mid, taskID int64,
	status, pn, ps int) (res *model.LaserResult, err error) {
	var total int
	if total, err = s.fkDao.LaserCount(c, appKey, platform, buvid, operator, taskID, mid, 0, 0, status); err != nil {
		log.Error("%v", err)
		return
	}
	if total == 0 {
		return
	}
	var ls []*appmdl.Laser
	if ls, err = s.fkDao.LaserList(c, appKey, platform, buvid, operator, taskID, mid, 0, 0, status, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	res = &model.LaserResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
		Items: ls,
	}
	return
}

// AppLaserAdd add laser
func (s *Service) AppLaserAdd(c context.Context, appKey, platform, buvid, logDate, userName, description string, mid int64) (res interface{}, err error) {
	var (
		appInfo *appmdl.APP
		laserId int64
	)
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if appInfo == nil {
		err = ecode.Error(ecode.RequestErr, "app_key没有找到对应的应用，请联系管理员。")
		return
	}
	if platform == "" {
		platform = appInfo.Platform
	}
	_ = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		laserId, err = s.fkDao.TxAddLaser(tx, appKey, platform, buvid, logDate, userName, description, appInfo.MobiApp, mid, appmdl.StatusQueuing, appmdl.StatusQueuing, appmdl.ChannelFawkes)
		return err
	})
	// 推送消息. 触达至用户
	msgId, pushError := s.LaserPush(utils.CopyTrx(c), appInfo, &appmdl.Laser{
		ID:      laserId,
		Buvid:   buvid,
		MID:     mid,
		MobiApp: appInfo.MobiApp,
		LogDate: logDate,
	})
	var status = model.LaserParseStatusRunning
	if pushError != nil {
		status = model.LaserParseStatusFailed
		log.Errorc(c, "AppLaserAdd error: %v", pushError)
	}
	_ = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		_, err = s.fkDao.TxUpLaserStatus(tx, laserId, status, "", "")
		err = s.fkDao.TxUpMsgId(tx, laserId, msgId)
		return err
	})
	return lasermdl.AddLaserResp{
		ID: laserId,
	}, err
}

func (s *Service) LaserPush(ctx context.Context, app *appmdl.APP, laser *app.Laser) (msgId int64, err error) {
	if len(app.LaserWebhook) == 0 {
		return s.LaserPushLogUpload(ctx, app.AppKey, laser)
	} else {
		return s.LaserPushByWebhook(ctx, app, laser)
	}
}

// AppLaserDel del laser.
func (s *Service) AppLaserDel(c context.Context, taskID int64) (err error) {
	_ = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if _, err = s.fkDao.TxDelLaser(tx, taskID); err != nil {
			log.Error("%v", err)
		}
		return err
	})
	return
}

// AppLaserReport business report laser status.
func (s *Service) AppLaserReport(c context.Context, taskID int64, status int, url, mobiApp, build, errMsg, md5, rawUposUri string) (err error) {
	var laser *appmdl.Laser
	if laser, err = s.fkDao.Laser(c, taskID); err != nil {
		log.Error("%v", err)
		return
	}
	if laser == nil {
		log.Error("AppLaserReport laser not found. task_id=%v mobi_app=%v", taskID, mobiApp)
		return
	}
	// 上传成功的任务. 不再重置其配置信息
	if laser.Status == appmdl.StatusUpSuccess {
		log.Error("%v", err)
		return
	}
	err = s.fkDao.Transact(c, func(tx *sql.Tx) (txError error) {
		if _, txError = s.fkDao.TxUpLaserStatus(tx, taskID, status, mobiApp, build); err != nil {
			log.Error("%v", err)
			return
		}
		if status == appmdl.StatusUpSuccess {
			if _, txError = s.fkDao.TxUpLaserURL(tx, taskID, url, md5, rawUposUri); err != nil {
				log.Error("%v", err)
				return
			}
		}
		if status == appmdl.StatusUpFaild {
			if _, txError = s.fkDao.TxUpLaserErrorMessage(tx, taskID, errMsg); err != nil {
				log.Error("%v", err)
			}
		}
		return
	})
	if err != nil {
		log.Error("%v", err)
		return
	}
	if status == appmdl.LaserCmdStatusSuccess {
		var msgText string
		if msgText, err = s.fkDao.TemplateAlter(laser, template.LaserReportTemp_WeChat); err != nil {
			log.Error("%v", err)
			return
		}
		_ = s.fkDao.WechatCardMessageNotify(
			"Laser 日志拉取成功",
			msgText,
			fmt.Sprintf("%v/#/laser/task-list?app_key=%v&task_id=%v", s.c.Host.Fawkes, laser.AppKey, laser.ID),
			"",
			laser.Operator,
			s.c.Comet.FawkesAppID)
	}
	return
}

// AppLaserReportSilence business report laser status.
func (s *Service) AppLaserReportSilence(c context.Context, taskID int64, status int, url, recallMobiApp, build, errorMessage string) (err error) {
	var (
		tx        *sql.Tx
		laserTask *appmdl.Laser
	)
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
	// 由于网关逻辑异常. 打捞数据时.会触发端上的静默上报. 将iphone粉版数据推上来. 这边做一层过滤
	if laserTask, err = s.fkDao.Laser(c, taskID); err != nil {
		log.Error("AppLaserReportSilence error: %v", err)
		return
	}
	if laserTask.AppKey != recallMobiApp {
		log.Error("AppLaserReportSilence mobiapp error; TaskID: %v", taskID)
		return
	}
	if _, err = s.fkDao.TxUpLaserSilenceStatus(tx, taskID, status, recallMobiApp, build); err != nil {
		log.Error("%v", err)
		return
	}
	if status == appmdl.StatusUpSuccess {
		if _, err = s.fkDao.TxUpLaserSilenceURL(tx, taskID, url); err != nil {
			log.Error("%v", err)
			return
		}
	}
	if status == appmdl.StatusUpFaild {
		if _, err = s.fkDao.TxUpLaserErrorMessage(tx, taskID, errorMessage); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

// LaserUser get laser user info.
func (s *Service) LaserUser(c context.Context, appKey, operator string, mid, startTime int64) (res *appmdl.Laser, err error) {
	var (
		activeLasers []*appmdl.Laser
		normalLasers []*appmdl.Laser
		appInfo      *appmdl.APP
	)
	// 1. 查询用户是否已经上报了laser主动反馈数据
	if activeLasers, err = s.fkDao.LaserActiveList(c, appKey, "", mid, 0, startTime, 0, 1, 10); err != nil {
		log.Error("%v", err)
		return
	}
	if len(activeLasers) > 0 {
		for _, active := range activeLasers {
			if active.URL != "" {
				res = active
				return
			}
		}
	}
	// 2. 若active表内不存在数据，则查询主动laser表内是否含有数据
	if normalLasers, err = s.fkDao.LaserList(c, appKey, "", "", "", 0, mid, startTime, 0, 0, 1, 20); err != nil {
		log.Error("%v", err)
		return
	}
	if len(normalLasers) > 0 {
		for _, laser := range normalLasers {
			if laser.SilenceURL != "" && laser.SilenceStatus == appmdl.StatusUpSuccess {
				res = laser
				return
			} else if laser.URL != "" && laser.Status == appmdl.StatusUpSuccess {
				res = laser
				return
			} else if laser.Status != appmdl.StatusUpFaild && laser.Status != appmdl.StatusSendFaild {
				res = laser
			}
		}
	}
	if res != nil {
		return
	}
	// 3. 若laser表不存在数据，则新建一个laser
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	logDate := time.Unix(startTime, 0).Format("2006-01-02")
	newLaser, err := s.AppLaserAdd(c, appKey, appInfo.Platform, "", logDate, operator, "", mid)
	resp := newLaser.(lasermdl.AddLaserResp)
	res = &appmdl.Laser{ID: resp.ID, AppKey: appKey, Platform: appInfo.Platform, MID: mid, Buvid: "", Email: "", LogDate: logDate, URL: "", Status: appmdl.StatusQueuing, Operator: operator, CTime: 0, MTime: 0, SilenceURL: "", SilenceStatus: appmdl.StatusQueuing, ParseStatus: 0, Channel: appmdl.ChannelBusinessApi, Description: "", MobiApp: appInfo.MobiApp, RecallMobiApp: "", Build: "", ErrorMessage: ""}
	return
}

//// AppLaserParseStatusUpdate update parse_status
//func (s *Service) AppLaserParseStatusUpdate(c context.Context, status int, laserID int64, laserType, operator, appKey string) (err error) {
//	var tx *sql.Tx
//	if tx, err = s.fkDao.BeginTran(c); err != nil {
//		log.Error("s.fkDao.BeginTran() error(%v)", err)
//		return
//	}
//	defer func() {
//		if r := recover(); r != nil {
//			//nolint:errcheck
//			tx.Rollback()
//			log.Error("%v", r)
//		}
//		if err != nil {
//			if err1 := tx.Rollback(); err1 != nil {
//				log.Error("tx.Rollback() error(%v)", err1)
//			}
//			return
//		}
//		if err = tx.Commit(); err != nil {
//			log.Error("tx.Commit() error(%v)", err)
//		}
//	}()
//
//	var (
//		ls        []*appmdl.Laser
//		laserInfo *appmdl.Laser
//		content   string
//		app       *appmdl.APP
//	)
//	if app, err = s.fkDao.AppPass(c, appKey); err != nil {
//		log.Error("%v", err)
//		return
//	}
//	template := "应用： %v(%v)\n" +
//		"任务ID：%v\n" +
//		"描述信息：%v\n" +
//		"操作人：%v"
//	if laserType == model.LaserTypeTask {
//		if err = s.fkDao.TxUpLaserParseStatus(tx, status, laserID); err != nil {
//			log.Error("%v", err)
//		}
//		if ls, err = s.fkDao.LaserList(c, appKey, "", "", "", laserID, 0, 0, 0, 0, 1, 20); err != nil {
//			log.Error("%v", err)
//			return
//		}
//		if len(ls) == 0 {
//			err = errors.New("laser not found")
//			return
//		}
//		laserInfo = ls[0]
//		content = fmt.Sprintf(template, app.Name, appKey, laserID, laserInfo.Description, operator)
//	} else {
//		if err = s.fkDao.TxUpActiveLaserParseStatus(tx, status, laserID); err != nil {
//			log.Error("%v", err)
//		}
//		content = fmt.Sprintf(template, app.Name, appKey, laserID, "-", operator)
//	}
//	var (
//		title, link, operators string
//	)
//	if operator == "caijian" || operator == "jinjianxiang" {
//		operators = operator
//	} else {
//		operators = fmt.Sprintf("%v|caijian|jinjianxiang", operator)
//	}
//	if status == model.LaserParseStatusSuccess {
//		title = "Laser日志解析成功"
//		link = fmt.Sprintf("http://fawkes.bilibili.co/#/laser/log-list?app_key=%v&lid=%v", appKey, laserID)
//	} else {
//		if status == model.LaserParseStatusFailed {
//			title = "Laser日志解析失败"
//		} else if status == model.LaserParseStatusWaiting {
//			title = "Laser日志解析任务开始"
//		}
//
//		url := "http://fawkes.bilibili.co/#/laser/%v?app_key=%v"
//		if laserType == model.LaserTypeTask {
//			link = fmt.Sprintf(url, "task-list", appKey)
//		} else {
//			link = fmt.Sprintf(url, "active-task-list", appKey)
//		}
//	}
//	if title != "" {
//		_ = s.fkDao.WechatCardMessageNotify(
//			title,
//			content,
//			link,
//			"",
//			operators,
//			s.c.Comet.FawkesAppID)
//	}
//	return
//}

//func (s *Service) AppLaserPendingList(c context.Context) (res *model.LaserPendingResult, err error) {
//	res = &model.LaserPendingResult{}
//	if res.LogUploadList, err = s.fkDao.LaserPendingAll(c); err != nil {
//		log.Error("%v", err)
//	}
//	return
//}
