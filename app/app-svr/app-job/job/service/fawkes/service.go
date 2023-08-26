package fawkes

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/conf/env"
	"go-common/library/log"

	"go-farm"

	"go-gateway/app/app-svr/app-job/job/conf"
	fkdao "go-gateway/app/app-svr/app-job/job/dao/fawkes"
	pushdao "go-gateway/app/app-svr/app-job/job/dao/push"
	"go-gateway/app/app-svr/app-job/job/model"
	jfkmdl "go-gateway/app/app-svr/app-job/job/model/fawkes"

	"github.com/robfig/cron"
)

// Service module service.
type Service struct {
	c                       *conf.Config
	fkDao                   *fkdao.Dao
	cron                    *cron.Cron
	running, runningSilence bool
	pushDao                 *pushdao.Dao
}

// New new a module service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		fkDao:   fkdao.New(c),
		cron:    cron.New(),
		pushDao: pushdao.New(c),
	}
	// 间隔一分钟
	if err := s.cron.AddFunc("@every 1m", s.laserPush); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc("@every 1m", s.laserPushSilence); err != nil {
		panic(err)
	}
	s.cron.Start()
	return
}

// Close is.
func (s *Service) Close() {
	s.cron.Stop()
}

// laserPush push laser.
func (s *Service) laserPush() {
	// 为什么把 s.c.FawkesLaser 的开关放到这里，
	// 是因为s.cron.Stop会给一个无缓冲的chan发送close信号，这个无缓冲的close是s.cron.Start()方法开启的goroutine消费这个chan。
	// 如果 cron.Start() 没执行，执行cron.Stop就会阻塞。
	// 之前为了保证 start和close行为一致，所以把s.c.FawkesLaser放到了Service对象里，防止热更新。
	if !model.EnvRun() || env.DeployEnv != "prod" || s.running {
		return
	}
	s.running = true
	defer func() {
		s.running = false
	}()
	ctx := context.Background()
	laser, err := s.fkDao.LaserAll(ctx)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	log.Info("laser %+v", laser)
	for _, tl := range laser {
		if tl.Buvid == "" && tl.MID == 0 {
			continue
		}
		lmsg := &jfkmdl.LaserMsg{
			Date:   tl.LogDate,
			TaskID: strconv.FormatInt(tl.ID, 10),
		}
		mb, err := json.Marshal(lmsg)
		if err != nil {
			log.Error("%+v", err)
			continue
		}
		var filters []string
		if tl.Platform != "" {
			filters = append(filters, fmt.Sprintf("platform==%v", tl.Platform))
		}
		if tl.MobiApp != "" {
			filters = append(filters, fmt.Sprintf("mobi_app==%v", tl.MobiApp))
		}
		if tl.Buvid != "" {
			filters = append(filters, fmt.Sprintf("buvid==%v", tl.Buvid))
		} else if tl.MID != 0 {
			filters = append(filters, fmt.Sprintf("mid==%v", tl.MID))
		}
		msg := string(mb)
		filter := strings.Join(filters, " AND ")
		if err = s.fkDao.PushAll(ctx, msg, filter); err != nil {
			log.Error("laserPush push broadcaset error %v ", err)
			if err = s.fkDao.LaserReportBroadCast(ctx, tl.ID, jfkmdl.StatusSendFaild, ""); err != nil {
				log.Error("laserPush msg(%v) filter(%v) push braodcast faild report err %v ", msg, filter, err)
			}
		} else if err = s.fkDao.LaserReportBroadCast(ctx, tl.ID, jfkmdl.StatusWaitSend, ""); err != nil {
			log.Error("laserPush msg(%v) filter(%v) push braodcast success report err %v ", msg, filter, err)
		}
		time.Sleep(time.Millisecond * 10)
	}
}

// laserPush push laser.
func (s *Service) laserPushSilence() {
	if !model.EnvRun() || env.DeployEnv != "prod" || s.runningSilence {
		return
	}
	s.runningSilence = true
	defer func() {
		s.runningSilence = false
	}()
	ctx := context.Background()
	laser, err := s.fkDao.LaserAllSilence(ctx)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	for _, tl := range laser {
		if tl == nil {
			continue
		}
		if tl.Buvid == "" && tl.MID == 0 {
			continue
		}
		pushParams := &model.PushParam{
			AppID:       s.c.Push.FawkesLaser.AppID,
			BusinessID:  s.c.Push.FawkesLaser.BusinessID,
			LinkType:    s.c.Push.FawkesLaser.LinkType,
			LinkValue:   fmt.Sprintf("%v,%v", tl.ID, tl.LogDate),
			PassThrough: 1, // 是否透传。 0:不透传, 1: 透传, 默认 0
		}
		if tl.MID != 0 {
			pushParams.MIDs = []int64{tl.MID}
		}
		if tl.Buvid != "" {
			pushParams.Buvids = []string{tl.Buvid}
		}
		pushParams.UUID = s.hash(pushParams)
		log.Info("laserPushSilence push silence param %+v token %v", pushParams, s.c.Push.FawkesLaser.Token)
		if err = s.pushDao.Push(ctx, pushParams, s.c.Push.FawkesLaser.Token); err != nil {
			log.Error("laserPushSilence push silence error %v ", err)
			if err = s.fkDao.LaserReportSilence(ctx, tl.ID, jfkmdl.StatusSendFaild, ""); err != nil {
				log.Error("laserPushSilence push silence faild report error %v ", err)
			}
		} else if err = s.fkDao.LaserReportSilence(ctx, tl.ID, jfkmdl.StatusWaitSend, ""); err != nil {
			log.Error("laserPushSilence push silence success report error %v ", err)
		}
		time.Sleep(time.Millisecond * 10)
	}
}

// hash get banner hash.
func (s *Service) hash(v *model.PushParam) (value string) {
	bs, err := json.Marshal(v)
	if err != nil {
		log.Error("json.Marshal error(%v)", err)
		return
	}
	value = strconv.FormatUint(farm.Hash64(bs), 10)
	return
}
