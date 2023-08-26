package service

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	rankmdl "go-gateway/app/web-svr/activity/job/model/rank"
	t "go-gateway/app/web-svr/activity/job/model/task"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	remixKey         = "remix"
	daySid           = 1000000
	remixIsOnRank    = true
	remixIsNotOnRank = false
)

// RemixEveryHour 鬼畜活动每小时脚本
func (s *Service) RemixEveryHour() {
	now := time.Now().Unix()
	if now < s.c.Remix.ActivityStart.Unix() || now > s.c.Remix.ActivityEnd.Unix() {
		return
	}
	s.remixRunning.Lock()
	defer s.remixRunning.Unlock()

	c := context.Background()
	task, childRule, sids, err := s.getActivitySidsChildTask(c, s.c.Remix.Sid)
	if err != nil {
		log.Error("s.getActivitySidsChildTask(%d) error(%v)", s.c.Remix.Sid, err)
		return
	}
	mids, midArchive, mapArchive, err := s.getRemixAllMidArchive(c, sids)

	err = s.remixtaskAndRank(c, s.c.Remix.Sid, task, childRule, mids, midArchive, mapArchive)
	if err != nil {
		log.Error("s.remixtaskAndRank(%d) error(%v)", task.ID, err)
		return
	}
	log.Info("Remix success()")

}

// remixtaskAndRank 任务和排行
func (s *Service) remixtaskAndRank(c context.Context, sid int64, task *t.Task, childRule []*t.Rule, mids []int64, midArchive rankmdl.ArchiveStatMap, mapArchive map[int64]rankmdl.ArchiveBatch) error {
	eg := errgroup.WithContext(c)
	var (
		allMapSidScoreBatch   map[int64]rankmdl.MidScoreMap
		childMapSidScoreBatch map[int64]rankmdl.ArchiveScoreMap
	)
	eg.Go(func(ctx context.Context) (err error) {
		// 积分计算
		allMapSidScoreBatch, err = s.remixRankScore(c, sid, midArchive)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		childMapSidScoreBatch, err = s.childRankScore(c, mapArchive)
		return err
	})

	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	eg1 := errgroup.WithContext(c)
	eg1.Go(func(ctx context.Context) (err error) {
		// 总榜+日榜排名计算
		return s.remixRank(c, sid, mids, midArchive, allMapSidScoreBatch)
	})
	eg1.Go(func(ctx context.Context) (err error) {
		// 分赛道排名计算
		return s.remixChildRank(c, sid, mapArchive, childMapSidScoreBatch)
	})
	eg1.Go(func(ctx context.Context) (err error) {
		// 计算任务
		return s.remixTaskHour(c, task, childRule, mids, midArchive)
	})

	if err := eg1.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return nil
}

// remixRankScore 计算分数
func (s *Service) remixRankScore(c context.Context, sid int64, midArchive rankmdl.ArchiveStatMap) (map[int64]rankmdl.MidScoreMap, error) {
	// 总榜
	mapSidScoreBatch := make(map[int64]rankmdl.MidScoreMap)
	mapSidScoreBatch[sid] = *midArchive.Score(remixScore)
	// 日榜
	dayScore, err := s.remixDayScore(c, sid, mapSidScoreBatch[sid])
	if err != nil {
		return nil, err
	}
	mapSidScoreBatch[daySid+sid] = dayScore
	return mapSidScoreBatch, nil
}

func (s *Service) childRankScore(c context.Context, mapArchive map[int64]rankmdl.ArchiveBatch) (map[int64]rankmdl.ArchiveScoreMap, error) {
	mapSidScoreBatch := make(map[int64]rankmdl.ArchiveScoreMap)
	// 分赛道榜
	for childSid, v := range mapArchive {
		midChildScoreMap := v.Score(remixScore)
		err := s.remixHistoryChildRank(c, midChildScoreMap, childSid)
		if err != nil {
			return nil, err
		}
		mapSidScoreBatch[childSid] = *midChildScoreMap
	}
	return mapSidScoreBatch, nil
}

