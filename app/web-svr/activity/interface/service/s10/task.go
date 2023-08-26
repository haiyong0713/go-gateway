package s10

import (
	"context"
	"time"

	"go-gateway/app/web-svr/activity/ecode"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/s10"

	"go-common/library/log"
	"go-common/library/stat/prom"
	"go-common/library/sync/errgroup.v2"
)

func (s *Service) taskProgress(ctx context.Context, mid int64) ([]int32, error) {
	timestamp := time.Now().Unix()
	wg := errgroup.WithContext(ctx)
	progress := make([]int32, len(s.staticConf.Tasks))
	for i, task := range s.staticConf.Tasks {
		index := i
		tmp := task.Caption
		switch tmp {
		case s10.S10ActSign:
			wg.Go(func(ctx context.Context) (err error) {
				progress[index], err = s.Signed(ctx, mid, timestamp)
				return
			})
		case s10.S10ActPred:
		default:
			wg.Go(func(ctx context.Context) (err error) {
				progress[index], err = s.dao.GetCounterRes(ctx, mid, timestamp, tmp, s.s10Act)
				return
			})
		}
	}
	err := wg.Wait()
	if err != nil {
		log.Errorc(ctx, "s10 wg.Wait() error(%v)", err)
		if progress, err = s.dao.TaskProgressCache(ctx, mid); err != nil {
			return nil, err
		}
		prom.BusinessInfoCount.Incr("s10:TaskProgressCache")
	} else {
		if timestamp&1 == 1 {
			if err = cache.Do(context.Background(), func(ctx context.Context) {
				s.dao.AddTaskProgressCache(ctx, mid, progress)
			}); err != nil {
				log.Errorc(ctx, "s10 s.cache.Do() error(%v)", err)
			}
		}
	}
	return progress, nil
}

func (s *Service) Tasks(ctx context.Context, mid int64) ([]*v1.TaskProgress, error) {
	var (
		err      error
		progress = make([]int32, len(s.staticConf.Tasks))
	)
	timestamp := time.Now().Unix()
	if err = s.s10GoodsTimePeriod(timestamp); err != nil {
		return nil, err
	}
	if mid > 0 {
		if progress, err = s.taskProgress(ctx, mid); err != nil {
			return nil, ecode.ActivityTasksProgressGetFail
		}
		if len(progress) != len(s.staticConf.Tasks) {
			log.Errorc(ctx, "s10 conflict tasks(%d) and taskProgress(%d)", len(s.staticConf.Tasks), len(progress))
			return nil, ecode.ActivityTasksProgressGetFail
		}
	}
	res := make([]*v1.TaskProgress, 0, len(s.staticConf.Tasks))
	for i, task := range s.staticConf.Tasks {
		res = append(res, &v1.TaskProgress{
			UniqID:   task.Caption,
			Status:   task.Total <= progress[i] && mid > 0 && task.Total != 0,
			Progress: &v1.TaskDetail{Completed: progress[i], MaxTimes: task.Total},
		})
	}
	return res, nil
}

func (s *Service) TaskPub(ctx context.Context, mid, timestamp int64, act string) (err error) {
	if err = s.s10PointsTimePeriod(timestamp); err != nil {
		return err
	}
	return s.dao.TaskPubDataBus(ctx, mid, timestamp, act)
}
