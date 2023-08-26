package service

import (
	"context"
	"runtime"
	"sync"
	"time"

	pb "go-gateway/app/web-svr/native-page/job/api"
	"go-gateway/app/web-svr/native-page/job/internal/dao"

	"github.com/BurntSushi/toml"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.NativePageJobServer), new(*Service)))

type Config struct {
	SponsoredUp *SponsoredUp
}

type SponsoredUp struct {
	Open bool
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	log.Info("progress-service-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}

// Service service.
type Service struct {
	cfg      *Config
	redisCfg *redis.Config
	dao      dao.Dao
	fanout   *fanout.Fanout
	// progressWorker
	progressWorker *fanout.Fanout
	// cache
	progressCache map[string]int64
	progCacheMu   sync.Mutex
	// railgun
	loadPageRelations       *railgun.Railgun
	loadProgressParamsExtra *railgun.Railgun
	upAutoAudit             *railgun.Railgun
	broadProgress           *railgun.Railgun
	broadClickProgress      *railgun.Railgun
	offlinePage             *railgun.Railgun
	onlinePage              *railgun.Railgun
	newTopicPage            *railgun.Railgun
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		cfg:            &Config{},
		dao:            d,
		progressWorker: fanout.New("progress-worker", fanout.Worker(runtime.NumCPU()), fanout.Buffer(1024)),
		progressCache:  make(map[string]int64),
		fanout:         fanout.New("service-cache"),
	}
	cf = s.Close
	if err = paladin.Get("application.toml").UnmarshalTOML(s.cfg); err != nil {
		return
	}
	if err = paladin.Watch("application.toml", s.cfg); err != nil {
		return
	}
	var (
		reCfg redis.Config
		reCt  paladin.Map
	)
	if err = paladin.Get("redis.toml").Unmarshal(&reCt); err != nil {
		return
	}
	if err = reCt.Get("Client").UnmarshalTOML(&reCfg); err != nil {
		return
	}
	s.redisCfg = &reCfg
	s.startCron()
	return
}

func (s *Service) startCron() {
	cfg := s.dao.GetCfg().CronInterval
	s.loadPageRelations = railgun.NewRailGun("刷新Native页父子关系", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: cfg.PageRelationCron}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			s.dao.LoadPageRelations(ctx)
			return railgun.MsgPolicyNormal
		}))
	s.loadPageRelations.Start()
	s.loadProgressParamsExtra = railgun.NewRailGun("刷新进度条配置参数", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: cfg.ProgressParamsExtraCron}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			s.dao.LoadProgressParamsExtra(ctx)
			return railgun.MsgPolicyNormal
		}))
	s.loadProgressParamsExtra.Start()
	s.upAutoAudit = railgun.NewRailGun("UP主发起活动自动过审", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: cfg.UpAutoAuditCron}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			s.UpAutoAudit(ctx)
			return railgun.MsgPolicyNormal
		}))
	s.upAutoAudit.Start()
	var uniqCfg = &railgun.CronUniqConfig{
		RedisLockConfig: &railgun.RedisLockConfig{
			Config:         s.redisCfg,
			KeyExpire:      xtime.Duration(time.Second * 30),
			RetryInterval:  xtime.Duration(time.Second),
			ErrMaxWaitTime: xtime.Duration(time.Second * 5),
		},
	}
	uniqCfg.Key = "natjob_lock_offline"
	s.offlinePage = railgun.NewRailGun("话题活动到期自动下线", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: cfg.UpDownCron}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			s.OfflinePage(ctx)
			return railgun.MsgPolicyNormal
		}, railgun.WithCronUnique(uniqCfg)))
	s.offlinePage.Start()
	uniqCfg.Key = "natjob_lock_online"
	s.onlinePage = railgun.NewRailGun("话题活动到期自动上线", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: cfg.UpDownCron}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			s.dao.OnlinePage(ctx)
			return railgun.MsgPolicyNormal
		}, railgun.WithCronUnique(uniqCfg)))
	s.onlinePage.Start()
	s.newTopicPage = railgun.NewRailGun("每日新增话题活动数据", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: cfg.NewPageCron}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			s.NewTopicPage(ctx)
			return railgun.MsgPolicyNormal
		}))
	s.newTopicPage.Start()
	s.broadProgress = railgun.NewRailGun("进度条组件整体维度推送", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: cfg.BroadProgressCron}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			s.broadcastProgress(ctx)
			return railgun.MsgPolicyNormal
		}))
	s.broadProgress.Start()
	s.broadClickProgress = railgun.NewRailGun("自定义组件进度模式整体维度推送", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: cfg.BroadClickProgressCron}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
			s.broadcastClickProgress(ctx)
			return railgun.MsgPolicyNormal
		}))
	s.broadClickProgress.Start()
	_ = s.fanout.Do(context.Background(), func(ctx context.Context) {
		s.cacheSponsoredUp(ctx)
	})
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.loadPageRelations.Close()
	s.loadProgressParamsExtra.Close()
	s.upAutoAudit.Close()
	s.offlinePage.Close()
	s.onlinePage.Close()
	s.newTopicPage.Close()
	s.broadProgress.Close()
	s.broadClickProgress.Close()
	s.progressWorker.Close()
	s.fanout.Close()
}
