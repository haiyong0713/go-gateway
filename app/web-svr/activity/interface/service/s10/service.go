package s10

import (
	"sync"
	"sync/atomic"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/s10"
	model "go-gateway/app/web-svr/activity/interface/model/s10"

	"go-common/library/sync/pipeline/fanout"
)

var (
	subTable bool
	cache    *fanout.Fanout
)

type Service struct {
	conf                          *conf.Config
	staticConf                    *Conf
	dao                           *s10.Dao
	bonuses                       atomic.Value
	goodsInfo                     atomic.Value
	robinLotteryInfos             sync.Map
	singleFlightForRestCount      sync.Map
	singleFlightForRoundRestCount sync.Map
	s10Act                        string
	whiteMap                      map[int64]struct{}
	whiteSwitch                   bool
	splitTab                      bool
	points                        int32
	unicomSecretkey               string
	mobileSecretkey               string
	flowRecvLimit                 bool
	flowSwitch                    bool
}

type Conf struct {
	Tasks []*model.Task
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		conf:            c,
		dao:             s10.New(c),
		s10Act:          c.S10Tasks.Act,
		staticConf:      &Conf{Tasks: c.S10Tasks.Task},
		whiteSwitch:     c.S10WhiteList.Switch,
		splitTab:        c.S10General.SplitTable,
		points:          c.S10General.Points,
		unicomSecretkey: c.S10General.UnicomSecretkey,
		mobileSecretkey: c.S10General.MobileSecretkey,
		flowRecvLimit:   c.S10General.FlowRecvLimit,
		flowSwitch:      c.S10General.FlowSwitch,
	}
	subTable = s.splitTab
	cache = fanout.New("cache", fanout.Worker(5), fanout.Buffer(1024))
	s.bonuses.Store(make(map[int32][]*model.Bonus))
	s.goodsInfo.Store(make(map[int32]*model.Bonus))
	s.whiteMap = make(map[int64]struct{}, len(c.S10WhiteList.List))
	for _, v := range c.S10WhiteList.List {
		s.whiteMap[v] = struct{}{}
	}
	go s.allGoodsProc()
	go s.userLottery()
	return s
}

func (s *Service) Close() {
	cache.Close()
}
