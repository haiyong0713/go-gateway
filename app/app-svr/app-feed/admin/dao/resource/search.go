package resource

import (
	"context"
	"strconv"

	"go-common/library/database/elastic"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	model "go-gateway/app/app-svr/app-feed/admin/model/resource"
)

// SearchCCLog is.
func (d *Dao) SearchCCLog(ctx context.Context, id int64) (*model.SearchLogResult, error) {
	index := []string{
		"log_audit_204_all",
	}
	r := d.es.NewRequest("log_audit").
		Index(index...).
		Fields("uid", "uname", "oid", "type", "action", "str_0", "str_1", "str_2", "int_0", "int_1", "int_2", "ctime", "extra_data").
		WhereEq("type", strconv.FormatInt(common.LogResourceCustomConfig, 10)).
		WhereEq("oid", strconv.FormatInt(id, 10)).
		Order("ctime", elastic.OrderDesc).
		Pn(1).
		Ps(100)

	result := &model.SearchLogResult{}
	if err := r.Scan(ctx, result); err != nil {
		log.Error("Failed to SearchUserAuditLog: Scan params(%s) error(%+v)", r.Params(), err)
		return nil, err
	}
	return result, nil
}
