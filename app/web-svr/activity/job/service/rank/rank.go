package rank

import (
	"context"
	"encoding/json"
	"fmt"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/sync/errgroup.v2"
	like "go-gateway/app/web-svr/activity/job/model/like"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
	sourcemdl "go-gateway/app/web-svr/activity/job/model/source"
	"os"
	"sort"
	"time"

	"github.com/pkg/errors"
)

const (
	// maxArcBatchLikeLimit 一次从表中获取稿件数量
	maxArcBatchLikeLimit = 1000
	// pathRankBatch 批量保存数量
	pathRankBatch = 2000
	// snappathRankBatch 批量保存数量
	snappathRankBatch = 100
	// concurrencyRankBatch 并发保存数量
	concurrencyRankBatch = 1
	// maxSetMidRankLength 一次设置用户数量
	maxSetMidRankLength = 50
	// concurrencyRedisSetMidRank redis并发量
	concurrencyRedisSetMidRank = 2
)

var rankCtx context.Context

func rankInit() {
	rankCtx = trace.SimpleServerTrace(context.Background(), "rankJobDetail")
}

// GetRankConfig 获取排行配置
func (s *Service) GetRankConfig(c context.Context, id int64) (rank *rankmdl.Rank, err error) {
	return s.rankDao.GetRankConfigByID(c, id)
}

// DoRank ...
func (s *Service) DoRank(c context.Context, rank *rankmdl.Rank, attributeType uint) (err error) {
	// 查询
	start := time.Now().Unix()
	log.Infoc(c, "do rank start start(%d) rankconfigID(%d) attributeType(%d)", start, rank.ID, attributeType)

	batch := rank.GetBatch(time.Now().Unix())
	// 获得上一次计算结果
	lastAllBatch, err := s.getLastBatch(c, rank, rankmdl.RankAttributeAll)
	if err != nil {
		log.Errorc(c, "s.getLastBatch error(%v)", err)
		return err
	}
	//  如果本次计算总榜，并且已经计算过，则结束
	if attributeType == rankmdl.RankAttributeAll && lastAllBatch == int64(batch) {
		return
	}
	var lastBatch int64
	var logAllID int64
	var logID int64

	if attributeType == rankmdl.RankAttributeAll {
		lastBatch = lastAllBatch
		logAllID, err = s.rankDao.InsertRankLog(c, rank.ID, int(batch), rank.GetRankAttributeType(rankmdl.RankAttributeAll), rankmdl.RankLogStateLearn)
		if err != nil {
			log.Errorc(c, "s.rankDao.InsertRankLog error(%v)", err)
			return err
		}
		logID = logAllID
	} else {
		lastBatch, err = s.getLastBatch(c, rank, attributeType)
		if err != nil {
			log.Errorc(c, "s.getLastBatch error(%v)", err)
			return
		}
		// 如果已经计算过，则结束
		if lastBatch == int64(batch) {
			return
		}
		if lastAllBatch != int64(batch) {
			// 检查上一次排行结果是否正常
			// 插入训练中
			logAllID, err = s.rankDao.InsertRankLog(c, rank.ID, int(batch), rank.GetRankAttributeType(rankmdl.RankAttributeAll), rankmdl.RankLogStateLearn)
			if err != nil {
				log.Errorc(c, "s.rankDao.InsertRankLog error(%v)", err)
				return err
			}
		}
		logID, err = s.rankDao.InsertRankLog(c, rank.ID, int(batch), rank.GetRankAttributeType(attributeType), rankmdl.RankLogStateLearn)
		if err != nil {
			log.Errorc(c, "s.rankDao.InsertRankLog error(%v)", err)
			return err
		}
	}

	log.Infoc(c, "rank lastAllBatch(%d) lastBatch(%d)", lastAllBatch, lastBatch)
	// 获取数据源
	subject, archives, err := s.GetSourceConfig(c, rank.ID, rank.SID, rank.SIDSource)
	if err != nil {
		log.Errorc(c, "s.GetSourceConfig error(%v)", err)
		return
	}
	// 根据稿件id获取稿件数据
	archive, err := s.getFilterArchive(c, archives, subject)
	if err != nil {
		log.Errorc(c, "s.getFilterArchive error(%v)", err)
		return
	}
	// 获取之前的结果
	historyAll, err := s.getHistoryRank(c, rank.ID, int(lastBatch), rank.GetRankAttributeType(rankmdl.RankAttributeAll))
	if err != nil {
		log.Errorc(c, "s.getHistoryRank error(%v) attribute(%d)", err, rank.GetRankAttributeType(rankmdl.RankAttributeAll))
		return
	}
	isNeedDiff := false
	if attributeType != rankmdl.RankAttributeAll {
		isNeedDiff = true
	}
	// 计算积分
	s.rankScore(c, rank, archive)
	// 分组
	archiveGroup := s.group(c, rank.RankType, archive)
	newArchiveGroup := make(map[int64]*sourcemdl.ArchiveGroup)
	for i, v := range archiveGroup {
		newArchiveGroup[i] = &sourcemdl.ArchiveGroup{
			MID:      v.MID,
			NewScore: v.NewScore,
			OID:      v.OID,
			Archive:  v.Archive,
		}
	}
	// 历史排行结果
	s.historyRankScore(c, archiveGroup, historyAll, rank.GetRankAttributeType(attributeType), isNeedDiff)
	if attributeType != rankmdl.RankAttributeAll {
		history, err := s.getHistoryRank(c, rank.ID, int(lastBatch), rank.GetRankAttributeType(attributeType))
		if err != nil {
			log.Errorc(c, "s.getHistoryRank error(%v) attribute(%d)", err, rank.GetRankAttributeType(attributeType))
			return err
		}
		s.historyRankScore(c, archiveGroup, history, rank.GetRankAttributeType(attributeType), rankmdl.NeedNotDiffScore)
	}
	rankResult := s.rankResult(c, archiveGroup)
	// 结果存储
	err = s.saveRankResult(c, rank, rankResult, rank.GetRankAttributeType(attributeType), batch, logID)
	// 如果本次不是总榜，总榜存储
	if attributeType != rankmdl.RankAttributeAll {
		s.historyRankScore(c, newArchiveGroup, historyAll, rank.GetRankAttributeType(rankmdl.RankAttributeAll), rankmdl.NeedNotDiffScore)
		newRankResult := s.rankResult(c, newArchiveGroup)
		err = s.saveRankResult(c, rank, newRankResult, rank.GetRankAttributeType(rankmdl.RankAttributeAll), batch, logAllID)
		if err != nil {
			log.Errorc(c, "s.saveRankResult error(%v)", err)
		}
	}
	if err != nil {
		return
	}
	end := time.Now().Unix()
	log.Infoc(c, "rank success start(%d) end(%d) rankconfigID(%d) attributeType(%d)", start, end, rank.ID, lastBatch)
	return
}

