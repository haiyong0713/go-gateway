package dao

import (
	"context"

	"go-common/library/database/elastic"
	gwconfig "go-gateway/app/app-svr/app-gw/management-job/internal/model/gateway-config"
	"go-gateway/app/app-svr/app-gw/management/audit"
)

func (d *dao) RawTaskLog(ctx context.Context, in *gwconfig.RawLogReq) (*gwconfig.LogReply, error) {
	req := d.es.NewRequest("log_audit").Index("log_audit_590_all").
		Fields("action", "business", "ctime", "extra_data", "oid", "str_0", "str_1", "str_2", "type", "uid", "uname").
		Order(in.Order, elastic.OrderDesc).
		WhereEq("action", audit.LogActionPush).
		WhereEq("str_0", in.Node).
		WhereEq("str_1", in.Gateway).
		WhereEq("type", in.ObjectType).
		Pn(in.Pn).Ps(in.Ps)
	taskLog := &gwconfig.LogReply{}
	if err := req.Scan(ctx, taskLog); err != nil {
		return nil, err
	}
	return taskLog, nil
}