// getRemixAllMidArchive 获取稿件信息, 全量mids，
func (s *Service) getRemixAllMidArchive(c context.Context, sids []int64) ([]int64, rankmdl.ArchiveStatMap, map[int64]rankmdl.ArchiveBatch, error) {
	var (
		mids          []int64
		midArchive    rankmdl.ArchiveStatMap
		mapMidArchive map[int64]rankmdl.ArchiveStatMap
		mapArchive    map[int64]rankmdl.ArchiveBatch
	)
	eg1 := errgroup.WithContext(c)
	eg1.Go(func(ctx context.Context) (err error) {
		mids, err = s.getAllMids(c, sids)
		return err
	})
	eg1.Go(func(ctx context.Context) (err error) {
		midArchive, mapMidArchive, err = s.getAllArchive(c, sids)
		mapArchive = s.getChildRemixArchive(c, mapMidArchive)
		return err
	})
	if err := eg1.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, nil, nil, err
	}
	return mids, midArchive, mapArchive, nil

}

// getChildRemixArchive 返回全量的用户信息+稿件，分赛道稿件信息
func (s *Service) getChildRemixArchive(c context.Context, mapMidArchive map[int64]rankmdl.ArchiveStatMap) map[int64]rankmdl.ArchiveBatch {
	res := make(map[int64]rankmdl.ArchiveBatch)
	for childSid, v := range mapMidArchive {
		res[childSid] = rankmdl.ArchiveBatch{}
		for _, midArchive := range v {
			for _, archive := range midArchive {
				res[childSid][archive.Aid] = archive
			}
		}
	}
	return res
}

// remixTaskHour 任务部分
func (s *Service) remixTaskHour(c context.Context, task *t.Task, childTask []*t.Rule, mids []int64, midArchive rankmdl.ArchiveStatMap) error {
	midRule := s.checkAllMidTask(c, childTask, mids, midArchive)
	count := s.countMidFinishTask(c, midRule)
	err := s.activityTaskRedis(c, task.ID, midRule, count)
	if err != nil {
		log.Error("s.activityTaskRedis(%d) error(%v)", task.ID, err)
		return err
	}
	// 如果是最后一个批次，落库
	hourString := time.Now().Format("2006010215")
	hourInt64, _ := strconv.ParseInt(hourString, 10, 64)
	if hourInt64 == s.c.Remix.LastBatch {
		err = s.lastTaskResultToDb(c, s.c.Remix.Sid, task.ID, midRule)
		log.Error("s.lastTaskResultToDb(%d) error(%v)", task.ID, err)
	}
	return err
}

// remixRank 排行处理
func (s *Service) remixRank(c context.Context, sid int64, mids []int64, mapMidArchive rankmdl.ArchiveStatMap, mapMidScoreBatch map[int64]rankmdl.MidScoreMap) error {
	mapSidScoreBatch := make(map[int64]rankmdl.MidScoreBatch)
	for id, v := range mapMidScoreBatch {
		topLength := s.c.Remix.TopRank
		mapSidScoreBatch[id] = s.remixRankResult(c, v, topLength)
		err := s.remixRedisRank(c, id, mids, mapMidArchive, mapSidScoreBatch[id], v)
		if err != nil {
			return err
		}
	}
	err := s.remixDbRank(c, mapSidScoreBatch[sid], sid)
	if err != nil {
		return err
	}
	return nil
}

// remixChildRank 分赛道处理
func (s *Service) remixChildRank(c context.Context, sid int64, mapArchive map[int64]rankmdl.ArchiveBatch, mapChildScoreBatch map[int64]rankmdl.ArchiveScoreMap) error {
	mapSidScoreBatch := make(map[int64]rankmdl.ArchiveScoreBatch)
	for childSid, v := range mapChildScoreBatch {
		topLength := s.c.Remix.ChildTopRank
		mapSidScoreBatch[childSid] = s.remixChildRankResult(c, v, topLength)
		err := s.remixRedisChildRank(c, childSid, mapArchive[childSid], mapSidScoreBatch[childSid])
		if err != nil {
			return err
		}
		// err = s.remixDbChildRank(c, mapSidScoreBatch[childSid], childSid)
		// if err != nil {
		// 	return err
		// }
	}
	return nil
}

