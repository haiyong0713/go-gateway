package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	xredis "go-common/library/cache/redis"
	"go-common/library/conf/env"
	"go-common/library/database/taishan"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-common/library/rate/limit/quota"
	"go-common/library/stat/prom"

	"go-gateway/app/app-svr/archive/job/conf"
	"go-gateway/app/app-svr/archive/job/dao/archive"
	codao "go-gateway/app/app-svr/archive/job/dao/control"
	locDao "go-gateway/app/app-svr/archive/job/dao/location"
	"go-gateway/app/app-svr/archive/job/dao/pgc"

	"go-gateway/app/app-svr/archive/job/dao/result"
	dbusmdl "go-gateway/app/app-svr/archive/job/model/databus"
	resmdl "go-gateway/app/app-svr/archive/job/model/result"
	"go-gateway/app/app-svr/archive/job/model/retry"

	"github.com/robfig/cron"
)

// Service service
type Service struct {
	c          *conf.Config
	closeRetry bool
	closeSub   bool
	archiveDao *archive.Dao
	resultDao  *result.Dao
	pgcDao     *pgc.Dao
	controlDao *codao.Dao
	locDao     *locDao.Dao
	redis      *xredis.Pool
	waiter     sync.WaitGroup
	//databus
	archiveResultPub     *databus.Databus
	videoupSubV2         *railgun.Railgun
	seasonNotifyArcSubV2 *railgun.Railgun
	steinsGateSubV2      *railgun.Railgun
	cacheSubV2           *railgun.Railgun
	internalSub          *railgun.Railgun
	//cron
	loadTypesCron        *railgun.Railgun
	syncCreativeTypeCron *railgun.Railgun
	checkConsumeCron     *railgun.Railgun

	Prom       *prom.Prom
	arcRedises []*xredis.Pool
	sArcRds    []*xredis.Pool
	upperRedis *xredis.Pool
	tNames     map[int32]string

	// databus channel
	ugcAidChan chan int64
	pgcAidChan chan int64
	// 年报视频
	nbAidChan     chan int64
	videoShotChan chan []int64
	videoFFChan   chan int64
	//稿件禁止项
	internalChan chan int64
	cron         *cron.Cron
	Taishan      *Taishan
	limiter      *limiter
}

type limiter struct {
	ugc   quota.Waiter
	ogv   quota.Waiter
	retry quota.Waiter
	other quota.Waiter
}

type Taishan struct {
	client   taishan.TaishanProxyClient
	tableCfg tableConfig
}

type tableConfig struct {
	Table string
	Token string
}

// New is archive service implementation.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
		limiter: &limiter{
			ugc:   quota.NewWaiter(c.Limiter.UGC),
			ogv:   quota.NewWaiter(c.Limiter.OGV),
			retry: quota.NewWaiter(c.Limiter.Retry),
			other: quota.NewWaiter(c.Limiter.Other),
		},
		archiveDao:       archive.New(c),
		controlDao:       codao.New(c),
		Prom:             prom.BusinessInfoCount,
		resultDao:        result.New(c),
		pgcDao:           pgc.New(c),
		locDao:           locDao.New(c),
		archiveResultPub: databus.New(c.ArchiveResultPub),
		redis:            xredis.NewPool(c.Redis),
		upperRedis:       xredis.NewPool(c.UpperRedis),
		pgcAidChan:       make(chan int64, 1024),
		ugcAidChan:       make(chan int64, 1024),
		nbAidChan:        make(chan int64, 1024),
		videoShotChan:    make(chan []int64, 1024),
		videoFFChan:      make(chan int64, 1024),
		internalChan:     make(chan int64, 1024),
		cron:             cron.New(),
	}
	var err error
	for _, re := range s.c.ArcRedises {
		s.arcRedises = append(s.arcRedises, xredis.NewPool(re))
	}
	for _, sards := range s.c.SimpleArcRedis {
		s.sArcRds = append(s.sArcRds, xredis.NewPool(sards))
	}
	zone := env.Zone
	if zone == "" {
		panic("env.Zone is empty")
	}
	t, err := taishan.NewClient(&warden.ClientConfig{Zone: zone})
	if err != nil {
		panic(fmt.Sprintf("taishan.NewClient error(%+v)", err))
	}
	s.Taishan = &Taishan{
		client: t,
		tableCfg: tableConfig{
			Table: c.Taishan.Table,
			Token: c.Taishan.Token,
		},
	}
	for i := 0; i < s.c.Custom.ChanSize; i++ {
		s.waiter.Add(1)
		// nolint:biligowordcheck
		go s.ugcConsumer()
		s.waiter.Add(1)
		// nolint:biligowordcheck
		go s.pgcConsumer()
		s.waiter.Add(1)
		// nolint:biligowordcheck
		go s.nbConsumer()
		s.waiter.Add(1)
		// nolint:biligowordcheck
		go s.videoShotproc()
		s.waiter.Add(1)
		// nolint:biligowordcheck
		go s.videoFFProc()
	}
	s.waiter.Add(1)
	// nolint:biligowordcheck
	go s.internalCacheProc()
	s.startDataBus(c)
	//s.waiter.Add(1)
	//go s.videoupConsumer()
	if err = s.loadTypes(); err != nil {
		panic(fmt.Sprintf("s.loadTypes error(%+v)", err))
	}
	s.startCron(c)
	s.waiter.Add(1)
	// nolint:biligowordcheck
	go s.retryproc(retry.FailList)
	s.waiter.Add(1)
	// nolint:biligowordcheck
	go s.retryproc(retry.FailVideoFFList)
	for i := 0; i < 4; i++ {
		s.waiter.Add(1)
		// nolint:biligowordcheck
		go s.retryproc(retry.FailVideoshotList)
	}
	s.waiter.Add(1)
	// nolint:biligowordcheck
	go s.retryproc(retry.FailInternalList)
	// nolint:biligowordcheck
	go s.rebuildVideoShot()
	return s
}

