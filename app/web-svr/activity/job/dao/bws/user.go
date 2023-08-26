package bws

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/bws"
)

const _BindUserSQL = "SELECT id,mid,`key` FROM act_bws_users WHERE mid>0 AND bid=? AND id>? ORDER BY id LIMIT ?"

func (d *Dao) BindUsers(c context.Context, bid, id int64, limit int) (list []*bws.User, err error) {
	rows, err := d.db.Query(c, _BindUserSQL, bid, id, limit)
	if err != nil {
		log.Error("BindUsers:dao.db.Query(%d,%d,%d) error(%v)", bid, id, limit, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &bws.User{}
		if err = rows.Scan(&r.ID, &r.Mid, &r.Key); err != nil {
			log.Error("BindUsers:rows.Scan error(%v)", err)
			return
		}
		list = append(list, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("BindUsers: rows.Err(%v)", err)
	}
	return
}