// remixDbRank db rank data
func (s *Service) remixDbRank(c context.Context, midScoreBatch rankmdl.MidScoreBatch, sid int64) (err error) {
	rankDb := make([]*rankmdl.DB, 0)
	mids := make([]int64, 0)
	for _, v := range midScoreBatch.Data {
		mids = append(mids, v.Mid)
	}
	hourString := time.Now().Format("2006010215")
	hour, _ := strconv.ParseInt(hourString, 10, 64)
	for i, v := range midScoreBatch.Data {
		if v != nil {
			var empty struct{}
			rank := &rankmdl.DB{
				Mid:          v.Mid,
				Score:        v.Score,
				Rank:         i + 1,
				Batch:        hour,
				SID:          sid,
				RemarkOrigin: empty,
			}
			rankDb = append(rankDb, rank)
		}
	}
	if len(rankDb) > 0 {
		err = s.rank.BatchAddRank(c, rankDb)
		if err != nil {
			log.Error("s.rank.BatchAddRank(%v) error(%v)", mids, err)
			err = errors.Wrapf(err, "s.rank.BatchAddRank %v", mids)
		}
	}
	return err
}

// remixDbChildRank db rank data
func (s *Service) remixDbChildRank(c context.Context, archiveScoreBatch rankmdl.ArchiveScoreBatch, sid int64) (err error) {
	rankDb := make([]*rankmdl.DB, 0)
	mids := make([]int64, 0)
	for _, v := range archiveScoreBatch.Data {
		mids = append(mids, v.Mid)
	}
	hourString := time.Now().Format("2006010215")
	hour, _ := strconv.ParseInt(hourString, 10, 64)
	for i, v := range archiveScoreBatch.Data {
		if v != nil {
			var empty struct{}
			rank := &rankmdl.DB{
				Mid:          v.Mid,
				Score:        v.Score,
				Rank:         i + 1,
				Batch:        hour,
				SID:          sid,
				RemarkOrigin: empty,
			}
			rankDb = append(rankDb, rank)
		}
	}
	if len(rankDb) > 0 {
		err = s.rank.BatchAddRank(c, rankDb)
		if err != nil {
			log.Error("s.rank.BatchAddRank(%v) error(%v)", mids, err)
			err = errors.Wrapf(err, "s.rank.BatchAddRank %v", mids)
		}
	}
	return err
}

// remixDayScore 日榜积分计算
func (s *Service) remixDayScore(c context.Context, sid int64, midArchive rankmdl.MidScoreMap) (rankmdl.MidScoreMap, error) {
	h := fmt.Sprintf("-%dh", 24) //加上24小时（后一天）
	curTime := time.Now()
	dh, _ := time.ParseDuration(h)
	lastDay := curTime.Add(dh).Format("2006010215")
	lastDayInt, _ := strconv.ParseInt(lastDay, 10, 64)
	lastRank, err := s.rank.GetRankListByBatch(c, sid, lastDayInt)
	if err != nil {
		log.Error("s.rank.GetRankListByBatch(%d,%d) error(%v)", sid, lastDayInt, err)
		return nil, err
	}
	thisTimesRank := make(map[int64]int64)
	for _, v := range midArchive {
		thisTimesRank[v.Mid] = v.Score
	}
	lastTimesRank := make(map[int64]int64)
	for _, v := range lastRank {
		lastTimesRank[v.Mid] = v.Score
	}

	dayMidScoreBatch := rankmdl.MidScoreMap{}
	for mid, v := range thisTimesRank {
		if last, ok := lastTimesRank[mid]; ok {
			dayMidScoreBatch[mid] = &rankmdl.MidScore{
				Mid:   mid,
				Score: v - last,
			}
			continue
		}
		dayMidScoreBatch[mid] = &rankmdl.MidScore{
			Mid:   mid,
			Score: v,
		}
	}
	return dayMidScoreBatch, nil

}

