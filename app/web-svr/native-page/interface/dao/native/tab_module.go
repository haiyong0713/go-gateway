package native

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go-common/library/log"

	v1 "go-gateway/app/web-svr/native-page/interface/api"
)

var (
	_tabModuleSQL = "SELECT `id`,`title`,`tab_id`,`ctime`,`mtime`,`state`,`active_img`,`inactive_img`,`category`,`pid`,`url`,`rank` FROM `act_tab_module` WHERE id in (%s)"
	_tabSortSQL   = "SELECT `id`,`rank` FROM `act_tab_module` WHERE `tab_id` =? AND `state` = 1 "
	_tabBindSQL   = "SELECT `id`,`pid` FROM `act_tab_module` WHERE `pid` in (%s) AND `category` = ? AND `state` = 1"
)

// RawNativeTabBind .
func (d *Dao) RawNativeTabBind(c context.Context, ids []int64, category int32) (map[int64]int64, error) {
	lenIDs := len(ids)
	if lenIDs == 0 {
		return nil, nil
	}
	var (
		param []string
		vals  []interface{}
	)
	for _, v := range ids {
		param = append(param, "?")
		vals = append(vals, v)
	}
	vals = append(vals, category)
	sqlStr := fmt.Sprintf(_tabBindSQL, strings.Join(param, ","))
	rows, e := d.db.Query(c, sqlStr, vals...)
	if e != nil {
		log.Error("RawNativeTabBind query ids(%+v)error(%v)", ids, e)
		return nil, e
	}
	defer rows.Close()
	rly := make(map[int64]int64)
	for rows.Next() {
		tmp := &v1.NativeTabModule{}
		if e := rows.Scan(&tmp.ID, &tmp.Pid); e != nil {
			log.Error("RawNativeTabModules scan ids(%+v)error(%v)", ids, e)
			return nil, e
		}
		rly[tmp.Pid] = tmp.ID
	}
	if e := rows.Err(); e != nil {
		log.Error("RawNativeTabModules rows.err ids(%+v)error(%v)", ids, e)
		return nil, e
	}
	return rly, nil
}

// RawNativeTabSort .
func (d *Dao) RawNativeTabSort(c context.Context, id int64) ([]int64, error) {
	rows, e := d.db.Query(c, _tabSortSQL, id)
	if e != nil {
		log.Error("RawNativeTabSort query ids(%+v)error(%v)", id, e)
		return nil, e
	}
	defer rows.Close()
	rly := make([]*v1.NativeTabModule, 0)
	for rows.Next() {
		tmp := &v1.NativeTabModule{}
		if e := rows.Scan(&tmp.ID, &tmp.Rank); e != nil {
			log.Error("RawNativeTabSort scaen ids(%+v)error(%v)", id, e)
			return nil, e
		}
		rly = append(rly, tmp)
	}
	if e := rows.Err(); e != nil {
		log.Error("RawNativeTabSort rows.err ids(%+v)error(%v)", id, e)
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

// RawNativeTab .
func (d *Dao) RawNativeTabModules(c context.Context, ids []int64) (map[int64]*v1.NativeTabModule, error) {
	lenIDs := len(ids)
	if lenIDs == 0 {
		return nil, nil
	}
	var (
		param []string
		vals  []interface{}
	)
	for _, v := range ids {
		param = append(param, "?")
		vals = append(vals, v)
	}
	sqlStr := fmt.Sprintf(_tabModuleSQL, strings.Join(param, ","))
	rows, e := d.db.Query(c, sqlStr, vals...)
	if e != nil {
		log.Error("RawNativeTabModules query ids(%+v)error(%v)", ids, e)
		return nil, e
	}
	defer rows.Close()
	rly := make(map[int64]*v1.NativeTabModule)
	for rows.Next() {
		tmp := &v1.NativeTabModule{}
		if e := rows.Scan(&tmp.ID, &tmp.Title, &tmp.TabID, &tmp.Ctime, &tmp.Mtime, &tmp.State, &tmp.ActiveImg, &tmp.InactiveImg, &tmp.Category, &tmp.Pid, &tmp.URL, &tmp.Rank); e != nil {
			log.Error("RawNativeTabModules scaen ids(%+v)error(%v)", ids, e)
			return nil, e
		}
		if tmp.IsOnline() {
			rly[tmp.ID] = tmp
		}
	}
	if e := rows.Err(); e != nil {
		log.Error("RawNativeTabModules rows.err ids(%+v)error(%v)", ids, e)
		return nil, e
	}
	return rly, nil
}
