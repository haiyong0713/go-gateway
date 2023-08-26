package web

import (
	"context"
	"time"

	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web-goblin/interface/conf"
	"go-gateway/app/web-svr/web-goblin/interface/dao/web"
	webmdl "go-gateway/app/web-svr/web-goblin/interface/model/web"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	artgrpc "git.bilibili.co/bapis/bapis-go/article/service"
	cheeseepgrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/episode"
	hisgrpc "git.bilibili.co/bapis/bapis-go/community/interface/history"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	livexroomgrpc "git.bilibili.co/bapis/bapis-go/live/xroom"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"

	"github.com/robfig/cron"
)

const _chCardTypeAv = "av"

// Service struct .
type Service struct {
	c                *conf.Config
	dao              *web.Dao
	tag              *web.TagRPCService
	arcGRPC          arcgrpc.ArchiveClient
	favGRPC          favgrpc.FavoriteClient
	hisGRPC          hisgrpc.HistoryClient
	episodeGRPC      episodegrpc.EpisodeClient
	artGRPC          artgrpc.ArticleGRPCClient
	livexroomGRPC    livexroomgrpc.RoomClient
	channelCards     map[int64][]*webmdl.ChCard
	accGRPC          accgrpc.AccountClient
	cheeseepGRPC     cheeseepgrpc.EpisodeClient
	outArcs          []*webmdl.OutArc
	baiduPushContent []byte
	// cache proc
	cache *fanout.Fanout
	cron  *cron.Cron
}

// New init .
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   web.New(c),
		tag:   web.NewTagRPC(c.TagRPC),
		cache: fanout.New("cache"),
		cron:  cron.New(),
	}
	var err error
	if s.arcGRPC, err = arcgrpc.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.favGRPC, err = favgrpc.NewClient(c.FavClient); err != nil {
		panic(err)
	}
	if s.hisGRPC, err = hisgrpc.NewClient(c.HisClient); err != nil {
		panic(err)
	}
	if s.episodeGRPC, err = episodegrpc.NewClient(c.EpisodeClient); err != nil {
		panic(err)
	}
	if s.artGRPC, err = artgrpc.NewClient(c.ArticleClient); err != nil {
		panic(err)
	}
	if s.livexroomGRPC, err = livexroomgrpc.NewClient(c.LivexroomClient); err != nil {
		panic(err)
	}
	if s.accGRPC, err = accgrpc.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.cheeseepGRPC, err = cheeseepgrpc.NewClient(c.CheeseepClient); err != nil {
		panic(err)
	}
	// nolint:biligowordcheck
	go s.chCardproc()
	s.initCron()
	return s
}

// Ping Service .
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}

// Close Service .
func (s *Service) Close() {
	s.dao.Close()
	s.cron.Stop()
}

func (s *Service) initCron() {
	s.loadOutArcs()
	if err := s.cron.AddFunc(s.c.Spec.OutArc, s.loadOutArcs); err != nil {
		panic(err)
	}
	s.loadBaiduArcContent()
	if err := s.cron.AddFunc(s.c.Spec.BaiduContent, s.loadBaiduArcContent); err != nil {
		panic(err)
	}
	s.cron.Start()
}

func (s *Service) chCardproc() {
	for {
		now := time.Now()
		cardMap, err := s.dao.ChCard(context.Background(), now)
		if err != nil {
			log.Error("chCardproc s.dao.ChCard() error(%v)", err)
			time.Sleep(time.Second)
		}
		l := len(cardMap)
		if l == 0 {
			time.Sleep(time.Duration(s.c.Rule.ChCardInterval))
			continue
		}
		tmp := make(map[int64][]*webmdl.ChCard, l)
		for channelID, card := range cardMap {
			for _, v := range card {
				if v.Type == _chCardTypeAv {
					tmp[channelID] = append(tmp[channelID], v)
				}
			}
		}
		s.channelCards = tmp
		time.Sleep(time.Duration(s.c.Rule.ChCardInterval))
	}
}
