package rank

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	rankmdl "go-gateway/app/web-svr/activity/admin/model/rank"
)

const (
	// concurrencyRankBatch 并发保存数量
	concurrencyRankBatch = 1
	// maxSetMidRankLength 一次设置用户数量
	maxSetMidRankLength = 50
	// concurrencyRedisSetMidRank redis并发量
	concurrencyRedisSetMidRank = 2
	// maxArcBatchLikeLimit 一次从表中获取稿件数量
	maxArcBatchLikeLimit = 1000
)

// Publish 发布结果
func (s *Service) Publish(c context.Context, id int64, attributeType int, batch int64) (err error) {
	var rank *rankmdl.Rank
	var rankResult = make([]*rankmdl.MidRank, 0)
	var rankMidResult = make([]*rankmdl.MidRank, 0)
	if rank, err = s.dao.GetRankConfigByID(c, id); err != nil {
		log.Errorc(c, "Get s.dao.GetRankConfigByID() failed. error(%v)", err)
		return
	}
	if rank == nil {
		return errors.New("can not find online rank")
	}
	rankOidList, err := s.historyRankByBatch(c, id, batch, attributeType)
	if err != nil {
		return err
	}
	var snapshot []*rankmdl.Snapshot
	if rankOidList == nil {
		return errors.New("can not find result")
	}
	mids := make([]int64, 0)
	top := rank.Top
	rankTypeUpAids := make(map[int64][]*rankmdl.AidScore)
	alreadIn := make(map[int64]struct{})
	alreadInAid := make(map[int64]struct{})
	if rankOidList != nil {
		var count int64
		//  up主维度
		if rank.RankType == rankmdl.RankTypeUp {
			for _, v := range rankOidList {
				if _, ok := alreadIn[v.OID]; ok {
					continue
				}
				alreadIn[v.OID] = struct{}{}
				if v.State == rankmdl.RankStateOnline && v.Rank > 0 && v.Score > 0 {
					count++
					if count > top {
						rankMidResult = append(rankMidResult, &rankmdl.MidRank{
							Mid:   v.OID,
							Score: v.Score,
							Rank:  0,
						})
					} else {
						rankMidResult = append(rankMidResult, &rankmdl.MidRank{
							Mid:   v.OID,
							Score: v.Score,
							Rank:  count,
						})
					}
					rankResult = append(rankResult, &rankmdl.MidRank{
						Mid:   v.OID,
						Score: v.Score,
						Rank:  count,
					})
					mids = append(mids, v.OID)
				} else {
					rankMidResult = append(rankMidResult, &rankmdl.MidRank{
						Mid:   v.OID,
						Score: 0,
						Rank:  0,
					})

				}

			}
			// 根据mid获取稿件信息
			if len(mids) > 0 {
				snapshot, err = s.dao.AllSnapshotByMids(c, id, mids, batch, attributeType)
				if err != nil {
					return err
				}
				if len(snapshot) > 0 {
					for _, v := range snapshot {
						if _, ok := rankTypeUpAids[v.MID]; !ok {
							rankTypeUpAids[v.MID] = make([]*rankmdl.AidScore, 0)
						}
						if len(rankTypeUpAids[v.MID]) < s.c.Rank.ArchiveLength {
							if v.Score > 0 && v.State == rankmdl.RankStateOnline {
								rankTypeUpAids[v.MID] = append(rankTypeUpAids[v.MID], &rankmdl.AidScore{Aid: v.AID, Score: v.Score})
							}

						}
					}
				}
				for i, v := range rankResult {
					if archive, ok := rankTypeUpAids[v.Mid]; ok {
						rankResult[i].Aids = archive
					}
				}
			}
		}
		if rank.RankType == rankmdl.RankTypeArchive {

			aids := make([]int64, 0)
			mapAidScore := make(map[int64]*rankmdl.Video)
			midRankMap := make(map[int64]*rankmdl.MidRank)

			for _, v := range rankOidList {
				if _, ok := alreadIn[v.OID]; ok {
					continue
				}
				alreadIn[v.OID] = struct{}{}
				aids = append(aids, v.OID)
			}
			if len(aids) > 0 {
				snapshot, err = s.dao.AllSnapshotByAids(c, id, aids, batch, attributeType)
				if err != nil {
					return err
				}
				if snapshot != nil {
					for _, v := range snapshot {
						if v.State == rankmdl.RankStateOnline && v.Score > 0 && v.Rank > 0 {
							mapAidScore[v.AID] = &rankmdl.Video{
								Aid:   v.AID,
								Score: v.Score,
								Mid:   v.MID,
							}
							if _, ok := midRankMap[v.MID]; !ok {
								midRankMap[v.MID] = &rankmdl.MidRank{
									Mid:  v.MID,
									Aids: make([]*rankmdl.AidScore, 0),
								}
							}
						}

					}
				}
			}
			for _, v := range rankOidList {
				if _, ok := alreadInAid[v.OID]; ok {
					continue
				}
				alreadInAid[v.OID] = struct{}{}
				if v.State == rankmdl.RankStateOnline && v.Rank > 0 && v.Score > 0 {
					// 未上榜，不展示稿件id
					if count >= top {
						continue
					}
					// 上榜稿件
					if archive, ok := mapAidScore[v.OID]; ok {
						aids := make([]*rankmdl.AidScore, 0)
						if archive.Score > 0 {
							aidInfo := &rankmdl.AidScore{
								Aid:   archive.Aid,
								Score: v.Score,
							}
							aids = append(aids, aidInfo)
							midRankMap[archive.Mid].Aids = append(midRankMap[archive.Mid].Aids, aidInfo)
							rankResult = append(rankResult, &rankmdl.MidRank{
								Mid:   archive.Mid,
								Score: v.Score,
								Rank:  count,
								Aids:  aids,
							})
							count++
						}

					}

				}
			}

			for _, v := range midRankMap {
				rankMidResult = append(rankMidResult, v)
			}
		}
	}

	rankName := fmt.Sprintf("%d_%d", rank.ID, attributeType)
	// // 榜单维度的写入
	// err = s.rankMidRedisSave(c, rankName, rankMidResult)
	// if err != nil {
	// 	log.Errorc(c, "s.rankMidRedisSave error(%v)", err)
	// 	return err
	// }
	rankOldResult := make([]*rankmdl.MidRank, 0)
	var newCount int64
	for i, v := range rankResult {
		index := i
		if len(v.Aids) > 0 {
			newCount++
			if newCount > top {
				break
			}
			rankResult[index].Rank = newCount
			rankOldResult = append(rankOldResult, rankResult[index])
		}
	}
	// 用户维度的写入
	err = s.rankRedisSave(c, rankName, rankOldResult)
	if err != nil {
		log.Errorc(c, "s.rankMidRedisSave error(%v)", err)
		return err
	}

	return nil
}

