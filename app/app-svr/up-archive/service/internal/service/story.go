package service

import (
	"context"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/up-archive/service/api"
)

// nolint:gocognit,gomnd
func (s *Service) ArcPassedStory(ctx context.Context, req *api.ArcPassedStoryReq) (*api.ArcPassedStoryReply, error) {
	if req.NextCount <= 0 && req.PrevCount <= 0 {
		return nil, ecode.RequestErr
	}
	if req.NextCount > 50 {
		req.NextCount = 50
	}
	if req.PrevCount > 50 {
		req.PrevCount = 50
	}
	total, err := s.dao.CacheArcPassedStoryTotal(ctx, req.Mid)
	if err != nil {
		log.Error("ArcPassedStory CacheArcPassedStoryTotal mid:%d error:%+v", req.Mid, err)
		return nil, err
	}
	if total == 0 {
		// 检查key是否存在，发回源消息
		nowTs := time.Now().Unix()
		_ = s.cache.Do(ctx, func(ctx context.Context) {
			exist, existErr := s.dao.CacheArcPassedStoryExists(ctx, req.Mid)
			if existErr != nil {
				log.Error("ArcPassedStory CacheArcPassedStoryExists mid:%d error:%+v", req.Mid, existErr)
				return
			}
			if !exist {
				if pubErr := s.dao.SendBuildCacheMsg(ctx, req.Mid, nowTs); pubErr != nil {
					log.Error("ArcPassedStory upArcPub.Send mid:%d error:%+v", req.Mid, pubErr)
				}
			}
		})
		return nil, ecode.NothingFound
	}
	if total == 1 {
		// 取2个数据检查是否空缓存
		var aids []int64
		aids, err = s.dao.CacheArcPassedStory(ctx, req.Mid, 0, 1, false)
		if err != nil {
			return nil, err
		}
		if len(aids) == 1 && aids[0] == -1 {
			// expire empty cache
			_ = s.cache.Do(ctx, func(ctx context.Context) {
				if expireErr := s.dao.ExpireEmptyArcPassedStory(ctx, req.Mid); expireErr != nil {
					log.Error("ArcPassedStory ExpireEmptyArcPassedStory mid:%d error:%+v", req.Mid, expireErr)
				}
			})
			return nil, ecode.NothingFound
		}
	}
	isAsc := req.Sort == "asc"
	var rank int64
	if req.Aid > 0 {
		rank, err = s.dao.CacheArcPassedStoryAidRank(ctx, req.Mid, req.Aid, isAsc)
		if err != nil {
			log.Error("ArcPassedStory CacheArcPassedStoryAidRankScore mid:%d aid:%d error:%+v", req.Mid, req.Aid, err)
		}
	}
	res := &api.ArcPassedStoryReply{Rank: rank, Total: total}
	if rank == 0 {
		rank = req.Rank
	}
	aidIndex := rank - 1
	if aidIndex < 0 {
		aidIndex = 0
	}
	start, end := func() (int64, int64) {
		var (
			prevStart, prevEnd, nextStart, nextEnd int64
			needPrev                               bool
		)
		if req.PrevCount > 0 && aidIndex > 0 {
			needPrev = true
			prevEnd = aidIndex
			prevStart = prevEnd - req.PrevCount
			if prevStart < 0 {
				prevStart = 0
			}
		}
		if req.NextCount > 0 {
			nextStart = aidIndex
			// 请求aid在列表中时跳过aid的rank
			if rank > 0 {
				nextStart++
			}
			nextEnd = nextStart + req.NextCount - 1
		}
		lastStart := nextStart
		if needPrev {
			lastStart = prevStart
		}
		lastEnd := prevEnd
		if nextEnd > prevEnd {
			lastEnd = nextEnd
		}
		return lastStart, lastEnd
	}()
	allAids, err := s.dao.CacheArcPassedStory(ctx, req.Mid, start, end, isAsc)
	if err != nil {
		log.Error("ArcPassedStory CacheArcPassedStory mid:%d start:%d end:%d isAsc:%v error:%+v", req.Mid, start, end, isAsc, err)
		return nil, err
	}
	for i, v := range allAids {
		if v == req.Aid {
			continue
		}
		resRank := start + int64(i) + 1
		if resRank < rank {
			res.PrevArcs = append(res.PrevArcs, &api.StoryArcs{Aid: v, Rank: resRank})
			continue
		}
		if resRank > rank {
			res.NextArcs = append(res.NextArcs, &api.StoryArcs{Aid: v, Rank: resRank})
		}
	}
	return res, nil
}
