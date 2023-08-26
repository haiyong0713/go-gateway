package service

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/sync/errgroup.v2"
	like "go-gateway/app/web-svr/activity/job/model/like"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
	"sort"
	"strconv"

	"github.com/pkg/errors"

	handwritemdl "go-gateway/app/web-svr/activity/job/model/handwrite"
	sourcemdl "go-gateway/app/web-svr/activity/job/model/source"

	"time"
)

var handwrite2021Ctx context.Context

func handwrite2021CtxInit() {
	handwrite2021Ctx = trace.SimpleServerTrace(context.Background(), "handwrite2021")
}

const (
	// maxMidBatchTaskLimit 一次从表中获取用户数量
	maxMidBatchTaskLimit = 1000
	// midTaskChannelLength db channel 长度
	midTaskChannelLength = 2
	// concurrencyMidTask ...
	concurrencyMidTask = 1
	// midTaskRecordbatch 一次记录用户获奖情况的数量
	midTaskRecordbatch = 1000
	// concurrencyMidTaskRecord 并发记录用户获奖情况
	concurrencyMidTaskRecord = 1
	// midTaskRecordDBbatch 一次记录用户获奖情况的数量
	midTaskRecordDBbatch = 1000
	// concurrencyMidTaskRecord 并发记录用户获奖情况
	concurrencyMidTaskDBRecord = 1
	// snappathRankBatch 批量保存数量
	snappathRankBatch = 100
)

// Handwrite2021 手书2021
func (s *Service) Handwrite2021() {
	s.handwrite2021Running.Lock()
	defer s.handwrite2021Running.Unlock()

	handwrite2021CtxInit()
	ctx := handwrite2021Ctx
	err := s.doHandwrite2021(ctx)
	if err != nil {
		// 错误处理
		err = s.sendWechat(ctx, "[手书任务]", fmt.Sprintf("%v", err), "zhangtinghua")
		if err != nil {
			log.Errorc(ctx, "handwrite2021 s.sendWechat (%v)", err)
		}
	}
}

func (s *Service) doHandwrite2021(ctx context.Context) error {
	start := time.Now()
	log.Infoc(ctx, "handwrite2021 start (%d)", start.Unix())
	var (
		isSnapShot bool
	)
	if start.Unix() > s.c.Handwrite2021.EndStatisticsTime.Unix() {
		// 如果已经过了统计时间，则获取快照数据
		isSnapShot = true
	}

	archives, err := s.getLikesSourceByType(ctx, isSnapShot)
	if err != nil {
		log.Errorc(ctx, "handwrite2021 s.getLikesSourceByType err(%v)", err)
		return err
	}

	midTask, god, tired1, tired2, tired3 := s.handwirte2021Task(ctx, archives)

	allTask, dbTask, err := s.handwrite2021Task(ctx, midTask)
	if err != nil {
		log.Errorc(ctx, "s.handwrite2021Task err(%v)", err)
		return err
	}
	err = s.handwrite2021TaskSave(ctx, allTask, dbTask, god, tired1, tired2, tired3)
	// 快照数据保存
	if !isSnapShot {
		batch := s.getBatch(time.Now().Unix())
		err = s.snapshotResultDB(ctx, archives, rankmdl.GetRankAttributeType(rankmdl.RankAttributeAll), batch)
		if err != nil {
			log.Errorc(ctx, "s.snapshotResultDB err(%v)", err)
			return err
		}
	}
	end := time.Now()
	spend := end.Unix() - start.Unix()
	log.Infoc(ctx, "handwrite2021 success() spend(%d)", spend)
	return nil
}

// GetBatch 返回batch
func (s *Service) getBatch(stime int64) int {
	hour, _, _, _, day, year, month := getDay(stime)
	lastBatchStr := fmt.Sprintf("%d%02d%02d%02d", year, month, day, hour)
	lastBatch, _ := strconv.Atoi(lastBatchStr)
	return lastBatch
}

