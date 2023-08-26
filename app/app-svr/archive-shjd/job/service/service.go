package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/aurora"
	"sync"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/env"
	"go-common/library/database/taishan"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"

	"go-gateway/app/app-svr/archive-shjd/job/conf"
	"go-gateway/app/app-svr/archive-shjd/job/dao"
	locDao "go-gateway/app/app-svr/archive-shjd/job/dao/location"
	"go-gateway/app/app-svr/archive-shjd/job/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model/archive"
	"go-gateway/app/app-svr/archive/service/model/videoshot"

	"github.com/robfig/cron"
)

const (
	_tableArchive   = "archive"
	_tableVideo     = "archive_video"
	_tableVideoShot = "archive_video_shot"
	_tableRedirect  = "archive_redirect"
	_tableInternal  = "archive_internal"
	_actionInsert   = "insert"
	_actionUpdate   = "update"
	_actionDelete   = "delete"
)

// Service service
type Service struct {
	c          *conf.Config
	waiter     sync.WaitGroup
	canal      *databus.Databus
	canalChan  chan *model.Message
	subMap     map[string]*databus.Databus
	subView    *databus.Databus
	subDm      *databus.Databus
	subReply   *databus.Databus
	subFav     *databus.Databus
	subCoin    *databus.Databus
	subShare   *databus.Databus
	subRank    *databus.Databus
	subLike    *databus.Databus
	subFollow  *databus.Databus
	notifyPub  *databus.Databus
	cacheSub   *databus.Databus
	statChan   chan *model.StatMsg
	statRedis  *redis.Pool   // stat-job 自用redis
	arcRedises []*redis.Pool // 需更新的arc-service redis集群，嘉定目前3个
	sArcRds    []*redis.Pool // archive mini缓存 只存储必要常用校验信息
	tNames     map[int32]string
	rds        *redis.Pool
	close      bool
	dao        *dao.Dao
	cron       *cron.Cron
	Taishan    *Taishan
	// 点赞aurora组件，用于处理消费多机房消息
	thumbupAurora *aurora.Aurora
	// 稿件stat railgun接入
	statRgs []*Railgun
	locDao  *locDao.Dao
}

type Taishan struct {
	client   taishan.TaishanProxyClient
	tableCfg tableConfig
}

type tableConfig struct {
	Table string
	Token string
}

const _thumbupDiscoveryID = "community.service.thumbup"

// New is archive service implementation.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:         c,
		canal:     databus.New(c.Databus),
		canalChan: make(chan *model.Message, 10240),
		rds:       redis.NewPool(c.Redis),
		subMap:    make(map[string]*databus.Databus),
		// databus
		subView:  databus.New(c.ViewSubRedis),
		subDm:    databus.New(c.DmSubRedis),
		subReply: databus.New(c.ReplySubRedis),
		subFav:   databus.New(c.FavSubRedis),
		subCoin:  databus.New(c.CoinSubRedis),
		subShare: databus.New(c.ShareSubRedis),
		subRank:  databus.New(c.RankSubRedis),
		//(灰度方式)废弃like旧方式消费
		subLike:   databus.New(c.LikeSubRedis),
		subFollow: databus.New(c.FollowSubRedis),
		notifyPub: databus.New(c.NotifyPub),
		cacheSub:  databus.New(c.CacheSub),
		tNames:    make(map[int32]string),
		statChan:  make(chan *model.StatMsg, 10240),
		dao:       dao.New(c),
		cron:      cron.New(),
		locDao:    locDao.New(c),
	}
	// 初始化aurora组件, 用于处理点赞多机房消息
	s.thumbupAurora = aurora.NewSimpleAurora(&aurora.SimpleConfig{DiscoveryID: _thumbupDiscoveryID})
	// 初始化点赞多机房消费railgun
	s.initStatRailgun()
	for _, re := range s.c.ArcRedises {
		s.arcRedises = append(s.arcRedises, redis.NewPool(re))
	}
	for _, sards := range s.c.SimpleArcRedis {
		s.sArcRds = append(s.sArcRds, redis.NewPool(sards))
	}
	var err error
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
	s.statRedis = redis.NewPool(s.c.StatRedis)
	s.subMap[model.TypeForView] = s.subView
	s.subMap[model.TypeForDm] = s.subDm
	s.subMap[model.TypeForReply] = s.subReply
	s.subMap[model.TypeForFav] = s.subFav
	s.subMap[model.TypeForCoin] = s.subCoin
	s.subMap[model.TypeForShare] = s.subShare
	s.subMap[model.TypeForRank] = s.subRank
	//(灰度方式)废弃like旧方式消费
	s.subMap[model.TypeForLike] = s.subLike
	s.subMap[model.TypeForFollow] = s.subFollow
	for i := 0; i < s.c.Custom.ProcCount; i++ {
		s.waiter.Add(1)
		// nolint:biligowordcheck
		go s.canalChanproc()
		s.waiter.Add(1)
		// nolint:biligowordcheck
		go s.statDealproc()
		log.Info("start statDealProc and canalChanProc idx = %d", i)
	}
	for k, d := range s.subMap {
		s.waiter.Add(1)
		// nolint:biligowordcheck
		go s.consumerproc(k, d)
	}
	s.waiter.Add(1)
	// nolint:biligowordcheck
	go s.canalproc()
	s.waiter.Add(1)
	// nolint:biligowordcheck
	go s.retryconsumer()
	if err = s.cron.AddFunc("@every 1m", s.monitorproc); err != nil {
		panic(fmt.Sprintf("cron add func monitorproc error(%+v)", err))
	}
	s.cron.Start()
	// nolint:biligowordcheck
	go s.initAllArcTaishan()
	return s
}

