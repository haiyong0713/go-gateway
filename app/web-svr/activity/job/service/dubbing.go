package service

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/job/model/dubbing"
	"go-gateway/app/web-svr/activity/job/model/rank"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank"
	t "go-gateway/app/web-svr/activity/job/model/task"
	"math"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	weekSid              = 3000000
	batchAidLimit        = 100
	oneSecondGetAidTimes = 2
	concurrencyOnce      = 1
)

var dubbingRankCtx context.Context

func dubbingRankCtxInit() {
	dubbingRankCtx = trace.SimpleServerTrace(context.Background(), "dubbing rank")
}

// DubbingRank 配音排行榜
func (s *Service) DubbingRank() {
	now := time.Now().Unix()
	if now < s.c.Dubbing.ActivityStart.Unix() || now > s.c.Dubbing.ActivityEnd.Unix() {
		return
	}
	s.dubbingRunning.Lock()
	defer s.dubbingRunning.Unlock()
	dubbingRankCtxInit()
	start := time.Now()
	log.Infoc(dubbingRankCtx, "dubbing rank start (%d)", start.Unix())

	task, childRule, sids, err := s.getChildTaskByTaskID(dubbingRankCtx, s.c.Dubbing.TaskID)
	if err != nil {
		log.Errorc(dubbingRankCtx, "s.getActivitySidsChildTask(%d) error(%v)", s.c.Dubbing.Sid, err)
		return
	}
	mids, mapSidMidArchive, err := s.getDubbingAllMidArchive(dubbingRankCtx, s.c.Dubbing.Sid, sids)
	err = s.dubbingtaskAndRank(dubbingRankCtx, s.c.Dubbing.Sid, task, childRule, mids, mapSidMidArchive)
	if err != nil {
		log.Errorc(dubbingRankCtx, "s.dubbingtaskAndRank(%d) error(%v)", task.ID, err)
		return
	}
	end := time.Now()
	spend := end.Unix() - start.Unix()
	log.Infoc(dubbingRankCtx, "Dubbing success() spend(%d)", spend)

}

// getDubbingAllMidArchive 获取稿件信息, 全量mids，
func (s *Service) getDubbingAllMidArchive(c context.Context, sid int64, sids []int64) ([]int64, map[int64]rankmdl.ArchiveStatMap, error) {
	var (
		mids               []int64
		midArchive         rankmdl.ArchiveStatMap
		mapChildSidArchive map[int64]rankmdl.ArchiveStatMap
	)
	eg1 := errgroup.WithContext(c)
	eg1.Go(func(ctx context.Context) (err error) {
		mids, err = s.getAllMids(c, sids)
		return err
	})
	eg1.Go(func(ctx context.Context) (err error) {
		midArchive, mapChildSidArchive, err = s.getAllArchive(c, sids)
		return err
	})
	if err := eg1.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, nil, err
	}
	mapChildSidArchive[sid] = midArchive
	return mids, mapChildSidArchive, nil

}

// dubbingtaskAndRank 任务和排行
func (s *Service) dubbingtaskAndRank(c context.Context, sid int64, task *t.Task, childRule []*t.Rule, mids []int64, midArchive map[int64]rankmdl.ArchiveStatMap) (err error) {

	allSidScoreBatch := make([]map[int64]rankmdl.MidScoreMap, 0)
	if midArchive == nil {
		return nil
	}
	// 计算总排行榜+分赛道的分数的各项分数
	for childSid, v := range midArchive {
		res, err := s.dubbingRankScore(c, childSid, v)
		if err != nil {
			return err
		}
		allSidScoreBatch = append(allSidScoreBatch, res)
	}

	allMapSidScoreBatch := make(map[int64]rankmdl.MidScoreMap)
	for _, v := range allSidScoreBatch {
		for k, value := range v {
			allMapSidScoreBatch[k] = value
		}
	}

	eg1 := errgroup.WithContext(c)
	eg1.Go(func(ctx context.Context) (err error) {
		// 总榜+日榜排名计算
		return s.dubbingRank(c, sid, mids, midArchive, allMapSidScoreBatch)
	})
	eg1.Go(func(ctx context.Context) (err error) {
		// 计算任务
		return s.dubbingTaskHour(c, task, childRule, mids, midArchive[sid])
	})
	if err := eg1.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return err
	}
	return nil
}

