package dao

import (
	"context"

	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-gateway/app/app-svr/app-free/job/internal/model"
)

const (
	_getSQL = "SELECT id,ip_start,ip_start_int,ip_end,ip_end_int,isp,is_bgp,business,state,success_time,cancel_time,ctime,mtime FROM free_record"
)

func NewDB() (db *sql.DB, cf func(), err error) {
	var (
		cfg sql.Config
		ct  paladin.TOML
	)
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("show").UnmarshalTOML(&cfg); err != nil {
		return
	}
	db = sql.NewMySQL(&cfg)
	cf = func() { db.Close() }
	return
}

func (d *dao) RawAllFreeRecords(ctx context.Context) (res []*model.FreeRecord, err error) {
	rows, err := d.db.Query(ctx, _getSQL)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &model.FreeRecord{}
		if err = rows.Scan(&r.ID, &r.IPStart, &r.IPStartInt, &r.IPEnd, &r.IPEndInt, &r.ISP, &r.IsBGP, &r.Business, &r.State, &r.SuccessTime, &r.CancelTime, &r.Ctime, &r.Mtime); err != nil {
			return
		}
		res = append(res, r)
	}
	err = rows.Err()
	return
}
