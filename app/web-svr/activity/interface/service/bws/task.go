package bws

import (
	"context"
	"math/rand"
	"sort"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) userTask(ctx context.Context, bid int64, userToken string) ([]*bwsmdl.UserTask, error) {
	todayInt, todayStr := todayDate()
	userTasks, err := s.dao.UserTasks(ctx, userToken, todayInt)
	if err != nil {
		return nil, err
	}
	if len(userTasks) == 0 {
		if err = func() error {
			// 无任务,领任务
			var cateMainTaskIDs, cateOtherTaskID, cateCatchTaskID []int64
			for _, v := range s.bwsAllTasks {
				if v != nil {
					switch v.Cate {
					case bwsmdl.TaskCateMain:
						cateMainTaskIDs = append(cateMainTaskIDs, v.ID)
					case bwsmdl.TaskCateOther:
						cateOtherTaskID = append(cateOtherTaskID, v.ID)
					case bwsmdl.TaskCateCatch:
						cateCatchTaskID = append(cateCatchTaskID, v.ID)
					default:
						continue
					}
				}
			}
			if len(cateMainTaskIDs) == 0 {
				return ecode.BwsNoMainTask
			}
			rand.Seed(time.Now().Unix())
			rand.Shuffle(len(cateMainTaskIDs), func(i, j int) {
				cateMainTaskIDs[i], cateMainTaskIDs[j] = cateMainTaskIDs[j], cateMainTaskIDs[i]
			})
			// 随机后取第一个task
			mainTaskID := cateMainTaskIDs[0]
			var addTasks []*bwsmdl.UserTask
			addTasks = append(addTasks, &bwsmdl.UserTask{TaskID: mainTaskID})
			for _, otherTaskID := range cateOtherTaskID {
				addTasks = append(addTasks, &bwsmdl.UserTask{TaskID: otherTaskID})
			}
			// 同步最近一天捕获任务进度
			lastUserTask, lastDay, err := s.dao.RawLastUserTasks(ctx, userToken)
			if err != nil {
				return err
			}
			if lastDay == todayInt {
				log.Warn("userTask has add task userToken:%s", userToken)
				return nil
			}
			for _, cateTaskID := range cateCatchTaskID {
				tmp := &bwsmdl.UserTask{TaskID: cateTaskID}
				for _, v := range lastUserTask {
					if v != nil && v.TaskID == cateTaskID {
						tmp.NowCount = v.NowCount
						tmp.UserState = v.UserState
						tmp.AwardState = v.AwardState
						break
					}
				}
				addTasks = append(addTasks, tmp)
			}
			isOk, err := s.dao.RequestLimit(ctx, bid, userToken, "AddUserTask", 1)
			if err == nil && !isOk {
				log.Warn("userTask token(%s) RequestLimit", userToken)
				return nil
			}
			if _, err = s.dao.AddUserTask(ctx, addTasks, todayInt, userToken); err != nil {
				return err
			}
			for _, v := range addTasks {
				userTasks = append(userTasks, v)
			}
			return nil
		}(); err != nil {
			return nil, err
		}
		s.cache.Do(ctx, func(ctx context.Context) {
			retry(func() error {
				return s.dao.DelCacheUserTasks(ctx, userToken, todayInt)
			})
		})
	}
	var pointIDs []int64
	for _, v := range userTasks {
		if v == nil || v.TaskID <= 0 {
			continue
		}
		if task, ok := s.bwsAllTasks[v.TaskID]; ok && task != nil {
			pointIDs = append(pointIDs, task.RuleIds...)
		}
	}
	var (
		userPoints []*bwsmdl.UserPointDetail
		pointMap   map[int64]*bwsmdl.Point
	)
	if len(pointIDs) > 0 {
		pointMap, err = s.dao.BwsPoints(ctx, pointIDs)
		if err != nil {
			log.Error("userTask s.dao.BwsPoints pointIDs(%v) error(%v)", pointIDs, err)
			return nil, err
		}
		userPoints, err = s.userLockPointsDay(ctx, bid, bwsmdl.ClockinType, userToken, todayStr)
		if err != nil {
			return nil, err
		}
	}
	log.Warn("userTask(%+v) userPoints:%+v pointMap:%+v userToken:%s", userTasks, userPoints, pointMap, userToken)
	var res []*bwsmdl.UserTask
	for _, v := range userTasks {
		if v == nil || v.TaskID <= 0 {
			continue
		}
		if task, ok := s.bwsAllTasks[v.TaskID]; ok && task != nil {
			tmp := &bwsmdl.UserTask{
				TaskID:      v.TaskID,
				NowCount:    v.NowCount,
				UserState:   v.UserState,
				AwardState:  v.AwardState,
				Title:       task.Title,
				FinishCount: task.FinishCount,
				OrderNum:    task.OrderNum,
			}
			// 任务完成now count修正为finish count
			if tmp.UserState == 1 {
				tmp.NowCount = tmp.FinishCount
			}
			if len(task.RuleIds) > 0 {
				for _, pointID := range task.RuleIds {
					point, ok := pointMap[pointID]
					if !ok || point == nil {
						continue
					}
					pointTmp := &bwsmdl.UserPointDetail{
						UserPoint:    &bwsmdl.UserPoint{Pid: point.ID},
						Name:         point.Name,
						Icon:         point.Icon,
						Fid:          point.Fid,
						Image:        point.Image,
						Unlocked:     point.Unlocked,
						LoseUnlocked: point.LoseUnlocked,
						LockType:     point.LockType,
						Dic:          point.Dic,
						Rule:         point.Rule,
						Bid:          point.Bid,
					}
					for _, userPoint := range userPoints {
						if userPoint == nil || userPoint.UserPoint == nil {
							continue
						}
						if userPoint.UserPoint.Pid == pointID {
							pointTmp = userPoint
						}
					}
					tmp.Points = append(tmp.Points, pointTmp)
				}
			}
			if len(tmp.Points) == 0 {
				tmp.Points = []*bwsmdl.UserPointDetail{}
			}
			res = append(res, tmp)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].OrderNum < res[j].OrderNum
	})
	return res, nil
}

