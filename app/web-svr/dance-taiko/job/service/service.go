package service

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/web-svr/dance-taiko/job/conf"
	"go-gateway/app/web-svr/dance-taiko/job/dao"
)

// Service struct .
type Service struct {
	c             *conf.Config
	dao           *dao.Dao
	waiter        *sync.WaitGroup
	danceBinlog   *databus.Databus
	fanout        *fanout.Fanout
	cron          *cron.Cron
	serviceClosed bool // 服务关闭，不启动新的结算逻辑，由新的job开始结算
}

// New init .
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:           c,
		dao:         dao.New(c),
		waiter:      new(sync.WaitGroup),
		danceBinlog: databus.New(c.DanceBinlogSub),
		fanout:      fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
		cron:        cron.New(),
	}

	// 定期进行分数结算
	if err := s.cron.AddFunc(s.c.Cfg.StatCron, s.gameStatDual); err != nil {
		panic(errors.Wrapf(err, "gameStat Cron "))
	}
	s.cron.Start()

	// 监听dance binlog
	s.waiter.Add(1)
	go s.syncGame()

	return s
}

// Ping Service .
func (s *Service) Ping(c context.Context) (err error) {
	return nil
}

// Close Service .
func (s *Service) Close() {
	s.serviceClosed = true
	s.waiter.Wait()
	s.dao.Close()
}
