package article

import (
	"go-common/library/conf/paladin.v2"
	"go-common/library/rate/limit/quota"
	"go-gateway/app/app-svr/hkt-note/service/conf"
	"go-gateway/app/app-svr/hkt-note/service/dao/article"
	"go-gateway/app/app-svr/hkt-note/service/dao/note"
)

type Service struct {
	c                      *conf.Config
	artDao                 *article.Dao
	noteDao                *note.Dao
	getAttachedRpidAllower quota.Allower
	arcTagAllower          quota.Allower
}

func New() (s *Service) {
	conf := &conf.Config{}
	if err := paladin.Get("hkt-note-service.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	s = &Service{
		c:                      conf,
		artDao:                 article.New(conf),
		noteDao:                note.New(conf),
		getAttachedRpidAllower: quota.NewAllower(&quota.AllowerConfig{ID: conf.GetAttachedRpidQuotaID}),
		arcTagAllower:          quota.NewAllower(&quota.AllowerConfig{ID: conf.ArcTagQuotaId}),
	}
	return
}

// Close resource.
func (s *Service) Close() {
	s.artDao.Close()
}
