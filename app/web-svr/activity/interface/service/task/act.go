package task

import (
	"context"
	"fmt"
	"go-common/library/sync/errgroup.v2"
	"sync"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	actmdl "go-gateway/app/web-svr/activity/interface/model/actplat"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"
	taskmdl "go-gateway/app/web-svr/activity/interface/model/task"
)

const (
	retry     = 3
	timeSleep = 100 * time.Millisecond
)

// TaskPre 任务前置操作
func (s *Service) TaskPre(ctx context.Context, mid int64, business string, activity string, risk interface{}, isInternal bool) (err error) {
	timeStamp := time.Now().Unix()
	task, err := s.actDao.GetTaskDetail(ctx, activity)
	if err != nil {
		log.Errorc(ctx, " s.actDao.GetTaskDetail err(%v)", err)
		return
	}
	if len(task) == 0 {
		return ecode.TaskCanNotFinish
	}
	for _, v := range task {
		if v.Counter == business {
			if !isInternal && v.IsFe != taskmdl.IsFeYes {
				return ecode.TaskCanNotFinish
			}
			if risk != nil {
				err = s.isRisk(ctx, v, mid, risk, timeStamp)
				if err != nil {
					return
				}
			}
			break
		}
	}
	return
}

// CardsActSend 集卡专用
func (s *Service) CardsActSend(ctx context.Context, mid int64, business string, activity string, timeStamp int64, risk *riskmdl.Base, isInternal bool) (err error) {
	// 风控
	var spRisk *riskmdl.Task
	if risk != nil {
		spRisk = &riskmdl.Task{
			Base:        *risk,
			Mid:         mid,
			ActivityUID: activity,
			Subscene:    business,
		}
	}
	err = s.TaskPre(ctx, mid, business, activity, spRisk, isInternal)
	if err != nil {
		log.Errorc(ctx, " s.TaskPre (%d,%s,%s,%+v) err(%v)", mid, business, activity, risk, err)
		return
	}
	return s.ActSend(ctx, mid, business, activity, timeStamp)
}

// ActSend
func (s *Service) ActSend(ctx context.Context, mid int64, business string, activity string, timeStamp int64) (err error) {
	t := time.Now().Unix()

	activityPoints := &actmdl.ActivityPoints{
		Timestamp: t,
		Mid:       mid,
		Source:    mid,
		Activity:  activity,
		Business:  business,
	}
	log.Infoc(ctx, "s.actDao.Send (%+v) time(%d)", activityPoints, timeStamp)
	err = s.actDao.Send(ctx, mid, activityPoints)
	if err != nil {
		log.Errorc(ctx, "s.actDao.Send end error data(%+v) err(%v) ", activityPoints, err)
		s.cache.SyncDo(ctx, func(ctx context.Context) {
			for i := 0; i < retry; i++ {
				err = s.actDao.Send(ctx, mid, activityPoints)
				if err == nil {
					return
				}
				log.Errorc(ctx, "retry info:%v", err)
				time.Sleep(timeSleep)
			}
			if err != nil {
				log.Errorc(ctx, "s.actDao.Send end error data(%+v) err(%v) ", activityPoints, err)
				content := fmt.Sprintf("timestamp:%d,mid:%d,source:%d,Activity:%s,Business:%s,err:%v", timeStamp, mid, mid, activity, business, err)
				err2 := s.wechatdao.SendWeChat(ctx, s.c.Wechat.PublicKey, "[任务推送失败 重试失败]", content, "zhangtinghua")
				if err2 != nil {
					log.Errorc(ctx, " s.wechatdao.SendWeChat(%v)", err2)
				}
			}

		})
	}
	return nil
}

// getCounter ...
func (s *Service) getCounter(c context.Context, mid int64, activity string, counter string) (count int64, err error) {
	count, err = s.actDao.GetCounter(c, mid, activity, counter)
	if err != nil {
		log.Errorc(c, "s.actDao.GetCounter counter(%s) err(%v)", counter, err)
		return
	}
	return
}

// ActSend
func (s *Service) Result(ctx context.Context, mid int64, activity string) (res *taskmdl.TaskReply, err error) {
	task, err := s.actDao.GetTaskDetail(ctx, activity)
	if err != nil {
		log.Errorc(ctx, " s.actDao.GetTaskDetail err(%v)", err)
		return
	}
	if task == nil {
		return
	}
	eg := errgroup.WithContext(ctx)
	taskList := make([]*taskmdl.TaskMember, 0)
	res = &taskmdl.TaskReply{}
	resTaskList := make([]*taskmdl.TaskDetail, 0)
	res.List = resTaskList
	var (
		mutex sync.Mutex
	)
	for _, t := range task {
		taskCounter := t.Counter
		isFe := t.IsFe

		taskFinishTimes := t.FinishTimes
		activity := t.Activity
		eg.Go(func(ctx context.Context) (err error) {
			counter, err := s.getCounter(ctx, mid, activity, taskCounter)
			if err != nil {
				log.Errorc(ctx, "s.getCounter err(%v)", err)
				return err
			}
			var state int

			if counter >= taskFinishTimes {
				state = taskmdl.StateFinish
				counter = taskFinishTimes
			}
			mutex.Lock()
			taskList = append(taskList, &taskmdl.TaskMember{
				Count:   counter,
				State:   state,
				Counter: taskCounter,
				IsFe:    isFe,
			})
			mutex.Unlock()
			return
		})

	}
	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "eg.Wait error(%v)", err)
		return res, err
	}
	taskMapList := make(map[string]*taskmdl.TaskMember)
	for _, v := range taskList {
		taskMapList[v.Counter] = v
	}
	for _, task := range task {
		link := task.Link
		var taskMember = &taskmdl.TaskMember{}
		member, ok := taskMapList[task.Counter]
		if ok {
			taskMember = member
			if taskMember.IsFe != taskmdl.IsFeYes {
				taskMember.Counter = ""
			}
		}
		taskDetail := &taskmdl.TaskDetail{
			Task: &taskmdl.SimpleTask{
				TaskName:    task.TaskName,
				LinkName:    task.LinkName,
				Desc:        task.Desc,
				Link:        link,
				FinishTimes: task.FinishTimes,
			},
			Member: taskMember,
		}
		resTaskList = append(resTaskList, taskDetail)
	}
	res.List = resTaskList
	return
}
