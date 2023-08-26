package common

import (
	"context"
	"time"

	arcClient "git.bilibili.co/bapis/bapis-go/archive/service"
	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"
	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	accdao "go-gateway/app/app-svr/app-feed/admin/dao/account"
	arcdao "go-gateway/app/app-svr/app-feed/admin/dao/archive"
	"go-gateway/app/app-svr/app-feed/admin/dao/article"
	resourceCard "go-gateway/app/app-svr/app-feed/admin/dao/card"
	"go-gateway/app/app-svr/app-feed/admin/dao/comic"
	"go-gateway/app/app-svr/app-feed/admin/dao/dynamic"
	"go-gateway/app/app-svr/app-feed/admin/dao/game"
	"go-gateway/app/app-svr/app-feed/admin/dao/live"
	medialist "go-gateway/app/app-svr/app-feed/admin/dao/media_list"
	"go-gateway/app/app-svr/app-feed/admin/dao/message"
	pgcdao "go-gateway/app/app-svr/app-feed/admin/dao/pgc"
	showdao "go-gateway/app/app-svr/app-feed/admin/dao/show"
	"go-gateway/app/app-svr/app-feed/admin/dao/vip"
)

var ctx = context.Background()

// Service is search service
type Service struct {
	showDao      *showdao.Dao
	pgcDao       *pgcdao.Dao
	accDao       *accdao.Dao
	arcDao       *arcdao.Dao
	client       *httpx.Client
	managerURL   string
	arcClient    arcClient.ArchiveClient
	ArcType      map[int32]*arcmdl.Tp
	tagClient    tagrpc.TagRPCClient
	GameDao      *game.Dao
	messageDao   *message.Dao
	liveDao      *live.Dao
	articleDao   *article.Dao
	comic        *comic.Dao
	dynamic      *dynamic.Dao
	vip          *vip.Dao
	feedUser     *conf.UserFeed
	cardDao      *resourceCard.Dao
	mediaListDao *medialist.Dao
}

// New new a search service
func New(c *conf.Config) (s *Service) {
	var (
		pgc *pgcdao.Dao
		err error
	)
	if pgc, err = pgcdao.New(c); err != nil {
		log.Error("pgcdao.New error(%v)", err)
		return
	}
	s = &Service{
		showDao:      showdao.New(c),
		pgcDao:       pgc,
		accDao:       accdao.New(c),
		arcDao:       arcdao.New(c),
		client:       httpx.NewClient(c.HTTPClient.Read),
		GameDao:      game.New(c),
		managerURL:   c.Host.Manager,
		messageDao:   message.New(c),
		liveDao:      live.New(c),
		articleDao:   article.New(c),
		comic:        comic.New(c),
		dynamic:      dynamic.New(c),
		cardDao:      resourceCard.New(c),
		vip:          vip.New(c),
		mediaListDao: medialist.New(c),
		feedUser:     c.UserFeed,
	}
	if s.arcClient, err = arcClient.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.tagClient, err = tagrpc.NewClient(c.TagGRPClient); err != nil {
		panic(err)
	}
	//nolint:biligowordcheck
	go s.LoadArchiveType()
	return
}

// LoadArchiveType load archive partition
func (s *Service) LoadArchiveType() {
	for {
		s.ArcType, _ = s.ArchiveTypeGrpc()
		time.Sleep(600 * time.Second)
	}
}
