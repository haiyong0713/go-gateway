package exporttask

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/conf/env"
	"go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/admin/model/exporttask"
	"sync"
	"time"
)

var (
	chanTask = make(chan *exporttask.ExportTask, 10)
	taskLock = sync.Mutex{}
)

type ExportTask interface {
	Do(context.Context, *sql.DB, map[string]string, *readerWriter) error
	Header(context.Context, map[string]string) ([]string, error)
}

func (s *Service) DoTask(c context.Context, task *exporttask.ExportTask) {
	chanTask <- task
}

func (s *Service) TaskDoProc() {
	ctx := context.Background()
	for task := range chanTask {
		s.doTask(ctx, task)
	}
}

func (s *Service) doTask(c context.Context, task *exporttask.ExportTask) (dataSet [][]string, err error) {
	taskLock.Lock()
	defer taskLock.Unlock()

	// 读取task配置
	conf := exportConf[task.TaskType]
	if conf == nil {
		log.Errorc(c, "doTask unknow task type[%v]", task.TaskType)
		return nil, nil
	}

	// 生成task data
	var data map[string]string
	if err = json.Unmarshal(task.Ext, &data); err != nil {
		log.Errorc(c, "doTask json.Unmarshal error[%v]", err)
		return
	}
	data["sid"] = task.SID
	data["start_time"] = task.StartTime.Time().Format("2006-01-02 15:04:05")
	data["end_time"] = task.EndTime.Time().Format("2006-01-02 15:04:05")

	// 更新task状态，锁定task
	task.State = exporttask.TaskStateDoing
	task.StartAt = xtime.Time(time.Now().Unix())
	task.Machine = env.IP
	if task.ID > 0 {
		updateRes := s.DB.Model(&exporttask.ExportTask{}).Where("id = ? AND state != ?", task.ID, exporttask.TaskStateFinish).Update(map[string]interface{}{
			"state":    task.State,
			"start_at": task.StartAt,
			"machine":  task.Machine,
		})
		if updateRes.Error != nil {
			log.Errorc(c, "doTask s.dao.DB.Update error[%v]", err)
			return nil, updateRes.Error
		}
		if updateRes.RowsAffected == 0 {
			log.Infoc(c, "doTask task[%v] already started.", *task)
			return
		}
	} else {
		if err = s.DB.Save(&task).Error; err != nil {
			log.Errorc(c, "doTask s.dao.DB.Save error[%v]", err)
			return
		}
	}
	defer func() {
		if err == nil {
			// 任务完成处理
			task.State = exporttask.TaskStateFinish
			task.EndAt = xtime.Time(time.Now().Unix())
			task.TimeCost = int64(task.EndAt.Time().Sub(task.StartAt.Time()).Seconds())
			s.DB.Save(&task)
			if task.DownURL == "" {
				SendWeChatTextMessage(c, []string{task.Author}, fmt.Sprintf("%d任务执行成功,数据产出为空", task.ID))
			} else {
				SendWeChatTextMessage(c, []string{task.Author}, fmt.Sprintf("%d任务执行成功,下载url:%s", task.ID, task.DownURL))
			}
		} else {
			log.Errorc(c, "%d任务(author:%s,sid:%s,stime:%v,etime:%v,ext:%s,type:%v)执行失败,原因%v",
				task.ID, task.Author, task.SID, task.StartTime, task.EndTime, task.Ext, task.TaskType, err)
			// 任务异常处理
			task.State = exporttask.TaskStateFail
			task.EndAt = xtime.Time(time.Now().Unix())
			task.TimeCost = int64(task.EndAt.Time().Sub(task.StartAt.Time()).Seconds())
			s.DB.Save(&task)
			errString := err.Error()
			if len([]rune(errString)) > 1000 {
				errString = string([]rune(errString)[0:1000])
			}
			SendWeChatTextMessage(c, []string{
				task.Author,
				"ouyangkeshou",
			}, fmt.Sprintf("%d任务(author:%s,sid:%s,stime:%v,etime:%v,ext:%s,type:%v)执行失败,原因%v",
				task.ID, task.Author, task.SID, task.StartTime, task.EndTime, task.Ext, task.TaskType, errString))
		}
	}()

	// 构建一个channel
	dataBuffer := NewReader()

	// 获取header信息
	var header []string
	header, err = conf.Execute.Header(c, data)
	if err != nil || len(header) == 0 {
		return
	}

	dataBuffer.Put([][]string{header})

	var errDo error
	// 异步执行任务，获取返回结果
	go func() {
		defer dataBuffer.Close()
		errDo = conf.Execute.Do(c, s.export, data, dataBuffer)
	}()

	// 保存到bfs
	task.DownURL, err = s.saveBoss(c, fmt.Sprintf("%s_%s_%s_%d.csv",
		task.SID,
		task.StartTime.Time().Format("20060102150405"),
		task.EndTime.Time().Format("20060102150405"),
		task.ID,
	), dataBuffer)

	if err == nil && errDo != nil {
		err = errDo
	}

	return
}
