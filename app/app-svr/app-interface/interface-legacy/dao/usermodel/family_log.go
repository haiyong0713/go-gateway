package usermodel

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/log"

	model "go-gateway/app/app-svr/app-interface/interface-legacy/model/family"
)

const (
	_batchAddLogsSQL = "INSERT INTO family_log (mid,operator,content) VALUES %s"
)

func (d *dao) AddFamilyLogs(ctx context.Context, items []*model.FamilyLog) error {
	if len(items) == 0 {
		return nil
	}
	parts := make([]string, 0, len(items))
	args := make([]interface{}, 0, len(items)*3)
	for _, item := range items {
		if item == nil {
			continue
		}
		parts = append(parts, "(?,?,?)")
		args = append(args, item.Mid, item.Operator, item.Content)
	}
	if _, err := d.db.Exec(ctx, fmt.Sprintf(_batchAddLogsSQL, strings.Join(parts, ",")), args...); err != nil {
		log.Error("Fail to batch create family_log, items=%+v error=%+v", items, err)
		return err
	}
	return nil
}
