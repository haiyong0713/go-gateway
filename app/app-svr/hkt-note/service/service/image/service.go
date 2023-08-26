package image

import (
	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/hkt-note/service/conf"
	imgDao "go-gateway/app/app-svr/hkt-note/service/dao/image"
)

type Service struct {
	c   *conf.Config
	dao *imgDao.Dao
}

func New() (s *Service) {
	conf := &conf.Config{}
	if err := paladin.Get("hkt-note-service.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	s = &Service{
		c:   conf,
		dao: imgDao.New(conf),
	}
	return
}
