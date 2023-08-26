package service

import (
	"context"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/job/model/rank"
	"time"

	"github.com/pkg/errors"
)

const (
	// maxSetMidRankLength 一次设置用户数量
	maxSetMidRankLength = 50
	// concurrencyRedisSetMidRank redis并发量
	concurrencyRedisSetMidRank = 2
	// pathRankBatch 批量保存数量
	pathRankBatch = 500
	// concurrencyRankBatch 并发保存数量
	concurrencyRankBatch = 1
)

// addRankBatch
func (s *Service) addRankBatch(c context.Context, rankDb []*rank.DB) (err error) {
	var times int
	patch := pathRankBatch
	concurrency := concurrencyRankBatch
	times = len(rankDb) / patch / concurrency
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(rankDb) {
					return nil
				}
				reqMids := rankDb[start:]
				end := start + patch
				if end < len(rankDb) {
					reqMids = rankDb[start:end]
				}
				if len(reqMids) > 0 {
					err = s.rank.BatchAddRank(c, reqMids)
					if err != nil {
						err = errors.Wrapf(err, "s.rank.BatchAddRank")
						return err
					}
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Error("eg.Wait error(%v)", err)
			return err
		}
	}
	return nil
}

// getRankListByBatchPatch
func (s *Service) getRankListByBatchPatch(c context.Context, sid int64, batch int64) ([]*rank.DB, error) {
	var (
		err   error
		times int
	)
	rankList := make([]*rank.DB, 0)
	for {
		rankBatchList, err := s.rank.GetRankListByBatchPatch(c, sid, batch, s.dubbingMysqlOffset(times), maxBatchLimit)
		if err != nil {
			log.Errorc(c, "s.rank.GetRankListByBatchPatch: error(%v)", err)
			break
		}
		rankList = append(rankList, rankBatchList...)
		if len(rankBatchList) < maxBatchLimit {
			break
		}
		time.Sleep(100 * time.Microsecond)
		times++
	}

	return rankList, err
}

// setMidRankBatch 批量设置用户维度缓存
func (s *Service) setMidRankBatch(c context.Context, rankNameKey string, midRank []*rank.Redis) (err error) {
	var times int
	patch := maxSetMidRankLength
	concurrency := concurrencyRedisSetMidRank
	times = len(midRank) / patch / concurrency
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(midRank) {
					return nil
				}
				reqMids := midRank[start:]
				end := start + patch
				if end < len(midRank) {
					reqMids = midRank[start:end]
				}
				if len(reqMids) > 0 {
					err := s.rank.SetMidRank(ctx, rankNameKey, reqMids)
					if err != nil {
						log.Errorc(c, " s.rank.SetMidRank: error(%v) batch(%d)", err, i)
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
