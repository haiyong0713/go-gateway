package rank

import (
	"fmt"
	"sync"
	"time"

	"go-gateway/app/web-svr/activity/interface/conf"
	dao "go-gateway/app/web-svr/activity/interface/dao/rank_v3"
	rankmdl "go-gateway/app/web-svr/activity/interface/model/rank_v3"
	"go-gateway/app/web-svr/activity/interface/service/account"
	"go-gateway/app/web-svr/activity/interface/service/archive"
	"go-gateway/app/web-svr/activity/interface/service/tag"
)

// Service ...
type Service struct {
	c          *conf.Config
	dao        *dao.Dao
	account    *account.Service
	archive    *archive.Service
	tag        *tag.Service
	rankResult *sync.Map
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		dao:     dao.New(c),
		account: account.New(c),
		archive: archive.New(c),
		tag:     tag.New(c),
	}

	// ctx := context.Background()
	s.rankResult = &sync.Map{}
	go s.updateRankResult()
	// go s.updateOgvLinkoop()
	return s
}

// setRankResultByRule ...
func (s *Service) setRankResultByRule(baseID, ruleID int64, result *rankmdl.ResultRank) {
	s.rankResult.Store(fmt.Sprintf("%d_%d", baseID, ruleID), result)
}

// getRankResultByRule ...
func (s *Service) getRankResultByRule(baseID, ruleID int64) (res *rankmdl.ResultRank, ok bool) {
	resRank, ok := s.rankResult.Load(fmt.Sprintf("%d_%d", baseID, ruleID))
	if !ok {
		return nil, ok
	}
	return resRank.(*rankmdl.ResultRank), true
}

func (s *Service) updateRankResult() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		s.deleteRankResult()
	}
}

func (s *Service) deleteRankResult() {
	s.rankResult.Range(func(k, v interface{}) bool {
		now := time.Now().Unix()
		res := v.(*rankmdl.ResultRank)
		if res.AddTime+2592000 < now {
			s.rankResult.Delete(k)
		}
		return true
	})
}

// Close ...
func (s *Service) Close() {
}
