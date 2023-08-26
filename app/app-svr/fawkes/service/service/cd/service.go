package cd

import (
	"context"
	"math/rand"
	"os"

	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	ossdao "go-gateway/app/app-svr/fawkes/service/dao/oss"
	gitSvr "go-gateway/app/app-svr/fawkes/service/service/gitlab"
	mdlSvr "go-gateway/app/app-svr/fawkes/service/service/modules"
	"go-gateway/app/app-svr/fawkes/service/tools/appstoreconnect"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/asaskevich/EventBus"
	"github.com/robfig/cron"
	"github.com/xanzy/go-gitlab"
)

// Service struct.
type Service struct {
	c                          *conf.Config
	fkDao                      *fkdao.Dao
	ossDao                     *ossdao.Dao
	appstoreClient             *appstoreconnect.Client
	gitlabClient               *gitlab.Client
	gitSvr                     *gitSvr.Service
	mdlSvr                     *mdlSvr.Service
	hanlderChan                []chan func()
	cron                       *cron.Cron
	event                      EventBus.Bus
	channelPackUploadCdnWorker *fanout.Fanout
}

// New new service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                          c,
		fkDao:                      fkdao.New(c),
		ossDao:                     ossdao.New(c),
		appstoreClient:             appstoreconnect.NewClient(c),
		gitSvr:                     gitSvr.New(c),
		mdlSvr:                     mdlSvr.New(c),
		gitlabClient:               gitlab.NewClient(nil, c.Gitlab.Token),
		cron:                       cron.New(),
		event:                      EventBus.New(),
		channelPackUploadCdnWorker: fanout.New("channel_pack_upload_cdn_worker", fanout.Worker(1), fanout.Buffer(1024)),
	}
	_ = s.gitlabClient.SetBaseURL(c.Gitlab.Host)
	if err := s.registerToAppstoreConnectClient(); err != nil {
		panic(err)
	}
	for i := 0; i < 5; i++ {
		s.hanlderChan = append(s.hanlderChan, make(chan func(), 512))
		// nolint:biligowordcheck
		go s.handlerproc(i)
	}
	if os.Getenv("CRON_SWITCH") == "on" {
		// 每 10 分钟刷一次上传状态
		if err := s.cron.AddFunc("0 1-51/10 * * * *", func() { _ = s.UploadStateProc() }); err != nil {
			panic(err)
		}
		// 每 10 分钟检查一次 appstore 包状态
		if err := s.cron.AddFunc("0 2-52/10 * * * *", func() { _ = s.PackStateProc() }); err != nil {
			panic(err)
		}
		// 每 10 分钟检查一次 review 状态
		if err := s.cron.AddFunc("0 3-53/10 * * * *", func() { _ = s.ReviewStateProc() }); err != nil {
			panic(err)
		}
		// 每 10 分钟刷一次使用人数
		if err := s.cron.AddFunc("0 4-54/10 * * * *", func() { _ = s.DistributeNumProc() }); err != nil {
			panic(err)
		}
		// 每分钟检查一次 betagroup 人数，超过阈值直接删除部分 tester
		if err := s.cron.AddFunc("0 * * * * *", func() { _ = s.DeleteTestersProc() }); err != nil {
			panic(err)
		}
		// 每 10 分钟检查一次过期的包，包过期后置为 Disable
		if err := s.cron.AddFunc("0 6-56/10 * * * *", func() { _ = s.DisableExpiredProc() }); err != nil {
			panic(err)
		}
		// 每 10 分钟检查一次是否有线上版更新
		if err := s.cron.AddFunc("0 7-57/10 * * * *", func() { _ = s.UpdateOnlineVersProc() }); err != nil {
			panic(err)
		}
	}
	// 定期更新所有机器的 appstore 信息
	if err := s.cron.AddFunc("0 8-58/10 * * * *", func() { _ = s.UpdateAppStoreConnectProc() }); err != nil {
		panic(err)
	}
	s.cron.Start()
	if err := s.event.SubscribeAsync(PackGenerateUpdateEvent, s.packGenerateUpdateAction, false); err != nil {
		panic(err)
	}
	if err := s.event.SubscribeAsync(SteadyPackAutoGenChannelPack, s.autoGenChannelPack, false); err != nil {
		panic(err)
	}
	if err := s.event.SubscribeAsync(PackGreyPushEvent, s.AddPackGreyHistory, false); err != nil {
		panic(err)
	}
	return
}

// AddPatchProc add patch build proc.
func (s *Service) AddHandlerProc(f func()) {
	i := rand.Intn(5)
	select {
	case s.hanlderChan[i] <- f:
	default:
		log.Warn("AddHandlerProc chan full")
	}
}

func (s *Service) handlerproc(i int) {
	for {
		f, ok := <-s.hanlderChan[i]
		if !ok {
			log.Warn("handlerproc exit")
			return
		}
		f()
	}
}

// Ping dao.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.fkDao.Ping(c); err != nil {
		log.Error("s.dao error(%v)", err)
	}
	return
}

// Close dao.
func (s *Service) Close() {
	s.fkDao.Close()
	_ = s.channelPackUploadCdnWorker.Close()
}
