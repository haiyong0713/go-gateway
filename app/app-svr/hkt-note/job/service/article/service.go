package article

import (
	"go-common/library/railgun"
	"sync"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/hkt-note/job/conf"
	"go-gateway/app/app-svr/hkt-note/job/dao/article"
	"go-gateway/app/app-svr/hkt-note/job/dao/note"
)

type Service struct {
	c                *conf.Config
	dao              *note.Dao
	artDao           *article.Dao
	waiter           *sync.WaitGroup
	closed           bool
	articleBinlogSub *databus.Databus
	replyDelRailGun  *railgun.Railgun
}

// nolint:biligowordcheck
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                c,
		dao:              note.New(c),
		artDao:           article.New(c),
		articleBinlogSub: databus.New(c.ArticleBinlogSub),
		waiter:           new(sync.WaitGroup),
		closed:           false,
	}
	// 监听删除评论databus
	s.initReplyDelRailGun(&railgun.DatabusV1Config{Config: s.c.ReplyDelSub})
	// 笔记binlog
	s.waiter.Add(1)
	go recoverFunc(s.consumeArticleBinlog)
	s.waiter.Add(1)
	go recoverFunc(s.retryArticleBinlog)
	return
}

// Close close the services
func (s *Service) Close() {
	s.closed = true
	log.Info("Close binlogSub!")
	s.articleBinlogSub.Close()
	log.Info("Close replyDelRailGun!")
	s.replyDelRailGun.Close()
	log.Info("Wait sync!")
	s.waiter.Wait()
	log.Info("Service Closed!")
}

func recoverFunc(f func()) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("panicError article Panic %+v", e)
		}
	}()
	f()
}