func (s *Service) getRedisData(c context.Context, midArchive rankmdl.ArchiveStatMap, midScore []*rankmdl.MidScore, isOnRank bool) []*rankmdl.Redis {
	rankRedis := make([]*rankmdl.Redis, 0)
	for i, v := range midScore {
		if v != nil {
			aids := make([]int64, 0)
			archive, ok := midArchive[v.Mid]
			if ok {
				aids = s.getTopViewAid(c, archive, s.c.Remix.TopView)
			}
			rank := 0
			if isOnRank {
				rank = i + 1
			}
			rankRedis = append(rankRedis, &rankmdl.Redis{
				Mid:   v.Mid,
				Score: v.Score,
				Rank:  rank,
				Aids:  aids,
			})
		}
	}
	return rankRedis
}

func (s *Service) childGetRedisData(c context.Context, archiveMap rankmdl.ArchiveBatch, archiveScore []*rankmdl.ArchiveScore) []*rankmdl.Redis {
	rankRedis := make([]*rankmdl.Redis, 0)
	for i, v := range archiveScore {
		if v != nil {
			aids := make([]int64, 0)
			aids = append(aids, v.Aid)
			if archive, ok := archiveMap[v.Aid]; ok {
				rankRedis = append(rankRedis, &rankmdl.Redis{
					Mid:   archive.Mid,
					Score: v.Score,
					Rank:  i + 1,
					Aids:  aids,
				})
			}
		}
	}
	return rankRedis
}

func (s *Service) remixRedisChildRank(c context.Context, sid int64, archiveBatch rankmdl.ArchiveBatch, archiveScoreBatch rankmdl.ArchiveScoreBatch) (err error) {
	key := strconv.FormatInt(sid, 10)
	rankRedis := s.childGetRedisData(c, archiveBatch, archiveScoreBatch.Data)
	if len(rankRedis) > 0 {
		err = s.rank.SetRank(c, key, rankRedis)
		if err != nil {
			log.Error("s.SetRank: error(%v)", err)
			err = errors.Wrapf(err, "s.SetRank")
			return err
		}
	}
	return nil
}

// remixRedisRank redis rank data,midArchive 用户稿件，midScoreBatch 排行榜中的用户，allMidScoreMap 所有用户
func (s *Service) remixRedisRank(c context.Context, sid int64, mids []int64, midArchive rankmdl.ArchiveStatMap, midScoreBatch rankmdl.MidScoreBatch, allMidScoreMap rankmdl.MidScoreMap) (err error) {
	key := strconv.FormatInt(sid, 10)
	rankRedis := s.getRedisData(c, midArchive, midScoreBatch.Data, remixIsOnRank)
	if len(rankRedis) > 0 {
		err = s.rank.SetRank(c, key, rankRedis)
		if err != nil {
			log.Error("s.SetRank: error(%v)", err)
			err = errors.Wrapf(err, "s.SetRank")
			return
		}
	}
	// 统计用户维度的积分情况
	if sid == s.c.Remix.Sid {
		midRankMap := make(map[int64]struct{})
		for _, v := range rankRedis {
			midRankMap[v.Mid] = struct{}{}
		}
		// 将已经计算好的排行结果取出
		allRankRedis := make([]*rankmdl.Redis, 0)
		allRankRedis = append(allRankRedis, rankRedis...)

		allMidScore := make([]*rankmdl.MidScore, 0)
		for mid, v := range allMidScoreMap {
			if _, ok := midRankMap[mid]; ok {
				continue
			}
			midRankMap[v.Mid] = struct{}{}
			allMidScore = append(allMidScore, v)
		}
		for _, v := range mids {
			if _, ok := midRankMap[v]; ok {
				continue
			}
			allMidScore = append(allMidScore, &rankmdl.MidScore{
				Mid: v,
			})
		}
		remainRankRedis := s.getRedisData(c, midArchive, allMidScore, remixIsNotOnRank)
		allRankRedis = append(allRankRedis, remainRankRedis...)

		err = s.rank.SetMidRank(c, key, allRankRedis)
		if err != nil {
			log.Error("s.rank.SetMidRank: error(%v)", err)
			err = errors.Wrapf(err, "s.rank.SetMidRank")
			return
		}

	}

	return err
}

// getTopViewAid ...
func (s *Service) getTopViewAid(c context.Context, archive []*rankmdl.ArchiveStat, length int) []int64 {
	var viewArchive rankmdl.ViewArchive
	viewArchive = archive
	sort.Sort(viewArchive)
	aids := make([]int64, 0)
	if len(viewArchive) < length {
		length = len(viewArchive)
	}
	for _, v := range viewArchive[:length] {
		aids = append(aids, v.Aid)
	}
	return aids

}

