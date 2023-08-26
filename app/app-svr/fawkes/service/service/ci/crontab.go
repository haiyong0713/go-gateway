package ci

import (
	"context"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"golang.org/x/sync/errgroup"
)

func (s *Service) crontab() {
	// 获取所有有效任务(待执行和执行中)
	crontabsAll, err := s.fkDao.CiCrontabAll(context.Background(), time.Now().Add(time.Hour*8))
	if err != nil {
		log.Error("%v", err)
		return
	}
	// 获取所有有效任务(待执行和执行中)
	var crontabs []*cimdl.Contab
	for _, crontab := range crontabsAll {
		if crontab.State == 1 || crontab.State == 0 {
			crontabs = append(crontabs, crontab)
		}
	}
	//关掉已停用和数据库不存在的任务
	s.crontabCIProc.Range(func(k, v interface{}) bool {
		var closeTask = 1
		for _, crontab := range crontabsAll {
			if k == crontab.ID && (crontab.State == 0 || crontab.State == 1) {
				closeTask = 0
				break
			}
		}
		if closeTask == 1 {
			log.Info("cron_ci close task cronID(%v)", k)
			s.crontabCIProc.Store(k, false)
		}
		return true
	})
	if len(crontabs) > 0 {
		for _, crontab := range crontabs {
			// 任务队列中存在则跳过
			if cstate, ok := s.crontabCIProc.Load(crontab.ID); ok && cstate.(bool) {
				continue
			}
			tick, err := time.ParseDuration(crontab.Tick)
			if err != nil {
				log.Error("%v", err)
				continue
			}
			// 待执行 状态的任务置为 执行中
			if crontab.State == cimdl.CronWait {
				// buildID 回写cron库
				var tx *sql.Tx
				if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
					log.Error("s.fkDao.BeginTran() error(%v)", err)
					return
				}
				defer func() {
					if r := recover(); r != nil {
						//nolint:errcheck
						tx.Rollback()
						log.Error("%v", r)
					}
				}()
				if _, err = s.fkDao.TxUpStatusCiCrontab(tx, crontab.ID, cimdl.CronRun); err != nil {
					log.Error("%v", err)
					//nolint:errcheck
					tx.Rollback()
					return
				}
				if err = tx.Commit(); err != nil {
					log.Error("tx.Commit() error(%v)", err)
					return
				}
			}
			log.Info("cron_ci start cron(%+v)", crontab)
			s.crontabCIProc.Store(crontab.ID, true)
			// nolint:biligowordcheck
			// go s.crontabCI(crontab, tick)
			go s.crontabCICommon(crontab, tick)
		}
	}
}

// func (s *Service) crontabCI(cc *cimdl.Contab, tick time.Duration) {
// 	for {
// 		if cstate, ok := s.crontabCIProc[cc.ID]; !ok {
// 			log.Error("cron_ci out off job cronID(%v) not exist", cc.ID)
// 			return
// 		} else if !cstate {
// 			delete(s.crontabCIProc, cc.ID)
// 			log.Error("cron_ci out off job cronID(%v) state false", cc.ID)
// 			return
// 		}
// 		// 当前时间与开始时间为周期整数倍才执行
// 		if (time.Now().Unix()-cc.STime)%int64(tick.Seconds()) == 0 {
// 		NEXT:
// 			var (
// 				buildID int64
// 				err     error
// 			)
// 			// 创建构建
// 			if buildID, err = s.CreateBuildPack(context.Background(), cc.AppKey, cc.PkgType, cc.GitType, cc.GitName, cc.Operator, cc.CIEnvVars, "", true); err != nil {
// 				log.Error("cronID(%v) cc.ID %v", cc.ID, err)
// 				time.Sleep(time.Second * 600)
// 				goto NEXT
// 			}
// 			log.Info("cron_ci cron(%d) get buildID(%d)", cc.ID, buildID)
// 			var tx *sql.Tx
// 			if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
// 				log.Error("s.fkDao.BeginTran() cronID(%v) error(%v)", cc.ID, err)
// 				time.Sleep(time.Second * 600)
// 				goto NEXT
// 			}
// 			defer func() {
// 				if r := recover(); r != nil {
// 					//nolint:errcheck
// 					tx.Rollback()
// 					log.Error("%v", r)
// 				}
// 			}()
// 			// 构建ID回写cron表
// 			if _, err = s.fkDao.TxUpBuildIDCiCrontab(tx, cc.ID, buildID); err != nil {
// 				log.Error("cronID(%v) %v", cc.ID, err)
// 				//nolint:errcheck
// 				tx.Rollback()
// 				time.Sleep(time.Second * 600)
// 				goto NEXT
// 			}
// 			if err = tx.Commit(); err != nil {
// 				log.Error("tx.Commit() error(%v)", err)
// 				return
// 			}
// 			// 执行构建
// 			var variables = map[string]string{
// 				"APP_KEY":     cc.AppKey,
// 				"PKG_TYPE":    strconv.Itoa(cc.PkgType),
// 				"BUILD_ID":    strconv.FormatInt(buildID, 10),
// 				"FAWKES":      "1",
// 				"FAWKES_USER": cc.Operator,
// 				"TASK":        "pack",
// 			}
// 			var envVarMap = make(map[string]string)
// 			if cc.CIEnvVars != "" {
// 				if err = json.Unmarshal([]byte(cc.CIEnvVars), &envVarMap); err != nil {
// 					log.Error("cronID(%v) json formatter error(%v)", cc.ID, err)
// 				}
// 				for key, value := range envVarMap {
// 					variables[key] = value
// 				}
// 			}
// 			if _, err = s.gitSvr.TriggerPipeline(context.Background(), cc.AppKey, cc.GitType, cc.GitName, variables); err != nil {
// 				log.Error("cronID(%v) %v", cc.ID, err)
// 			}
// 		}
// 		// 必须大于500ms&&小于等于1s
// 		time.Sleep(time.Second)
// 	}
// }

