package service

import (
	"context"
	"fmt"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/job/model/college"
	"go-gateway/app/web-svr/activity/job/model/rank"
	"strconv"
	"sync"
	"time"

	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/pkg/errors"
)

// collegeAllMidRank 校园用户计算排行，以及校园总分
func (s *Service) collegeAllMidRankAndScore(c context.Context, collegeInfo *college.College, version int) (int64, error) {
	mapMidInfo, err := s.getCollegeAllMidScore(c, collegeInfo.ID)

	if err != nil {
		log.Errorc(c, "s.getCollegeAllMidScore(%d) err(%v)", collegeInfo.ID, err)
		return 0, err
	}
	// 设置历史排行
	s.collegeSetMidHistoryRank(c, collegeInfo.ID, mapMidInfo, version)
	collegeScore := s.collegeCountScore(c, collegeInfo, mapMidInfo)
	midInfoList := make([]*rank.CollegeMidInfo, 0)
	for _, v := range mapMidInfo {
		midInfoList = append(midInfoList, v)
	}
	midRankData := rank.CollegeMidScore{}
	midRankData.Data = midInfoList
	midRankData.TopLength = s.c.College.MidTopLength
	rank.Sort(&midRankData)
	lastScore := s.collegeGetMidLastScore(c, midRankData.Data)
	// 排行榜存储
	redisRank, err := s.collegeMidRankResultSave(c, collegeInfo.ID, version, midRankData.Data, s.collegeMidRankSid(c, collegeInfo.ID))
	if err != nil {
		log.Errorc(c, "s.collegeMidRankResultSave(%d) err(%v)", collegeInfo.ID, err)
		return 0, err
	}
	// 用户维度积分记录
	err = s.collegeMidPersonalRank(c, mapMidInfo, redisRank, collegeInfo.ID, version, lastScore)
	if err != nil {
		log.Errorc(c, " s.collegeMidPersonalRank(%d) error(%s)", collegeInfo.ID, err)
		return 0, err
	}
	return collegeScore, nil
}

// collegeMidPersonalRank 用户维度,版本控制，缓存
func (s *Service) collegeMidPersonalRank(c context.Context, mapAllMidInfo map[int64]*rank.CollegeMidInfo, midRank []*rank.Redis, collegeID int64, version int, lastScore int64) error {
	maptopMidRank := make(map[int64]*rank.Redis)
	maptopRankScore := make(map[int]*rank.Redis)
	for _, v := range midRank {
		maptopMidRank[v.Mid] = v
		maptopRankScore[v.Rank] = v
	}
	allMidPersonal := make([]*college.Personal, 0)
	for _, v := range mapAllMidInfo {
		personal := &college.Personal{
			MID:       v.MID,
			Score:     v.Score,
			CollegeID: collegeID,
		}
		// 在排行榜中
		if midInfo, ok := maptopMidRank[v.MID]; ok {
			personal.Rank = midInfo.Rank
			// 计算与上一个排名的差距
			if lastRankScore, ok := maptopRankScore[midInfo.Rank-1]; ok {
				if lastRankScore.Score-personal.Score > 0 {
					personal.Diff = lastRankScore.Score - personal.Score
				}
			}
			allMidPersonal = append(allMidPersonal, personal)
			continue
		}
		// 不在排行榜中
		if lastScore-personal.Score > 0 {
			personal.Diff = lastScore - personal.Score
		}

		allMidPersonal = append(allMidPersonal, personal)
	}
	// 批量缓存用户维度的情况
	err := s.collegeMidPersonalRankSave(c, collegeID, allMidPersonal, version)
	if err != nil {
		log.Errorc(c, "s.collegeMidPersonalRankSave(%d) error(%v)", collegeID, err)
		return err
	}
	return nil
}

