package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"go-common/library/cache/redis"
	"go-common/library/conf/env"
	"go-common/library/log"
	"go-common/library/queue/databus"

	"go-gateway/app/app-svr/archive-honor/service/api"
	"go-gateway/app/app-svr/archive-honor/service/conf"
	"go-gateway/app/app-svr/archive-honor/service/dao"
	arcdao "go-gateway/app/app-svr/archive-honor/service/dao/archive"
	dcdao "go-gateway/app/app-svr/archive-honor/service/dao/dynamic"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

// Service is
type Service struct {
	c          *conf.Config
	d          *dao.Dao
	dynamicDao *dcdao.Dao
	arcDao     *arcdao.Dao
	redis      *redis.Pool
	// wait
	waiter     sync.WaitGroup
	honorSub   *databus.Databus
	rankSub    *databus.Databus
	closeSub   bool
	closeRetry bool
	Feature    *feature.Feature
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		d:          dao.New(c),
		dynamicDao: dcdao.New(c),
		arcDao:     arcdao.New(c),
		redis:      redis.NewPool(c.Redis),
		Feature:    feature.New(nil),
	}
	// nolint:biligowordcheck
	if env.DeployEnv == env.DeployEnvProd || env.DeployEnv == env.DeployEnvUat {
		s.honorSub = databus.New(c.ArchiveHonorSub)
		s.rankSub = databus.New(c.StatRankSub)
		s.waiter.Add(1)
		go s.consumerProc()
		s.waiter.Add(1)
		go s.consumerRankProc()
		s.waiter.Add(1)
		go s.retryproc()
	}
	return
}

// Close resource.
func (s *Service) Close() {
	s.closeRetry = true
	s.closeSub = true
	s.d.Close()
	if s.honorSub != nil {
		s.honorSub.Close()
	}
	if s.rankSub != nil {
		s.rankSub.Close()
	}
	s.waiter.Wait()
}

// consumerProc consumer honor msg
func (s *Service) consumerProc() {
	defer s.waiter.Done()
	for {
		var (
			msg *databus.Message
			ok  bool
			err error
		)
		if s.closeSub {
			log.Error("s.honorSub.messages closed")
			return
		}
		if msg, ok = <-s.honorSub.Messages(); !ok {
			log.Error("s.honorSub.messages closed")
			return
		}
		_ = msg.Commit()
		m := &api.HonorMsg{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", msg.Value, err)
			continue
		}
		log.Info("got honor message key(%s) value(%s) ", msg.Key, msg.Value)
		if m.Aid <= 0 {
			log.Error("aid(%d) <= 0 message(%s)", m.Aid, msg.Value)
			continue
		}
		if _, ok := api.ValidType[m.Type]; !ok {
			log.Error("unknown Type(%d) message(%s)", m.Type, msg.Value)
			continue
		}
		//处理热门消息
		if m.Type == api.TypeHot {
			m.Desc = api.HotDesc
			m.URL = api.HotURL
		}
		switch m.Action {
		case api.ActionUpdate:
			s.HonorUpdate(context.Background(), m.Aid, m.Type, m.URL, m.Desc, m.NaUrl)
		case api.ActionDel:
			s.HonorDel(context.Background(), m.Aid, m.Type)
		default:
			log.Error("unknown Action(%s) message(%s)", m.Action, msg.Value)
			continue
		}
	}
}

// consumerRankProc consumer rank msg
func (s *Service) consumerRankProc() {
	defer s.waiter.Done()
	for {
		var (
			msg *databus.Message
			ok  bool
			err error
		)
		if s.closeSub {
			log.Error("s.rankSub.messages closed")
			return
		}
		if msg, ok = <-s.rankSub.Messages(); !ok {
			log.Error("s.rankSub.messages closed")
			return
		}
		_ = msg.Commit()
		m := &api.StatMsg{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", msg.Value, err)
			continue
		}
		log.Info("got honor message key(%s) value(%s) ", msg.Key, msg.Value)
		if m.Aid <= 0 || m.Type != "archive_his" {
			log.Warn("ArcStat message(%s) error", msg.Value)
			continue
		}
		desc := fmt.Sprintf("全站排行榜最高第%d名", m.Count)
		s.HonorUpdate(context.Background(), m.Aid, api.TypeRank, api.RankURL, desc, "")
	}
}

// SendMsg send message
func (s *Service) SendMsg(c context.Context, aid int64, url string) {
	arc, err := s.arcDao.Arc(c, aid)
	if err != nil {
		log.Error("SendMsg Fail aid(%d) Arc err(%+v)", aid, err)
		return
	}
	msgKey, err := s.dynamicDao.GetMsgKey(c, arc.Title, arc.Author.Name)
	if err != nil {
		log.Error("SendMsg Fail aid(%d) upID(%d) GetMsgKey err(%+v)", aid, arc.Author.Mid, err)
		return
	}
	if err := s.dynamicDao.SendMsg(context.Background(), uint64(arc.Author.Mid), msgKey); err != nil {
		log.Error("SendMsg Fail aid(%d) upID(%d) SendMsg err(%+v) msgKey(%d)", aid, arc.Author.Mid, err, msgKey)
		return
	}
	log.Info("SendMsg Success aid(%d) upID(%d)", aid, arc.Author.Mid)
}
