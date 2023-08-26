package native

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	v1 "go-gateway/app/web-svr/native-page/interface/api"

	"github.com/pkg/errors"
)

var (
	_dynamicsSQL   = "select id,module_id,state,select_type,ctime,mtime,`class_type`,`class_id` from native_dynamic_ext where id in (%s)"
	_dynamicIDsSQL = "SELECT id FROM native_dynamic_ext WHERE module_id =? AND `state` = 1"
)

func (d *Dao) RawNativeDynamicIDs(c context.Context, moduleID int64) ([]int64, error) {
	rows, e := d.db.Query(c, _dynamicIDsSQL, moduleID)
	if e != nil {
		log.Error("RawNativeDynamicIDs query ids(%+v)error(%v)", moduleID, e)
		return nil, e
	}
	defer rows.Close()
	rly := make([]int64, 0)
	for rows.Next() {
		tmp := sql.NullInt64{}
		if e := rows.Scan(&tmp); e != nil {
			log.Error("RawNativeDynamicIDs scan ids(%+v)error(%v)", moduleID, e)
			return nil, e
		}
		rly = append(rly, tmp.Int64)
	}
	if e := rows.Err(); e != nil {
		log.Error("RawNativeDynamicIDs rows.err ids(%+v)error(%v)", moduleID, e)
		return nil, e
	}
	sort.Slice(rly, func(i, j int) bool {
		return rly[i] < rly[j]
	})
	return rly, nil
}

// Clicks .
func (d *Dao) Dynamics(c context.Context, ids []int64) (list map[int64]*v1.NativeDynamicExt, err error) {
	if len(ids) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_dynamicsSQL, xstr.JoinInts(ids)))
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	list = make(map[int64]*v1.NativeDynamicExt)
	for rows.Next() {
		t := &v1.NativeDynamicExt{}
		if err = rows.Scan(&t.ID, &t.ModuleID, &t.State, &t.SelectType, &t.Ctime, &t.Mtime, &t.ClassType, &t.ClassID); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		list[t.ID] = t
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
	}
	return
}

// RawNativeClicks .
func (d *Dao) RawNativeDynamics(c context.Context, ids []int64) (list map[int64]*v1.NativeDynamicExt, err error) {
	res, err := d.Dynamics(c, ids)
	if err != nil {
		return
	}
	list = make(map[int64]*v1.NativeDynamicExt)
	for _, v := range res {
		if v.IsOnline() {
			list[v.ID] = v
		}
	}
	return
}
