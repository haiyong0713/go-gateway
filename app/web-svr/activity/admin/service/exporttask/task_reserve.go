package exporttask

import (
	"context"
	"fmt"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/component"
	"go-gateway/app/web-svr/activity/admin/model"
	"strings"
)

const (
	stateQueryBatch = 1000
)

type taskExportReserve struct {
	reserveSQL   *taskExportSQL
	userStatSQL  *taskExportSQL
	idx          map[string]int
	baseFiledNum int
}

func (t *taskExportReserve) formatWithStateData(c context.Context, reserve []map[string]string, state []map[string]string, writer *readerWriter) error {
	// 拼装输出数据
	reserveMap := make(map[string][]string)
	for _, one := range reserve {
		p := make([]string, t.baseFiledNum+len(t.idx), t.baseFiledNum+len(t.idx))
		p[0] = one["mid"]
		p[1] = one["id"]
		p[2] = formatTimeString(one["ctime"])
		p[3] = one["state"]
		reserveMap[p[0]] = p
	}
	// 遍历积分信息输出
	for _, one := range state {
		if _, ok := reserveMap[one["mid"]]; !ok {
			continue
		}
		if _, ok := t.idx[one["task_id"]]; !ok {
			continue
		}
		reserveMap[one["mid"]][t.idx[one["task_id"]]] = one["cnt"]
	}
	// 按预约输出组织数据输出
	dataSet := make([][]string, 0, len(reserveMap))
	for _, one := range reserve {
		dataSet = append(dataSet, reserveMap[one["mid"]])
	}
	writer.Put(dataSet)
	return nil
}

func (t *taskExportReserve) Do(c context.Context, db *sql.DB, data map[string]string, writer *readerWriter) error {
	var reserveDo func([]map[string]string) error
	if len(t.idx) > 0 {
		reserveDo = func(reserve []map[string]string) error {
			total := len(reserve)
			var state = make([]map[string]string, 0, total)
			for i := 0; i < total; i += stateQueryBatch {
				var tmp []map[string]string
				if i+stateQueryBatch < total {
					tmp = reserve[i : i+stateQueryBatch]
				} else {
					tmp = reserve[i:]
				}
				midList := make([]string, 0, len(tmp))
				for _, r := range tmp {
					midList = append(midList, r["mid"])
				}
				data["mid_list"] = strings.Join(midList, ",")
				// 有积分信息，查询积分信息
				err := t.userStatSQL.GetData(c, db, data, func(s []map[string]string) error {
					state = append(state, s...)
					return nil
				})
				if err != nil {
					return err
				}
			}
			return t.formatWithStateData(c, reserve, state, writer)
		}
	} else {
		reserveDo = func(reserve []map[string]string) error {
			// 无积分信息，直接进行构造
			return t.formatWithStateData(c, reserve, []map[string]string{}, writer)
		}
	}
	// 查询预约数据
	return t.reserveSQL.GetData(c, db, data, reserveDo)
}

func (t *taskExportReserve) getTaskIDList(c context.Context, sid string) ([]string, error) {
	list := make([]*model.SubjectRule, 0, 10)
	if err := component.GlobalOrm.Where("sid=?", sid).Where("state != 3").Find(&list).Error; err != nil {
		log.Errorc(c, "getTaskIDList db.Where(sid:%d).Find error(%v)", sid, err)
		return nil, err
	}
	var taskIDs []string
	for _, v := range list {
		if v != nil && v.TaskID > 0 {
			taskIDs = append(taskIDs, fmt.Sprint(v.TaskID))
		}
	}
	return taskIDs, nil
}

func (t *taskExportReserve) Header(c context.Context, data map[string]string) ([]string, error) {
	// 获取taskid
	taskID, err := t.getTaskIDList(c, data["sid"])
	if err != nil {
		return nil, err
	}
	header := []string{
		"mid",
		"id",
		"日期",
		"状态(1:预约,0:取消预约)",
	}
	t.baseFiledNum = len(header)
	// 拼装header和积分信息列
	t.idx = map[string]int{}
	for i, id := range taskID {
		t.idx[id] = i + t.baseFiledNum
		header = append(header, fmt.Sprintf("行为%d", i+1))
	}
	return header, nil
}