func getDay(stime int64) (int, int, int, int, int, int, int) {
	hour := time.Unix(stime, 0).Hour()
	minute := time.Unix(stime, 0).Minute()
	second := time.Unix(stime, 0).Second()
	week := int(time.Unix(stime, 0).Weekday())
	day := int(time.Unix(stime, 0).Day())
	year := time.Unix(stime, 0).Year()
	month := int(time.Unix(stime, 0).Month())
	return hour, minute, second, week, day, year, month
}

// snapshotResultDB 保存db排行结果
func (s *Service) snapshotResultDB(c context.Context, archives []*sourcemdl.Archive, attributeType int, batch int) error {
	if archives == nil {
		return nil
	}
	snapshotAll := make([]*rankmdl.Snapshot, 0)
	for _, arc := range archives {
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
			RankAttribute: attributeType,
			Score:         arc.Score,
			Batch:         batch,
			ArcCtime:      arc.Ctime,
			PubTime:       arc.PubTime,
		}
		if arc.IsNormal() {
			result.State = rankmdl.SnapshotStateNormal
		}
		snapshotAll = append(snapshotAll, result)
	}
	err := s.addSnapshotBatchDB(c, snapshotAll, batch, int(attributeType))
	if err != nil {
		log.Errorc(c, "s.addSnapshotBatchD err(%v)", err)
		return err
	}
	return nil
}