func (s *Service) monitorproc() {
	// nolint:gomnd
	if l := len(s.canalChan); l > 50 {
		log.Error("日志告警 archive-job-shjd canal consume is slow than produce,current channel size(%d)", l)
	}
}

// nolint:gocognit
func (s *Service) canalChanproc() {
	defer s.waiter.Done()
	for {
		m, ok := <-s.canalChan
		if !ok {
			log.Info("canalChanproc closed")
			return
		}
		log.Info("got canal message table(%s) action(%s) old(%s) new(%s)", m.Table, m.Action, m.Old, m.New)
		var err error
		switch m.Table {
		case _tableArchive:
			var (
				old *model.Archive
				nw  *model.Archive
			)
			switch m.Action {
			case _actionInsert:
				if err = json.Unmarshal(m.New, &nw); err != nil {
					log.Error("json.Unmarshal(%s) error(%+v)", m.New, err)
					continue
				}
			case _actionUpdate:
				if err = json.Unmarshal(m.Old, &old); err != nil {
					log.Error("json.Unmarshal(%s) error(%+v)", m.Old, err)
					continue
				}
				if err = json.Unmarshal(m.New, &nw); err != nil {
					log.Error("json.Unmarshal(%s) error(%+v)", m.New, err)
					continue
				}
			default:
				log.Warn("got unknown action(%s)", m.Action)
				continue
			}
			if mtime, _ := nw.MTime.UnixValue(); mtime != 0 && time.Now().Unix()-mtime > s.c.Custom.CanalAlertTime {
				log.Error("日志告警 UpdateCache canal too late aid(%d)", nw.AID)
			}
			s.UpdateCache(old, nw, m.Action)
		case _tableVideo:
			var video *model.Video
			if err = json.Unmarshal(m.New, &video); err != nil {
				log.Error("json.Unmarshal(%s) error(%+v)", m.New, err)
				continue
			}
			switch m.Action {
			case _actionInsert, _actionUpdate:
				s.UpdateVideoCache(context.Background(), video.AID, video.CID)
			case _actionDelete:
				s.DelVideoCache(context.Background(), video.AID, video.CID)
			default:
				bs, _ := json.Marshal(m)
				log.Error("unknown action(%s) message(%s)", m.Action, bs)
			}
		case _tableVideoShot:
			var v *videoshot.Videoshot
			if err = json.Unmarshal(m.New, &v); err != nil {
				log.Error("json.Unmarshal(%s) error(%+v)", m.New, err)
				continue
			}
			switch m.Action {
			case _actionInsert, _actionUpdate:
				s.addVideoShotCache(context.Background(), v.Cid, v.Count, v.HDCount, v.SdCount, v.HDImg, v.SdImg)
			case _actionDelete:
				s.addVideoShotCache(context.Background(), v.Cid, 0, 0, 0, "", "")
			default:
				log.Warn("table(%s) action(%s) item(%+v) skiped", _tableVideoShot, m.Action, v)
			}
		case _tableRedirect:
			var v *archive.ArcRedirect
			if err = json.Unmarshal(m.New, &v); err != nil {
				log.Error("redirect json.Unmarshal(%s) error(%+v)", m.New, err)
				continue
			}
			switch m.Action {
			case _actionInsert, _actionUpdate:
				s.delRedirectCache(context.Background(), v.Aid)
			}
		case _tableInternal:
			var v *arcgrpc.ArcInternal
			if err = json.Unmarshal(m.New, &v); err != nil {
				log.Error("arc_internal json.Unmarshal(%s) error(%+v)", m.New, err)
				continue
			}
			switch m.Action {
			case _actionInsert, _actionUpdate: //不支持物理删除数据
				s.internalCacheHandler(context.Background(), v.Aid)
			default:
				log.Warn("table(%s) action(%s) item(%+v) skiped", _tableInternal, m.Action, v)
			}
		default:
			log.Warn("table(%s) skiped", m.Table)
		}
	}
}

func (s *Service) canalproc() {
	defer s.waiter.Done()
	msgs := s.canal.Messages()
	for {
		msg, ok := <-msgs
		if !ok || s.close {
			close(s.canalChan)
			log.Info("s.closed databus canal")
			return
		}
		var (
			m   = &model.Message{}
			err error
		)
		_ = msg.Commit()
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		s.canalChan <- m
	}
}

// Ping check status
func (s *Service) Ping() (err error) {
	return
}

// Close is
func (s *Service) Close() (err error) {
	s.close = true
	s.cron.Stop()
	time.Sleep(5 * time.Second)
	// 关闭railgun消费
	s.closeStatRailgun()
	s.canal.Close()
	s.cacheSub.Close()
	s.waiter.Wait()
	return
}