// historyRankByBatch 获取历史排行
func (s *Service) historyRankByBatch(c context.Context, id int64, rankBatch int64, attributeType int) ([]*rankmdl.OidResult, error) {
	var (
		batch int
	)
	list := make([]*rankmdl.OidResult, 0)
	for {
		rankList, err := s.dao.AllOidRank(c, id, rankBatch, attributeType, s.mysqlOffset(batch), maxArcBatchLikeLimit)
		if err != nil {
			log.Errorc(c, "s.dao.LikeList: error(%v)", err)
			return nil, err
		}
		if len(rankList) > 0 {
			list = append(list, rankList...)
		}
		if len(rankList) < maxArcBatchLikeLimit {
			break
		}
		time.Sleep(100 * time.Microsecond)
		batch++
	}
	return list, nil
}

// mysqlOffset count mysql offset
func (s *Service) mysqlOffset(batch int) int {
	return batch * maxArcBatchLikeLimit
}

// rankRedisSave 设置榜单维度的缓存
func (s *Service) rankRedisSave(c context.Context, rankNameKey string, midRank []*rankmdl.MidRank) (err error) {
	return s.dao.SetRank(c, rankNameKey, midRank)
}

// rankRedisSave 批量设置用户维度缓存
func (s *Service) rankMidRedisSave(c context.Context, rankNameKey string, midRank []*rankmdl.MidRank) (err error) {
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
					err := s.dao.SetMidRank(ctx, rankNameKey, reqMids)
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
