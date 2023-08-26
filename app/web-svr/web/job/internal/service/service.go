package service

import (
	"context"
	"time"

	"go-common/library/conf/paladin"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	archiveapi "go-gateway/app/app-svr/archive/service/api"
	dynamicapi "go-gateway/app/web-svr/dynamic/service/api/v1"
	pb "go-gateway/app/web-svr/web/job/api"
	"go-gateway/app/web-svr/web/job/internal/dao"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/robfig/cron"
)

const (
	_retryTimes = 3
	_retrySleep = 100 * time.Millisecond
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.WebJobBMServer), new(*Service)))

type Config struct {
	ArchiveClient *warden.ClientConfig
	DynamicClient *warden.ClientConfig
	Cron          struct {
		WebTop          string
		RankIndex       string
		RankRecommend   string
		LpRankRecommend string
		RankRegion      string
		RankTag         string
		RankList        string
		OnlineList      string
		NewList         string
		ArcType         string
		RegionList      string
		PopularSeries   string
	}
	RankRids struct {
		Recommend   []int64
		FirstRegion []int64
		Offline     []int32
		NoOriginal  []int32
		RankV2Rids  []int64
		RankOldRids []int64
		DayAll      []int32
		LandingPage []string
	}
	Rule struct {
		RcmdMinCnt int
	}
	ActPlatCfg *railgun.SingleConfig
	PopularAct struct {
		Activity   string
		Counter    string
		RankSwitch bool
	}
}

// Service service.
type Service struct {
	ac             *Config
	dynamicGRPC    dynamicapi.DynamicClient
	archiveGRPC    archiveapi.ArchiveClient
	dao            dao.Dao
	cron           *cron.Cron
	arcTypes       map[int32]*archiveapi.Tp
	actPlatRailGun *railgun.Railgun
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:   &Config{},
		dao:  d,
		cron: cron.New(),
	}
	cf = s.Close
	err = paladin.Get("application.toml").UnmarshalTOML(&s.ac)
	if err != nil {
		return nil, nil, err
	}
	if s.archiveGRPC, err = archiveapi.NewClient(s.ac.ArchiveClient); err != nil {
		return nil, nil, err
	}
	if s.dynamicGRPC, err = dynamicapi.NewClient(s.ac.DynamicClient); err != nil {
		return nil, nil, err
	}
	if err = s.loadArcTypes(); err != nil {
		return nil, nil, err
	}
	err = s.initCron()
	if err != nil {
		return nil, nil, err
	}
	var (
		dc         paladin.Map
		actPlatSub *databus.Config
	)
	if err = paladin.Get("databus.toml").Unmarshal(&dc); err != nil {
		return
	}
	if err = dc.Get("ActPlatSub").UnmarshalTOML(&actPlatSub); err != nil {
		return
	}
	s.initActPlatRailGun(&railgun.DatabusV1Config{Config: actPlatSub}, s.ac.ActPlatCfg)
	return
}

func (s *Service) initCron() error {
	if err := s.cron.AddFunc(s.ac.Cron.ArcType, s.cronArcTypes); err != nil {
		return err
	}
	if s.ac.Cron.WebTop != "" {
		if err := s.cron.AddFunc(s.ac.Cron.WebTop, s.setWebTop); err != nil {
			return err
		}
	}
	if s.ac.Cron.RankIndex != "" {
		if err := s.cron.AddFunc(s.ac.Cron.RankIndex, s.setRankIndex); err != nil {
			return err
		}
	}
	if s.ac.Cron.RankRecommend != "" {
		if err := s.cron.AddFunc(s.ac.Cron.RankRecommend, s.setRankRecommend); err != nil {
			return err
		}
	}
	if s.ac.Cron.LpRankRecommend != "" {
		if err := s.cron.AddFunc(s.ac.Cron.LpRankRecommend, s.setLpRankRecommend); err != nil {
			return err
		}
	}
	if s.ac.Cron.RankRegion != "" {
		if err := s.cron.AddFunc(s.ac.Cron.RankRegion, s.setFirstRankRegion); err != nil {
			return err
		}
		if err := s.cron.AddFunc(s.ac.Cron.RankRegion, s.setSecondRankRegion); err != nil {
			return err
		}
	}
	if s.ac.Cron.RankTag != "" {
		if err := s.cron.AddFunc(s.ac.Cron.RankTag, s.setRankTag); err != nil {
			return err
		}
	}
	if s.ac.Cron.RankList != "" {
		if err := s.cron.AddFunc(s.ac.Cron.RankList, s.setRankList); err != nil {
			return err
		}
	}
	if s.ac.Cron.OnlineList != "" {
		if err := s.cron.AddFunc(s.ac.Cron.OnlineList, s.setOnlineAids); err != nil {
			return err
		}
	}
	if s.ac.Cron.NewList != "" {
		if err := s.cron.AddFunc(s.ac.Cron.NewList, s.setNewListFirstRegion); err != nil {
			return err
		}
		if err := s.cron.AddFunc(s.ac.Cron.NewList, s.setNewListSecondRegion); err != nil {
			return err
		}
	}

	if s.ac.Cron.PopularSeries != "" {
		if err := s.cron.AddFunc(s.ac.Cron.PopularSeries, s.setPopularSeries); err != nil {
			return err
		}
	}
	if s.ac.Cron.RegionList != "" {
		s.setRegionList()
		if err := s.cron.AddFunc(s.ac.Cron.RegionList, s.setRegionList); err != nil {
			return err
		}
	}
	s.cron.Start()
	return nil
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.cron.Stop()
	s.dao.Close()
	s.actPlatRailGun.Close()
}

func retry(callback func() error) error {
	var err error
	for i := 0; i < _retryTimes; i++ {
		if err = callback(); err == nil {
			return nil
		}
		time.Sleep(_retrySleep)
	}
	return err
}
