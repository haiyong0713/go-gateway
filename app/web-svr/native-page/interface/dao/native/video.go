package native

import (
	"context"
	"fmt"
	"sort"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	v1 "go-gateway/app/web-svr/native-page/interface/api"

	"github.com/pkg/errors"
)

var (
	_videosSQL   = "select id,module_id,state,sort_type,rank,ctime,mtime,sort_name,category from native_video_ext where id in (%s)"
	_videosIDSQL = "SELECT id,rank FROM native_video_ext WHERE module_id = ? AND state = 1"
)

func (d *Dao) RawNativeVideoIDs(c context.Context, moduleID int64) ([]int64, error) {
	rows, e := d.db.Query(c, _videosIDSQL, moduleID)
	if e != nil {
		log.Error("RawNativeVideos query ids(%+v)error(%v)", moduleID, e)
		return nil, e
	}
	defer rows.Close()
	rly := make([]*v1.NativeVideoExt, 0)
	for rows.Next() {
		tmp := &v1.NativeVideoExt{}
		if e := rows.Scan(&tmp.ID, &tmp.Rank); e != nil {
			log.Error("RawNativeVideoIDs scan ids(%+v)error(%v)", moduleID, e)
			return nil, e
		}
		rly = append(rly, tmp)
	}
	if e := rows.Err(); e != nil {
		log.Error("RawNativeVideoIDs rows.err ids(%+v)error(%v)", moduleID, e)
		return nil, e
	}
	sort.Slice(rly, func(i, j int) bool {
		return rly[i].Rank < rly[j].Rank
	})
	var res []int64
	for _, v := range rly {
		res = append(res, v.ID)
	}
	return res, nil
}

// NatVideos .
func (d *Dao) NatVideos(c context.Context, ids []int64) (list map[int64]*v1.NativeVideoExt, err error) {
	if len(ids) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_videosSQL, xstr.JoinInts(ids)))
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	list = make(map[int64]*v1.NativeVideoExt)
	for rows.Next() {
		t := &v1.NativeVideoExt{}
		if err = rows.Scan(&t.ID, &t.ModuleID, &t.State, &t.SortType, &t.Rank, &t.Ctime, &t.Mtime, &t.SortName, &t.Category); err != nil {
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

// RawNativeVideos .
func (d *Dao) RawNativeVideos(c context.Context, ids []int64) (list map[int64]*v1.NativeVideoExt, err error) {
	res, err := d.NatVideos(c, ids)
	if err != nil {
		return
	}
	list = make(map[int64]*v1.NativeVideoExt)
	for _, v := range res {
		if v.IsOnline() {
			list[v.ID] = v
		}
	}
	return
}
