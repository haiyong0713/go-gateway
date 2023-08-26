package service

import (
	"context"
	"sync"
	"time"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-common/library/queue/databus"

	"go-gateway/app/app-svr/steins-gate/job/internal/dao"

	"github.com/BurntSushi/toml"
	"github.com/robfig/cron"
)

// Conf is
var Conf = &Config{}

// Service service.
type Service struct {
	dao           *dao.Dao
	daoClosed     bool
	arcNotifySub  *databus.Databus
	evaluationSub *databus.Databus
	waiter        *sync.WaitGroup
	conf          *Config
	cron          *cron.Cron
}

// Config is
type Config struct {
	Message     *Message
	Elimination *Elimination
}

// Set set config and decode.
func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	*c = tmp
	return nil
}

// Message is
type Message struct {
	Title   string
	Content string
	MC      string
}

type Elimination struct {
	CronExpr    string
	PeriodValid int // day
}

// New new a service and return.
func New() (s *Service) {
	var (
		databusCfg struct {
			ArcNotify        *databus.Config
			SteinsEvaluation *databus.Config
		}
	)
	checkErr(paladin.Watch("application.toml", Conf))
	checkErr(paladin.Get("databus.toml").UnmarshalTOML(&databusCfg))
	s = &Service{
		dao:           dao.New(),
		arcNotifySub:  databus.New(databusCfg.ArcNotify),
		evaluationSub: databus.New(databusCfg.SteinsEvaluation),
		waiter:        new(sync.WaitGroup),
		conf:          Conf,
		cron:          cron.New(),
	}
	if err := s.cron.AddFunc(s.conf.Elimination.CronExpr, s.removeHvarRec); err != nil {
		panic(err)
	}
	s.cron.Start()
	s.waiter.Add(1)
	//nolint:biligowordcheck
	go s.arcConsumeproc()
	s.waiter.Add(1)
	//nolint:biligowordcheck
	go s.evalConsumeproc()
	return s
}

// Close close the resource.
func (s *Service) Close() {
	s.daoClosed = true          // the dao is logically closed
	time.Sleep(2 * time.Second) // waiter 2 seconds to let the tasks stop
	s.arcNotifySub.Close()      // close channel & databus
	s.evaluationSub.Close()
	s.waiter.Wait()
	s.dao.Close() // close the dao physically
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *Service) removeHvarRec() {
	var c = context.Background()
	allowed, err := s.dao.GetHvarLock(c)
	if err != nil {
		return
	}
	if !allowed {
		log.Error("RemoveHvarRec locked")
		return
	}
	//nolint:errcheck
	defer s.dao.DelHvarLock(c) // 释放锁，避免err场景下死锁
	if err := s.dao.RemoveHvarRec(s.conf.Elimination.PeriodValid); err != nil {
		log.Error("%+v", err)
		return
	}

}
