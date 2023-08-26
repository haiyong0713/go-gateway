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
	_clicksSQL    = "select id,module_id,state,left_x,left_y,width,length,link,ctime,mtime,type,foreign_id,unfinished_image,finished_image,tip,optional_image,ext from native_click where id in (%s)"
	_clickSortSQL = "SELECT id FROM native_click WHERE module_id = ? AND `state`=1"
)

func (d *Dao) RawNativeClickIDs(c context.Context, id int64) ([]int64, error) {
	rows, e := d.db.Query(c, _clickSortSQL, id)
	if e != nil {
		log.Error("RawNativeClickIDs query ids(%+v)error(%v)", id, e)
		return nil, e
	}
	defer rows.Close()
	rly := make([]int64, 0)
	for rows.Next() {
		tmp := sql.NullInt64{}
		if e := rows.Scan(&tmp); e != nil {
			log.Error("RawNativeClickIDs scaen ids(%+v)error(%v)", id, e)
			return nil, e
		}
		rly = append(rly, tmp.Int64)
	}
	if e := rows.Err(); e != nil {
		log.Error("RawNativeClickIDs rows.err ids(%+v)error(%v)", id, e)
		return nil, e
	}
	sort.Slice(rly, func(i, j int) bool {
		return rly[i] < rly[j]
	})
	return rly, nil
}

// Clicks .
func (d *Dao) Clicks(c context.Context, ids []int64) (list map[int64]*v1.NativeClick, err error) {
	if len(ids) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_clicksSQL, xstr.JoinInts(ids)))
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	list = make(map[int64]*v1.NativeClick)
	for rows.Next() {
		t := &v1.NativeClick{}
		if err = rows.Scan(&t.ID, &t.ModuleID, &t.State, &t.Leftx, &t.Lefty, &t.Width, &t.Length, &t.Link, &t.Ctime, &t.Mtime, &t.Type, &t.ForeignID, &t.UnfinishedImage, &t.FinishedImage, &t.Tip, &t.OptionalImage, &t.Ext); err != nil {
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
func (d *Dao) RawNativeClicks(c context.Context, ids []int64) (list map[int64]*v1.NativeClick, err error) {
	res, err := d.Clicks(c, ids)
	if err != nil {
		return
	}
	list = make(map[int64]*v1.NativeClick)
	for _, v := range res {
		if v.IsOnline() {
			list[v.ID] = v
		}
	}
	return
}