// databus接railgun
func (s *Service) startDataBus(c *conf.Config) {
	//稿件aid分发
	s.videoupSubV2 = railgun.NewRailGun("VideoupSub-稿件id分发", c.VideoUpSubV2Config.Cfg,
		railgun.NewDatabusV1Inputer(&railgun.DatabusV1Config{Config: c.VideoupSub}),
		railgun.NewSingleProcessor(c.VideoUpSubV2Config.SingleConfig, s.VideoUpUnpack, s.VideoUpDo))
	s.videoupSubV2.Start()
	//合集稿件
	s.seasonNotifyArcSubV2 = railgun.NewRailGun("合集稿件", c.SeasonNotifyArcSubV2Config.Cfg,
		railgun.NewDatabusV1Inputer(&railgun.DatabusV1Config{Config: c.SeasonNotifyArcSub}),
		railgun.NewSingleProcessor(c.SeasonNotifyArcSubV2Config.SingleConfig, s.SeasonNotifyUnpack, s.SeasonNotifyDo))
	s.seasonNotifyArcSubV2.Start()
	//互动视频
	s.steinsGateSubV2 = railgun.NewRailGun("互动视频", c.SteinsGateSubV2Config.Cfg,
		railgun.NewDatabusV1Inputer(&railgun.DatabusV1Config{Config: c.SteinsGateSub}),
		railgun.NewSingleProcessor(c.SteinsGateSubV2Config.SingleConfig, s.SteinsGateUnpack, s.SteinsGateDo))
	s.steinsGateSubV2.Start()
	//缓存-cachesubproc
	s.cacheSubV2 = railgun.NewRailGun("稿件缓存", c.CacheSubV2Config.Cfg,
		railgun.NewDatabusV1Inputer(&railgun.DatabusV1Config{Config: c.CacheSub}),
		railgun.NewSingleProcessor(c.CacheSubV2Config.SingleConfig, s.CacheSubProcUnpack, s.CacheSubProcDo))
	s.cacheSubV2.Start()
	//缓存稿件禁止项：海外禁止
	s.internalSub = railgun.NewRailGun("稿件部分禁止项缓存", c.InternalSubConfig.Cfg,
		railgun.NewDatabusV1Inputer(&railgun.DatabusV1Config{Config: c.InternalSub}),
		railgun.NewSingleProcessor(c.InternalSubConfig.SingleConfig, s.InternalSubUnpack, s.InternalSubProcDo))
	s.internalSub.Start()
}

