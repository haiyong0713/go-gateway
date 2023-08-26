package college

import (
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao/college"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"github.com/jinzhu/gorm"
)

// Service ...
type Service struct {
	c       *conf.Config
	college college.Dao
	tagRPC  tagrpc.TagRPCClient
	DB      *gorm.DB
	AidDB   *gorm.DB
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		college: college.New(c),
	}
	s.DB = s.college.GetDB()
	s.AidDB = s.college.GetAidDB()
	var err error
	if s.tagRPC, err = tagrpc.NewClient(c.TagClient); err != nil {
		panic(err)
	}

	return s
}

// Close ...
func (s *Service) Close() {
	s.college.Close()
}
