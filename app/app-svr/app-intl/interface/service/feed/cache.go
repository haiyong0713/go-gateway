package feed

import (
	"context"
	"fmt"
	"hash/crc32"
	"math/rand"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-intl/interface/model"
)

// indexCache is.
func (s *Service) indexCache(c context.Context, mid int64, count int) (rs []*ai.Item, err error) {
	var (
		pos, nextPos int
	)
	cache := s.rcmdCache
	if len(cache) < count {
		return
	}
	if pos, err = s.rcmd.PositionCache(c, mid); err != nil {
		return
	}
	rs = make([]*ai.Item, 0, count)
	if pos < len(cache)-count {
		nextPos = pos + count
		rs = append(rs, cache[pos:nextPos]...)
	} else if pos < len(cache) {
		nextPos = count - (len(cache) - pos)
		rs = append(rs, cache[pos:]...)
		rs = append(rs, cache[:nextPos]...)
	} else {
		nextPos = count
		rs = append(rs, cache[:nextPos]...)
	}
	s.addCache(func() {
		if err = s.rcmd.AddPositionCache(context.Background(), mid, nextPos); err != nil {
			log.Error("s.rcmd.AddPositionCache err(%+v)", err)
		}
	})
	return
}

// recommendCache is.
func (s *Service) recommendCache(count int) (rs []*ai.Item) {
	cache := s.rcmdCache
	index := len(cache)
	if count > 0 && count < index {
		index = count
	}
	rs = make([]*ai.Item, 0, index)
	for _, idx := range rand.Perm(len(cache))[:index] {
		rs = append(rs, cache[idx])
	}
	return
}

// group is.
func (s *Service) group(mid int64, buvid string) (group int) {
	if mid == 0 && buvid == "" {
		group = -1
		return
	}
	if mid != 0 {
		if v, ok := s.groupCache[mid]; ok {
			group = v
			return
		}
		group = int(mid % 20)
		return
	}
	// group = int(crc32.ChecksumIEEE([]byte(buvid)) % 20) 老的buvid实验组逻辑
	// ai新的buvid实验组规则 https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661001258044
	group = int(crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s_1CF61D5DE42C7852", buvid))) % 4)
	return
}

// loadRcmdCache is.
// nolint:gomnd
func (s *Service) loadRcmdCache() {
	is, err := s.rcmd.RcmdCache(context.Background())
	if err != nil {
		log.Error("%+v", err)
	}
	if len(is) >= 50 {
		for _, i := range is {
			i.Goto = model.GotoAv
		}
		s.rcmdCache = is
		return
	}
	aids, err := s.rcmd.Hots(context.Background())
	if err != nil {
		log.Error("%+v", err)
	}
	if len(aids) < 50 && len(s.rcmdCache) != 0 {
		return
	}
	s.rcmdCache = s.fromAids(aids)
}

// fromAids is.
func (s *Service) fromAids(aids []int64) (is []*ai.Item) {
	is = make([]*ai.Item, 0, len(aids))
	for _, aid := range aids {
		i := &ai.Item{
			ID:   aid,
			Goto: model.GotoAv,
		}
		is = append(is, i)
	}
	return
}

// rcmdproc is.
func (s *Service) rcmdproc() {
	for {
		time.Sleep(s.tick)
		s.loadRcmdCache()
	}
}

// loadRankCache is.
func (s *Service) loadRankCache() {
	rank, err := s.rank.AllRank(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.rankCache = rank
}

// rankproc is.
func (s *Service) rankproc() {
	for {
		time.Sleep(s.tick)
		s.loadRankCache()
	}
}

// loadUpCardCache is.
func (s *Service) loadUpCardCache() {
	follow, err := s.card.Follow(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.followCache = follow
}

// upCardproc is.
func (s *Service) upCardproc() {
	for {
		time.Sleep(s.tick)
		s.loadUpCardCache()
	}
}

// loadGroupCache is.
func (s *Service) loadGroupCache() {
	group, err := s.rcmd.Group(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.groupCache = group
}

// groupproc is.
func (s *Service) groupproc() {
	for {
		time.Sleep(s.tick)
		s.loadGroupCache()
	}
}

func (s *Service) loadFawkes() {
	fv, err := s.fawkes.FawkesVersion(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	if len(fv) > 0 {
		s.FawkesVersionCache = fv
	}
}

func (s *Service) loadFawkesProc() {
	for {
		time.Sleep(time.Duration(s.c.Custom.FawkesTick))
		s.loadFawkes()
	}
}

func (s *Service) loadConvergeCache() {
	converge, err := s.rsc.ConvergeCards(context.Background())
	if err != nil {
		log.Error("s.rsc.ConvergeCards err is %+v", err)
		return
	}
	s.convergeCache = converge
}

// nolint:staticcheck
func (s *Service) loadSpecialCache() {
	special, err := s.rsc.SpecialCards(context.Background())
	if err != nil {
		log.Error("s.rsc.SpecialCards err is %+v", err)
		return
	}
	var roomIDs []int64
	idm := map[int64]int64{}
	for _, sp := range special {
		if sp.Goto == model.GotoLive && sp.Pid != 0 {
			roomIDs = append(roomIDs, sp.Pid)
			idm[sp.Pid] = sp.ID
		}
	}
	if len(special) > 0 {
		s.specialCache = special
	}
}

func (s *Service) convergeproc() {
	for {
		time.Sleep(time.Duration(s.c.Custom.Tick))
		s.loadConvergeCache()
	}
}

func (s *Service) tabproc() {
	for {
		time.Sleep(time.Minute * 1)
		s.loadTabCache()
	}
}

func (s *Service) specialproc() {
	for {
		time.Sleep(time.Duration(s.c.Custom.Tick))
		s.loadSpecialCache()
	}
}
