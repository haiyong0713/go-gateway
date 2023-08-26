package service

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"go-common/library/conf/paladin"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/dao"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	blizzardDao "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/dao"
	qqDao "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/dao"
	"time"
)

var Provider = wire.NewSet(New, qqDao.Init, blizzardDao.Init, wire.Bind(new(pb.ArchivePushServer), new(*Service)))

// Service service.
type Service struct {
	Cfg         *model.ApplicationConfig
	dao         *dao.Dao
	qqDAO       *qqDao.Dao
	blizzardDAO *blizzardDao.Dao
}

// New new a service and return.
func New(d *dao.Dao, qqDAO *qqDao.Dao, blizzardDAO *blizzardDao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		Cfg:         &model.ApplicationConfig{},
		dao:         d,
		qqDAO:       qqDAO,
		blizzardDAO: blizzardDAO,
	}
	cf = s.Close
	s.loadConfig()
	configChangeCh := paladin.WatchEvent(context.Background(), "application.toml")
	go func() {
		// 监听配置文件变更
		for {
			sig := <-configChangeCh
			if sig.Event >= 0 {
				log.Info("config reload")
				s.loadConfig()
			}
		}
	}()

	// 每分钟检查batch并将未推送的推送
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			if s.Cfg.Debug {
				log.Info("Ticker: CheckAndPushTodoArchives Start %s", time.Now().Format(model.DefaultTimeLayout))
			}
			if err := s.dao.LockBatchTodo(); err != nil {
				log.Error("Ticker: CheckAndPushTodoArchives LockBatchTodo error %v", err)
				continue
			}
			s.CheckAndPushTodoArchives()
			if err := s.dao.UnlockBatchTodo(); err != nil {
				log.Error("Ticker: CheckAndPushTodoArchives UnlockBatchTodo error %v", err)
			}
			if s.Cfg.Debug {
				log.Info("Ticker CheckAndPushTodoArchives End")
			}
		}
	}()
	return
}

func (s *Service) loadConfig() {
	if err := paladin.Get("application.toml").UnmarshalTOML(s.Cfg); err != nil {
		log.Error("[IMPORTANT] Service: loadConfig paladin.Get(application.toml).UnmarshalTOML Error %v", err)
		return
	}
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}