// dubbingTaskHour 任务部分
func (s *Service) dubbingTaskHour(c context.Context, task *t.Task, childTask []*t.Rule, mids []int64, midArchive rankmdl.ArchiveStatMap) error {
	midRule := s.checkAllMidTask(c, childTask, mids, midArchive)
	count := s.countMidFinishTask(c, midRule)
	err := s.activityTaskRedis(c, task.ID, midRule, count)
	if err != nil {
		log.Errorc(c, "s.activityTaskRedis(%d) error(%v)", task.ID, err)
		return err
	}
	// 如果是最后一个批次，落库
	hourString := time.Now().Format("2006010215")
	hourInt64, _ := strconv.ParseInt(hourString, 10, 64)
	if hourInt64 == s.c.Dubbing.LastBatch {
		err = s.lastTaskResultToDb(c, s.c.Dubbing.Sid, task.ID, midRule)
		log.Errorc(c, "s.lastTaskResultToDb(%d) error(%v)", task.ID, err)
	}
	return err
}

// dubbingRank 排行处理
func (s *Service) dubbingRank(c context.Context, sid int64, mids []int64, mapMidArchive map[int64]rankmdl.ArchiveStatMap, mapMidScoreBatch map[int64]rankmdl.MidScoreMap) error {
	mapSidScoreBatch := make(map[int64]rankmdl.MidScoreBatch)
	rankSidPersonal := make(map[int64][]*rank.Redis)
	for id, v := range mapMidScoreBatch {
		topLength := len(v)
		mapSidScoreBatch[id] = s.dubbingRankResult(c, v, topLength)
		var (
			rankRedis []*rank.Redis
			err       error
		)
		if s.dubbingSidIsWeek(c, id) {
			normalSid := id - weekSid
			if _, ok := mapMidArchive[normalSid]; !ok {
				continue
			}
			rankRedis, err = s.dubbingRedisRank(c, id, mids, mapMidArchive[normalSid], mapSidScoreBatch[id], v)
		} else {
			rankRedis, err = s.dubbingRedisRank(c, id, mids, mapMidArchive[id], mapSidScoreBatch[id], v)

		}

		if err != nil {
			return err
		}
		rankSidPersonal[id] = rankRedis
		if !s.dubbingSidIsWeek(c, id) {
			err = s.dubbingDbRank(c, mapSidScoreBatch[id], id)
			if err != nil {
				return err
			}
		}
	}
	// 开始用户维度的存储
	err := s.dubbingSidMidRankSet(c, mids, rankSidPersonal, mapSidScoreBatch, mapMidScoreBatch)
	if err != nil {
		log.Errorc(c, "s.dubbingSidMidRankSet: error(%v)", err)
		return err
	}
	return nil
}

