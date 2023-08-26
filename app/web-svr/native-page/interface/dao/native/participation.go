package native

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	xsql "go-common/library/database/sql"
	"go-common/library/xstr"

	v1 "go-gateway/app/web-svr/native-page/interface/api"
)

var (
	_participationSQL = "SELECT id,module_id,state,m_type,image,title,rank,foreign_id,ctime,mtime,up_type,ext FROM native_participation_ext WHERE id in (%s)"
	_rawNatPartIDs    = "SELECT id,rank FROM native_participation_ext WHERE module_id=? AND state=1 ORDER BY rank"
)

func (d *Dao) RawNativePart(c context.Context, ids []int64) (list map[int64]*v1.NativeParticipationExt, err error) {
	if len(ids) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_participationSQL, xstr.JoinInts(ids)))
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	list = make(map[int64]*v1.NativeParticipationExt)
	for rows.Next() {
		tmp := &v1.NativeParticipationExt{}
		if err = rows.Scan(&tmp.ID, &tmp.ModuleID, &tmp.State, &tmp.MType, &tmp.Image, &tmp.Title, &tmp.Rank, &tmp.ForeignID, &tmp.Ctime,
			&tmp.Mtime, &tmp.UpType, &tmp.Ext); err != nil {
			return
		}
		list[tmp.ID] = tmp
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
	}
	return
}

// RawNatPartIDs get native_participation_ext ids
func (d *Dao) NatPartIDsSearch(c context.Context, pid int64) (res []*v1.NativeParticipationExt, err error) {
	rows, err := d.db.Query(c, _rawNatPartIDs, pid)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		tmp := &v1.NativeParticipationExt{}
		if err = rows.Scan(&tmp.ID, &tmp.Rank); err != nil {
			return
		}
		res = append(res, tmp)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawNatPartIDs rows.Err")
	}
	return
}
