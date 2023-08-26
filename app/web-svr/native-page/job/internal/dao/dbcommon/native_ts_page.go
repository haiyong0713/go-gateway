package dbcommon

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"
	natGRPC "go-gateway/app/web-svr/native-page/interface/api"

	"go-gateway/app/web-svr/native-page/admin/model/native"
)

const (
	pagingAuditSQL = "select `id`,`pid`,`audit_type`,`state`,`audit_time` from `native_ts_page` where `audit_type`=? and state=? order by `id` limit ?, ?"
)

func (d *Dao) PagingAutoAuditTsPages(c context.Context, pn, ps int64) ([]*natGRPC.NativeTsPage, error) {
	var offset int64
	if pn >= 1 {
		offset = (pn - 1) * ps
	}
	rows, err := d.db.Query(c, pagingAuditSQL, native.TsAuditAuto, native.TsWaitOnline, offset, ps)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*natGRPC.NativeTsPage{}, nil
		}
		log.Errorc(c, "Fail to query pagingAuditSQL, sql=%s error=%+v", pagingAuditSQL, err)
		return nil, err
	}
	defer rows.Close()
	list := make([]*natGRPC.NativeTsPage, 0, ps)
	for rows.Next() {
		t := &natGRPC.NativeTsPage{}
		err = rows.Scan(&t.Id, &t.Pid, &t.AuditType, &t.State, &t.AuditTime)
		if err != nil {
			log.Errorc(c, "Fail to scan NativeTsPage row, error=%+v", err)
			continue
		}
		list = append(list, t)
	}
	err = rows.Err()
	if err != nil {
		log.Errorc(c, "Fail to get NativeTsPage rows, error=%+v", err)
		return nil, err
	}
	return list, nil
}
