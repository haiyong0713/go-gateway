package selected

import (
	"context"
	"time"

	"go-common/library/conf/env"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/archive"
	"go-gateway/app/app-svr/app-feed/admin/dao/danmu"
	"go-gateway/app/app-svr/app-feed/admin/dao/elastic"
	"go-gateway/app/app-svr/app-feed/admin/dao/favorite"
	"go-gateway/app/app-svr/app-feed/admin/dao/push"
	"go-gateway/app/app-svr/app-feed/admin/dao/selected"
	model "go-gateway/app/app-svr/app-feed/admin/model/selected"

	"github.com/robfig/cron"
)

// Service is egg service
type Service struct {
	c           *conf.Config
	dao         *selected.Dao
	esDao       *elastic.Dao
	arcDao      *archive.Dao
	pushDao     *push.Dao
	favDao      *favorite.Dao
	cronLoad    *cron.Cron
	SeriesInUse []*model.SelPreview
	damuDao     *danmu.Dao
}

// New new a egg service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:        c,
		dao:      selected.New(c),
		esDao:    elastic.New(c),
		arcDao:   archive.New(c),
		pushDao:  push.New(c),
		favDao:   favorite.New(c),
		cronLoad: cron.New(),
		damuDao:  danmu.New(c),
	}

	var err error
	// 定时刷新每周必看内容
	if err = s.loadSeriesInUse(); err != nil {
		log.Error("【日志报警】每周必看loadSeriesInUse failed")
		panic(err)
	}
	err = s.cronLoad.AddFunc("@every 1m", func() {
		if err2 := s.loadSeriesInUse(); err2 != nil {
			log.Error("【日志报警】每周必看loadSeriesInUse failed")
		}
	})
	if err != nil {
		panic(err)
	}
	// 定时创建 每周下一期每周必看空白数据
	err = s.cronLoad.AddFunc(s.c.WeeklySelected.NewSerieCron, func() {
		var (
			lock bool
			e    error
		)
		// 错误重试：
		for i := 0; i < 5; i++ {
			lock, e = s.dao.GetCronJobLock(context.Background(), "NewSerieCron")
			if e == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if !lock {
			return
		}
		log.Warn("NewSerie Start")
		if err2 := s.newSerie(model.SERIE_TYPE_WEEKLY_SELECTED); err2 != nil {
			log.Error("【日志报警】每周必看 下一期空白每周必看创建失败 error(%+v)", err2)
		}
	})
	if err != nil {
		panic(err)
	}
	// 每周必看18:00发布
	err = s.cronLoad.AddFunc(s.c.WeeklySelected.PublishCron, func() {
		// 禁止预发环境定时发布
		if env.DeployEnv == env.DeployEnvPre {
			return
		}
		var (
			lock bool
			e    error
		)
		log.Warn("PublishWeekly GetCronJobLock")
		// 错误重试：
		for i := 0; i < 5; i++ {
			lock, e = s.dao.GetCronJobLock(context.Background(), "PublishWeekly")
			if e == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if !lock {
			log.Warn("PublishWeekly GetCronJobLock lock=%+v", lock)
			return
		}
		log.Warn("PublishWeekly Start")
		if e = s.PublishWeekly(context.Background()); e != nil {
			log.Error("【日志报警】每周必看 发布失败 error(%+v)", e)
		}
	})
	if err != nil {
		panic(err)
	}
	// 每周必看18:00发布时会将每周必看在热门分类入口设置的位置调到类入口设置的位置调到 RankIndex，周一18点恢复到之前的位置
	err = s.cronLoad.AddFunc(s.c.WeeklySelected.RollBackRankCron, func() {
		var (
			lock bool
			e    error
		)
		// 错误重试：
		for i := 0; i < 5; i++ {
			lock, e = s.dao.GetCronJobLock(context.Background(), "rollbackEntranceRank")
			if e == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if !lock {
			return
		}
		log.Warn("RollBackRank Start")
		if err2 := s.rollbackEntranceRank(); err2 != nil {
			log.Error("【日志报警】每周必看 热门分类入口配置位置 恢复失败 error(%+v)", err2)
		}
	})
	if err != nil {
		panic(err)
	}
	s.cronLoad.Start()
	return
}
