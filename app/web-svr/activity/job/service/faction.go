package service

import (
	"context"
	"time"

	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/job/model/like"
)

func (s *Service) SetFactionLikes() {
	go func() {
		s.loadFactionLikes()
	}()
}

func (s *Service) loadFactionLikes() {
	now := time.Now()
	if now.Unix() > s.c.Faction.Etime.Unix() {
		return
	}
	if len(s.c.Faction.Sids) == 0 {
		log.Warn("len(s.c.Faction.Sids) == 0")
		return
	}
	log.Warn("loadFactionLikes start")
	ctx := context.Background()
	data := make([]*like.Faction, 0, len(s.c.Faction.Sids))
	for _, sid := range s.c.Faction.Sids {
		likeArcs, err := s.loadLikeList(ctx, sid, _retryTimes)
		if err != nil {
			log.Error("loadFactionLikes s.loadLikeList sid(%d) error(%v)", sid, err)
			return
		}
		var aids []int64
		for _, v := range likeArcs {
			if v != nil && v.Wid > 0 {
				blocked := func() bool {
					for _, blockAid := range s.c.Faction.BlockAids {
						if v.Wid == blockAid {
							return true
						}
					}
					return false
				}()
				if blocked {
					continue
				}
				aids = append(aids, v.Wid)
			}
		}
		aidsLen := len(aids)
		if aidsLen == 0 {
			log.Warn("loadFactionLikes sid(%d) len(aids) == 0", sid)
			return
		}
		var archives []*arcmdl.Arc
		for i := 0; i < aidsLen; i += _aidBulkSize {
			time.Sleep(10 * time.Millisecond)
			var partAids []int64
			if i+_aidBulkSize > aidsLen {
				partAids = aids[i:]
			} else {
				partAids = aids[i : i+_aidBulkSize]
			}
			partArcs, err := s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: partAids})
			if err != nil {
				log.Error("loadFactionLikes s.arcClient.Arcs partAids(%v) error(%v)", partAids, err)
				continue
			}
			for _, v := range partArcs.GetArcs() {
				if v != nil && v.IsNormal() {
					blocked := func() bool {
						for _, blockMid := range s.c.Faction.BlockMids {
							if v.Author.Mid == blockMid {
								return true
							}
						}
						return false
					}()
					if blocked {
						continue
					}
					archives = append(archives, v)
				}
			}
		}
		tmp := &like.Faction{Sid: sid}
		upScoreMap := make(map[int64]int64)
		for _, v := range archives {
			if v != nil && v.IsNormal() {
				tmp.Score += int64(v.Stat.Like)
				upScoreMap[v.Author.Mid] += int64(v.Stat.Like)
			}
		}
		var maxScore, maxMid int64
		for mid, score := range upScoreMap {
			if score > maxScore {
				maxScore = score
				maxMid = mid
			}
		}
		tmp.List = []*like.FactionAcc{{Mid: maxMid, Score: maxScore}}
		data = append(data, tmp)
	}
	if err := s.dao.SetFactionCache(ctx, data, now.Format("2006010215")); err != nil {
		log.Error("loadFactionLikes SetFactionCache data(%+v) err(%+v)", data, err)
		return
	}
	log.Warn("loadFactionLikes success data(%+v)", data)
}
