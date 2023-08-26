package note

import (
	"go-common/library/conf/env"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/hkt-note/job/conf"
	"go-gateway/app/app-svr/hkt-note/job/dao/article"
	"go-gateway/app/app-svr/hkt-note/job/dao/note"
	"sync"
)

type Service struct {
	c                *conf.Config
	dao              *note.Dao
	artDao           *article.Dao
	waiter           *sync.WaitGroup
	closed           bool
	noteBinlogSub    *databus.Databus
	infocV2Log       infocV2.Infoc
	noteAddRailGun   *railgun.Railgun
	noteAuditRailGun *railgun.Railgun

	hotArcBotPush *railgun.Railgun
}

//nolint:biligowordcheck
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:             c,
		dao:           note.New(c),
		artDao:        article.New(c),
		noteBinlogSub: databus.New(c.NoteBinlogSub),
		waiter:        new(sync.WaitGroup),
		closed:        false,
	}
	var err error
	s.infocV2Log, err = infocV2.New(c.InfocV2)
	if err != nil {
		if env.DeployEnv == env.DeployEnvProd {
			panic(err)
		}
	}
	// 监听更新笔记详情databus
	s.initNoteAddRailGun(&railgun.DatabusV1Config{Config: s.c.NoteNotifySub})
	// 监听提交敏感词审核databus
	s.initNoteAuditRailGun(&railgun.DatabusV1Config{Config: s.c.NoteAuditSub})
	// 笔记binlog
	s.waiter.Add(1)
	go recoverFunc(s.consumeNoteBinlog)

	s.waiter.Add(1)
	go recoverFunc(s.retryDetail)

	s.waiter.Add(1)
	go recoverFunc(s.retryNoteList)

	s.waiter.Add(1)
	go recoverFunc(s.retryNoteListRem)

	s.waiter.Add(1)
	go recoverFunc(s.retryDetailDB)

	s.waiter.Add(1)
	go recoverFunc(s.retryContDBDel)

	s.waiter.Add(1)
	go recoverFunc(s.retryDetailDBDel)

	s.waiter.Add(1)
	go recoverFunc(s.retryUser)

	s.waiter.Add(1)
	go recoverFunc(s.retryAid)

	s.waiter.Add(1)
	go recoverFunc(s.retryContent)

	s.waiter.Add(1)
	go recoverFunc(s.retryDelCache)

	s.waiter.Add(1)
	go recoverFunc(s.retryAudit)

	s.waiter.Add(1)
	go recoverFunc(s.retryArtContDB)

	s.waiter.Add(1)
	go recoverFunc(s.retryArtDetailDB)

	s.waiter.Add(1)
	go recoverFunc(s.retryArtBinlog)

	var uniqCfg = &railgun.CronUniqConfig{
		RedisLockConfig: &railgun.RedisLockConfig{
			Config: s.c.Redis.Config,
			Key:    "hot_arc_bot_push_flag",
		},
	}
	if env.DeployEnv == env.DeployEnvProd {
		s.hotArcBotPush = railgun.NewRailGun("热门稿件定时push", nil,
			railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: c.NoteCfg.HotArcBotPushCron}),
			railgun.NewCronProcessor(nil, s.HotArcPushBot,
				railgun.WithCronUnique(uniqCfg)))
		s.hotArcBotPush.Start()
	}
	return
}

// Close close the services
func (s *Service) Close() {
	s.closed = true
	log.Info("Close binlogSub!")
	s.noteBinlogSub.Close()
	log.Info("Wait sync!")
	s.waiter.Wait()
	log.Info("Service Closed!")
	s.noteAuditRailGun.Close()
	s.noteAddRailGun.Close()
	s.hotArcBotPush.Close()
}

func recoverFunc(f func()) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("panicError note Panic %+v", e)
		}
	}()
	f()
}
