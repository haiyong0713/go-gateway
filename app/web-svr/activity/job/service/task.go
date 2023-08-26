package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	l "go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/job/model/match"
	rankMdl "go-gateway/app/web-svr/activity/job/model/rank"
	t "go-gateway/app/web-svr/activity/job/model/task"

	"github.com/pkg/errors"
)

const (
	midLimit = 500
)

// upSubject update act_subject cache .
func (s *Service) upDoTask(c context.Context, arch *l.Archive) (err error) {
	if err = s.dao.DoTask(c, s.c.Image.TaskArchiveID, arch.Mid); err != nil {
		log.Error(" s.dao.DoTask(%d,%d) error(%+v)", arch.MissionID, arch.Mid, err)
		return
	}
	log.Info("upDotask success s.dao.Dotask(%d,%d)", arch.MissionID, arch.Mid)
	return
}

func (s *Service) singleDoTask(c context.Context, mid, taskID int64) (err error) {
	if err = s.dao.DoTask(c, taskID, mid); err != nil {
		log.Error("s.dao.DoTask(%d,%d) error(%v)", taskID, mid, err)
		return
	}
	log.Info("singleDoTask success s.dao.DoTask(%d,%d)", taskID, mid)
	return
}

func (s *Service) upTaskStat(ctx context.Context, msg *match.Message) {
	if msg.Action != match.ActInsert {
		return
	}
	v := &l.TaskUserLog{}
	if err := json.Unmarshal(msg.New, v); err != nil {
		log.Errorc(ctx, "upTaskStat json.Unmarshal() msg.New:%s,error(%v)", string(msg.New), err)
		return
	}
	affected, err := s.dao.TaskStateIncr(ctx, v.ForeignID, v.BusinessID, v.TaskID)
	if err != nil {
		log.Errorc(ctx, "upTaskStat TaskStateIncr task:%v error:%v", v, err)
		return
	}
	if affected == 0 {
		if err = s.dao.TaskStateAdd(ctx, v.ForeignID, v.BusinessID, v.TaskID, 1); err != nil {
			log.Errorc(ctx, "upTaskStat TaskStateAdd task:%v error:%v", v, err)
			return
		}
	}
	if err = s.dao.TaskStateCacheIncr(ctx, v.ForeignID, v.BusinessID, v.TaskID); err != nil {
		log.Errorc(ctx, "upTaskStat TaskStateCacheIncr task:%v error:%v", v, err)
		return
	}
}

// getChildTask
func (s *Service) getChildTaskByTaskID(c context.Context, taskID int64) (*t.Task, []*t.Rule, []int64, error) {
	task, err := s.getTaskInfoByTaskID(c, taskID)

	if err != nil {
		log.Errorc(c, "s.getTaskInfoByTaskID(%d) error(%v)", taskID, err)
		time.Sleep(time.Second)
		task, err = s.getTaskInfoByTaskID(c, taskID)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	var (
		childTask []*t.Rule
		likeIDs   []int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		childTask, err = s.getChildTask(c, task.ID)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		likeIDs, err = s.getLikesSID(c, task)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, nil, nil, err
	}
	return task, childTask, likeIDs, nil
}

// getActivitySidsChildTask
func (s *Service) getActivitySidsChildTask(c context.Context, sid int64) (*t.Task, []*t.Rule, []int64, error) {
	task, err := s.getTaskInfoByForeignID(c, sid)
	if err != nil {
		log.Error("s.getTaskInfoByForeignID(%d) error(%v)", sid, err)
		return nil, nil, nil, err
	}
	var (
		childTask []*t.Rule
		likeIDs   []int64
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		childTask, err = s.getChildTask(c, task.ID)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		likeIDs, err = s.getLikesSID(c, task)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, nil, nil, err
	}
	return task, childTask, likeIDs, nil
}

// activityTaskRedis 任务结果缓存redis
func (s *Service) activityTaskRedis(c context.Context, taskID int64, midsRule map[int64][]*t.MidRule, count int64) (err error) {
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		return s.dao.SetTaskCount(c, taskID, count)
	})
	eg.Go(func(ctx context.Context) (err error) {
		return s.activityTaskMidStatuBatch(c, taskID, midsRule)
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return err
	}
	return nil
}

