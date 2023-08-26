package pgc

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	pgcdao "go-gateway/app/app-svr/app-feed/admin/dao/pgc"
	"go-gateway/app/app-svr/app-feed/admin/util"

	epgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

// Service is egg service
type Service struct {
	pgc      *pgcdao.Dao
	userFeed *conf.UserFeed
}

// New new a egg service
func New(c *conf.Config) (s *Service) {
	var (
		b   *pgcdao.Dao
		err error
	)
	if b, err = pgcdao.New(c); err != nil {
		log.Error("pgcdao.New error(%v)", err)
		return
	}
	s = &Service{
		pgc:      b,
		userFeed: c.UserFeed,
	}
	return
}

// GetSeason get season from pgc
func (s *Service) GetSeason(c context.Context, seasonIDs []int32) (seasonCards map[int32]*seasongrpc.CardInfoProto, err error) {
	if seasonCards, err = s.pgc.CardsInfoReply(c, seasonIDs); err != nil {
		log.Error("%+v", err)
	}
	if len(seasonCards) == 0 {
		err = fmt.Errorf("无效pgc卡片ID(%v)"+util.ErrorPersonFmt, seasonIDs, s.userFeed.Pgc)
		return
	}
	return
}

// GetEp get ep from pgc
func (s *Service) GetEp(c context.Context, epIds []int32) (res map[int32]*epgrpc.EpisodeCardsProto, err error) {
	if res, err = s.pgc.CardsEpInfoReply(c, epIds); err != nil {
		log.Error("%+v", err)
	}
	if len(res) == 0 {
		err = fmt.Errorf("无效pgc卡片ID(%v)"+util.ErrorPersonFmt, epIds, s.userFeed.Pgc)
	}
	return
}
