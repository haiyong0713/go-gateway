package topic

import (
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	accountdao "go-gateway/app/app-svr/app-dynamic/interface/dao/account"
	activitydao "go-gateway/app/app-svr/app-dynamic/interface/dao/activity"
	archivedao "go-gateway/app/app-svr/app-dynamic/interface/dao/archive"
	channeldao "go-gateway/app/app-svr/app-dynamic/interface/dao/channel"
	dyndao "go-gateway/app/app-svr/app-dynamic/interface/dao/dynamicV2"
	nativePagedao "go-gateway/app/app-svr/app-dynamic/interface/dao/nativePage"
	topicdao "go-gateway/app/app-svr/app-dynamic/interface/dao/topic"
)

type Service struct {
	c             *conf.Config
	nativePageDao *nativePagedao.Dao
	activityDao   *activitydao.Dao
	topicDao      *topicdao.Dao
	accountDao    *accountdao.Dao
	archiveDao    *archivedao.Dao
	dynDao        *dyndao.Dao
	channelDao    *channeldao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:             c,
		nativePageDao: nativePagedao.New(c),
		activityDao:   activitydao.New(c),
		topicDao:      topicdao.New(c),
		accountDao:    accountdao.New(c),
		archiveDao:    archivedao.New(c),
		dynDao:        dyndao.New(c),
		channelDao:    channeldao.New(c),
	}
	return
}