func (s *Service) remixRankResult(c context.Context, midScoreMap rankmdl.MidScoreMap, topLength int) rankmdl.MidScoreBatch {
	var midScoreBatch = rankmdl.MidScoreBatch{}
	for _, v := range midScoreMap {
		midScoreBatch.Data = append(midScoreBatch.Data, v)
	}
	midScoreBatch.TopLength = topLength
	rankmdl.Sort(&midScoreBatch)
	return midScoreBatch
}

func (s *Service) remixChildRankResult(c context.Context, archiveScoreMap rankmdl.ArchiveScoreMap, topLength int) rankmdl.ArchiveScoreBatch {
	var archiveScoreBatch = rankmdl.ArchiveScoreBatch{}
	for _, v := range archiveScoreMap {
		archiveScoreBatch.Data = append(archiveScoreBatch.Data, v)
	}
	archiveScoreBatch.TopLength = topLength
	rankmdl.Sort(&archiveScoreBatch)
	return archiveScoreBatch
}

func (s *Service) remixHistoryRank(c context.Context, midScoreMap *rankmdl.MidScoreMap, sid int64) error {
	key := strconv.FormatInt(sid, 10)
	historyRank, err := s.rank.GetRank(c, key)
	if err != nil {
		err = errors.Wrapf(err, "s.rank.GetRank(%s)", key)
		return err
	}
	if historyRank != nil {
		for _, v := range historyRank {
			if _, ok := (*midScoreMap)[v.Mid]; ok {
				(*midScoreMap)[v.Mid].History = v.Rank
			}
		}
	}
	return nil
}

func (s *Service) remixHistoryChildRank(c context.Context, scoreMap *rankmdl.ArchiveScoreMap, sid int64) error {
	key := strconv.FormatInt(sid, 10)
	historyRank, err := s.rank.GetRank(c, key)
	if err != nil {
		err = errors.Wrapf(err, "s.rank.GetRank(%s)", key)
		return err
	}
	if historyRank != nil {
		for _, v := range historyRank {
			if _, ok := (*scoreMap)[v.Mid]; ok {
				(*scoreMap)[v.Mid].History = v.Rank
			}
		}
	}
	return nil
}

// remixScore 积分计算公式
func remixScore(arc *rankmdl.ArchiveStat) int64 {
	if arc.View == 0 {
		return 0
	}
	return remixGetPlayScore(arc) + remixGetFavScore(arc) + remixGetTopicScore(arc) + remixGetCoinScore(arc)
}

// remixGetPlayScore 获取播放分数
func remixGetPlayScore(arc *rankmdl.ArchiveStat) int64 {
	videos := float64(arc.Videos)
	views := float64(arc.View)
	pRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", (4/(videos+3))), 64)
	aRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", ((200000+views)/(2*views))), 64)
	if aRevise > 1 {
		aRevise = 1
	}
	return int64(math.Floor(views*pRevise*aRevise + 0.5))
}

// remixGetFavScore 收藏分
func remixGetFavScore(arc *rankmdl.ArchiveStat) int64 {
	return int64(arc.Fav) * 20
}

// remixBRevise ...
func remixBRevise(arc *rankmdl.ArchiveStat) float64 {
	fav := float64(arc.Fav)
	coin := float64(arc.Coin)
	views := float64(arc.View)
	reply := float64(arc.Reply)
	bRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", ((coin*10+fav*20)/(views+coin*10+reply*50))), 64)
	if bRevise > 1 {
		return 1
	}
	return bRevise
}

// remixGetTopicScore 获取讨论分
func remixGetTopicScore(arc *rankmdl.ArchiveStat) int64 {
	reply := float64(arc.Reply)
	return int64(math.Floor(reply*50*remixBRevise(arc) + 0.5))
}

// remixGetCoinScore ...
func remixGetCoinScore(arc *rankmdl.ArchiveStat) int64 {
	coin := float64(arc.Coin)
	return int64(math.Floor(coin*10*remixBRevise(arc) + 0.5))
}