func (s *Service) activityTaskMidStatuBatch(c context.Context, taskID int64, midsRule map[int64][]*t.MidRule) error {
	var times int
	patch := 100
	concurrency := 1
	mids := make([]*t.MidRuleBatch, 0)
	for mid, v := range midsRule {
		mids = append(mids, &t.MidRuleBatch{Mid: mid, MidRule: v})
	}
	times = len(mids) / patch / concurrency
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
					reqMidRule := make(map[int64][]*t.MidRule)
					for _, v := range reqMids {
						reqMidRule[v.Mid] = v.MidRule
					}
					err := s.dao.ActivityTaskMidStatus(c, taskID, reqMidRule)
					if err != nil {
						log.Errorc(c, " s.dao.ActivityTaskMidStatus: error(%v)", err)
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

// getActivityTaskCount 获取活动任务完成人数
func (s *Service) getActivityTaskCount(c context.Context, taskID int64) (count int64, err error) {
	return s.dao.GetTaskCount(c, taskID)
}

// activityTaskDB db
func (s *Service) activityTaskDB(c context.Context, sid, taskID int64, midsRule map[int64][]*t.MidRule) error {
	finishMidsRule := s.getFinishAllMidTask(c, midsRule)
	var usersState = make([]*t.UserState, 0)
	for mid, v := range finishMidsRule {
		userState := &t.UserState{
			MID:        mid,
			BusinessID: t.BusinessAct,
			ForeignID:  sid,
			TaskID:     taskID,
			Count:      len(v),
			Finish:     t.HasFinish,
		}
		usersState = append(usersState, userState)
	}
	if len(usersState) > 0 {
		return s.taskUserStateUpBatch(c, sid, usersState)
	}
	return nil
}

func (s *Service) taskUserStateUpBatch(c context.Context, sid int64, mids []*t.UserState) error {
	var times int
	patch := 100
	concurrency := 1

	times = len(mids) / patch / concurrency
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
					err := s.dao.TaskUserStateUp(c, sid, reqMids)
					if err != nil {
						log.Errorc(c, " s.dao.ActivityTaskMidStatus: error(%v)", err)
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

// getAllArchive 获得所有子活动的稿件，所有mid下的所有稿件+赛道稿件
func (s *Service) getAllArchive(c context.Context, likeIDs []int64) (rankMdl.ArchiveStatMap, map[int64]rankMdl.ArchiveStatMap, error) {
	var midsArchive = rankMdl.ArchiveStatMap{}
	var mapMidsArchive = make(map[int64]rankMdl.ArchiveStatMap)
	allAidMap := make(map[int64]struct{})
	for _, sid := range likeIDs {
		sidMidArchive, err := s.midArchiveInfo(c, sid)
		if err != nil {
			log.Errorc(c, "s.midArchiveInfo(%v) error(%v)", sid, err)
			return nil, nil, err
		}
		if sidMidArchive != nil {
			newSidMidArchive := make(rankMdl.ArchiveStatMap)
			for mid, v := range *sidMidArchive {
				newSidMidArchive[mid] = v
				if v != nil && len(v) > 0 {
					aids := make([]*rankMdl.ArchiveStat, 0)
					for _, archive := range v {
						if _, ok := allAidMap[archive.Aid]; ok {
							continue
						}
						aids = append(aids, archive)
						allAidMap[archive.Aid] = struct{}{}
					}
					if _, ok := midsArchive[mid]; ok {
						midsArchive[mid] = append(midsArchive[mid], aids...)
						continue
					}
					midsArchive[mid] = aids
				}
			}
			mapMidsArchive[sid] = newSidMidArchive
		}
	}
	return midsArchive, mapMidsArchive, nil
}

// getAllMids 获得所有mid
func (s *Service) getAllMids(c context.Context, likeIDs []int64) ([]int64, error) {
	mids := make([]int64, 0)
	var offset int
	for {
		itemsList, err := s.dao.AllDistinctMidBySids(c, likeIDs, offset, midLimit)
		if err != nil {
			log.Errorc(c, "s.college.GetAllCollege error(%v)", err)
			return nil, err
		}
		if len(itemsList) > 0 {
			for _, v := range itemsList {
				mids = append(mids, v.Mid)
			}
		}
		if len(itemsList) < midLimit {
			break
		}
		offset += midLimit
	}
	return mids, nil
}

// getTaskInfoByForeignID 根据sid 获得task info
func (s *Service) getTaskInfoByForeignID(c context.Context, sid int64) (*t.Task, error) {
	task, err := s.dao.GetTaskByForeignID(c, sid, t.BusinessAct)
	if err != nil {
		log.Errorc(c, "s.dao.GetTaskByForeignID(%d) error(%v)", sid, err)
		return nil, err
	}
	return task, nil
}

// getTaskInfoByTaskID 根据sid 获得task info
func (s *Service) getTaskInfoByTaskID(c context.Context, taskID int64) (*t.Task, error) {
	task, err := s.dao.GetTaskByTaskID(c, taskID)
	if err != nil {
		log.Errorc(c, "s.dao.GetTaskByTaskID(%d) error(%v)", taskID, err)
		return nil, err
	}
	return task, nil
}

// getChildTask 获取子任务
func (s *Service) getChildTask(c context.Context, taskID int64) (taskChild []*t.Rule, err error) {
	taskChild, err = s.dao.GetChildTask(c, taskID)
	if err != nil {
		log.Error("s.dao.GetChildTask(%d) error(%v)", taskID, err)
	}
	return
}

// countMidFinishTask 统计完成任务的人数
func (s *Service) countMidFinishTask(c context.Context, midsRule map[int64][]*t.MidRule) int64 {
	var count int64
	for _, midRule := range midsRule {
		var allFinish = t.IsNotFinish
		if midRule != nil {
			for _, v := range midRule {
				allFinish += v.State
			}
			if allFinish == len(midRule) {
				count++
			}
		}
	}
	return count
}

// getFinishAllMidTask 获得所有完成任务的
func (s *Service) getFinishAllMidTask(c context.Context, midsRule map[int64][]*t.MidRule) map[int64][]*t.MidRule {
	res := make(map[int64][]*t.MidRule)
	if midsRule == nil {
		return res
	}
	for mid, midRule := range midsRule {
		var allFinish = t.IsNotFinish
		if midRule != nil {
			for _, v := range midRule {
				allFinish += v.State
			}
			if allFinish == len(midRule) {
				res[mid] = midRule
			}
		}
	}
	return res
}

// checkAllMidTask 校验所有mid任务完成情况
func (s *Service) checkAllMidTask(c context.Context, childTask []*t.Rule, mids []int64, midArchive rankMdl.ArchiveStatMap) map[int64][]*t.MidRule {
	midTask := make(map[int64][]*t.MidRule, 0)
	for _, mid := range mids {
		if v, ok := midArchive[mid]; ok {
			midTask[mid] = s.checkMidTask(c, mid, childTask, v)
			continue
		}
		midTask[mid] = s.checkMidTask(c, mid, childTask, nil)
	}
	return midTask
}

// checkMidTask 验证任务是否完成
func (s *Service) checkMidTask(c context.Context, mid int64, childTask []*t.Rule, midArchive []*rankMdl.ArchiveStat) []*t.MidRule {
	res := make([]*t.MidRule, 0)
	for _, v := range childTask {
		count, finish := v.ChildTaskFunc(midArchive)
		var state int
		if finish {
			state = t.IsFinish
		}
		midRule := &t.MidRule{
			Object: v.Object,
			MID:    mid,
			State:  state,
			Count:  count,
		}
		res = append(res, midRule)
	}
	return res
}

// getLikesSID 获得关联likes sid
func (s *Service) getLikesSID(c context.Context, task *t.Task) ([]int64, error) {
	res := make([]int64, 0)
	if task.IsMultiSource() {
		subjectChild, err := s.dao.SubjectChild(c, task.ForeignID)
		if err != nil {
			err = errors.Wrapf(err, "s.getLikesSID")
			log.Error("s.getLikesSID task:%v error:%v", task, err)
		}
		if subjectChild != nil {
			res = subjectChild.ChildIdsList
		}
		return res, nil
	}
	res = append(res, task.ForeignID)
	return res, nil
}

// lastTaskResultToDb 最后结果统计db
func (s *Service) lastTaskResultToDb(c context.Context, sid, taskID int64, midsRule map[int64][]*t.MidRule) (err error) {
	err = s.activityTaskDB(c, sid, taskID, midsRule)
	if err != nil {
		return err
	}
	return nil
}

// taskUserState 获取用户任务状态
func (s *Service) taskUserState(c context.Context, foreignID, taskID int64, mids []int64) (map[int64]*t.UserState, error) {
	res := make(map[int64]*t.UserState)
	midState, err := s.dao.GetUserTaskState(c, taskID, foreignID, mids)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.GetUserTaskState")
		log.Error("s.dao.GetUserTaskState taskID:%d error:%v", taskID, err)
		return nil, err
	}
	if midState != nil && len(midState) > 0 {
		for _, v := range midState {
			if v.Finish == t.HasFinish {
				res[v.MID] = v
			}
		}
	}
	return res, nil
}

// taskUserStateByRedis 获取用户任务完成状态
func (s *Service) taskUserStateByRedis(c context.Context, taskID int64, mids []int64) (map[int64]*t.UserState, error) {
	res := make(map[int64]*t.UserState)
	reply, err := s.dao.GetActivityTaskMidStatus(c, taskID, mids)
	if err != nil {
		err = errors.Wrapf(err, "s.dao.GetActivityTaskMidStatus")
		log.Error("s.dao.GetActivityTaskMidStatus taskID:%d, error:%v", taskID, err)
		return nil, err
	}
	finish := s.getFinishAllMidTask(c, reply)
	if finish != nil && len(finish) > 0 {
		for mid := range finish {
			res[mid] = &t.UserState{
				MID:    mid,
				TaskID: taskID,
				Finish: t.HasFinish,
			}
		}
	}
	return res, nil
}
