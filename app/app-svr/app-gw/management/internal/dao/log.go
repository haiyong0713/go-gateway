package dao

import (
	"context"

	"go-common/library/database/elastic"
	"go-gateway/app/app-svr/app-gw/management/internal/model"
)

func (d *dao) ListLog(ctx context.Context, node, gateway string, object, pn, ps int64) (*model.ListLogReply, error) {
	req := d.es.NewRequest("log_audit").Index("log_audit_590_all").
		Fields("action", "business", "ctime", "extra_data", "str_0", "str_1", "str_2", "str_3", "str_4", "str_5", "oid", "type", "uid", "uname").
		WhereEq("str_0", node).
		WhereEq("str_1", gateway)
	if object != 0 {
		req.WhereEq("type", object)
	}
	req.Pn(int(pn)).Ps(int(ps)).Order("ctime", elastic.OrderDesc)
	listLog := &model.ListLogReply{}
	if err := req.Scan(ctx, listLog); err != nil {
		return nil, err
	}
	return listLog, nil
}