// Rank 计算排行结果
func (s *Service) Rank(rank *rankmdl.Rank, attributeType uint) {
	rankInit()
	c := rankCtx
	if rank == nil {
		return
	}
	if int64(rank.Etime) < time.Now().Unix() {
		return
	}
	err := s.DoRank(c, rank, attributeType)
	if err != nil {
		// 错误处理
		s.sendWechat(c, "[排行榜]", fmt.Sprintf("%v", err), "zhangtinghua")
	}

}

// getLastBatch 获取上一次完成的排行结果
func (s *Service) getLastBatch(c context.Context, rank *rankmdl.Rank, attributeType uint) (int64, error) {
	rankAttributeType := rank.GetRankAttributeType(attributeType)
	rankLog, err := s.rankDao.GetRankLogOrderByTimeAll(c, rank.ID, rankAttributeType)
	if err != nil {
		log.Errorc(c, "s.rankDao.GetRankLogOrderByTimeAll error(%v)", err)
		return 0, err
	}
	if rankLog == nil {
		return 0, nil
	}
	if rankLog.State == rankmdl.RankLogStateLearn {
		log.Errorc(c, "last still learn rankID(%d) rankAttributeType (%d)", rank.ID, rankAttributeType)
		return 0, errors.New(fmt.Sprintf("上一轮还在训练 rankID(%d) rankAttributeType (%d)", rank.ID, rankAttributeType))
	}

	return rankLog.Batch, nil
}

