package note

import (
	"go-common/library/conf/paladin.v2"
	"go-common/library/rate/limit/quota"
	"go-gateway/app/app-svr/hkt-note/service/conf"
	"go-gateway/app/app-svr/hkt-note/service/dao/article"
	"go-gateway/app/app-svr/hkt-note/service/dao/note"

	"github.com/robfig/cron"
)

type Service struct {
	c                 *conf.Config
	dao               *note.Dao
	artDao            *article.Dao
	politicsUpMap     map[int64]struct{}
	cron              *cron.Cron
	ArcsForbidAllower quota.Allower
}

func New() (s *Service) {
	conf := &conf.Config{}
	if err := paladin.Get("hkt-note-service.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	s = &Service{
		c:                 conf,
		dao:               note.New(conf),
		artDao:            article.New(conf),
		politicsUpMap:     make(map[int64]struct{}),
		cron:              cron.New(),
		ArcsForbidAllower: quota.NewAllower(&quota.AllowerConfig{ID: conf.ArcsForbidQuotaID}),
	}
	var err error
	s.loadFeaPolitics()
	if err = s.cron.AddFunc(conf.NoteCfg.ForbidCfg.FeaCron, s.loadFeaPolitics); err != nil {
		panic(err)
	}
	s.cron.Start()
	return
}

// Close resource.
func (s *Service) Close() {
	s.dao.Close()
}
