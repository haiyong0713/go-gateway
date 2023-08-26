package v1

import (
	"context"

	searchadm "go-gateway/app/app-svr/app-feed/admin/model/search"
	pb "go-gateway/app/app-svr/app-search/api/v1"
	"go-gateway/app/app-svr/app-search/configs"
	"go-gateway/app/app-svr/app-search/internal/dao/v1"
	"go-gateway/app/app-svr/app-search/internal/model/search"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/robfig/cron"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.SearchServer), new(*Service)))

// Service service.
type Service struct {
	dao v1.Dao

	c *configs.Config
	// config
	seasonNum             int
	movieNum              int
	seasonShowMore        int
	movieShowMore         int
	upUserNum             int
	uvLimit               int
	userNum               int
	userVideoLimit        int
	userVideoLimitMix     int
	biliUserNum           int
	biliUserVideoLimit    int
	biliUserVideoLimitMix int
	iPadSearchBangumi     int
	iPadSearchFt          int
	// cron
	cron *cron.Cron
	// ai hot archive cache
	hotAids         map[int64]struct{}
	searchTipsCache map[int64]*search.SearchTips
	specialCache    map[int64]*searchadm.SpreadConfig
	systemNotice    map[int64]*search.SystemNotice
}

// New new a service and return.
func New(d v1.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		c:   d.GetConfig(),
		dao: d,
	}
	cf = s.Close
	// configs
	s.seasonNum = s.c.Search.SeasonNum
	s.movieNum = s.c.Search.MovieNum
	s.seasonShowMore = s.c.Search.SeasonMore
	s.movieShowMore = s.c.Search.MovieMore
	s.upUserNum = s.c.Search.UpUserNum
	s.uvLimit = s.c.Search.UVLimit
	s.userNum = s.c.Search.UpUserNum
	s.userVideoLimit = s.c.Search.UVLimit
	s.userVideoLimitMix = s.c.Search.UserVideoLimitMix
	s.biliUserNum = s.c.Search.BiliUserNum
	s.biliUserVideoLimit = s.c.Search.BiliUserVideoLimit
	s.biliUserVideoLimitMix = s.c.Search.BiliUserVideoLimitMix
	s.iPadSearchBangumi = s.c.Search.IPadSearchBangumi
	s.iPadSearchFt = s.c.Search.IPadSearchFt
	// cache
	s.searchTipsCache = map[int64]*search.SearchTips{}
	s.specialCache = map[int64]*searchadm.SpreadConfig{}
	s.cron = cron.New()
	s.hotAids = make(map[int64]struct{})
	s.initCron()
	s.cron.Start()
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {}