// cronjob接railgun
func (s *Service) startCron(c *conf.Config) {
	//更新archive_type
	s.loadTypesCron = railgun.NewRailGun("更新archive_type", c.LoadTypesCronConfig.Cfg,
		railgun.NewCronInputer(c.LoadTypesCronConfig.CronInputConfig),
		railgun.NewCronProcessor(c.LoadTypesCronConfig.CronProcConfig, func(ctx context.Context) railgun.MsgPolicy {
			_ = s.loadTypes()
			return railgun.MsgPolicyNormal
		}))
	s.loadTypesCron.Start()
	//更新"creative_type"
	s.syncCreativeTypeCron = railgun.NewRailGun("更新creative_type", c.SyncCreativeTypeCronConfig.Cfg,
		railgun.NewCronInputer(c.SyncCreativeTypeCronConfig.CronInputConfig),
		railgun.NewCronProcessor(c.SyncCreativeTypeCronConfig.CronProcConfig, func(ctx context.Context) railgun.MsgPolicy {
			s.syncCreativeType()
			return railgun.MsgPolicyNormal
		}))
	s.syncCreativeTypeCron.Start()
	//检查consumer数据
	if env.DeployEnv == env.DeployEnvProd {
		s.checkConsumeCron = railgun.NewRailGun("检查consumer数据", c.CheckConsumeCronConfig.Cfg,
			railgun.NewCronInputer(c.CheckConsumeCronConfig.CronInputConfig),
			railgun.NewCronProcessor(c.CheckConsumeCronConfig.CronProcConfig, func(ctx context.Context) railgun.MsgPolicy {
				s.checkConsume()
				return railgun.MsgPolicyNormal
			}))
		s.checkConsumeCron.Start()
	}
}

// nolint:bilirailguncheck
func (s *Service) sendNotify(upInfo *resmdl.ArchiveUpInfo) {
	if upInfo == nil {
		return
	}
	var (
		nw  []byte
		old []byte
		err error
		msg *dbusmdl.Message
		c   = context.TODO()
		rt  = &retry.Info{}
	)
	if nw, err = json.Marshal(resmdl.FromNotifyArc(upInfo.Nw)); err != nil {
		log.Error("json.Marshal(%+v) error(%+v)", upInfo.Nw, err)
		return
	}
	if old, err = json.Marshal(resmdl.FromNotifyArc(upInfo.Old)); err != nil {
		log.Error("json.Marshal(%+v) error(%+v)", upInfo.Old, err)
		return
	}
	msg = &dbusmdl.Message{Action: upInfo.Action, Table: upInfo.Table, New: nw, Old: old}
	if err = s.archiveResultPub.Send(c, strconv.FormatInt(upInfo.Nw.Aid, 10), msg); err != nil {
		log.Error("s.archiveResultPub.Send(%+v) error(%+v)", msg, err)
		rt.Action = retry.FailDatabus
		rt.Data.Aid = upInfo.Nw.Aid
		rt.Data.DatabusMsg = upInfo
		s.PushFail(c, rt, retry.FailList)
		return
	}
	msgStr, _ := json.Marshal(msg)
	log.Info("sendNotify(%s) successed", msgStr)
}

// check consumer stat
func (s *Service) checkConsume() {
	if l := len(s.ugcAidChan); l > s.c.Custom.MonitorSize {
		log.Error(fmt.Sprintf("日志告警 视频过审消息堆积预警 UGC size(%d)", l))
	}
	if l := len(s.pgcAidChan); l > s.c.Custom.MonitorSize {
		log.Error(fmt.Sprintf("日志告警 视频过审消息堆积预警 OGV size(%d)", l))
	}
	if l := len(s.nbAidChan); l > s.c.Custom.MonitorSize {
		log.Error(fmt.Sprintf("日志告警 视频过审消息堆积预警 年报 size(%d)", l))
	}
}

// Close kafaka consumer close.
func (s *Service) Close() (err error) {
	s.closeSub = true
	s.cron.Stop()
	time.Sleep(2 * time.Second)
	close(s.videoShotChan)
	s.closeRetry = true
	s.waiter.Wait()
	//rail_gun
	s.videoupSubV2.Close()
	s.seasonNotifyArcSubV2.Close()
	s.steinsGateSubV2.Close()
	s.cacheSubV2.Close()
	s.internalSub.Close()
	return
}
