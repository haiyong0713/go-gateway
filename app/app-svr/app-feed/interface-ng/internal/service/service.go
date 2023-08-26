package service

import (
	"context"
	"go-common/library/conf/paladin"
	"go-common/library/log"
	"go-common/library/log/infoc"
	xtime "go-common/library/time"
	pb "go-gateway/app/app-svr/app-feed/interface-ng/api"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	"go-gateway/app/app-svr/app-feed/interface-ng/internal/dao"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/robfig/cron"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.AppFeedNGServer), new(*Service)))

var RequestCount uint64

// Service service.
type Service struct {
	ac                    *paladin.Map
	dao                   dao.Dao
	cron                  *cron.Cron
	customConfig          *CustomConfig
	HotAidSetFunc         func() sets.Int64
	dispatchMid2GroupFunc func(int64) (int, bool)
	autoplayMidSetFunc    func(int64) bool
	followModeSetFunc     func(int64) bool
	// infoc
	logCh chan interface{}
}

// Custom config
type CustomConfig struct {
	AutoPlayMids           []int64
	TransferSwitch         bool
	AutoRefreshTime        xtime.Duration
	Inline                 *InlineConfig
	GotoStoryDislikeReason string
	ShowInfoc              *infoc.Config
	Prefer4GAutoPlay       bool
	CmAdSwitch             bool
}

type InlineConfig struct {
	ShowInlineDanmaku int
}

// New is
func New(d dao.Dao) (*Service, func(), error) {
	return newService(d)
}

func newService(d dao.Dao) (*Service, func(), error) {
	s := &Service{
		ac:           &paladin.TOML{},
		dao:          d,
		cron:         cron.New(),
		customConfig: new(CustomConfig),
		// infoc
		logCh: make(chan interface{}, 1024),
	}
	closeFn := s.Close
	if err := paladin.Watch("application.toml", s.ac); err != nil {
		return nil, nil, err
	}
	if err := s.ac.Get("customConfig").UnmarshalTOML(&s.customConfig); err != nil {
		panic(err)
	}
	s.loadCronAndFunc()
	s.cron.Start()
	//nolint:biligowordcheck
	go s.infocproc()
	return s, closeFn, nil
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {}

func (s *Service) HotAidSet() (func(), func() sets.Int64) {
	out := sets.Int64{}
	load := func() {
		aidSet, err := s.dao.Recommend().Hots(context.Background())
		if err != nil {
			log.Error("%+v", err)
			return
		}
		out = aidSet
	}
	hotAidSet := func() sets.Int64 {
		return out
	}
	return load, hotAidSet
}

func (s *Service) dispatchMidToGroup() (func(), func(int64) (int, bool)) {
	groupMidToGroup := map[int64]int{}
	load := func() {
		group, err := s.dao.Recommend().Group(context.Background())
		if err != nil {
			log.Error("%+v", err)
			return
		}
		groupMidToGroup = group
	}
	dispatch := func(mid int64) (int, bool) {
		group, ok := groupMidToGroup[mid]
		return group, ok
	}
	return load, dispatch
}

func (s *Service) autoplayMids() (func(), func(int64) bool) {
	autoplayMidSet := sets.Int64{}
	load := func() {
		autoplayMidSet.Insert(s.customConfig.AutoPlayMids...)
	}
	autoplayFunc := func(mid int64) bool {
		return autoplayMidSet.Has(mid)
	}
	return load, autoplayFunc
}

func (s *Service) followModeSet() (func(), func(int64) bool) {
	followModeSet := sets.Int64{}
	load := func() {
		reply, err := s.dao.Recommend().FollowModeSet(context.Background())
		if err != nil {
			log.Error("Failed to get follow mode: %+v", err)
			return
		}
		followModeSet = reply
	}
	followModFunc := func(mid int64) bool {
		return followModeSet.Has(mid)
	}
	return load, followModFunc
}

func (s *Service) loadCronAndFunc() {
	hotAidCronFunc, hotAidSetFunc := s.HotAidSet()
	s.HotAidSetFunc = hotAidSetFunc
	autoplayMidsCron, autoplayMidsFunc := s.autoplayMids()
	s.autoplayMidSetFunc = autoplayMidsFunc
	dispatchCron, dispatchFunc := s.dispatchMidToGroup()
	s.dispatchMid2GroupFunc = dispatchFunc
	followModeCron, followModeFunc := s.followModeSet()
	s.followModeSetFunc = followModeFunc

	if err := s.cron.AddFunc("@every 1m", hotAidCronFunc); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc("@every 1m", autoplayMidsCron); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc("@every 1m", dispatchCron); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc("@every 1m", followModeCron); err != nil {
		panic(err)
	}
}
