package service

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"go-common/library/aurora"
	"go-common/library/cache/redis"
	"go-common/library/conf/env"
	"go-common/library/log"
	"go-common/library/queue/databus"

	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/stat/job/conf"
	"go-gateway/app/app-svr/stat/job/dao"
	"go-gateway/app/app-svr/stat/job/model"

	"github.com/robfig/cron"
)

// Service is stat job service.
type Service struct {
	c *conf.Config
	// dao
	dao *dao.Dao
	// wait
	waiter sync.WaitGroup
	closed bool
	// databus
	subRedisMap map[string]*databus.Databus
	subMonitor  map[string]*model.Monitor
	// channel: mq中获取到的message会不断向各个channel中放，在channel的对端会有一个routine去处理数据
	statChan   chan *model.StatMsg
	mu         sync.Mutex
	arcClient  arcmdl.ArchiveClient
	arcRedises []*redis.Pool
	statRedis  *redis.Pool
	cron       *cron.Cron
	// 点赞aurora组件，用于处理消费多机房消息
	thumbupAurora *aurora.Aurora
	// 稿件stat railgun接入
	statRgs []*Railgun
}

const _thumbupDiscoveryID = "community.service.thumbup"

// New is stat-job service implementation.
// nolint:biligowordcheck
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
		// dao
		dao:         dao.New(c),
		subRedisMap: make(map[string]*databus.Databus),
		subMonitor:  make(map[string]*model.Monitor),
		statChan:    make(chan *model.StatMsg, 10240),
		cron:        cron.New(),
	}
	// 初始化aurora组件, 用于处理点赞多机房消息
	s.thumbupAurora = aurora.NewSimpleAurora(&aurora.SimpleConfig{DiscoveryID: _thumbupDiscoveryID})
	// 初始化点赞多机房消费railgun
	s.initStatRailgun()
	// 初始化redis逻辑
	for _, re := range s.c.ArcRedises {
		s.arcRedises = append(s.arcRedises, redis.NewPool(re))
	}
	s.statRedis = redis.NewPool(s.c.StatRedis)
	var err error
	if s.arcClient, err = arcmdl.NewClient(c.ArchiveGRPC); err != nil {
		panic(err)
	}
	// view
	s.subRedisMap[model.TypeForView] = databus.New(c.ViewSubRedis)
	s.subMonitor[model.TypeForView] = &model.Monitor{Topic: c.ViewSubRedis.Topic, Count: 0}
	// dm
	s.subRedisMap[model.TypeForDm] = databus.New(c.DmSubRedis)
	s.subMonitor[model.TypeForDm] = &model.Monitor{Topic: c.DmSubRedis.Topic, Count: 0}
	// reply
	s.subRedisMap[model.TypeForReply] = databus.New(c.ReplySubRedis)
	s.subMonitor[model.TypeForReply] = &model.Monitor{Topic: c.ReplySubRedis.Topic, Count: 0}
	// fav
	s.subRedisMap[model.TypeForFav] = databus.New(c.FavSubRedis)
	s.subMonitor[model.TypeForFav] = &model.Monitor{Topic: c.FavSubRedis.Topic, Count: 0}
	// coin
	s.subRedisMap[model.TypeForCoin] = databus.New(c.CoinSubRedis)
	s.subMonitor[model.TypeForCoin] = &model.Monitor{Topic: c.CoinSubRedis.Topic, Count: 0}
	// share
	s.subRedisMap[model.TypeForShare] = databus.New(c.ShareSubRedis)
	s.subMonitor[model.TypeForShare] = &model.Monitor{Topic: c.ShareSubRedis.Topic, Count: 0}
	// rank
	s.subRedisMap[model.TypeForRank] = databus.New(c.RankSubRedis)
	// (灰度方式)废弃like旧方式消费
	s.subRedisMap[model.TypeForLike] = databus.New(c.LikeSubRedis)
	s.subMonitor[model.TypeForLike] = &model.Monitor{Topic: c.LikeSubRedis.Topic, Count: 0}
	// follow
	s.subRedisMap[model.TypeForFollow] = databus.New(c.FollowSubRedis)
	// s.subMonitor[model.TypeForFollow] = &model.Monitor{Topic: c.FollowSubRedis.Topic, Count: 0}

	// 添加 新的点赞多机房消息 服务自检监控
	s.subMonitor[model.TypeForLikeYLF] = &model.Monitor{Topic: c.LikeYLFRailgun.Databus.Topic, Count: 0}
	s.subMonitor[model.TypeForLikeJD] = &model.Monitor{Topic: c.LikeJDRailgun.Databus.Topic, Count: 0}

	for i := 0; i < s.c.Custom.ProcCount; i++ {
		s.waiter.Add(1)
		go s.statDealproc()
		log.Info("start statDealProc(%d)", i)
	}
	if env.DeployEnv == env.DeployEnvProd {
		go s.monitorproc()
	}
	for k, d := range s.subRedisMap {
		s.waiter.Add(1)
		go s.consumerproc(k, d)
	}
	if err = s.cron.AddFunc("0 0 2 * * *", s.HappyBaby); err != nil {
		panic(err)
	}
	s.cron.Start()
	return
}