func (s *Service) doPointTask(ctx context.Context, userToken string, cate string, pointID, day int64, count int) error {
	userTasks, err := s.dao.UserTasks(ctx, userToken, day)
	if err != nil {
		log.Error("doTask UserTasks userToken:%s day:%d error:%v", userToken, day, err)
		return err
	}
	tasks := make(map[int64]*bwsmdl.Task)
	for _, v := range userTasks {
		if v != nil && v.TaskID > 0 {
			if data, ok := s.bwsAllTasks[v.TaskID]; ok && data != nil {
				tasks[data.ID] = data
			}
		}
	}
	if cate == bwsmdl.TaskCateCatch {
		// 重置
		tasks = make(map[int64]*bwsmdl.Task)
		for _, v := range s.bwsAllTasks {
			if v != nil && v.Cate == bwsmdl.TaskCateCatch {
				tasks[v.ID] = v
			}
		}
	}
	doTasks := tasks
	// 非捕获任务匹配任务下point id
	if cate != bwsmdl.TaskCateCatch {
		// 重置
		doTasks = make(map[int64]*bwsmdl.Task)
		for _, v := range tasks {
			if v == nil || len(v.RuleIds) == 0 {
				continue
			}
			for _, id := range v.RuleIds {
				if id == pointID {
					doTasks[v.ID] = v
				}
			}
		}
	}
	if len(doTasks) == 0 {
		log.Warn("doTask userToken:%s pointID:%d cate:%s day:%d no match taskID", userToken, pointID, cate, day)
		return nil
	}
	eg := errgroup.WithContext(ctx)
	for _, v := range doTasks {
		tmp := v
		eg.Go(func(ctx context.Context) (err error) {
			userTaskMap := make(map[int64]*bwsmdl.UserTask, len(userTasks))
			for _, v := range userTasks {
				userTaskMap[v.TaskID] = v
			}
			if err = s.doTask(ctx, tmp.ID, day, int64(count), userToken, tmp, userTaskMap[tmp.ID]); err != nil {
				log.Error("doPointTask taskID:%d day:%d userToken:%s error:%v", tmp.ID, day, userToken, err)
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *Service) doTask(ctx context.Context, taskID, day, count int64, userToken string, task *bwsmdl.Task, userTask *bwsmdl.UserTask) error {
	if task == nil {
		log.Warn("task nil taskID:%d", taskID)
		return nil
	}
	if userTask != nil {
		if userTask.UserState == bwsmdl.TaskUserFinish {
			log.Warn("doTask userToken:%s taskID:%d day:%d task finish", userToken, taskID, day)
			return nil
		}
		totalCount := userTask.NowCount + count
		var finish int64
		if totalCount >= task.FinishCount {
			count = totalCount - userTask.NowCount + 1
			finish = 1
		}
		if _, err := s.dao.UpUserTask(ctx, userToken, taskID, day, count, finish); err != nil {
			return err
		}
		if finish == 1 {
			// s.cache.Do(ctx, func(ctx context.Context) {
			// 	taskMid, _, err := s.keyToMid(ctx, s.c.Bws.Bws2020Bid, userToken)
			// 	if err == nil && taskMid > 0 {
			// 		s.AwardTask(ctx, taskMid, taskID)
			// 	}
			// })
		}
	} else {
		if _, err := s.dao.AddUserTask(ctx, []*bwsmdl.UserTask{{TaskID: taskID}}, day, userToken); err != nil {
			return err
		}
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		retry(func() error {
			return s.dao.DelCacheUserTasks(ctx, userToken, day)
		})
	})
	return nil
}

func (s *Service) AwardTask(ctx context.Context, mid, taskID int64) error {
	bid := s.c.Bws.Bws2020Bid
	userToken, err := s.midToKey(ctx, bid, mid)
	if err != nil {
		return err
	}
	data, ok := s.bwsAllTasks[taskID]
	if !ok || data == nil {
		return xecode.NothingFound
	}
	todayInt, _ := todayDate()
	userTasks, err := s.dao.UserTasks(ctx, userToken, todayInt)
	if err != nil {
		log.Error("AwardTask UserTasks userToken:%s day:%d error:%v", userToken, todayInt, err)
		return err
	}
	var userTask *bwsmdl.UserTask
	for _, v := range userTasks {
		if v != nil && v.TaskID == taskID {
			userTask = v
			break
		}
	}
	if userTask == nil || userTask.UserState != bwsmdl.TaskUserFinish {
		return ecode.ActivityTaskNotFinish
	}
	if userTask.AwardState == bwsmdl.TaskHasAward {
		return ecode.ActivityHasAward
	}
	affected, err := s.dao.AwardUserTask(ctx, userToken, taskID, todayInt)
	if err != nil {
		log.Error("AwardTask AwardUserTask userToken:%s taskID:%d day:%d error:%v", userToken, taskID, todayInt, err)
		return err
	}
	if affected <= 0 {
		return ecode.ActivityHasAward
	}
	if _, err = s.dao.AddLotteryTimes(ctx, userToken, taskID); err != nil {
		log.Error("AwardTask AddLotteryTimes userToken:%s taskID:%d error:%v", userToken, taskID, err)
		return err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		retry(func() error {
			return s.dao.DelCacheUserTasks(ctx, userToken, todayInt)
		})
		retry(func() error {
			return s.dao.DelCacheUserLotteryTimes(ctx, userToken)
		})
	})
	return nil
}

func (s *Service) loadAllTaskIDs() {
	cateTasks, err := s.dao.RawTaskList(context.Background())
	if err != nil {
		log.Error("loadAllMainTasks RawMainTask error:%v", err)
		return
	}
	if len(cateTasks) == 0 {
		log.Error("loadAllMainTasks RawMainTask error:%v", err)
		return
	}
	s.bwsAllTasks = cateTasks
}
