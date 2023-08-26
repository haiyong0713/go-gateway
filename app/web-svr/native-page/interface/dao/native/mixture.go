package native

import (
	"context"
	"fmt"
	"strings"

	xsql "go-common/library/database/sql"

	v1 "go-gateway/app/web-svr/native-page/interface/api"

	"github.com/pkg/errors"
)

var (
	_mixturesSQL     = "SELECT `id`,`module_id`,`state`,`m_type`,`rank`,`ctime`,`mtime`,`foreign_id`,`reason` FROM native_mixture_ext WHERE id IN (%s) AND `state` = 1"
	_mixModuleSQL    = "SELECT `id`,`rank` FROM `native_mixture_ext` WHERE `module_id` = ? AND `m_type` = ? AND `state` = 1"
	_mixAllModuleSQL = "SELECT `id`,`rank` FROM `native_mixture_ext` WHERE `module_id` = ?  AND `state` = 1"
)

// NatMixIDsSearch .
func (d *Dao) NatMixIDsSearch(c context.Context, ModuleID int64, MType int32) (list []*v1.NativeMixtureExt, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _mixModuleSQL, ModuleID, MType); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		t := &v1.NativeMixtureExt{}
		if err = rows.Scan(&t.ID, &t.Rank); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		list = append(list, t)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
	}
	return
}

// NatAllMixIDsSearch .
func (d *Dao) NatAllMixIDsSearch(c context.Context, ModuleID int64) (list []*v1.NativeMixtureExt, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _mixAllModuleSQL, ModuleID); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	for rows.Next() {
		t := &v1.NativeMixtureExt{}
		if err = rows.Scan(&t.ID, &t.Rank); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		list = append(list, t)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
	}
	return
}

// NatMixtures .
func (d *Dao) RawNativeMixtures(c context.Context, ids []int64) (list map[int64]*v1.NativeMixtureExt, err error) {
	var (
		rows *xsql.Rows
		sqls []string
		args []interface{}
	)

	if len(ids) == 0 {
		return
	}
	for _, k := range ids {
		sqls = append(sqls, "?")
		args = append(args, k)
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_mixturesSQL, strings.Join(sqls, ",")), args...); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	list = make(map[int64]*v1.NativeMixtureExt)
	for rows.Next() {
		t := &v1.NativeMixtureExt{}
		if err = rows.Scan(&t.ID, &t.ModuleID, &t.State, &t.MType, &t.Rank, &t.Ctime, &t.Mtime, &t.ForeignID, &t.Reason); err != nil {
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