// dubbingSidMidRankSet
func (s *Service) dubbingSidMidRankSet(c context.Context, mids []int64, rankSidPersonal map[int64][]*rank.Redis, mapSidScoreBatch map[int64]rankmdl.MidScoreBatch, mapMidScoreBatch map[int64]rankmdl.MidScoreMap) (err error) {
	mapLastScore := make(map[int64]int64)
	for sid, v := range mapSidScoreBatch {
		mapLastScore[sid] = s.dubbingGetMidLastScore(c, sid, v.Data)
	}
	allMidSidScore := make(map[int64]*dubbing.MapMidDubbingScore)
	for _, v := range mids {
		score := make(map[int64]*dubbing.RedisDubbing)
		allMidSidScore[v] = &dubbing.MapMidDubbingScore{Mid: v, Score: score}
	}
	// // 将已经计算好的排行结果取出
	for sid, allMidScoreMap := range mapMidScoreBatch {
		midRankMap := make(map[int64]struct{})
		for _, v := range rankSidPersonal[sid] {
			midRankMap[v.Mid] = struct{}{}
		}
		allRankRedis := make(map[int64]*dubbing.RedisDubbing)
		for _, v := range rankSidPersonal[sid] {
			allRankRedis[v.Mid] = &dubbing.RedisDubbing{
				Score: v.Score,
				Diff:  0,
				Rank:  v.Rank,
				Mid:   v.Mid,
			}
		}
		for _, mid := range mids {
			if midScore, ok := allMidScoreMap[mid]; ok {
				if _, ok := midRankMap[mid]; ok {
					continue
				}
				if mapLastScore[sid]-midScore.Score > 0 {
					midScore.Diff = mapLastScore[sid] - midScore.Score
				}
				midRankMap[midScore.Mid] = struct{}{}
				allRankRedis[midScore.Mid] = &dubbing.RedisDubbing{
					Score: midScore.Score,
					Diff:  midScore.Diff,
					Mid:   midScore.Mid,
				}
				continue
			}
			allRankRedis[mid] = &dubbing.RedisDubbing{
				Score: 0,
				Diff:  mapLastScore[sid],
				Mid:   mid,
			}

		}

		for mid, v := range allMidSidScore {
			if score, ok := allRankRedis[mid]; ok {
				if allMidSidScore[mid].Score == nil {
					allMidSidScore[mid].Score = make(map[int64]*dubbing.RedisDubbing)
					allMidSidScore[mid].Mid = mid
				}
				allMidSidScore[mid].Score[sid] = &dubbing.RedisDubbing{
					Mid:   v.Mid,
					Score: score.Score,
					Diff:  score.Diff,
					Rank:  score.Rank,
				}
			}
		}

	}

	redisBatch := make([]*dubbing.MapMidDubbingScore, 0)
	for _, v := range allMidSidScore {
		redisBatch = append(redisBatch, v)
	}
	if len(redisBatch) > 0 {
		for _, v := range redisBatch {
			err = s.dubbing.SetDubbingMidScore(c, v.Mid, v)
			if err != nil {
				log.Errorc(c, "s.dubbingSetMidRankBatch: error(%v)", err)
				err = errors.Wrapf(err, "s.dubbingSetMidRankBatch")
				time.Sleep(time.Second)
				err = s.dubbing.SetDubbingMidScore(c, v.Mid, v)
				if err != nil {
					log.Errorc(c, "s.dubbingSetMidRankBatch retry: error(%v)", err)
				}
			}
		}

	}
	return nil
}

// setMidRankBatch 批量设置用户维度缓存

// dubbingDbRank db rank data
func (s *Service) dubbingDbRank(c context.Context, midScoreBatch rankmdl.MidScoreBatch, sid int64) (err error) {
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
		err = s.addRankBatch(c, rankDb)
		if err != nil {
			log.Errorc(c, "s.rank.BatchAddRank(%v) error(%v)", mids, err)
			err = errors.Wrapf(err, "s.rank.BatchAddRank %v", mids)
		}
	}
	return err
}

// dubbingGetMidLastScore 获取最后一名分数
func (s *Service) dubbingGetMidLastScore(c context.Context, sid int64, midScore []*rankmdl.MidScore) int64 {
	if midScore != nil {
		length := len(midScore)
		if sid == s.c.Dubbing.Sid || sid == s.c.Dubbing.Sid+weekSid {
			if len(midScore) > s.c.Dubbing.TopRank {
				length = s.c.Dubbing.TopRank
			}
		} else {
			if len(midScore) > s.c.Dubbing.ChildTopRank {
				length = s.c.Dubbing.ChildTopRank
			}
		}
		if length > 0 {
			return midScore[length-1].Score
		}
	}
	return 0
}

// dubbingSidIsWeek 是否是周榜
func (s *Service) dubbingSidIsWeek(c context.Context, sid int64) bool {
	if sid == s.c.Dubbing.Sid {
		return false
	}
	for _, v := range s.c.Dubbing.ChildSid {
		if sid == v {
			return false
		}
	}
	return true
}

func (s *Service) dubbingGetRedisData(c context.Context, sid int64, midArchive rankmdl.ArchiveStatMap, midScore []*rankmdl.MidScore, isOnRank bool) (rankRedis []*rankmdl.Redis, err error) {
	rankRedis = make([]*rankmdl.Redis, 0)
	getWeekAid, err := s.getLastWeekArchiveScore(c, sid-weekSid)
	if err != nil {
		log.Errorc(c, "s.getLastWeekArchiveScore (%d)", sid-weekSid)
	}
	for i, v := range midScore {
		if v != nil {
			aids := make([]int64, 0)
			archive, ok := midArchive[v.Mid]

			if ok {
				if s.dubbingSidIsWeek(c, sid) {
					aids = s.dubbingGetTopScoreArchiveWeek(c, archive, getWeekAid)
				} else {
					aids = s.dubbintGetTopScoreArchive(c, archive)
				}
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
				Diff:  v.Diff,
			})
		}
	}
	return rankRedis, nil
}

