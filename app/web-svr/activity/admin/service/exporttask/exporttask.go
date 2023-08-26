package exporttask

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/exporttask"
	"gopkg.in/go-playground/validator.v9"
)

var (
	validate = validator.New()
)

func (s *Service) ExportTaskAdd(c context.Context, req *exporttask.ReqExportTaskAdd, data map[string]string) (interface{}, error) {
	if conf, ok := exportConf[req.TaskType]; !ok {
		return nil, ecode.Error(ecode.RequestErr, "任务类型未定义")
	} else {
		// 数据校验
		for k, v := range conf.Params {
			if v.Validate != "" {
				if err := validate.Var(data[k], v.Validate); err != nil {
					log.Errorc(c, "ExportTaskAdd validate.Var(%s) value[%v] error[%v]", k, data[k], err)
					return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("Validate key[%v] error[%v]", k, err.Error()))
				}
			}
		}
	}

	task := new(exporttask.ExportTask)
	task.Author = req.Author
	task.State = exporttask.TaskStateAdd
	task.Ext, _ = json.Marshal(data)
	task.SID = req.SID
	task.TaskType = req.TaskType
	task.StartTime = req.StartTime
	task.EndTime = req.EndTime
	// 创建任务
	err := s.DB.Model(&exporttask.ExportTask{}).Create(&task).Error
	if err != nil {
		return nil, err
	}
	if req.Synchronize {
		var dataSet [][]string
		dataSet, err = s.doTask(c, task)
		if req.ExportType == exporttask.ExportTypeJson {
			return dataSet, err
		}
		return task, err
	} else {
		s.DoTask(c, task)
	}
	return task, err
}

func (s *Service) ExportTaskState(c context.Context, taskID int64) (interface{}, error) {
	task := new(exporttask.ExportTask)
	return task, s.DB.Model(&exporttask.ExportTask{}).Where("id = ?", taskID).First(&task).Error
}

func (s *Service) ExportTaskRedo(c context.Context, taskID int64) (interface{}, error) {
	task := new(exporttask.ExportTask)
	err := s.DB.Model(&exporttask.ExportTask{}).Where("id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	if task.State != exporttask.TaskStateFinish {
		s.DoTask(c, task)
	}
	return task, nil
}

func (s *Service) ExportTaskList(c context.Context, req *exporttask.ReqExportTaskList) (interface{}, error) {
	list := new([]*exporttask.ExportTask)
	source := s.DB.Model(&exporttask.ExportTask{}).Where("sid = ? AND author = ?", req.SID, req.Author)
	if req.Type > 0 {
		source = source.Where("task_type = ?", req.Type)
	}
	var count int
	if err := source.Count(&count).Error; err != nil {
		return nil, err
	}
	if err := source.Order("id desc").Offset(req.PageSize*req.Page - req.PageSize).Limit(req.PageSize).Find(&list).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"count": count,
		"list":  list,
	}, nil
}
