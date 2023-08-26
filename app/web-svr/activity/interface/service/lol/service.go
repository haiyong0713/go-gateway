package lol

import (
	"context"
	"time"

	coinapi "git.bilibili.co/bapis/bapis-go/community/service/coin"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/lol"
	esmdl "go-gateway/app/web-svr/activity/interface/model/esports_model"
	guemdl "go-gateway/app/web-svr/activity/interface/model/guess"
	lolmdl "go-gateway/app/web-svr/activity/interface/model/lol"
)

// Service service
type Service struct {
	c          *conf.Config
	dao        *lol.Dao
	coinClient coinapi.CoinClient
	// fanout
	cache *fanout.Fanout
	// s10 contest
	contestID     map[int64]struct{}
	contestDetail map[int64]*esmdl.ContestCard
	mainID        map[int64]struct{}
	detailOptions map[int64][]*lolmdl.DetailOption
}

var (
	unSettlementContestIDList map[int64]int64
)

func init() {
	unSettlementContestIDList = make(map[int64]int64, 0)
}

// New new
func New(c *conf.Config) *Service {
	s := &Service{
		c:             c,
		dao:           lol.New(c),
		cache:         fanout.New("cache", fanout.Worker(5), fanout.Buffer(1024)),
		contestID:     make(map[int64]struct{}),
		contestDetail: make(map[int64]*esmdl.ContestCard),
		mainID:        make(map[int64]struct{}),
		detailOptions: make(map[int64][]*lolmdl.DetailOption),
	}
	var err error
	if s.coinClient, err = coinapi.NewClient(c.S10Client.Coin); err != nil {
		panic(err)
	}

	go s.loadContestProc()
	go asyncUnSettlementContestIDList()

	return s
}

func asyncUnSettlementContestIDList() {
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			if d, err := lol.UnSettlementContestIDList(context.Background()); err == nil {
				unSettlementContestIDList = d
			}
		}
	}
}

// loadCourseListProc load contest detail info from mc.
func (s *Service) loadContestProc() {
	for {
		s.loadContestDetail()
		s.loadContestList()
		s.loadDetailOptions()
		time.Sleep(3 * time.Second)
	}
}

func (s *Service) loadContestDetail() {
	contestDetail := make(map[int64]*esmdl.ContestCard)
	res, err := s.dao.ContestListDetail(context.Background())
	if err != nil {
		log.Error("loadContestListProc res(%s) error(%+v)", res, err)
		return
	}
	if len(res) == 0 {
		log.Error("loadContestListProc res(%s) error(%+v)", res, err)
		return
	}
	for _, detail := range res {
		for _, info := range detail {
			contestDetail[info.Contest.ID] = info
		}
	}
	if len(contestDetail) > 0 {
		s.contestDetail = contestDetail
	}
}

// loadCourseProc load oids from mc.
func (s *Service) loadContestList() {
	contestID := make(map[int64]struct{})
	res, err := s.dao.ContestList(context.Background())
	if err != nil {
		log.Error("loadContestProc res(%s) error(%+v)", res, err)
		return
	}
	if len(res) == 0 {
		log.Error("loadContestProc res(%s) error(%+v)", res, err)
		return
	}
	for _, oid := range res {
		contestID[oid] = struct{}{}
	}
	if len(contestID) > 0 {
		s.contestID = contestID
	}
	s.getMainID(context.Background(), res)
}

// GetMainID get mainID from binlog.
func (s *Service) getMainID(c context.Context, oids []int64) {
	var (
		mainID = make(map[int64]struct{})
		list   []*guemdl.MainGuess
		err    error
	)
	if list, err = s.dao.MainList(c, oids); err != nil {
		return
	}
	for _, l := range list {
		if _, ok1 := s.contestID[l.Oid]; ok1 {
			if _, ok2 := s.mainID[l.ID]; !ok2 {
				mainID[l.ID] = struct{}{}
			}
		}
	}
	if len(mainID) > 0 {
		s.mainID = mainID
	}
}

// loadDetailOptions load guess main details from mc.
func (s *Service) loadDetailOptions() {
	res, err := s.dao.GuessDetailOptions(context.Background())
	if err != nil {
		log.Error("loadContestProc  loadDetailOptions s.dao.GuessDetailOptions() error(%+v)", err)
		return
	}
	if len(res) == 0 {
		log.Error("loadContestProc  loadDetailOptions s.dao.GuessDetailOptions() empty")
		return
	}
	s.detailOptions = res
}

// Close close dao.
func (s *Service) Close() {
	s.cache.Close()
	s.dao.Close()
}
