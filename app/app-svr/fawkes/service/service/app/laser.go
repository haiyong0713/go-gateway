package app

import (
	"context"
	"fmt"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func (s *Service) AppLaserCmdReport(c context.Context, taskID int64, status int, recallMobiApp, build, url, errorMsg, result, md5, rawUposUri string) (err error) {
	var (
		laserCmdInfo *appmdl.LaserCmd
		needUpdate   = true
	)
	if laserCmdInfo, err = s.fkDao.LaserCmdInfo(c, taskID); err != nil || laserCmdInfo == nil {
		log.Error("%v", err)
		return
	}
	// 如果状态已经成功了 或者 状态为 （4：业务到达）且laser状态为（2: 待上传）  则不进行更新操作
	if laserCmdInfo.Status == appmdl.LaserCmdStatusSuccess || (status == appmdl.LaserCmdStatusReceiveSuccess && laserCmdInfo.Status != appmdl.LaserCmdStatusWaiting) {
		needUpdate = false
	}
	if needUpdate {
		if err = s.fkDao.TxUpLaserCmd(c, taskID, status, recallMobiApp, build, url, errorMsg, result, md5, rawUposUri); err != nil {
			log.Error("AppLaserCmdReport error: %v", err)
			return
		}
	}
	if status == appmdl.LaserCmdStatusSuccess || status == appmdl.LaserCmdStatusSendFail || status == appmdl.LaserCmdStatusFail || status == appmdl.LaserCmdStatusUnsupport {
		var titleStatus string
		template := "- 应用： %v\n" +
			"- 任务ID：%v\n" +
			"- Action：%v\n" +
			"- 平台：%v\n" +
			"- mid：%v\n" +
			"- buvid：%v\n" +
			"- ErrMsg：%v\n" +
			"- 描述信息：%v\n" +
			"- 操作人：%v"
		if status == appmdl.LaserCmdStatusSuccess {
			titleStatus = "执行成功"
		} else {
			titleStatus = "执行失败"
		}
		content := fmt.Sprintf(template, laserCmdInfo.AppKey, taskID, laserCmdInfo.Action, laserCmdInfo.Platform, laserCmdInfo.MID, laserCmdInfo.Buvid, errorMsg, laserCmdInfo.Description, laserCmdInfo.Operator)
		if err = s.fkDao.WechatCardMessageNotify(
			fmt.Sprintf("【Laser指令下发通知】%v", titleStatus),
			content,
			fmt.Sprintf("%v/#/laser/laser-command-list?app_key=%v&task_id=%v", s.c.Host.Fawkes, laserCmdInfo.AppKey, taskID),
			"",
			laserCmdInfo.Operator,
			s.c.Comet.FawkesAppID); err != nil {
			log.Error("WechatCardMessageNotify error(%v)", err)
		}
	}
	return
}
