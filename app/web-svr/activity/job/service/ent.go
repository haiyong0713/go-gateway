package service

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/interface/model/like"
	tagmdl "go-main/app/community/tag/service/api"

	"go-common/library/sync/errgroup.v2"
)

const _aidBulkSize = 50

func (s *Service) entRank() {
	ctx := context.Background()
	entSize := 2500
	list, err := s.webDataList(ctx, s.c.Ent.Vid, 0, entSize, _retryTimes)
	if err != nil {
		log.Error("entRankproc s.webDataList(%d,%d,%d) error(%+v)", s.c.Ent.Vid, 0, _objectPieceSize, err)
		return
	}
	var tagEntDatas = make(map[int64][]*like.EntData)
	for _, val := range list {
		if val == nil || val.Data == "" {
			continue
		}
		entData := new(like.EntData)
		if err = json.Unmarshal([]byte(val.Data), entData); err != nil {
			log.Error("entRankproc json.Unmarshal(%v) error(%+v)", val.Data, err)
			continue
		}
		tagEntDatas[entData.TagID] = append(tagEntDatas[entData.TagID], entData)
	}
	for _, v := range s.c.Ent.UpArcSids {
		s.singleEntRank(ctx, v.ArcSid, v.UpSid, tagEntDatas)
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *Service) singleEntRank(c context.Context, arcSid, upSid int64, entData map[int64][]*like.EntData) {
	likeArcs, err := s.loadLikeList(c, arcSid, _retryTimes)
	if err != nil {
		log.Error("entRankproc arcs s.loadLikeList sid(%d) error(%v)", arcSid, err)
		return
	}
	upLids, err := s.loadLikeList(c, upSid, _retryTimes)
	if err != nil || len(upLids) == 0 {
		log.Error("entRankproc ups s.loadLikeList sid(%d) error(%v) or len == 0", upSid, err)
		return
	}
	upLidMap := make(map[int64]struct{}, len(upLids))
	for _, v := range upLids {
		if v != nil && v.ID > 0 {
			upLidMap[v.ID] = struct{}{}
		}
	}
	var aids []int64
	for _, v := range likeArcs {
		if v != nil && v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	if len(aids) == 0 {
		noScoreUpList := make(map[int64]*like.LidLikeRes)
		err = s.setEntUpRank(c, upSid, noScoreUpList, _retryTimes)
		if err != nil {
			log.Error("entRankproc s.setEntUpRank sid(%d) error(%v)", upSid, err)
			return
		}
		log.Warn("entRankproc s.setEntUpRank sid(%d) success", upSid)
		return
	}
	archives, tags, err := s.archiveWithTag(c, aids)
	if err != nil {
		log.Error("entRankproc s.archiveWithTag aids(%v) error(%v)", aids, err)
		return
	}
	type AidScore struct {
		Aid   int64
		Score float64
	}
	var rankArcs []*AidScore
	upArcStats := make(map[int64][]arcmdl.Stat)
	for _, v := range likeArcs {
		arc, ok := archives[v.Wid]
		if !ok || arc == nil || !arc.IsNormal() {
			continue
		}
		rankArcs = append(rankArcs, &AidScore{
			Aid:   arc.Aid,
			Score: float64(arc.Stat.View)/7 + float64(arc.Stat.Like)/2 + float64(arc.Stat.Fav) + float64(arc.Stat.Coin), //播放/7+点赞/2+收藏+硬币
		})
		arcTags, ok := tags[v.Wid]
		if !ok || len(arcTags) == 0 {
			continue
		}
		var (
			firstEntItems []*like.EntData
			tagHitCount   int64
		)
		for _, tag := range arcTags {
			if tag == nil {
				continue
			}
			entItems, ok := entData[tag.Tid]
			if ok && len(entItems) > 0 {
				// 只取第一次命中的匹配项
				if tagHitCount == 0 {
					firstEntItems = entItems
				}
				tagHitCount++
			}
		}
		// 匹配到2个tag，不计算该稿件分数
		if tagHitCount > 1 {
			continue
		}
		for _, item := range firstEntItems {
			if item != nil && item.Lid > 0 {
				if _, ok := upLidMap[item.Lid]; ok {
					upArcStats[item.Lid] = append(upArcStats[item.Lid], arc.Stat)
					break
				}
			}
		}
	}
	sort.Slice(rankArcs, func(i, j int) bool {
		return rankArcs[i].Score > rankArcs[j].Score
	})
	var rankAids []int64
	for _, v := range rankArcs {
		rankAids = append(rankAids, v.Aid)
	}
	if len(rankAids) > _rankViewPieceSize {
		rankAids = rankAids[:_rankViewPieceSize]
	}
	err = s.setViewRank(c, arcSid, rankAids, "ent", _retryTimes)
	if err != nil {
		log.Error("entRankproc s.setViewRank sid(%d) error(%v)", arcSid, err)
	}
	log.Warn("entRankproc s.setViewRank sid(%d) success", arcSid)
	upScores := make(map[int64]*like.LidLikeRes, len(upLids))
	for _, v := range upLids {
		stats, ok := upArcStats[v.ID]
		if !ok {
			continue
		}
		totalStat := new(struct {
			View  float64
			Like  float64
			Count float64
		})
		for _, stat := range stats {
			totalStat.View += float64(stat.View)
			totalStat.Like += float64(stat.Like)
			totalStat.Count++
		}
		// 热度：（相关稿件总播放X0.1+相关稿件点赞）X0.6+相关稿件数量X50X0.4
		upScores[v.ID] = &like.LidLikeRes{
			Lid:   v.ID,
			Score: int64((totalStat.View*0.1+totalStat.Like)*0.6 + totalStat.Count*50*0.4),
		}
	}
	err = s.setEntUpRank(c, upSid, upScores, _retryTimes)
	if err != nil {
		log.Error("entRankproc s.setEntUpRank sid(%d) error(%v)", upSid, err)
		return
	}
	log.Warn("entRankproc s.setEntUpRank sid(%d) success", upSid)
}

// archiveWithTag get archives and tags.
func (s *Service) archiveWithTag(c context.Context, aids []int64) (archives map[int64]*arcmdl.Arc, arcTags map[int64][]*tagmdl.Resource, err error) {
	var (
		mutex   = sync.Mutex{}
		aidsLen = len(aids)
	)
	group := errgroup.WithContext(c)
	archives = make(map[int64]*arcmdl.Arc, aidsLen)
	arcTags = make(map[int64][]*tagmdl.Resource, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		group.Go(func(ctx context.Context) (err error) {
			var arcs *arcmdl.ArcsReply
			arg := &arcmdl.ArcsRequest{Aids: partAids}
			if arcs, err = s.arcClient.Arcs(ctx, arg); err != nil {
				log.Error("s.arcClient.Arcs(%v) error(%v)", partAids, err)
				return
			}
			mutex.Lock()
			for _, v := range arcs.Arcs {
				archives[v.Aid] = v
			}
			mutex.Unlock()
			return
		})
		group.Go(func(ctx context.Context) error {
			reply, tagErr := s.tagClient.ResTags(ctx, &tagmdl.ResTagsReq{Oids: aids, Type: 3})
			if tagErr != nil {
				log.Error("s.tagClient.ResTags aids(%v) error(%v)", aids, err)
				return nil
			}
			if reply != nil {
				mutex.Lock()
				for oid, v := range reply.Resource {
					arcTags[oid] = v.Resource
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	err = group.Wait()
	return
}

func (s *Service) setEntUpRank(c context.Context, sid int64, data map[int64]*like.LidLikeRes, retryTime int) (err error) {
	for i := 0; i < retryTime; i++ {
		if err = s.dao.SetEntUpRank(c, sid, data); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}