// addRankBatchDB
func (s *Service) addRankBatchDB(c context.Context, rank *rankmdl.Rank, rankDb []*rankmdl.OidResult, snapshotDb []*rankmdl.Snapshot, lastBatch, attribute int, logID int64) (err error) {
	startTrans := time.Now().Unix()
	tx, err := s.rankDao.BeginTran(c)
	if err != nil {
		log.Errorc(c, "begin trans err")
		return err
	}
	defer func() {
		endTrans := time.Now().Unix()
		log.Infoc(c, "trans time (%d)", endTrans-startTrans)
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "tx.Rollback()  %v", r)
			err = errors.New(fmt.Sprintf("保存失败 rankID(%d) attribute (%d)", rank.ID, attribute))
			return
		}
		if err != nil {
			if err = tx.Rollback(); err != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
	}()
	err = s.saveOidRank(c, tx, rank, rankDb)
	if err != nil {
		log.Errorc(c, "s.saveOidRank err(%v)", err)
		return err
	}

	err = s.saveSnapshotRank(c, tx, rank, snapshotDb)
	if err != nil {
		log.Errorc(c, "s.saveSnapshotRank err(%v)", err)
		return err
	}

	// 更新一次更新时间
	err = s.rankDao.UpdateRankLog(c, tx, logID, rankmdl.RankLogStateDone)
	if err != nil {
		log.Errorc(c, "s.rankDao.UpdateRankLog err(%v)", err)

		return err
	}
	return nil
}

