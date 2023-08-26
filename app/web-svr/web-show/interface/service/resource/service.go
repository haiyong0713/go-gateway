package resource

import (
	"context"

	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	recrpc "go-gateway/app/app-svr/resource/service/rpc/client"
	"go-gateway/app/web-svr/web-show/interface/conf"
	"go-gateway/app/web-svr/web-show/interface/dao/ad"
	"go-gateway/app/web-svr/web-show/interface/dao/bangumi"
	"go-gateway/app/web-svr/web-show/interface/dao/data"
	"go-gateway/app/web-svr/web-show/interface/dao/live"
	resdao "go-gateway/app/web-svr/web-show/interface/dao/resource"
	seasondao "go-gateway/app/web-svr/web-show/interface/dao/season"
	rsmdl "go-gateway/app/web-svr/web-show/interface/model/resource"

	accclient "git.bilibili.co/bapis/bapis-go/account/service"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	vugrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/robfig/cron"
)

// Service define web-show service
type Service struct {
	c          *conf.Config
	resdao     *resdao.Dao
	bangumiDao *bangumi.Dao
	recrpc     *recrpc.Service
	adDao      *ad.Dao
	dataDao    *data.Dao
	seasonDao  *seasondao.Dao
	liveDao    *live.Dao
	// cache
	asgCache       map[int][]*rsmdl.Assignment // resID => assignments
	bossAsgCache   map[int][]*rsmdl.Assignment
	urlMonitor     map[int]map[string]string // pf=>map[rs.name=>url]
	videoCache     map[int64][][]*rsmdl.VideoAD
	posCache       map[string]*rsmdl.Position // resID=>srcIDs
	defBannerCache *rsmdl.Assignment
	locGRPC        locgrpc.LocationClient
	arcGRPC        arcgrpc.ArchiveClient
	vuGRPC         vugrpc.VideoUpOpenClient
	// grpc
	accClient     accclient.AccountClient
	channelClient channelgrpc.ChannelRPCClient
	// crontab
	cron *cron.Cron
	// running
	resRunning       bool
	videoAdRunning   bool
	checkDiffRunning bool
}

// New return service object
func New(c *conf.Config) *Service {
	s := &Service{
		c:            c,
		adDao:        ad.New(c),
		resdao:       resdao.New(c),
		bangumiDao:   bangumi.New(c),
		dataDao:      data.New(c),
		seasonDao:    seasondao.New(c),
		liveDao:      live.New(c),
		asgCache:     make(map[int][]*rsmdl.Assignment),
		bossAsgCache: make(map[int][]*rsmdl.Assignment),
		videoCache:   make(map[int64][][]*rsmdl.VideoAD),
		posCache:     make(map[string]*rsmdl.Position),
		cron:         cron.New(),
	}
	s.recrpc = recrpc.New(c.RPCClient2.Resource)
	var err error
	if s.locGRPC, err = locgrpc.NewClient(c.LocationGRPC); err != nil {
		panic(err)
	}
	if s.arcGRPC, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(err)
	}
	if s.vuGRPC, err = vugrpc.NewClient(c.VideoUpGRPC); err != nil {
		panic(err)
	}
	if s.accClient, err = accclient.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.channelClient, err = channelgrpc.NewClient(c.ChannelGRPC); err != nil {
		panic(err)
	}
	if err := s.init(); err != nil {
		log.Error("%+v", err)
	}
	s.cron.Start()
	return s
}

func (s *Service) init() (err error) {
	s.loadRes()
	s.loadVideoAd()
	s.checkDiff()
	if err = s.cron.AddFunc(s.c.Cron.LoadRes, s.loadRes); err != nil {
		log.Error("adService.Load, err (%v)", err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadVideoAd, s.loadVideoAd); err != nil {
		log.Error("adService.LoadVideo, err (%v)", err)
	}
	if err = s.cron.AddFunc(s.c.Cron.CheckDiff, s.checkDiff); err != nil {
		log.Error("adService.LoadVideo, err (%v)", err)
	}
	return
}

// Close close service
func (s *Service) Close() {
	s.resdao.Close()
}

// Ping ping service
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.resdao.Ping(c); err != nil {
		log.Error("s.resDap.Ping err(%v)", err)
		return
	}
	if err = s.adDao.Ping(c); err != nil {
		log.Error("s.adDao.Ping err(%v)", err)
	}
	return
}