// awardResult 获奖结果
func (s *Service) collegeMidPersonalRankSave(c context.Context, collegeID int64, collegePersonal []*college.Personal, version int) (err error) {
	var times int
	patch := personalInfoBatch
	concurrency := concurrencyPersonal
	times = len(collegePersonal) / patch / concurrency
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(collegePersonal) {
					return nil
				}
				reqMids := collegePersonal[start:]
				end := start + patch
				if end < len(collegePersonal) {
					reqMids = collegePersonal[start:end]
				}
				if len(reqMids) > 0 {
					err = s.college.SetMidPersonal(c, reqMids, version)
					if err != nil {
						log.Errorc(c, "s.college.SetMidPersonal(c)")
						err = errors.Wrapf(err, "s.college.SetMidPersonal")
						time.Sleep(time.Second)
						err = s.college.SetMidPersonal(c, reqMids, version)
						if err != nil {
							time.Sleep(time.Second)
							log.Errorc(c, "s.college.SetMidPersonal(c) retries")
							err = s.college.SetMidPersonal(c, reqMids, version)
							if err != nil {
								time.Sleep(time.Second)
								err = s.college.SetMidPersonal(c, reqMids, version)
								if err != nil {
									log.Errorc(c, "s.college.SetMidPersonal(c) retries twice")
									return err
								}
							}
							return nil
						}
					}
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Errorc(c, "eg.Wait error(%v)", err)
			return err
		}
	}
	return nil
}

// redisRank redis rank data
func (s *Service) collegeRedisMidRank(c context.Context, collegeID int64, version int, collegeMidScore []*rank.CollegeMidInfo) (rankRedis []*rank.Redis, err error) {
	rankRedis = make([]*rank.Redis, 0)
	for i, v := range collegeMidScore {
		if v != nil {
			rankRedis = append(rankRedis, &rank.Redis{
				Mid:   v.MID,
				Score: v.Score,
				Rank:  i + 1,
			})
		}
	}
	if len(rankRedis) > 0 {
		err = s.rank.SetRank(c, s.collegeMidRankKey(c, collegeID, version), rankRedis)
		if err != nil {
			log.Errorc(c, "s.rank.SetRank: error(%v)", err)
			err = errors.Wrapf(err, "s.SetRank")
		}
	}
	return rankRedis, err
}

// collegeMidRankSid 获取活动id
func (s *Service) collegeMidRankSid(c context.Context, collegeID int64) int64 {
	sidStr := fmt.Sprintf("%d%05d", s.c.College.MIDSID, collegeID)
	sid, _ := strconv.ParseInt(sidStr, 10, 64)
	return sid
}

// collegeGetMidLastScore 获取最后一名分数
func (s *Service) collegeGetMidLastScore(c context.Context, collegeMidScore []*rank.CollegeMidInfo) int64 {
	if collegeMidScore != nil {
		length := len(collegeMidScore)
		if length > 0 {
			return collegeMidScore[length-1].Score
		}
	}
	return 0
}

// collegeRankResultSave 排名结果保存
func (s *Service) collegeMidRankResultSave(c context.Context, collegeID int64, version int, collegeMidScore []*rank.CollegeMidInfo, sid int64) ([]*rank.Redis, error) {
	eg := errgroup.WithContext(c)
	redisRank := make([]*rank.Redis, 0)
	// redis 存储
	eg.Go(func(ctx context.Context) (err error) {
		redisRank, err = s.collegeRedisMidRank(c, collegeID, version, collegeMidScore)
		return
	})
	// mysql 存储
	eg.Go(func(ctx context.Context) error {
		return s.collegeDbMidRank(c, collegeMidScore, sid)
	})

	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, err
	}
	return redisRank, nil
}

// dbRank db rank data
func (s *Service) collegeDbMidRank(c context.Context, collegeMidScore []*rank.CollegeMidInfo, sid int64) (err error) {
	rankDb := make([]*rank.DB, 0)
	mids := make([]int64, 0)
	for _, v := range collegeMidScore {
		mids = append(mids, v.MID)
	}
	hourString := time.Now().Format("2006010215")
	hour, _ := strconv.ParseInt(hourString, 10, 64)
	for i, v := range collegeMidScore {
		if v != nil {
			rank := &rank.DB{
				Mid:   v.MID,
				Score: v.Score,
				Rank:  i + 1,
				Batch: hour,
				SID:   sid,
			}
			rankDb = append(rankDb, rank)
		}
	}
	if len(rankDb) > 0 {
		err = s.rank.BatchAddRank(c, rankDb)
		if err != nil {
			log.Errorc(c, "s.rank.BatchAddRank(%v) error(%v)", mids, err)
			err = errors.Wrapf(err, "s.rank.BatchAddRank %v", mids)
		}
	}
	return err
}

func (s *Service) collegeMidRankKey(c context.Context, collegeID int64, version int) string {
	return fmt.Sprintf("college_mid:%d:%d", version, collegeID)
}

// collegeCountScore
func (s *Service) collegeCountScore(c context.Context, collegeInfo *college.College, mapMidInfo map[int64]*rank.CollegeMidInfo) int64 {
	var score int64
	for _, v := range mapMidInfo {
		score += v.Score
	}
	// 调整分
	score += collegeInfo.Score
	return score
}

