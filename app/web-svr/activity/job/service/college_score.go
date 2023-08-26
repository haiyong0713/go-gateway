package service

import (
	"context"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/job/model/college"
	"time"
)

var collegeScoreCtx context.Context

const (
	// concurrencySetMidALlScore 并发设置用户总积分
	concurrencySetMidALlScore = 1
	// batchSetMidScore 批量设置用户积分
	batchSetMidScore = 100
)

func collegeScoreCtxInit() {
	collegeScoreCtx = trace.SimpleServerTrace(context.Background(), "collegeScore")
}

// CollegeScore 开学季用户积分写回
func (s *Service) CollegeScore() {
	s.collegeScoreRunning.Lock()
	defer s.collegeScoreRunning.Unlock()
	start := time.Now()
	collegeScoreCtxInit()
	log.Errorc(collegeScoreCtx, "college score start (%d)", start.Unix())
	collegeList, err := s.getAllCollege(collegeScoreCtx)
	if err != nil {
		log.Errorc(collegeScoreCtx, "s.getAllCollege(c) err(%v)", err)
		return
	}
	err = s.collegeAllMidScore(collegeScoreCtx, collegeList)
	if err != nil {
		log.Errorc(collegeScoreCtx, "s.collegeAllMidScore err(%v)", err)
		return
	}
	log.Errorc(collegeScoreCtx, "collegeScore success()")

}

func (s *Service) collegeAllMidScore(c context.Context, collegeList []*college.College) (err error) {
	for _, v := range collegeList {
		err = s.collegeMidScoreUpdate(c, v)
		if err != nil {
			log.Errorc(c, "s.collegeMidScoreUpdate error twice")
			return err
		}
	}
	return nil
}

// collegeMidScoreUpdate 校园用户计算排行，以及校园总分
func (s *Service) collegeMidScoreUpdate(c context.Context, collegeInfo *college.College) (err error) {
	mapMidInfo, err := s.getCollegeAllMidScore(c, collegeInfo.ID)
	if mapMidInfo != nil && len(mapMidInfo) > 0 {
		mids := make([]*college.Personal, 0)
		for _, v := range mapMidInfo {
			mids = append(mids, &college.Personal{MID: v.MID, Score: v.Score})
		}
		err = s.concurrencyUpdateMidScore(c, mids)
		if err != nil {
			log.Errorc(c, "s.concurrencyUpdateMidScore error(%v)", err)
			time.Sleep(time.Second)
			err = s.concurrencyUpdateMidScore(c, mids)
			if err != nil {
				log.Errorc(c, "s.concurrencyUpdateMidScore retry error(%v)", err)
			}
		}
	}
	return err

}

// concurrencyUpdateMidScore 并发更新用户积分
func (s *Service) concurrencyUpdateMidScore(c context.Context, mids []*college.Personal) error {
	var times int
	concurrency := concurrencySetMidALlScore
	patch := batchSetMidScore
	times = len(mids) / concurrency
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
			eg.Go(func(ctx context.Context) error {
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
					_, err := s.college.UpdateCollegeMidScore(ctx, reqMids)
					if err != nil {
						log.Errorc(c, "s.college.UpdateCollegeMidScore error(%v) batch(%d)", err, i)
						return err
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
