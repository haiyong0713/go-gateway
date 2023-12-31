package show

import (
	"context"

	"go-common/library/log"

	api "go-gateway/app/app-svr/resource/service/api/v1"
)

var (
	_auditSQL = "SELECT mobi_app,build FROM audit"
)

// Audit get audit.
func (d *Dao) Audit(c context.Context) (res map[string][]int, err error) {
	res = make(map[string][]int)
	rows, err := d.db.Query(c, _auditSQL)
	if err != nil {
		log.Error("d.audit error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			mobiApp string
			build   int
		)
		if err = rows.Scan(&mobiApp, &build); err != nil {
			log.Error("d.audit rows.Scan error(%v)", err)
			res = nil
			return
		}
		res[mobiApp] = append(res[mobiApp], build)
	}
	err = rows.Err()
	return
}

// AppAudits get all audit build.
func (d *Dao) AppAudits(c context.Context) (res []*api.Audit, err error) {
	rows, err := d.db.Query(c, _auditSQL)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &api.Audit{}
		if err = rows.Scan(&a.MobiApp, &a.Build); err != nil {
			log.Error("%+v", err)
			return
		}
		res = append(res, a)
	}
	err = rows.Err()
	return
}