func (s *Service) collegeSetMidHistoryRank(c context.Context, collegeID int64, mapMidInfo map[int64]*rank.CollegeMidInfo, version int) error {
	historyRank, err := s.rank.GetRank(c, s.collegeMidRankKey(c, collegeID, version))
	if err != nil {
		err = errors.Wrapf(err, "s.rank.GetRank")
		return err
	}
	if historyRank != nil {
		for _, v := range historyRank {
			if _, ok := (mapMidInfo)[v.Mid]; ok {
				(mapMidInfo)[v.Mid].History = v.Rank
			}
		}
	}
	return nil
}

// getCollegeAllMidScore 获取某个校园所有用户积分
func (s *Service) getCollegeAllMidScore(c context.Context, collegeID int64) (map[int64]*rank.CollegeMidInfo, error) {
	mids, err := s.getCollegeAllMid(c, collegeID)
	if err != nil {
		log.Errorc(c, "s.getCollegeAllMid(%d)", collegeID)
		return nil, err
	}
	return s.concurrencyGetCollegeMidScore(c, mids)
}

// getCollegeAllMid 获取校园所有用户id
func (s *Service) getCollegeAllMid(c context.Context, collegeID int64) ([]*college.MidInfo, error) {
	var (
		err   error
		batch int
	)
	midList := make([]*college.MidInfo, 0)
	for {
		midBatchList, err := s.college.GetCollegeMidByBatch(c, collegeID, s.collegeMysqlOffset(batch), maxBatchLimit)
		if err != nil {
			log.Errorc(c, "s.college.GetCollegeMidByBatch: error(%v)", err)
			break
		}
		midList = append(midList, midBatchList...)
		if len(midBatchList) < maxBatchLimit {
			break
		}
		time.Sleep(100 * time.Microsecond)
		batch++
	}
	return midList, err
}

// collegeMysqlOffset count mysql offset
func (s *Service) collegeMysqlOffset(batch int) int {
	return batch * maxBatchLimit
}

// concurrencyGetCollegeMidScore 并发获取一个学校所有mid的积分
func (s *Service) concurrencyGetCollegeMidScore(c context.Context, mids []*college.MidInfo) (map[int64]*rank.CollegeMidInfo, error) {
	var times int
	var mutex sync.Mutex
	concurrency := concurrencyGetMidScore
	times = len(mids) / concurrency
	midScoreBatch := make([]*rank.CollegeMidInfo, 0)
	for index := 0; index <= times; index++ {
		// 这个轮次的开始时
		startTime := time.Now().UnixNano() / 1e6
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
			eg.Go(func(ctx context.Context) error {
				start := i*concurrency + b
				if start >= len(mids) {
					return nil
				}
				reqMid := mids[start].MID
				score, err := s.getMidScore(c, reqMid)
				if err != nil {
					log.Errorc(c, "s.getMidScore: error(%v)", err)
					if xecode.EqualError(xecode.Deadline, err) {
						time.Sleep(time.Second)
						score, err = s.getMidScore(c, reqMid)
						if err != nil {
							return err
						}
					} else {
						return err
					}
				}
				midScore := &rank.CollegeMidInfo{
					Score: score,
					MID:   mids[start].MID,
				}
				mutex.Lock()
				midScoreBatch = append(midScoreBatch, midScore)
				mutex.Unlock()
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Errorc(c, "eg.Wait error(%v)", err)
			return nil, err
		}
		endTime := time.Now().UnixNano() / 1e6
		waitTime := s.getWaitTime(startTime, endTime)
		if waitTime > 0 {
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
		}
	}
	mapMidInfo := make(map[int64]*rank.CollegeMidInfo)
	for _, v := range midScoreBatch {
		if v != nil {
			mapMidInfo[v.MID] = v
		}
	}
	return mapMidInfo, nil
}

// getMidScore 获取用户积分
func (s *Service) getMidScore(c context.Context, mid int64) (int64, error) {
	result, err := s.actplatClient.GetFormulaResult(c, &actplatapi.GetFormulaResultReq{
		Activity: s.c.College.MidActivity,
		Formula:  s.c.College.MidFormula,
		Mid:      mid,
	})
	if result == nil && err != nil {
		log.Errorc(c, "s.actplatClient.GetFormulaResult mid(%d) err(%v)", mid, err)
		return 0, err
	}
	return result.Result, nil
}
