package native

import (
	"context"
	"sort"

	"go-common/library/log"
	v1 "go-gateway/app/web-svr/native-page/interface/api"
)

var (
	_actIDsSQL = "SELECT rank,page_id FROM native_act WHERE module_id = ? AND state = 1"
)

func (d *Dao) RawNativeActIDs(c context.Context, moduleID int64) ([]int64, error) {
	rows, e := d.db.Query(c, _actIDsSQL, moduleID)
	if e != nil {
		log.Error("RawNativeActIDs query ids(%+v)error(%v)", moduleID, e)
		return nil, e
	}
	defer rows.Close()
	rly := make([]*v1.NativeAct, 0)
	for rows.Next() {
		tmp := &v1.NativeAct{}
		if e := rows.Scan(&tmp.Rank, &tmp.PageID); e != nil {
			log.Error("RawNativeActIDs scan ids(%+v)error(%v)", moduleID, e)
			return nil, e
		}
		rly = append(rly, tmp)
	}
	if e := rows.Err(); e != nil {
		log.Error("RawNativeActIDs rows.err ids(%+v)error(%v)", moduleID, e)
		return nil, e
	}
	sort.Slice(rly, func(i, j int) bool {
		return rly[i].Rank < rly[j].Rank
	})
	var res []int64
	for _, v := range rly {
		res = append(res, v.PageID)
	}
	return res, nil
}