func (s *Service) monitorproc() {
	for {
		time.Sleep(90 * time.Second)
		s.mu.Lock()
		for _, mo := range s.subMonitor {
			if mo.Count == 0 {
				log.Error("日志告警 stat-job topic(%s) 没消费！！！！", mo.Topic)
			}
			mo.Count = 0
		}
		s.mu.Unlock()
	}
}

// consumerproc consumer all topic
func (s *Service) consumerproc(k string, d *databus.Databus) {
	defer s.waiter.Done()
	var msgs = d.Messages()
	for {
		var (
			err error
			ok  bool
			msg *databus.Message
			now = time.Now().Unix()
		)
		msg, ok = <-msgs
		if !ok || s.closed {
			log.Info("databus(%s) consumer exit", k)
			return
		}
		_ = msg.Commit()
		var ms = &model.StatCount{}
		if err = json.Unmarshal(msg.Value, ms); err != nil {
			log.Error("ArcStat json.Unmarshal(%s) error(%v)", string(msg.Value), err)
			continue
		}
		if ms.Aid <= 0 || (ms.Type != "archive" && ms.Type != "archive_his") {
			log.Warn("ArcStat message(%s) type is not archive nor archive_his, abort", msg.Value)
			continue
		}
		if now-ms.TimeStamp > 8*60*60 { // 太老的消息就不处理了，只处理8个小时以内的消息
			log.Warn("ArcStat topic(%s) message(%s) too early", msg.Topic, msg.Value)
			continue
		}
		stat := &model.StatMsg{Aid: ms.Aid, Type: k, Ts: ms.TimeStamp}
		switch k {
		case model.TypeForView:
			stat.Click = ms.Count
			stat.Platform = ms.Platform
		case model.TypeForDm:
			stat.DM = ms.Count
		case model.TypeForReply:
			stat.Reply = ms.Count
		case model.TypeForFav:
			stat.Fav = ms.Count
		case model.TypeForCoin:
			stat.Coin = ms.Count
		case model.TypeForShare:
			stat.Share = ms.Count
		case model.TypeForRank:
			stat.HisRank = ms.Count // 只有少数稿件有hisRank，nowRank已经停用了
		case model.TypeForLike:
			//在实验白名单+灰度内，不处理，交由新消息处理
			_, inWhitelist := s.c.LikeRailgunWhitelist[strconv.FormatInt(ms.Aid, 10)]
			if inWhitelist || ms.Aid%10000 < s.c.LikeRailgunGray {
				continue
			}
			stat.Like = ms.Count
		case model.TypeForFollow: // 新增ogv追番数据
			stat.Follow = ms.Count
		default:
			log.Error("unknow type(%s) message(%s)", k, msg.Value)
			continue
		}
		s.mu.Lock()
		if _, ok := s.subMonitor[k]; ok {
			s.subMonitor[k].Count++
		}
		s.mu.Unlock()
		s.statChan <- stat
	}
}

// Close Databus consumer close.
func (s *Service) Close() (err error) {
	s.closed = true
	time.Sleep(2 * time.Second)
	log.Info("start close job")
	// close arc stat
	for k, d := range s.subRedisMap {
		d.Close()
		log.Info("databus(%s) cloesed", k)
	}
	// 关闭railgun消费
	s.closeStatRailgun()
	close(s.statChan)
	log.Info("end close job")
	s.waiter.Wait()
	return
}

// Ping check server ok
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}
