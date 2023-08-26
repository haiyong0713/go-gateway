package feed

import (
	"context"
	"fmt"
	"hash/crc32"
	"math/rand"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-feed/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

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

func (s *Service) group(mid int64, buvid string) (group int) {
	if mid == 0 && buvid == "" {
		group = -1
		return
	}
	if mid != 0 {
		group = int(mid % 20)
		return
	}
	// group = int(crc32.ChecksumIEEE([]byte(buvid)) % 20) 老的buvid实验组逻辑
	// ai新的buvid实验组规则 https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661001258044
	group = int(crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s_1CF61D5DE42C7852", buvid))) % 4)
	return
}

func (s *Service) loadRcmdCache() {
	is, err := s.rcmd.RcmdCache(context.Background())
	if err != nil {
		log.Error("%+v", err)
	}
	//nolint:gomnd
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
	if len(aids) < 50 && len(s.rcmdCache) != 0 { // 从http接口获得的aid较少或有错误，暂不更新内存
		return
	}
	if is, err = s.fromArchvies(aids); err != nil {
		log.Error("%+v", err)
		return
	}
	s.rcmdCache = is
}

func (s *Service) fromArchvies(aids []int64) (is []*ai.Item, err error) {
	var as map[int64]*arcgrpc.Arc
	if as, err = s.arc.Archives(context.Background(), aids, 0, "", ""); err != nil {
		return
	}
	is = make([]*ai.Item, 0, len(aids))
	for _, aid := range aids {
		a, ok := as[aid]
		if !ok || a == nil || !a.IsNormal() {
			continue
		}
		is = append(is, &ai.Item{ID: aid, Goto: model.GotoAv, Archive: a})
	}
	return
}

func (s *Service) loadRankCache() {
	rank, err := s.rank.AllRank(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.rankCache = rank
}

func (s *Service) loadConvergeCache() {
	if s.c.Custom.ResourceDegradeSwitch {
		return
	}
	converge, err := s.rsc.ConvergeCards(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.convergeCache = converge
}

func (s *Service) loadDownloadCache() {
	if s.c.Custom.ResourceDegradeSwitch {
		return
	}
	download, err := s.rsc.DownLoadCards(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.downloadCache = download
}

func (s *Service) loadSpecialCache() {
	if s.c.Custom.ResourceDegradeSwitch {
		return
	}
	special, err := s.rsc.SpecialCards(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	idm := map[int64]int64{}
	for _, sp := range special {
		if sp.Goto == model.GotoLive && sp.Pid != 0 {
			idm[sp.Pid] = sp.ID
		}
	}
	if len(special) > 0 {
		s.specialCache = special
	}
}

func (s *Service) loadFollowModeList() {
	list, err := s.rcmd.FollowModeList(context.Background())
	if err != nil {
		log.Error("%+v", err)
		if list, err = s.rcmd.FollowModeListCache(context.Background()); err != nil {
			log.Error("%+v", err)
			return
		}
	} else {
		s.addCache(func() {
			if err := s.rcmd.AddFollowModeListCache(context.Background(), list); err != nil {
				log.Error("Failed to AddFollowModeListCache: %+v", err)
			}
		})
	}
	log.Warn("loadFollowModeList list len(%d)", len(list))
	s.followModeList = list
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

func (s *Service) loadUpCardCache() {
	if s.c.Custom.ResourceDegradeSwitch {
		return
	}
	follow, err := s.rsc.Follow(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.followCache = follow
}

func (s *Service) loadLiveCardCache() {
	liveCard, err := s.lv.Card(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.liveCardCache = liveCard
}

func (s *Service) liveUpRcmdCard(_ context.Context, ids ...int64) (cardm map[int64][]*live.Card, upIDs []int64) {
	if len(ids) == 0 {
		return
	}
	cardm = make(map[int64][]*live.Card, len(ids))
	for _, id := range ids {
		if card, ok := s.liveCardCache[id]; ok {
			cardm[id] = card
			for _, c := range card {
				if c.UID != 0 {
					upIDs = append(upIDs, c.UID)
				}
			}
		}
	}
	return
}

func (s *Service) loadAutoPlayMid() {
	tmp := map[int64]struct{}{}
	for _, mid := range s.c.Custom.AutoPlayMids {
		tmp[mid] = struct{}{}
	}
	s.autoplayMidsCache = tmp
}

func (s *Service) loadRcmdHotCache() {
	tmp, err := s.rcmd.RecommendHot(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.hotAids = tmp
}