// dubbingRedisRank redis rank data,midArchive 用户稿件，midScoreBatch 排行榜中的用户，allMidScoreMap 所有用户
func (s *Service) dubbingRedisRank(c context.Context, sid int64, mids []int64, midArchive rankmdl.ArchiveStatMap, midScoreBatch rankmdl.MidScoreBatch, allMidScoreMap rankmdl.MidScoreMap) (rankRedis []*rank.Redis, err error) {
	key := strconv.FormatInt(sid, 10)
	rankRedis, err = s.dubbingGetRedisData(c, sid, midArchive, midScoreBatch.Data, remixIsOnRank)
	if err != nil {
		log.Errorc(c, "s.dubbingGetRedisData error(%v)", err)
		return nil, err
	}
	if len(rankRedis) > 0 {
		if sid == s.c.Dubbing.Sid || sid == s.c.Dubbing.Sid+weekSid {
			if len(rankRedis) > s.c.Dubbing.TopRank {
				rankRedis = rankRedis[:s.c.Dubbing.TopRank]
			}
		} else {
			if len(rankRedis) > s.c.Dubbing.ChildTopRank {
				rankRedis = rankRedis[:s.c.Dubbing.ChildTopRank]
			}
		}
		err = s.rank.SetRank(c, key, rankRedis)
		if err != nil {
			log.Errorc(c, "s.SetRank: error(%v)", err)
			err = errors.Wrapf(err, "s.SetRank")
			return
		}
	}
	return
}

func (s *Service) dubbingRankResult(c context.Context, midScoreMap rankmdl.MidScoreMap, topLength int) rankmdl.MidScoreBatch {
	var midScoreBatch = rankmdl.MidScoreBatch{}
	for _, v := range midScoreMap {
		midScoreBatch.Data = append(midScoreBatch.Data, v)
	}
	midScoreBatch.TopLength = topLength
	rankmdl.Sort(&midScoreBatch)
	return midScoreBatch
}

// dubbingArcScore 计算每个稿件的分数
func (s *Service) dubbingArcScore(c context.Context, midArchive rankmdl.ArchiveStatMap) map[int64]int64 {
	aidScore := make(map[int64]int64)
	if midArchive != nil {
		for _, archiveList := range midArchive {
			if archiveList != nil {
				for _, value := range archiveList {
					aidScore[value.Aid] = dubbingScore(value)
				}
			}
		}
	}
	return aidScore
}

// getLastWeedBatch 获取上周batch
func (s *Service) getLastWeekBatch(c context.Context) int64 {
	curTime := time.Now()
	day := s.c.Dubbing.LastDay
	lastWeekTime := curTime.AddDate(0, 0, -day)
	if lastWeekTime.Unix() < s.c.Dubbing.FirstTime.Unix() {
		lastWeekTime = s.c.Dubbing.FirstTime
	}
	lastWeek := lastWeekTime.Format("2006010215")
	lastWeekInt, _ := strconv.ParseInt(lastWeek, 10, 64)
	return lastWeekInt
}

// getLastWeekArchiveScore 获取上周批次稿件积分
func (s *Service) getLastWeekArchiveScore(c context.Context, sid int64) (map[int64]int64, error) {
	lastWeekInt := s.getLastWeekBatch(c)
	var count int
	archiveScore := make(map[int64]int64)
	for {
		count++
		startTime := time.Now().UnixNano() / 1e6
		res, err := s.dubbing.GetArchiveScore(c, sid, lastWeekInt, count)
		if err != nil || res == nil {
			log.Errorc(c, "s.dubbing.GetArchiveScore (%v)", err)
			return nil, nil
		}
		if len(res) > 0 {
			for k, v := range res {
				archiveScore[k] = v
			}
		}
		if len(res) < batchAidLimit {
			break
		}
		endTime := time.Now().UnixNano() / 1e6
		waitTime := s.getWaitTime(startTime, endTime)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
		}
		if count == tagMaxCount {
			log.Errorc(c, "tag get aid list count equal max")
			return archiveScore, nil
		}
	}
	return archiveScore, nil
}

