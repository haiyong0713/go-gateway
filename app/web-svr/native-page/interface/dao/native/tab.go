package native

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/log"

	v1 "go-gateway/app/web-svr/native-page/interface/api"
)

var (
	_tabSQL = "SELECT `id`,`title`,`stime`,`ctime`,`mtime`,`state`,`etime`,`bg_type`,`bg_img`,`bg_color`,`icon_type`,`active_color`,`inactive_color` FROM `act_tab` WHERE `id` IN (%s)"
)

// RawNativeTab .
func (d *Dao) RawNativeTabs(c context.Context, ids []int64) (map[int64]*v1.NativeActTab, error) {
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
	sqlStr := fmt.Sprintf(_tabSQL, strings.Join(param, ","))
	rows, e := d.db.Query(c, sqlStr, vals...)
	if e != nil {
		log.Error("RawNativeTabs query ids(%+v)error(%v)", ids, e)
		return nil, e
	}
	defer rows.Close()
	res := make(map[int64]*v1.NativeActTab)
	for rows.Next() {
		rly := &v1.NativeActTab{}
		if e := rows.Scan(&rly.ID, &rly.Title, &rly.Stime, &rly.Ctime, &rly.Mtime, &rly.State, &rly.Etime, &rly.BgType, &rly.BgImg, &rly.BgColor, &rly.IconType, &rly.ActiveColor, &rly.InactiveColor); e != nil {
			log.Error("RawNativeTabs id(%v) error(%v)", ids, e)
			return nil, e
		}
		res[rly.ID] = rly
	}
	if e := rows.Err(); e != nil {
		return nil, e
	}
	return res, nil
}
