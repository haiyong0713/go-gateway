package exporttask

import (
	"context"
	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/jinzhu/gorm"
	"go-common/library/database/bfs"
	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/admin/component"
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao"
)

var (
	arcClient arcapi.ArchiveClient
	accClient accapi.AccountClient
)

// Service biz service def.
type Service struct {
	c      *conf.Config
	export *sql.DB
	DB     *gorm.DB
	bfs    *bfs.BFS
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:      c,
		export: component.ExportDB,
		DB:     dao.New(c).DB,
		bfs:    bfs.New(c.BFS),
	}
	var err error
	if arcClient, err = arcapi.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if accClient, err = accapi.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	go s.TaskDoProc()
	httpClient = bm.NewClient(c.HTTPClient)
	go GetMemberInfoProc()
	return s
}

// Ping check dao health.
func (s *Service) Ping(c context.Context) (err error) {
	return s.export.Ping(c)
}

// Wait wait all closed.
func (s *Service) Wait() {}

// Close close all dao.
func (s *Service) Close() {
	s.DB.Close()
}