// setThisWeekArchiveScore 设置稿件积分
func (s *Service) setThisWeekArchiveScore(c context.Context, sid int64, rankBatch int64, data map[int64]int64) (err error) {
	if data == nil || len(data) == 0 {
		return
	}
	var times int
	patch := batchAidLimit
	concurrency := concurrencyOnce
	mids := make([]*rank.ArchiveScore, 0)
	for k, v := range data {
		mids = append(mids, &rank.ArchiveScore{Aid: k, Score: v})
	}
	times = len(mids) / patch / concurrency
	var count int
	for index := 0; index <= times; index++ {
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
			start := i*patch*concurrency + b*patch
			if start >= len(mids) {
				return nil
			}
			reqMids := mids[start:]
			end := start + patch
			if end < len(mids) {
				reqMids = mids[start:end]
			}
			if len(reqMids) > 0 {
				count++
				reqMap := make(map[int64]int64)
				for _, v := range reqMids {
					if v != nil {
						reqMap[v.Aid] = v.Score
					}
				}
				err := s.dubbing.AddArchiveScore(c, sid, rankBatch, count, reqMap)
				if err != nil {
					log.Errorc(c, " s.dubbing.AddArchiveScore: error(%v) batch(%d)", err, batch)
					return err
				}
			}
		}
	}
	return nil
}

// dubbintGetTopScoreArchive 获取稿件排序
func (s *Service) dubbintGetTopScoreArchive(c context.Context, archive []*rankmdl.ArchiveStat) []int64 {
	length := s.c.Dubbing.TopView

	var archiveMap = rank.ArchiveBatch{}
	for _, v := range archive {
		archiveMap[v.Aid] = v
	}
	archiveScore := archiveMap.Score(dubbingScore)
	var archiveScoreBatch = rankmdl.ArchiveScoreBatch{}
	for _, v := range *archiveScore {
		archiveScoreBatch.Data = append(archiveScoreBatch.Data, v)
	}
	archiveScoreBatch.TopLength = length
	rank.Sort(&archiveScoreBatch)
	aids := make([]int64, 0)
	if len(archiveScoreBatch.Data) < length {
		length = len(archiveScoreBatch.Data)
	}
	for _, v := range archiveScoreBatch.Data[:length] {
		aids = append(aids, v.Aid)
	}
	return aids
}

// dubbingScore 积分计算公式
func dubbingScore(arc *rankmdl.ArchiveStat) int64 {
	if arc.View == 0 {
		return 0
	}
	return dubbingGetPlayScore(arc) + dubbingGetOtherScore(arc)
}

// dubbingGetPlayScore 获取播放分数
func dubbingGetPlayScore(arc *rankmdl.ArchiveStat) int64 {
	videos := float64(arc.Videos)
	views := float64(arc.View)
	pRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", (4/(videos+3))), 64)
	aRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", ((300000+views)/(2*views))), 64)
	if aRevise > 1 {
		aRevise = 1
	}
	return int64(math.Floor(views*pRevise*aRevise + 0.5))
}

// dubbingGetOtherScore 获取播放分数
func dubbingGetOtherScore(arc *rankmdl.ArchiveStat) int64 {
	fav := int64(arc.Fav)
	coin := int64(arc.Coin)
	views := int64(arc.View)
	like := int64(arc.Like)
	share := int64(arc.Share)
	return (like*5 + coin*10 + fav*20 + share*50) * (like*5 + coin*10 + fav*20 + share*50) / (views + like*5 + coin*10 + fav*20 + share*50)
}

