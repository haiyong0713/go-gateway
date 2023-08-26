package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model"
)

const sql4InsertUpdateKnowledgeTask = `
INSERT INTO %s (mid,%s) 
VALUES %s ON DUPLICATE KEY UPDATE
%s=values(%s);
`

// UpdateUserKnowTask .
func (d *Dao) UpdateUserKnowTask(ctx context.Context, isBack int64, table, field string, historyList []int64) (err error) {
	if field == "" || len(historyList) == 0 {
		return
	}
	var (
		rowsValue []interface{}
		rowsParam []string
	)
	for _, mid := range historyList {
		rowsParam = append(rowsParam, "(?,?)")
		rowsValue = append(rowsValue, mid, isBack)
	}
	sql := fmt.Sprintf(sql4InsertUpdateKnowledgeTask, table, field, strings.Join(rowsParam, ","), field, field)
	if err = d.DB.Exec(sql, rowsValue...).Error; err != nil {
		log.Errorc(ctx, "UpdateUserKnowTask:db.Exec table(%s) field(%s) error(%+v)", table, field, err)
		return
	}
	return
}

func (d *Dao) RawKnowledgeConfig(ctx context.Context, configID int64) (res *model.KnowConfigInfo, err error) {
	dbRes := new(model.KnowConfig)
	if err = d.DB.Where("id = ?", configID).Where("is_deleted = ?", 0).First(dbRes).Error; err != nil {
		log.Error("RawKnowledgeConfig s.DB.Where(id ,%d).First() error(%v)", configID, err)
		return
	}
	if dbRes.ConfigDetails == "" {
		return
	}
	tmpDetail := &model.KnowConfigDetail{}
	err = json.Unmarshal([]byte(dbRes.ConfigDetails), tmpDetail)
	if err != nil {
		log.Errorc(ctx, "RawKnowledgeConfig json.Unmarshal() error(%+v)", err)
		return
	}
	res = &model.KnowConfigInfo{
		ID:            dbRes.ID,
		ConfigDetails: tmpDetail,
	}
	return
}