// saveSnapshotRank 保存快照
func (s *Service) saveSnapshotRank(c context.Context, tx *xsql.Tx, rank *rankmdl.Rank, rankDb []*rankmdl.Snapshot) (err error) {
	var times int
	patch := snappathRankBatch
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
					snapshotstart := time.Now().Unix()
					err = s.rankDao.BatchAddSnapshotRank(c, tx, rank.ID, reqMids)
					snapshotend := time.Now().Unix()
					log.Infoc(c, "saveSnapshotRank time (%d)", snapshotend-snapshotstart)
					if err != nil {
						err = errors.Wrapf(err, "s.rank.BatchAddSnapshotRank")
						log.Errorc(c, "s.rank.BatchAddSnapshotRank error(%v)", err)
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

// saveOidRank 保存分组排行
func (s *Service) saveOidRank(c context.Context, tx *xsql.Tx, rank *rankmdl.Rank, rankDb []*rankmdl.OidResult) (err error) {
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
					batchAddOidRankstart := time.Now().Unix()
					err = s.rankDao.BatchAddOidRank(c, tx, rank.ID, reqMids)
					batchAddOidRankend := time.Now().Unix()
					log.Infoc(c, "batchAddOidRankend time (%d)", batchAddOidRankend-batchAddOidRankstart)
					if err != nil {
						err = errors.Wrapf(err, "s.rank.BatchAddOidRank")
						log.Errorc(c, "s.rank.BatchAddOidRank(%v)", err)
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

// saveRankResult 保存排行榜结果
func (s *Service) saveRankResult(c context.Context, rank *rankmdl.Rank, archiveGroup *sourcemdl.OidArchiveGroup, attributeType int, batch int, logID int64) error {
	var err1, err2 error
	if rank.IsAutoPublish() {
		err1 = s.rankResultRedis(c, rank, attributeType, archiveGroup)
		if err1 != nil {
			log.Errorc(c, "s.rankResultRedis error(%v)", err1)
		}
	}
	err2 = s.rankResultDB(c, rank, archiveGroup, attributeType, batch, logID)
	if err2 != nil {
		log.Errorc(c, "s.rankResultRedis error(%v)", err2)
	}
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

// rankResultRedis 保存redis排行结果
func (s *Service) rankResultRedis(c context.Context, rank *rankmdl.Rank, attributeType int, archiveGroup *sourcemdl.OidArchiveGroup) error {
	midRank, allRank := s.rankMidResult(c, rank, archiveGroup)

	rankName := fmt.Sprintf("%d_%d", rank.ID, attributeType)
	// 榜单维度的写入
	err := s.rankMidRedisSave(c, rankName, midRank)
	if err != nil {
		log.Errorc(c, "s.rankMidRedisSave error(%v)", err)
		return err
	}
	// 用户维度的写入
	err = s.rankRedisSave(c, rankName, allRank)
	if err != nil {
		log.Errorc(c, "s.rankMidRedisSave error(%v)", err)
		return err
	}
	return nil
}

// rankRedisSave 设置榜单维度的缓存
func (s *Service) rankRedisSave(c context.Context, rankNameKey string, midRank []*rankmdl.MidRank) (err error) {
	return s.rankDao.SetRank(c, rankNameKey, midRank)
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
					err := s.rankDao.SetMidRank(ctx, rankNameKey, reqMids)
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

// rankMidResult ...
func (s *Service) rankMidResult(c context.Context, rank *rankmdl.Rank, archiveGroup *sourcemdl.OidArchiveGroup) ([]*rankmdl.MidRank, []*rankmdl.MidRank) {
	// 用户维度的榜单
	midRank := make([]*rankmdl.MidRank, 0)
	// 榜单
	allRank := make([]*rankmdl.MidRank, 0)
	if archiveGroup.Data == nil {
		return nil, nil
	}
	top := rank.Top
	// 用户维度分组
	if rank.RankType == rankmdl.RankTypeUp {
		var count int64
		for _, v := range archiveGroup.Data {
			if v.Score < 0 {
				v.Score = 0
			}
			count++
			if count > top || v.Score == 0 {
				midRank = append(midRank, &rankmdl.MidRank{
					Mid:   v.OID,
					Score: v.Score,
					Rank:  0,
				})
				continue
			}
			midRank = append(midRank, &rankmdl.MidRank{
				Mid:   v.OID,
				Score: v.Score,
				Rank:  count,
			})
			if v.Archive == nil {
				continue
			}
			aids := make([]*rankmdl.AidScore, 0)
			for _, archive := range v.Archive {
				if len(aids) >= s.c.Rank.ArchiveLength {
					break
				}
				if archive.IsNormal() && archive.Score > 0 {
					aids = append(aids, &rankmdl.AidScore{Aid: archive.Aid, Score: archive.Score})
				}
			}
			allRank = append(allRank, &rankmdl.MidRank{
				Mid:   v.OID,
				Score: v.Score,
				Rank:  count,
				Aids:  aids,
			})
		}
		return midRank, allRank
	}
	// 稿件维度分组
	var count int64
	midRankMap := make(map[int64]*rankmdl.MidRank)
	for _, v := range archiveGroup.Data {
		if v.Archive == nil {
			continue
		}
		for _, archive := range v.Archive {
			if _, ok := midRankMap[archive.Mid]; !ok {
				midRankMap[archive.Mid] = &rankmdl.MidRank{Mid: archive.Mid}
				midRankMap[archive.Mid].Aids = make([]*rankmdl.AidScore, 0)
			}
		}
		count++
		// 未上榜，不展示稿件id
		if count > top {
			continue
		}
		// 上榜稿件
		var allNormal = true
		for _, archive := range v.Archive {
			if archive.IsNormal() && archive.Score > 0 {
				midRankMap[archive.Mid].Aids = append(midRankMap[archive.Mid].Aids, &rankmdl.AidScore{Aid: archive.Aid, Score: archive.Score})
				continue
			}
			allNormal = false
		}
		if allNormal {
			aids := make([]*rankmdl.AidScore, 0)
			aids = append(aids, &rankmdl.AidScore{
				Aid:   v.OID,
				Score: v.Score,
			})
			allRank = append(allRank, &rankmdl.MidRank{
				Mid:   v.MID,
				Score: v.Score,
				Rank:  count,
				Aids:  aids,
			})
		}
	}
	for _, v := range midRankMap {
		midRank = append(midRank, v)
	}
	return midRank, allRank

}

// rankResultDB 保存db排行结果
func (s *Service) rankResultDB(c context.Context, rank *rankmdl.Rank, archiveGroup *sourcemdl.OidArchiveGroup, attributeType int, batch int, logID int64) error {
	if archiveGroup == nil {
		return errors.New("archive group nil")
	}
	if logID == 0 {
		return nil
	}
	oidRank := make([]*rankmdl.OidResult, 0)
	oidRankAll := make([]*rankmdl.OidResult, 0)
	snapshotRank := make([]*rankmdl.Snapshot, 0)
	snapshotRankAll := make([]*rankmdl.Snapshot, 0)
	var r int64
	for _, v := range archiveGroup.Data {
		var hasNormal = false
		snapshot := make([]*rankmdl.Snapshot, 0)
		if v.Archive != nil {
			var rankNum int64
			for _, arc := range v.Archive {
				rankNum++
				result := &rankmdl.Snapshot{
					MID:           arc.Mid,
					AID:           arc.Aid,
					TID:           arc.TypeID,
					View:          arc.View,
					Danmaku:       arc.Danmaku,
					Reply:         arc.Reply,
					Fav:           arc.Fav,
					Coin:          arc.Coin,
					Share:         arc.Share,
					Like:          arc.Like,
					Videos:        arc.Videos,
					Rank:          rankNum,
					RankAttribute: attributeType,
					Score:         arc.Score,
					Batch:         batch,
				}
				if arc.IsNormal() {
					result.State = rankmdl.SnapshotStateNormal
					hasNormal = true
				}
				snapshot = append(snapshot, result)
			}
		}
		oidDetail := &rankmdl.OidResult{
			OID:           v.OID,
			Score:         v.Score,
			RankAttribute: int(attributeType),
			Batch:         batch,
		}
		if hasNormal {
			r++
			oidDetail.Rank = r
			oidDetail.State = rankmdl.OidStateNormal
		}
		oidRankAll = append(oidRankAll, oidDetail)
		snapshotRankAll = append(snapshotRankAll, snapshot...)

		if r > s.c.Rank.TopLength {
			continue
		}
		oidRank = append(oidRank, oidDetail)
		snapshotRank = append(snapshotRank, snapshot...)
	}

	err := s.addRankBatchDB(c, rank, oidRank, snapshotRank, batch, int(attributeType), logID)
	if err != nil {
		return err
	}

	// 文件存储全部结果
	err = s.addRankBatchFile(c, rank, oidRankAll, snapshotRankAll, batch, int(attributeType))
	if err != nil {
		log.Errorc(c, "s.addRankBatchFile (%v)", err)
	}
	return nil
}

// addRankBatchFile 文件存储
func (s *Service) addRankBatchFile(c context.Context, rank *rankmdl.Rank, rankDb []*rankmdl.OidResult, snapshotDb []*rankmdl.Snapshot, lastBatch, attribute int) (err error) {
	path := fmt.Sprintf("./data/rank/rank_id_%d", rank.ID)
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Errorc(c, "os.MkdirAll (%v)", err)
		return
	}
	err = s.snapshotFile(c, path, snapshotDb, lastBatch, attribute)
	if err != nil {
		log.Errorc(c, "os.snapshotFile (%v)", err)
		return
	}
	err = s.oidFile(c, path, rankDb, lastBatch, attribute)
	if err != nil {
		log.Errorc(c, "os.oidFile (%v)", err)
		return
	}
	return err
}

// snapshotFile
func (s *Service) snapshotFile(c context.Context, path string, snapshotDb []*rankmdl.Snapshot, lastBatch, attribute int) error {
	fileName := fmt.Sprintf("%s/lastBatch_%d_type_%d_snapshot", path, lastBatch, attribute)
	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		log.Errorc(c, "snapshotFile os.Open(%s) error(%v)", fileName, err)
		return err
	}
	for _, one := range snapshotDb {
		b, err := json.Marshal(one)
		if err != nil {
			log.Errorc(c, "snapshotFile json.Marshal(%v) error(%v)", one, err)
			return err
		}
		if _, err := f.Write(b); err != nil {
			log.Errorc(c, "snapshotFile f.Write(%s) error(%v)", b, err)
			return err
		}
		if _, err := f.WriteString("\n"); err != nil {
			log.Errorc(c, "snapshotFile f.WriteString(\\n) error(%v)", err)
			return err
		}
	}
	return nil
}

// oidFile
func (s *Service) oidFile(c context.Context, path string, rankDb []*rankmdl.OidResult, lastBatch, attribute int) error {
	fileName := fmt.Sprintf("%s/lastBatch_%d_type_%d_oidResult", path, lastBatch, attribute)
	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		log.Errorc(c, "oidFile os.Open(%s) error(%v)", fileName, err)
		return err
	}
	for _, one := range rankDb {
		b, err := json.Marshal(one)
		if err != nil {
			log.Errorc(c, "oidFile json.Marshal(%v) error(%v)", one, err)
			return err
		}
		if _, err := f.Write(b); err != nil {
			log.Errorc(c, "oidFile f.Write(%s) error(%v)", b, err)
			return err
		}
		if _, err := f.WriteString("\n"); err != nil {
			log.Errorc(c, "oidFile f.WriteString(\\n) error(%v)", err)
			return err
		}
	}
	return nil
}

// rankResult ...
func (s *Service) rankResult(c context.Context, archive map[int64]*sourcemdl.ArchiveGroup) *sourcemdl.OidArchiveGroup {
	var rankArchive = &sourcemdl.OidArchiveGroup{}
	if archive == nil {
		return nil
	}
	rankArchive.Data = make([]*sourcemdl.ArchiveGroup, 0)
	for _, v := range archive {
		sort.Sort(v.Archive)
		rankArchive.Data = append(rankArchive.Data, v)
	}
	rankArchive.TopLength = len(rankArchive.Data)
	rankmdl.Sort(rankArchive)
	return rankArchive
}

// historyRank 处理历史排名及积分
func (s *Service) historyRankScore(c context.Context, archive map[int64]*sourcemdl.ArchiveGroup, historyRank []*rankmdl.OidResult, attributeType int, needDiff bool) {
	historyRankMap := make(map[int64]*rankmdl.OidResult)
	for _, v := range historyRank {
		historyRankMap[v.OID] = v
	}
	for oid := range archive {
		if history, ok := historyRankMap[oid]; ok {
			archive[oid].HistoryRank = history.Rank
			archive[oid].HistoryScore = history.Score
			if needDiff {
				archive[oid].Score = archive[oid].NewScore - history.Score
				if archive[oid].Score < 0 {
					archive[oid].Score = 0
				}
				archive[oid].NewScore = archive[oid].Score
				continue
			}
		}
		archive[oid].Score = archive[oid].NewScore
	}
}

// getHistoryRank 获取历史排行
func (s *Service) getHistoryRank(c context.Context, id int64, lastBatch, attributeType int) ([]*rankmdl.OidResult, error) {
	res := make([]*rankmdl.OidResult, 0)
	res, err := s.historyRankByBatch(c, id, lastBatch, attributeType)
	if err != nil {
		return res, err
	}
	return res, nil
}

// historyRankByBatch 获取历史排行
func (s *Service) historyRankByBatch(c context.Context, id int64, rankBatch, attributeType int) ([]*rankmdl.OidResult, error) {
	var (
		batch int
	)
	list := make([]*rankmdl.OidResult, 0)
	for {
		rankList, err := s.rankDao.AllOidRank(c, id, rankBatch, attributeType, s.mysqlOffset(batch), maxArcBatchLikeLimit)
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

func (s *Service) rankScore(c context.Context, rank *rankmdl.Rank, archive []*sourcemdl.Archive) {
	for _, v := range archive {
		v.ScoreFunc(func(arc *sourcemdl.Archive) int64 {
			if !arc.IsNormal() {
				return 0
			}
			var score int64
			score = arc.Danmaku*rank.RatioStruct.Danmaku + arc.Reply*rank.RatioStruct.Reply + arc.Fav*rank.RatioStruct.Fav + arc.Like*rank.RatioStruct.Like + arc.Share*rank.RatioStruct.Share + arc.Coin*rank.RatioStruct.Coin + arc.View*rank.RatioStruct.View
			revise := (200000 + arc.View) / 2
			if revise > 1 {
				revise = 1
			}
			if rank.RatioStruct.Revise == 1 {
				score = score * revise
			}
			return score
		})
	}
}

func (s *Service) getFilterArchive(c context.Context, archives []*like.Like, subject *like.ActSubject) (list []*sourcemdl.Archive, err error) {
	// 根据稿件id获取稿件数据

	archive, err := s.source.ArchiveInfoDetailFilter(c, archives, true)
	if err != nil {
		log.Errorc(c, "s.source.ArchiveInfoDetailFilter error(%v)", err)
		return
	}

	archiveList, err := s.source.FilterArchive(c, subject, archive)
	if err != nil {
		log.Errorc(c, "s.source.FilterArchive error(%v)", err)
		return
	}
	return archiveList, nil
}
