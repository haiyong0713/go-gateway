package popular

import (
	"time"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/conf/env"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/archive"
	"go-gateway/app/app-svr/app-feed/admin/dao/article"
	"go-gateway/app/app-svr/app-feed/admin/dao/message"
	showdao "go-gateway/app/app-svr/app-feed/admin/dao/show"
)

// Service is search service
type Service struct {
	showDao    *showdao.Dao
	messageDao *message.Dao
	arrDao     *archive.Dao
	artDao     *article.Dao
	tagClient  tagrpc.TagRPCClient
}

// New new a search service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		showDao:    showdao.New(c),
		messageDao: message.New(c),
		arrDao:     archive.New(c),
		artDao:     article.New(c),
	}
	//只会在预发机器跑通知任务
	if env.DeployEnv == env.DeployEnvPre {
		log.Info("Notify start")
		//nolint:biligowordcheck
		go s.Notify()
	}
	var err error
	if s.tagClient, err = tagrpc.NewClient(c.TagGRPClient); err != nil {
		panic(err)
	}
	return
}

// Notify notify job
func (s *Service) Notify() {
	for {
		if err := s.NotifyUp(); err != nil {
			log.Error("popular service NotifyUp error(%+v)", err)
		}
		log.Info("Notify running")
		time.Sleep(300 * time.Second)
	}
}
