package service

import (
	"context"
	"time"

	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	model "go-gateway/app/web-svr/activity/job/model/knowledge_task"
)

const _retryDBTimes = 5

func (s *Service) UserKnowledgeTaskCalculate(logDate string) {
	var (
		taskId    int64
		calcCount int
	)
	ctx := context.Background()
	if logDate == "" {
		log.Infoc(ctx, "UserKnowledgeTaskCalculate logDate is empty")
		return
	}
	nowTime := time.Now()
	key := "knowledge_lock_" + logDate
	if err := s.knowTaskDao.RsSetNX(ctx, key, 43200); err != nil {
		log.Infoc(ctx, "UserKnowledgeTaskCalculate error(%+v)", err)
		return
	}
	log.Infoc(ctx, "UserKnowledgeTaskCalculate begin %v", nowTime.Format("2006-01-02 15:04:05"))
	for {
		knowCalculates, err := s.knowledgeCalculateList(ctx, logDate, taskId, s.c.KnowledgeTask.KnowTaskBatchNum)
		if err != nil {
			// 配制告警 KnowledgeCalculateAlarm
			log.Errorc(ctx, "KnowledgeCalculateAlarm  s.knowTaskDao.RawKnowledgeCalcList offset(%d) logDate(%s) calcCount(%d) error(%+v)", taskId, logDate, calcCount, err)
			break
		}
		if len(knowCalculates) == 0 {
			log.Infoc(ctx, "UserKnowledgeTaskCalculate s.knowTaskDao.RawKnowledgeCalcList success taskID(%d)", taskId)
			break
		}
		// 更新插入数据库
		if err = s.insertUpdateUserKnowTask(ctx, knowCalculates); err != nil {
			// 配制告警 KnowledgeCalculateAlarm
			log.Errorc(ctx, "KnowledgeCalculateAlarm s.knowTaskDao.InsertUpdateUserKnowTask() offset(%d) logDate(%s) calcCount(%d) error(%+v)", taskId, logDate, calcCount, err)
			break
		}
		taskId = knowCalculates[len(knowCalculates)-1].Id
		calcCount += len(knowCalculates)
		log.Infoc(ctx, "UserKnowledgeTaskCalculate log_date(%s) taskID(%d) calc(%d)", logDate, taskId, calcCount)
	}
	log.Infoc(ctx, "UserKnowledgeTaskCalculate end %v since(%v) log_date(%s) calc(%d)", time.Now().Format("2006-01-02 15:04:05"), time.Since(nowTime).Seconds(), logDate, calcCount)
}

func (s *Service) insertUpdateUserKnowTask(ctx context.Context, list []*model.KnowledgeTaskCalc) (err error) {
	if err = retry.WithAttempts(ctx, "user_knowledge_task_calculate", _retryDBTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		err = s.knowTaskDao.InsertUpdateUserKnowTask(ctx, list)
		return err
	}); err != nil {
		log.Errorc(ctx, "insertUpdateUserKnowTask s.knowTaskDao.RawKnowledgeCalcList() error(%+v)", err)
		return
	}
	return
}

func (s *Service) knowledgeCalculateList(ctx context.Context, logDate string, taskId, limit int64) (res []*model.KnowledgeTaskCalc, err error) {
	if err = retry.WithAttempts(ctx, "user_knowledge_task_calculate", _retryDBTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		res, err = s.knowTaskDao.RawKnowledgeCalcList(ctx, logDate, taskId, limit)
		return err
	}); err != nil {
		log.Errorc(ctx, "knowledgeCalculateList s.knowTaskDao.RawKnowledgeCalcList() offset(%d)  error(%+v)", taskId, err)
		return
	}
	return
}

func (s *Service) DeleteKnowledgeCalculate() {
	// 得到上两天日期
	t := time.Now()
	newTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	logDate := newTime.AddDate(0, 0, -2).Format("20060102")
	// 删除历史db记录
	s.DeleteKnowledgeCalculateDB(logDate)
}

func (s *Service) DeleteKnowledgeCalculateDB(logDate string) {
	var (
		err      error
		count    int64
		delCount int64
		errCount int
	)
	ctx := context.Background()
	timeSleep := time.Duration(s.c.KnowledgeTask.DelSleep)
	if timeSleep <= 0 {
		timeSleep = time.Second
	}
	nowTime := time.Now()
	log.Errorc(ctx, "DeleteKnowledgeCalculateDB begin %v", nowTime.Format("2006-01-02 15:04:05"))
	for {
		time.Sleep(timeSleep)
		if err = retry.WithAttempts(ctx, "delete_knowledge_calculate", _retryDBTimes, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
			count, err = s.knowTaskDao.DeleteKnowledgeCalculate(ctx, logDate, s.c.KnowledgeTask.DelKnowCalcBatchNum)
			return err
		}); err != nil {
			log.Errorc(ctx, "DelKnowledgeCalculateAlarm s.knowTaskDao.DeleteKnowledgeCalculate() logDate(%s) delCount(%d) error(%+v)", logDate, delCount, err)
			errCount++
			if errCount >= 3 {
				break
			}
			continue
		}
		delCount += count
		if count < int64(s.c.KnowledgeTask.DelKnowCalcBatchNum) || count == 0 {
			break
		}
	}
	log.Errorc(ctx, "DeleteKnowledgeCalculateDB end %v since(%v) logDate(%s) delCount(%d)", time.Now().Format("2006-01-02 15:04:05"), time.Since(nowTime).Seconds(), logDate, delCount)
}
