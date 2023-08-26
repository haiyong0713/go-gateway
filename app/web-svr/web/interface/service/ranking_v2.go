package service

import (
	"context"
	"fmt"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/model"
)

var _emptyRankV2Arc = make([]*model.RankV2Arc, 0)

// nolint: gocognit
func (s *Service) RankingV2(ctx context.Context, typ, rid int64) (*model.RankV2, error) {
	if rid > 0 {
		var ridCheck bool
		for _, v := range s.c.Rule.RankV2Rids {
			if rid == v {
				ridCheck = true
				break
			}
		}
		if !ridCheck {
			log.Warn("RankingV2 rid:%d not match", rid)
			return nil, xecode.RequestErr
		}
	}
	rankData, err := func() (*model.RankV2, error) {
		memData, memOk := s.rankingV2Data[rankingV2MemKey(typ, rid)]
		if !memOk || memData == nil {
			return nil, fmt.Errorf("RankingV2 typ:%d rid:%d not found", typ, rid)
		}
		var aids []int64
		for _, v := range memData.List {
			if v == nil || v.Aid == 0 {
				continue
			}
			aids = append(aids, v.Aid)
			for i, other := range v.Others {
				if other == nil || other.Aid == 0 {
					continue
				}
				if i >= _rankOtherLimit {
					break
				}
				aids = append(aids, other.Aid)
			}
		}
		if len(aids) == 0 {
			return nil, fmt.Errorf("RankingV2 typ:%d rid:%d list nil", typ, rid)
		}
		arcs, arcErr := s.batchArchives(ctx, filterRepeatedInt64(aids))
		if arcErr != nil {
			return nil, arcErr
		}
		res := &model.RankV2{Note: memData.Note}
		for _, item := range memData.List {
			if item == nil || item.Aid == 0 {
				continue
			}
			arc, ok := arcs[item.Aid]
			if !ok || arc == nil || !arc.IsNormal() {
				continue
			}
			model.ClearAttrAndAccess(arc)
			tmp := &model.RankV2Arc{Arc: arc, Bvid: s.avToBv(arc.Aid), Score: 0}
			for _, other := range item.Others {
				if other == nil || other.Aid == 0 {
					continue
				}
				otherArc, ok := arcs[other.Aid]
				if !ok || otherArc == nil || !otherArc.IsNormal() {
					continue
				}
				tmp.Others = append(tmp.Others, &model.RankV2OtherArc{Arc: otherArc, Bvid: s.avToBv(otherArc.Aid), Score: 0})
			}
			res.List = append(res.List, tmp)
		}
		return res, nil
	}()
	if err != nil {
		log.Error("RankingV2 typ:%d rid:%d error:%v", typ, rid, err)
		//get from remote cache
		res, err := s.dao.RankingV2BakCache(ctx, typ, rid)
		if err != nil {
			log.Error("日志告警 RankingV2 RankingV2BakCache typ:%d rid:%d error:%v", typ, rid, err)
			return &model.RankV2{List: _emptyRankV2Arc}, nil
		}
		return res, nil
	}
	if err := s.cache.Do(ctx, func(ctx context.Context) {
		if err := s.dao.SetRankingV2BakCache(ctx, typ, rid, rankData); err != nil {
			log.Error("%+v", err)
		}
	}); err != nil {
		log.Error("%+v", err)
	}
	return rankData, nil
}

func (s *Service) loadRankingV2Data() {
	ctx := context.Background()
	tmp := make(map[string]*model.RankV2Cache)
	for _, rid := range s.c.Rule.RankV2Rids {
		memKey := rankingV2MemKey(0, rid)
		tmp[memKey] = s.rankingV2Data[memKey]
		cacheData, cacheErr := s.dao.RankingV2Cache(ctx, 0, rid)
		if cacheErr != nil {
			log.Error("日志告警 loadRankingV2Data RankingV2Cache rid:%d error:%v", rid, cacheErr)
			continue
		}
		if cacheCnt := len(cacheData.List); cacheCnt < s.c.Rule.MinRankCount {
			log.Error("日志告警 loadRankingV2Data len RankingV2Cache rid:%d count:%d", rid, cacheCnt)
			continue
		}
		tmp[memKey] = cacheData
	}
	for _, typ := range model.RankV2Types {
		memKey := rankingV2MemKey(typ, 0)
		tmp[memKey] = s.rankingV2Data[memKey]
		cacheData, cacheErr := s.dao.RankingV2Cache(ctx, typ, 0)
		if cacheErr != nil {
			log.Error("loadRankingV2Data RankingV2Cache typ:%d error:%v", typ, cacheErr)
			continue
		}
		if cacheCnt := len(cacheData.List); cacheCnt < s.c.Rule.MinRankCount {
			log.Error("日志告警 loadRankingV2Data len RankingV2Cache typ:%d count:%d", typ, cacheCnt)
			continue
		}
		tmp[rankingV2MemKey(typ, 0)] = cacheData
	}
	s.rankingV2Data = tmp
}

func rankingV2MemKey(tye, rid int64) string {
	return fmt.Sprintf("%d_%d", tye, rid)
}
