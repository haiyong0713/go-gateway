package playurl

import (
	infoc2 "go-common/library/log/infoc.v2"
	"go-gateway/app/app-svr/app-car/interface/conf"
	bgmdao "go-gateway/app/app-svr/app-car/interface/dao/bangumi"
	playerdao "go-gateway/app/app-svr/app-car/interface/dao/playurl"
)

type Service struct {
	c       *conf.Config
	bgm     *bgmdao.Dao
	player  *playerdao.Dao
	infocV2 infoc2.Infoc
}

func New(c *conf.Config) *Service {
	s := &Service{
		c:      c,
		bgm:    bgmdao.New(c),
		player: playerdao.New(c),
	}
	infocV2, err := infoc2.New(nil)
	if err != nil {
		s.infocV2 = infocV2
	}
	return s
}
