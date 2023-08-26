package dao

import (
	"context"

	"go-gateway/app/app-svr/archive-honor/service/api"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
)

const (
	_honorsByAidSQL = "SELECT aid,`type`,url,`desc`,na_url FROM archive_honor WHERE aid=? AND `status`=?"
	_delHonorSQL    = "UPDATE archive_honor SET `status`=0 WHERE aid=? AND `type`=?"
	_inHonorSQL     = "INSERT INTO archive_honor (aid,`type`,url,`desc`,`status`,na_url) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE url=?,`desc`=?,`status`=?,na_url=?"
)

// honorsByAid is
func (d *Dao) honorsByAid(c context.Context, aid int64) (res map[int32]*api.Honor, err error) {
	var rows *xsql.Rows
	res = make(map[int32]*api.Honor)
	if rows, err = d.db.Query(c, _honorsByAidSQL, aid, api.StatusForNormal); err != nil {
		log.Error("d.db.Query aid(%d) error(%v)", aid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		honor := &api.Honor{}
		if err = rows.Scan(&honor.Aid, &honor.Type, &honor.Url, &honor.Desc, &honor.NaUrl); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		res[honor.Type] = honor
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// delHonor is
func (d *Dao) delHonor(c context.Context, aid int64, typ int32) (rows int64, err error) {
	res, err := d.db.Exec(c, _delHonorSQL, aid, typ)
	if err != nil {
		log.Error("d.db.Exec aid(%d) type(%d) error(%v)", aid, typ, err)
		return
	}
	return res.RowsAffected()
}

// UpHonor is
func (d *Dao) UpHonor(c context.Context, aid int64, typ int32, url, desc, naUrl string) (rows int64, err error) {
	res, err := d.db.Exec(c, _inHonorSQL, aid, typ, url, desc, api.StatusForNormal, naUrl, url, desc, api.StatusForNormal, naUrl)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
