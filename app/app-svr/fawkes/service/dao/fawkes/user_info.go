package fawkes

import (
	"context"

	mdl "go-gateway/app/app-svr/fawkes/service/model/app"
)

const _tableName = "user_info"

// UserInfo get by username list
func (d *Dao) UserInfo(c context.Context, userNames []string) (res []*mdl.UserInfo, err error) {
	if len(userNames) == 0 {
		return
	}
	d.ORMDB.Table(_tableName).Where("user_name IN (?)", userNames).Find(&res)
	return
}
