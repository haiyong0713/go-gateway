package cards

import (
	"context"
	"sync/atomic"
	"time"

	"go-common/library/log"
	http "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/web-svr/activity/interface/conf"
	act "go-gateway/app/web-svr/activity/interface/dao/actplat"
	dao "go-gateway/app/web-svr/activity/interface/dao/cards"
	likedao "go-gateway/app/web-svr/activity/interface/dao/like"
	"go-gateway/app/web-svr/activity/interface/dao/wechat"
	cardsmdl "go-gateway/app/web-svr/activity/interface/model/cards"
	"go-gateway/app/web-svr/activity/interface/service/account"
	likeApi "go-gateway/app/web-svr/activity/interface/service/like"
	lotteryApi "go-gateway/app/web-svr/activity/interface/service/lottery"

	databusv1 "go-common/library/queue/databus"
)

// Service ...
type Service struct {
	c              *conf.Config
	dao            *dao.Dao
	likeDao        *likedao.Dao
	actDao         *act.Dao
	config         *atomic.Value
	httpClient     *http.Client
	actPlatDatabus *databusv1.Databus
	lotterySvr     *lotteryApi.Service
	likeSvr        *likeApi.Service
	allTask        []*cardsmdl.Task
	followMid      []*cardsmdl.FollowMid
	ogvLink        []*cardsmdl.OgvLink
	account        *account.Service
	wechatdao      *wechat.Dao
	cache          *fanout.Fanout
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:              c,
		dao:            dao.New(c),
		likeDao:        likedao.New(c),
		config:         &atomic.Value{},
		httpClient:     http.NewClient(c.HTTPClient),
		actPlatDatabus: databusv1.New(c.DataBus.ActPlatPub),
		lotterySvr:     lotteryApi.New(c),
		likeSvr:        likeApi.New(c),
		actDao:         act.New(c),
		wechatdao:      wechat.New(c),
		account:        account.New(c),
		cache:          fanout.New("cards", fanout.Worker(1), fanout.Buffer(1024)),
	}

	ctx := context.Background()
	err := s.initTask(ctx)
	if err != nil {
		log.Errorc(ctx, "New init task error: %v", err)
		panic(err)
	}
	err = s.initFollowMid(ctx)
	if err != nil {
		log.Errorc(ctx, "New init follow mid error: %v", err)
		panic(err)
	}
	err = s.initOgvLink(ctx)
	if err != nil {
		log.Errorc(ctx, "New init ogv link error: %v", err)
		panic(err)
	}
	go s.updateTaskLoop()
	go s.updateMidFollowLoop()
	go s.updateOgvLinkoop()
	return s
}

func (s *Service) updateTaskLoop() {
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		err := s.initTask(ctx)
		if err != nil {
			continue
		}
	}
}

func (s *Service) updateMidFollowLoop() {
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		err := s.initFollowMid(ctx)
		if err != nil {
			continue
		}
	}
}

func (s *Service) updateOgvLinkoop() {
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		err := s.initOgvLink(ctx)
		if err != nil {
			continue
		}
	}
}

// Close ...
func (s *Service) Close() {
	s.actPlatDatabus.Close()
}