// addSnapshotBatchDB ...
func (s *Service) addSnapshotBatchDB(c context.Context, snapshot []*rankmdl.Snapshot, batch, attribute int) (err error) {
	var times int
	startTrans := time.Now().Unix()
	patch := snappathRankBatch
	concurrency := concurrencyRankBatch
	times = len(snapshot) / patch / concurrency
	tx, err := s.rankv2.BeginTran(c)
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
			err = errors.New(fmt.Sprintf("保存失败  rankID(%d) attribute (%d)", s.c.Handwrite2021.RankID, attribute))
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
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(snapshot) {
					return nil
				}
				reqMids := snapshot[start:]
				end := start + patch
				if end < len(snapshot) {
					reqMids = snapshot[start:end]
				}
				if len(reqMids) > 0 {
					snapshotstart := time.Now().Unix()
					err = s.rankv2.BatchAddSnapshotRank(c, tx, s.c.Handwrite2021.RankID, reqMids)
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

// getLikesSourceByType 获取数据源数据
func (s *Service) getLikesSourceByType(c context.Context, isSnapShot bool) (list []*sourcemdl.Archive, err error) {
	// 获取数据源
	subject, archives, err := s.rankSvr.GetSourceConfig(c, s.c.Handwrite2021.RankID, s.c.Handwrite2021.Sid, rankmdl.SIDSourceAid)
	if err != nil {
		log.Errorc(c, "handwrite2021 s.GetSourceConfig error(%v)", err)
		return
	}
	list = make([]*sourcemdl.Archive, 0)
	//  如果是快照数据
	if isSnapShot {
		// 根据稿件id获取稿件数据
		list, err = s.getFilterSnapShotArchive(c, archives, subject)
		if err != nil {
			log.Errorc(c, "handwrite2021 s.getFilterSnapShotArchive error(%v)", err)
		}
		return list, err

	}
	// 根据稿件id获取稿件数据
	list, err = s.getFilterArchive(c, archives, subject)
	if err != nil {
		log.Errorc(c, "handwrite2021 s.getFilterSnapShotArchive error(%v)", err)
	}
	return list, err

}

// getFilterSnapShotArchive 获取快照数据
func (s *Service) getFilterSnapShotArchive(c context.Context, archives []*like.Like, subject *like.ActSubject) (list []*sourcemdl.Archive, err error) {
	// 根据稿件id获取稿件数据
	archive, err := s.sourceSvr.ArchiveInfoDetailFromSnapshotFilter(c, s.c.Handwrite2021.RankID, s.c.Handwrite2021.LastBatch, rankmdl.GetRankAttributeType(rankmdl.RankAttributeAll), archives, false)
	if err != nil {
		log.Errorc(c, "handwrite2021 s.source.ArchiveInfoDetailFromSnapshotFilter error(%v)", err)
		return
	}
	archiveList, err := s.sourceSvr.FilterArchive(c, subject, archive)
	if err != nil {
		log.Errorc(c, "handwrite2021 s.source.FilterArchive error(%v)", err)
		return
	}
	return archiveList, nil
}

// getFilterArchive 获取稿件数据
func (s *Service) getFilterArchive(c context.Context, archives []*like.Like, subject *like.ActSubject) (list []*sourcemdl.Archive, err error) {
	// 根据稿件id获取稿件数据
	archive, err := s.sourceSvr.ArchiveInfoDetailFilter(c, archives, false)
	if err != nil {
		log.Errorc(c, "handwrite2021 s.source.ArchiveInfoDetailFilter error(%v)", err)
		return
	}
	archiveList, err := s.sourceSvr.FilterArchive(c, subject, archive)
	if err != nil {
		log.Errorc(c, "handwrite2021 s.source.FilterArchive error(%v)", err)
		return
	}
	return archiveList, nil
}

// handwirte2021Task 用户任务完成情况
func (s *Service) handwirte2021Task(c context.Context, list []*sourcemdl.Archive) (map[int64]map[int64]*handwritemdl.MidTask, int64, int64, int64, int64) {
	// 分组
	archiveGroup := s.rankSvr.Group(c, rankmdl.RankTypeUp, list)
	midArchiveGroup := make(map[int64]*sourcemdl.ArchiveGroup)
	if archiveGroup != nil {
		for mid, v := range archiveGroup {
			sort.Sort(v.Archive)
			midArchiveGroup[mid] = v
		}
	}
	midTaskResult := make(map[int64]map[int64]*handwritemdl.MidTask)
	var (
		godCount, tired1Count, tired2Count, tired3Count int
	)
	// 统计任务情况
	for _, v := range midArchiveGroup {
		if v.Archive != nil {
			archiveNew := make(sourcemdl.ArchiveBatch, 0)
			for _, arc := range v.Archive {
				if arc.IsNormal() && int64(arc.PubTime) < s.c.Handwrite2021.LastPubTime.Unix() {
					archiveNew = append(archiveNew, arc)
				}
			}
			godTask := s.hanwriteGodTask(c, v.MID, archiveNew)
			tiredTask1, tiredTask2, tiredTask3 := s.hanwriteTiredTask(c, v.MID, archiveNew)
			godCount += godTask.FinishCount
			tired1Count += tiredTask1.FinishCount
			tired2Count += tiredTask2.FinishCount
			tired3Count += tiredTask3.FinishCount
			midTaskResult[v.MID] = make(map[int64]*handwritemdl.MidTask)
			midTaskResult[v.MID][handwritemdl.TaskTypeGod] = godTask
			midTaskResult[v.MID][handwritemdl.TaskTypeTiredLevel1] = tiredTask1
			midTaskResult[v.MID][handwritemdl.TaskTypeTiredLevel2] = tiredTask2
			midTaskResult[v.MID][handwritemdl.TaskTypeTiredLevel3] = tiredTask3
		}
	}
	return midTaskResult, int64(godCount), int64(tired1Count), int64(tired2Count), int64(tired3Count)
}

// handwrite2021TaskSave ...
func (s *Service) handwrite2021TaskSave(c context.Context, midTaskAll []*handwritemdl.MidTaskAll, midTaskDB []*handwritemdl.MidTaskDB, godCount, tired1Count, tired2Count, tired3Count int64) error {
	eg := errgroup.WithContext(c)
	// redis中保存
	eg.Go(func(ctx context.Context) error {
		err := s.handwrite2021TaskSaveRedis(c, midTaskAll)
		if err != nil {
			log.Errorc(c, "s.handwrite2021TaskSaveRedis err(%v)", err)
		}
		return err
	})
	// db 中保存
	eg.Go(func(ctx context.Context) error {
		err := s.handwrite2021TaskSaveDB(c, midTaskDB)
		if err != nil {
			log.Errorc(c, "s.handwrite2021TaskSaveDB err(%v)", err)
		}
		return err
	})
	// 总数统计
	eg.Go(func(ctx context.Context) error {
		award := &handwritemdl.AwardCountNew{
			God:         godCount,
			TiredLevel1: tired1Count,
			TiredLevel2: tired2Count,
			TiredLevel3: tired3Count,
		}
		err := s.handWrite.SetTaskCount(c, award)
		if err != nil {
			log.Errorc(c, "s.SetTaskCount err(%v)", err)
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("handwrite2021TaskSave eg.Wait error(%v)", err)
		return err
	}
	return nil
}

// awardResult 获奖结果
func (s *Service) handwrite2021TaskSaveDB(c context.Context, midTaskAll []*handwritemdl.MidTaskDB) error {
	var times int
	patch := midTaskRecordDBbatch
	concurrency := concurrencyMidTaskDBRecord
	times = len(midTaskAll) / patch / concurrency
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(midTaskAll) {
					return nil
				}
				reqMids := midTaskAll[start:]
				end := start + patch
				if end < len(midTaskAll) {
					reqMids = midTaskAll[start:end]
				}
				if len(reqMids) > 0 {
					err := s.handWrite.BatchAddTask(c, reqMids)
					if err != nil {
						err = errors.Wrapf(err, "s.handWrite.BatchAddTask")
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

// awardResult 获奖结果
func (s *Service) handwrite2021TaskSaveRedis(c context.Context, midTaskAll []*handwritemdl.MidTaskAll) error {
	var times int
	patch := midTaskRecordbatch
	concurrency := concurrencyMidTaskRecord
	times = len(midTaskAll) / patch / concurrency
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			b := batch
			i := index
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(midTaskAll) {
					return nil
				}
				reqMids := midTaskAll[start:]
				end := start + patch
				if end < len(midTaskAll) {
					reqMids = midTaskAll[start:end]
				}
				if len(reqMids) > 0 {
					mapMidTaskAll := make(map[int64]*handwritemdl.MidTaskAll)
					for _, v := range reqMids {
						mapMidTaskAll[v.Mid] = v
					}
					err := s.handWrite.AddMidTask(c, mapMidTaskAll)
					if err != nil {
						err = errors.Wrapf(err, "s.handWrite.AddMidTask")
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

// handwrite2021Task ...
func (s *Service) handwrite2021Task(c context.Context, midTask map[int64]map[int64]*handwritemdl.MidTask) ([]*handwritemdl.MidTaskAll, []*handwritemdl.MidTaskDB, error) {
	mids := make([]int64, 0)
	for mid := range midTask {
		mids = append(mids, mid)
	}
	godTaskOld, err := s.handwrite2021MidTaskGet(c, mids, handwritemdl.TaskTypeGod)
	if err != nil {
		log.Errorc(c, "handwrite2021 s.handwrite2021MidTaskGet godTaskOld (%v)", err)
		return nil, nil, err
	}
	tired1TaskOld, err := s.handwrite2021MidTaskGet(c, mids, handwritemdl.TaskTypeTiredLevel1)
	if err != nil {
		log.Errorc(c, "handwrite2021 s.handwrite2021MidTaskGet  tired1TaskOld (%v)", err)
		return nil, nil, err
	}
	tired2TaskOld, err := s.handwrite2021MidTaskGet(c, mids, handwritemdl.TaskTypeTiredLevel2)
	if err != nil {
		log.Errorc(c, "handwrite2021 s.handwrite2021MidTaskGet  tired2TaskOld (%v)", err)
		return nil, nil, err
	}
	tired3TaskOld, err := s.handwrite2021MidTaskGet(c, mids, handwritemdl.TaskTypeTiredLevel3)
	if err != nil {
		log.Errorc(c, "handwrite2021 s.handwrite2021MidTaskGet  tired3TaskOld (%v)", err)
		return nil, nil, err
	}
	newTask := make([]*handwritemdl.MidTaskDB, 0)
	newTaskAll := make([]*handwritemdl.MidTaskAll, 0)
	now := time.Now().Unix()
	for mid, v := range midTask {
		if v != nil {
			midNewTaskAll := &handwritemdl.MidTaskAll{Mid: mid}
			if godTask, ok := v[handwritemdl.TaskTypeGod]; ok {
				if godOldTask, ok1 := godTaskOld[mid]; ok1 {
					newTaskItem := s.oldNewTaskUpdate(c, mid, now, handwritemdl.TaskTypeGod, godOldTask, godTask)
					if newTaskItem.Mid > 0 {
						newTask = append(newTask, newTaskItem)
					}
				} else {
					newTaskItem := s.oldNewTaskUpdate(c, mid, now, handwritemdl.TaskTypeGod, nil, godTask)
					newTask = append(newTask, newTaskItem)
				}
				midNewTaskAll.God = godTask.FinishCount
			}
			if tiredTask1, ok := v[handwritemdl.TaskTypeTiredLevel1]; ok {
				if tiredOldTask, ok1 := tired1TaskOld[mid]; ok1 {
					newTaskItem := s.oldNewTaskUpdate(c, mid, now, handwritemdl.TaskTypeTiredLevel1, tiredOldTask, tiredTask1)
					if newTaskItem.Mid > 0 {
						newTask = append(newTask, newTaskItem)
					}
				} else {
					newTaskItem := s.oldNewTaskUpdate(c, mid, now, handwritemdl.TaskTypeTiredLevel1, nil, tiredTask1)
					newTask = append(newTask, newTaskItem)
				}
				midNewTaskAll.TiredLevel1 = tiredTask1.FinishCount

			}
			if tiredTask2, ok := v[handwritemdl.TaskTypeTiredLevel2]; ok {
				if tiredOldTask, ok1 := tired2TaskOld[mid]; ok1 {
					newTaskItem := s.oldNewTaskUpdate(c, mid, now, handwritemdl.TaskTypeTiredLevel2, tiredOldTask, tiredTask2)
					if newTaskItem.Mid > 0 {
						newTask = append(newTask, newTaskItem)
					}
				} else {
					newTaskItem := s.oldNewTaskUpdate(c, mid, now, handwritemdl.TaskTypeTiredLevel2, nil, tiredTask2)
					newTask = append(newTask, newTaskItem)
				}
				midNewTaskAll.TiredLevel2 = tiredTask2.FinishCount

			}
			if tiredTask3, ok := v[handwritemdl.TaskTypeTiredLevel3]; ok {
				if tiredOldTask, ok1 := tired3TaskOld[mid]; ok1 {
					newTaskItem := s.oldNewTaskUpdate(c, mid, now, handwritemdl.TaskTypeTiredLevel3, tiredOldTask, tiredTask3)
					if newTaskItem.Mid > 0 {
						newTask = append(newTask, newTaskItem)
					}
				} else {
					newTaskItem := s.oldNewTaskUpdate(c, mid, now, handwritemdl.TaskTypeTiredLevel3, nil, tiredTask3)
					newTask = append(newTask, newTaskItem)
				}
				midNewTaskAll.TiredLevel3 = tiredTask3.FinishCount
			}
			newTaskAll = append(newTaskAll, midNewTaskAll)

		}
	}
	return newTaskAll, newTask, nil

}

func (s *Service) oldNewTaskUpdate(c context.Context, mid, now int64, taskType int, old *handwritemdl.MidTaskDB, new *handwritemdl.MidTask) *handwritemdl.MidTaskDB {
	var newTaskItem *handwritemdl.MidTaskDB
	var finishTime int64
	if old != nil {
		if new.FinishCount != old.FinishCount {
			finishTime = now
		} else {
			finishTime = old.FinishTime
		}
	} else {
		finishTime = now
	}
	if new.FinishCount == 0 {
		finishTime = 0
	}
	newTaskItem = &handwritemdl.MidTaskDB{
		Mid:              mid,
		TaskType:         taskType,
		FinishCount:      new.FinishCount,
		FinishTime:       finishTime,
		TaskDetailStruct: new.TaskDetail,
	}

	return newTaskItem
}

// hanwriteGodTask 神仙模式
func (s *Service) hanwriteGodTask(c context.Context, mid int64, archiveGroup []*sourcemdl.Archive) (midTask *handwritemdl.MidTask) {
	midTask = &handwritemdl.MidTask{}
	midTask.TaskType = handwritemdl.TaskTypeGod
	midTask.Mid = mid
	midTask.TaskDetail = make([]int64, 0)
	for _, v := range archiveGroup {
		if v.Coin >= s.c.Handwrite2021.Coin {
			midTask.FinishCount++
			midTask.TaskDetail = append(midTask.TaskDetail, v.Aid)
		}
	}
	return midTask
}

// hanwriteTiredTask tired模式
func (s *Service) hanwriteTiredTask(c context.Context, mid int64, archiveGroup []*sourcemdl.Archive) (*handwritemdl.MidTask, *handwritemdl.MidTask, *handwritemdl.MidTask) {
	midTask1 := &handwritemdl.MidTask{}
	midTask1.TaskType = handwritemdl.TaskTypeTiredLevel1
	midTask1.Mid = mid
	midTask1.TaskDetail = make([]int64, 0)

	if len(archiveGroup) >= 1 {
		if archiveGroup[0].View > s.c.Handwrite2021.View1 {
			midTask1.FinishCount = 1
			midTask1.TaskDetail = append(midTask1.TaskDetail, archiveGroup[0].Aid)
		}
	}
	midTask2 := &handwritemdl.MidTask{}
	midTask2.TaskType = handwritemdl.TaskTypeTiredLevel2
	midTask2.Mid = mid
	midTask2.TaskDetail = make([]int64, 0)
	if len(archiveGroup) >= 2 {
		if archiveGroup[0].View+archiveGroup[1].View > s.c.Handwrite2021.View2 {
			midTask2.FinishCount = 1
			midTask2.TaskDetail = append(midTask2.TaskDetail, archiveGroup[0].Aid, archiveGroup[1].Aid)
		}
	}

	midTask3 := &handwritemdl.MidTask{}
	midTask3.TaskType = handwritemdl.TaskTypeTiredLevel1
	midTask3.Mid = mid
	midTask3.TaskDetail = make([]int64, 0)
	if len(archiveGroup) >= 3 {
		if archiveGroup[0].View+archiveGroup[1].View+archiveGroup[2].View > s.c.Handwrite2021.View3 {
			midTask3.FinishCount = 1
			midTask3.TaskDetail = append(midTask3.TaskDetail, archiveGroup[0].Aid, archiveGroup[1].Aid, archiveGroup[2].Aid)
		}
	}
	return midTask1, midTask2, midTask3
}

// handwrite2021MidTaskGet 用户任务完成情况
func (s *Service) handwrite2021MidTaskGet(c context.Context, mids []int64, taskType int) (map[int64]*handwritemdl.MidTaskDB, error) {
	eg := errgroup.WithContext(c)
	midTaskInfo := make(map[int64]*handwritemdl.MidTaskDB)
	channel := make(chan []*handwritemdl.MidTaskDB, midTaskChannelLength)
	eg.Go(func(ctx context.Context) error {
		return s.handwrite2021MidTaskGetChannel(c, mids, taskType, channel)
	})
	eg.Go(func(ctx context.Context) (err error) {
		midTaskInfo, err = s.handwrite2021OutChannel(c, channel)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "handwrite2021 s.handwrite2021MidTaskGet")
		return nil, err
	}
	return midTaskInfo, nil
}

func (s *Service) handwrite2021MidTaskGetChannel(c context.Context, mids []int64, taskType int, channel chan []*handwritemdl.MidTaskDB) error {
	var times int
	patch := maxMidBatchTaskLimit
	concurrency := concurrencyMidTask
	times = len(mids) / patch / concurrency
	defer close(channel)
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
					reply, err := s.handWrite.MidTask(c, reqMids, taskType)
					if err != nil {
						log.Error("s.arcClient.Arcs: error(%v)", err)
						return err
					}
					channel <- reply
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

func (s *Service) handwrite2021OutChannel(c context.Context, channel chan []*handwritemdl.MidTaskDB) (res map[int64]*handwritemdl.MidTaskDB, err error) {
	midTaskInfo := make(map[int64]*handwritemdl.MidTaskDB)
	for v := range channel {
		for _, task := range v {
			midTaskInfo[task.Mid] = task
		}
	}
	return midTaskInfo, nil
}
