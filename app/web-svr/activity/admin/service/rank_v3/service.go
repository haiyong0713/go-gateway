package rank

import (
	"context"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/admin/conf"
	likeDao "go-gateway/app/web-svr/activity/admin/dao"
	rankDao "go-gateway/app/web-svr/activity/admin/dao/rank_v3"
	"go-gateway/app/web-svr/activity/admin/service/account"
	"go-gateway/app/web-svr/activity/admin/service/archive"
	"go-gateway/app/web-svr/activity/admin/service/tag"
	"time"

	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"
	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

// Service struct
type Service struct {
	c         *conf.Config
	dao       rankDao.Dao
	likedao   *likeDao.Dao
	account   *account.Service
	archive   *archive.Service
	tag       *tag.Service
	tagRPC    tagrpc.TagRPCClient
	arcClient arcgrpc.ArchiveClient
	cache     *fanout.Fanout
}

// Close service
func (s *Service) Close() {
	if s.dao != nil {
		s.dao.Close()
	}
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		dao:     rankDao.New(c),
		likedao: likeDao.New(c),
		account: account.New(c),
		archive: archive.New(c),
		tag:     tag.New(c),
		cache:   fanout.New("rank", fanout.Worker(1), fanout.Buffer(1024)),
	}
	var err error
	if s.arcClient, err = arcgrpc.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.tagRPC, err = tagrpc.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	ctx := context.Background()

	s.changeRankState(ctx)

	go s.updateRankStateLoop()
	return
}

func (s *Service) updateRankStateLoop() {
	ctx := context.Background()
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		err := s.changeRankState(ctx)
		if err != nil {
			continue
		}
	}
}