// dubbingGetTopScoreArchiveWeek 获取稿件周榜排序
func (s *Service) dubbingGetTopScoreArchiveWeek(c context.Context, archive []*rankmdl.ArchiveStat, weekArchive map[int64]int64) []int64 {
	length := s.c.Dubbing.TopView
	var archiveMap = rank.ArchiveBatch{}
	for _, v := range archive {
		archiveMap[v.Aid] = v
	}
	archiveScore := archiveMap.Score(dubbingScore)
	if archiveScore != nil {
		for k := range *archiveScore {
			aid := (*archiveScore)[k].Aid
			if weekArchive != nil {
				if score, ok := weekArchive[aid]; ok {
					(*archiveScore)[k].Score -= score
				}
			}

		}
	}
	var archiveScoreBatch = rankmdl.ArchiveScoreBatch{}
	for _, v := range *archiveScore {
		archiveScoreBatch.Data = append(archiveScoreBatch.Data, v)
	}
	archiveScoreBatch.TopLength = length
	rank.Sort(&archiveScoreBatch)
	aids := make([]int64, 0)
	if len(archiveScoreBatch.Data) < length {
		length = len(archiveScoreBatch.Data)
	}
	for _, v := range archiveScoreBatch.Data[:length] {
		aids = append(aids, v.Aid)
	}
	return aids
}

// dubbingRankScore 计算分数
func (s *Service) dubbingRankScore(c context.Context, sid int64, midArchive rankmdl.ArchiveStatMap) (map[int64]rankmdl.MidScoreMap, error) {
	// 总榜
	mapSidScoreBatch := make(map[int64]rankmdl.MidScoreMap)
	mapSidScoreBatch[sid] = *midArchive.Score(dubbingScore)
	err := s.dubbingHistoryChildRank(c, mapSidScoreBatch[sid], sid)

	// 稿件积分
	if sid == s.c.Dubbing.Sid {
		archiveScore := make(map[int64]int64)
		if midArchive != nil && len(midArchive) > 0 {
			for _, archive := range midArchive {
				for _, v := range archive {
					archiveScore[v.Aid] = v.Score
				}
			}
		}
		hourString := time.Now().Format("2006010215")
		hour, _ := strconv.ParseInt(hourString, 10, 64)
		err := s.setThisWeekArchiveScore(c, sid, hour, archiveScore)
		if err != nil {
			log.Errorc(c, "s.setThisWeekArchiveScore error(%v)", err)
		}
	}

	// 周榜
	dayScore, err := s.dubbingWeekScore(c, sid, mapSidScoreBatch[sid])
	if err != nil {
		return nil, err
	}
	if dayScore != nil {
		mapSidScoreBatch[weekSid+sid] = dayScore
	}
	return mapSidScoreBatch, nil
}

// dubbingWeekScore 周榜积分计算
func (s *Service) dubbingWeekScore(c context.Context, sid int64, midArchive rankmdl.MidScoreMap) (rankmdl.MidScoreMap, error) {
	lastWeekInt := s.getLastWeekBatch(c)
	lastRank, err := s.getRankListByBatchPatch(c, sid, lastWeekInt)
	if err != nil {
		log.Errorc(c, "s.getRankListByBatchPatch(%d,%d) error(%v)", sid, lastWeekInt, err)
		// return nil, err
	}
	dayMidScoreBatch := rankmdl.MidScoreMap{}
	if lastRank == nil || len(lastRank) == 0 {
		return nil, nil
	}
	thisTimesRank := make(map[int64]int64)
	for _, v := range midArchive {
		thisTimesRank[v.Mid] = v.Score
	}
	lastTimesRank := make(map[int64]int64)
	for _, v := range lastRank {
		lastTimesRank[v.Mid] = v.Score
	}
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

// collegeMysqlOffset count mysql offset
func (s *Service) dubbingMysqlOffset(batch int) int {
	return batch * maxBatchLimit
}

// dubbingHistoryChildRank
func (s *Service) dubbingHistoryChildRank(c context.Context, scoreMap rankmdl.MidScoreMap, sid int64) error {
	key := strconv.FormatInt(sid, 10)
	historyRank, err := s.rank.GetRank(c, key)
	if err != nil {
		err = errors.Wrapf(err, "s.rank.GetRank(%s)", key)
		return err
	}
	if historyRank != nil {
		for _, v := range historyRank {
			if _, ok := scoreMap[v.Mid]; ok {
				scoreMap[v.Mid].History = v.Rank
			}
		}
	}
	return nil
}
