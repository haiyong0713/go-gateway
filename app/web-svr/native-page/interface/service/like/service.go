package like

import (
	"context"
	"time"

	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	channelapi "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	topicapi "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	fligrpc "git.bilibili.co/bapis/bapis-go/filter/service"
	media "git.bilibili.co/bapis/bapis-go/pgc/service/media"
	upapi "git.bilibili.co/bapis/bapis-go/up-archive/service"
	"go-common/library/sync/pipeline/fanout"

	arccli "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/native-page/interface/conf"
	accdao "go-gateway/app/web-svr/native-page/interface/dao/account"
	actdao "go-gateway/app/web-svr/native-page/interface/dao/act"
	aegisdao "go-gateway/app/web-svr/native-page/interface/dao/aegis"
	busdao "go-gateway/app/web-svr/native-page/interface/dao/business"
	cardao "go-gateway/app/web-svr/native-page/interface/dao/cartoon"
	"go-gateway/app/web-svr/native-page/interface/dao/dynamic"
	dynvotedao "go-gateway/app/web-svr/native-page/interface/dao/dynamic-vote"
	"go-gateway/app/web-svr/native-page/interface/dao/esports"
	"go-gateway/app/web-svr/native-page/interface/dao/favorite"
	gadao "go-gateway/app/web-svr/native-page/interface/dao/game"
	hmtchanneldao "go-gateway/app/web-svr/native-page/interface/dao/hmt-channel"
	"go-gateway/app/web-svr/native-page/interface/dao/like"
	"go-gateway/app/web-svr/native-page/interface/dao/live"
	"go-gateway/app/web-svr/native-page/interface/dao/lottery"
	nat "go-gateway/app/web-svr/native-page/interface/dao/native"
	"go-gateway/app/web-svr/native-page/interface/dao/pgc"
	platdao "go-gateway/app/web-svr/native-page/interface/dao/plat"
	populardao "go-gateway/app/web-svr/native-page/interface/dao/popular"
	reldao "go-gateway/app/web-svr/native-page/interface/dao/relation"
	replydao "go-gateway/app/web-svr/native-page/interface/dao/reply"
	scoredao "go-gateway/app/web-svr/native-page/interface/dao/score"
	shopdao "go-gateway/app/web-svr/native-page/interface/dao/shop"
	spacedao "go-gateway/app/web-svr/native-page/interface/dao/space"
	tagmdl "go-gateway/app/web-svr/native-page/interface/dao/tag"
	"go-gateway/app/web-svr/native-page/interface/dao/uprating"
)

// Service struct
type Service struct {
	c               *conf.Config
	dao             *like.Dao
	dynamicDao      *dynamic.Dao
	natDao          *nat.Dao
	pgcDao          *pgc.Dao
	lottDao         *lottery.Dao
	liveDao         *live.Dao
	favDao          *favorite.Dao
	tagDao          *tagmdl.Dao
	replyDao        *replydao.Dao
	busDao          *busdao.Dao
	popularDao      *populardao.Dao
	hmtChannelDao   *hmtchanneldao.Dao
	actDao          *actdao.Dao
	platDao         *platdao.Dao
	spaceDao        *spacedao.Dao
	relDao          *reldao.Dao
	accDao          *accdao.Dao
	gameDao         *gadao.Dao
	shopDao         *shopdao.Dao
	upratingDao     *uprating.Dao
	cartoonDao      *cardao.Dao
	aegisDao        *aegisdao.Dao
	dynvoteDao      *dynvotedao.Dao
	scoreDao        *scoredao.Dao
	esportsDao      *esports.Dao
	arcClient       arccli.ArchiveClient
	fliClient       fligrpc.FilterClient
	upClient        upapi.UpArchiveClient
	artClient       artapi.ArticleGRPCClient
	characterClient media.CharacterClient
	topicClient     topicapi.TopicClient
	channelClient   channelapi.ChannelRPCClient
	cache           *fanout.Fanout
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:             c,
		cache:         fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
		dao:           like.New(c),
		liveDao:       live.New(c),
		dynamicDao:    dynamic.New(c),
		lottDao:       lottery.New(c),
		popularDao:    populardao.New(c),
		pgcDao:        pgc.New(c),
		natDao:        nat.New(c),
		tagDao:        tagmdl.New(c),
		replyDao:      replydao.New(c),
		gameDao:       gadao.New(c),
		favDao:        favorite.New(c),
		busDao:        busdao.New(c),
		cartoonDao:    cardao.New(c),
		actDao:        actdao.New(c),
		relDao:        reldao.New(c),
		accDao:        accdao.New(c),
		platDao:       platdao.New(c),
		spaceDao:      spacedao.New(c),
		shopDao:       shopdao.New(c),
		upratingDao:   uprating.New(c),
		aegisDao:      aegisdao.New(c),
		hmtChannelDao: hmtchanneldao.New(c),
		dynvoteDao:    dynvotedao.NewDao(c),
		scoreDao:      scoredao.NewDao(c),
		esportsDao:    esports.NewDao(c),
	}

	var err error
	if s.arcClient, err = arccli.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.fliClient, err = fligrpc.NewClient(c.FliClient); err != nil {
		panic(err)
	}
	if s.upClient, err = upapi.NewClient(c.UpClient); err != nil {
		panic(err)
	}
	if s.artClient, err = artapi.NewClient(c.ArtClient); err != nil {
		panic(err)
	}
	if s.characterClient, err = media.NewClientCharacter(c.CharGRPC); err != nil {
		panic(err)
	}
	if s.topicClient, err = topicapi.NewClient(c.TopicClient); err != nil {
		panic(err)
	}
	if s.channelClient, err = channelapi.NewClient(c.ChannelClient); err != nil {
		panic(err)
	}
	return
}

// Close service
func (s *Service) Close() {
	s.cache.Close()
	s.lottDao.Close()

	// set timeout as 2 seconds, make sure that kafka consumers exit
	//     >>> in order to increase the missing of reserve data
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	<-ctx.Done()
}

// Ping service
func (s *Service) Ping(c context.Context) (err error) {
	return
}
