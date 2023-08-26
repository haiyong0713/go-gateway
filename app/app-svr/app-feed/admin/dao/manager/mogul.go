package manager

import (
	"context"

	"go-gateway/app/app-svr/app-feed/admin/model/manager"
)

// AppMogulLogList is.
func (d *Dao) AppMogulLogList(ctx context.Context, param *manager.AppMogulLogParam) (*manager.AppMogulLogReply, error) {
	query := d.DB.Table("app_mogul_log")
	if param.Mid != 0 {
		query = query.Where("mid = ?", param.Mid)
	}
	if param.Path != "" {
		query = query.Where("path = ?", param.Path)
	}
	if param.Etime != 0 {
		query = query.Where("request_time <= ?", param.Etime)
	}
	if param.Stime != 0 {
		query = query.Where("request_time >= ?", param.Stime)
	}
	var total int
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []*manager.AppMogulLog
	query = query.Order("request_time DESC").
		Offset(int((param.Pn - 1) * param.Ps)).Limit(int(param.Ps))
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.Duration != 0 {
			item.DurationHuman = item.Duration.String()
		}
	}
	reply := &manager.AppMogulLogReply{
		Items: items,
		Page: &manager.Page{
			Num:   param.Pn,
			Size:  param.Ps,
			Total: total,
		},
	}
	return reply, nil
}