func (s *Service) crontabCICommon(cc *cimdl.Contab, tick time.Duration) {
	for {
		if cstate, ok := s.crontabCIProc.Load(cc.ID); !ok {
			log.Error("cron_ci out off job cronID(%v) not exist", cc.ID)
			return
		} else if !cstate.(bool) {
			s.crontabCIProc.Delete(cc.ID)
			log.Error("cron_ci out off job cronID(%v) state false", cc.ID)
			return
		}
		// 当前时间与开始时间为周期整数倍才执行
		if (time.Now().Unix()-cc.STime)%int64(tick.Seconds()) == 0 {
		NEXT:
			var (
				buildID, resignBuildID int64
				err                    error
			)
			if buildID, resignBuildID, err = s.CreateBuildPackCommon(context.Background(), cc.AppKey, cc.Send, cc.PkgType, cc.GitType, cc.GitName, cc.Operator, cc.CIEnvVars, "【定时出包】", "", true, 0, []int64{}); err != nil {
				log.Error("CreateBuildPackCommon error %v", err)
				time.Sleep(time.Second * 600)
				// goto NEXT
				return
			}
			var tx *sql.Tx
			if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
				log.Error("s.fkDao.BeginTran() cronID(%v) error(%v)", cc.ID, err)
				time.Sleep(time.Second * 600)
				goto NEXT
			}
			defer func() {
				if r := recover(); r != nil {
					//nolint:errcheck
					tx.Rollback()
					log.Error("%v", r)
				}
			}()
			// 构建ID回写cron表
			if resignBuildID != 0 {
				buildID = resignBuildID
			}
			if _, err = s.fkDao.TxUpBuildIDCiCrontab(tx, cc.ID, buildID); err != nil {
				log.Error("cronID(%v) %v", cc.ID, err)
				//nolint:errcheck
				tx.Rollback()
				time.Sleep(time.Second * 600)
				goto NEXT
			}
			if err = tx.Commit(); err != nil {
				log.Error("tx.Commit() error(%v)", err)
				return
			}
		}
		// 必须大于500ms&&小于等于1s
		time.Sleep(time.Second)
	}
}

func (s *Service) Broadcast(c context.Context, btype string, param *model.Broadcast) (err error) {
	g, ctx := errgroup.WithContext(c)
	var weChatContent string
	buildID := param.Param.(*cimdl.HookParam).BuildID
	if param.Username != "" {
		g.Go(func() (err error) {
			if weChatContent, err = s.formContent(c, btype, buildID); err != nil {
				log.Errorc(c, "formContent error(%v), build_id = %v", err, buildID)
				return
			}
			if err = s.fkDao.WechatEPNotify(weChatContent, param.Username); err != nil {
				log.Errorc(c, "WechatEPNotify error(%v), build_id = %v", err, buildID)
				return
			}
			log.Warnc(c, "Broadcast WechatEPNotify success build_id = %v", buildID)
			return nil
		})
	}
	if param.Hook != nil {
		g.Go(func() (err error) {
			if err = s.fkDao.Hook(ctx, param.Param, param.Hook); err != nil {
				log.Errorc(c, "Hook error(%v), build_id = %v", err, buildID)
				return
			}
			log.Warnc(c, "Broadcast Hook success build_id = %v", buildID)
			return nil
		})
	}
	if param.Bots != "" {
		g.Go(func() (err error) {
			if weChatContent, err = s.formContent(c, btype, buildID); err != nil {
				log.Errorc(c, "formContent error(%v), build_id = %v", err, buildID)
				return
			}
			if err = s.NotifyBot(c, weChatContent, param.Bots); err != nil {
				log.Errorc(c, "NotifyBot error(%v), build_id = %v", err, buildID)
				return
			}
			log.Warnc(c, "Broadcast NotifyBot success build_id = %v", buildID)
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		log.Errorc(c, "g.Wait() error(%v), build_id = %v", err, buildID)
	}
	return
}

// NotifyBot
func (s *Service) NotifyBot(c context.Context, weChatContent, botIds string) (err error) {
	var webhooks []string
	msgContent := &appmdl.Text{
		Content: weChatContent,
	}
	for _, botId := range strings.Split(botIds, ",") {
		var (
			robotId int64
			robot   *appmdl.Robot
		)
		if robotId, err = strconv.ParseInt(botId, 10, 64); err != nil {
			log.Errorc(c, "strconv.ParseInt error %v", err)
			return
		}
		if robot, err = s.fkDao.AppRobotInfoById(c, robotId); err != nil {
			log.Errorc(c, "AppRobotInfoById error(%v)", err)
			return
		}
		if robot == nil {
			log.Warnc(c, "robot is nil")
			continue
		}
		webhooks = append(webhooks, robot.WebHook)
	}
	for _, webhook := range webhooks {
		if err = s.fkDao.RobotNotify(webhook, msgContent); err != nil {
			log.Errorc(c, "RobotNotify error(%v), webhookURL %v notify %v", err, webhook, msgContent)
		}
	}
	return
}

func (s *Service) formContent(c context.Context, btype string, buildID int64) (content string, err error) {
	if btype == "ci" {
		var (
			app *appmdl.APP
			ci  *cimdl.BuildPack
		)
		if ci, app, err = s.notifyBuildInfo(c, buildID); err != nil {
			log.Errorc(c, "notifyBuildInfo error(%v)", err)
			return
		}
		if content, err = s.combineWeChatContent(c, app, ci); err != nil {
			log.Errorc(c, "combineWeChatContent error(%v), build_id = %v", err, buildID)
		}
	}
	return
}
