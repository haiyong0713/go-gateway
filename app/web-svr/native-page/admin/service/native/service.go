package native

import (
	"context"

	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/native-page/admin/conf"
	"go-gateway/app/web-svr/native-page/admin/dao"
	natDao "go-gateway/app/web-svr/native-page/admin/dao/native"
	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"

	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
)

// Service struct
type Service struct {
	c     *conf.Config
	dao   *natDao.Dao
	addao *dao.Dao
	// chan
	cloudInfoc    infoc.Infoc
	cache         *fanout.Fanout
	actplatClient actplatapi.ActPlatClient
}

// Close service
func (s *Service) Close() {
	if s.dao != nil {
		s.dao.Close()
	}
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   natDao.New(c),
		addao: dao.New(c),
		cache: fanout.New("cache"),
	}
	var err error
	if s.cloudInfoc, err = infoc.New(c.InfocConf.CloudInfoc); err != nil {
		panic(err)
	}
	if s.actplatClient, err = actplatapi.NewClient(c.PlatGRPC); err != nil {
		panic(err)
	}
	return
}

// infoc
func (s *Service) infocSave(i interface{}) {
	switch v := i.(type) {
	case natmdl.NtCloudInfo:
		payload := infoc.NewLogStream(s.c.InfocConf.CloudLogID, v.Status, v.TopicID, v.ActivityName, v.ApplyDate, v.BeginDate, v.Ctime, v.Mid)
		_ = s.cloudInfoc.Info(context.Background(), payload)
	default:
		log.Warn("infocproc can't process the type")
	}
}
