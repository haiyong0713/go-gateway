package mod

import (
	"go-common/library/database/boss"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/railgun"

	"go-common/library/conf/paladin.v2"

	"go-gateway/app/app-svr/fawkes/job/internal/dao"
	"go-gateway/app/app-svr/fawkes/job/internal/model/mod"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

// Service struct.
type Service struct {
	ac          *paladin.Map
	dao         dao.Dao
	boss        *boss.Boss
	versionChan chan *mod.Version
	closeChan   chan struct{}
	folder      string
	modRailgun  *railgun.Railgun
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	var config = struct {
		Folder           string
		BossConfig       *boss.Config
		ModRailgunConfig *railgun.SingleConfig
		ModBinlogConfig  *databus.Config
	}{}
	if err = paladin.Get("application.toml").UnmarshalTOML(&config); err != nil {
		panic(err)
	}
	s = &Service{
		ac:          &paladin.TOML{},
		versionChan: make(chan *mod.Version, 64),
		closeChan:   make(chan struct{}),
		dao:         d,
		folder:      config.Folder,
		boss:        boss.New(config.BossConfig),
	}
	if err = paladin.Watch("application.toml", s.ac); err != nil {
		panic(err)
	}
	s.initModRailgun(&railgun.DatabusV1Config{Config: config.ModBinlogConfig}, config.ModRailgunConfig)
	cf = s.Close
	return
}

// Close close the resource.
func (s *Service) Close() {
	close(s.closeChan)
	s.modRailgun.Close()
	log.Warn("fawkes-job mod close exit")
}
