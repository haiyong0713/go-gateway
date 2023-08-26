package native

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go-common/library/log"
	"go-common/library/xstr"

	v1 "go-gateway/app/web-svr/native-page/interface/api"
)

const (
	_pagesDynSQL         = "SELECT `id`,`pid`,`stime`,`validity`,`ctime`,`mtime`,`tids`,`small_card`,`big_card`,`square_title`,`dynamic`,`dyn_id` FROM native_page_dyn WHERE pid in (%s)"
	_addPageDynSQL       = "INSERT INTO `native_page_dyn` (`pid`, `dynamic`) VALUES (?,?)"
	_updateDynDynamicSQL = "UPDATE `native_page_dyn` SET `dynamic`=? WHERE `id`=?"
	_updateDynDynIDSQL   = "UPDATE `native_page_dyn` SET `dyn_id`=? WHERE `id`=?"
)

// RawNativePages .
func (d *Dao) RawNativePagesExt(c context.Context, ids []int64) (map[int64]*v1.NativePageDyn, error) {
	if len(ids) == 0 {
		return make(map[int64]*v1.NativePageDyn), nil
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_pagesDynSQL, xstr.JoinInts(ids)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make(map[int64]*v1.NativePageDyn)
	for rows.Next() {
		t := &v1.NativePageDyn{}
		if err = rows.Scan(&t.Id, &t.Pid, &t.Stime, &t.Validity, &t.Ctime, &t.Mtime, &t.Tids, &t.SmallCard, &t.BigCard, &t.SquareTitle, &t.Dynamic, &t.DynId); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return nil, err
		}
		list[t.Pid] = t
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
		return nil, err
	}
	return list, nil
}

func (d *Dao) AddNativePageDyn(c context.Context, pid int64, dynamic string) (int64, error) {
	rly, err := d.db.Exec(c, _addPageDynSQL, pid, dynamic)
	if err != nil {
		log.Error("Fail to add native_page_dyn, pid=%d dynamic=%s error=%+v", pid, dynamic, err)
		return 0, err
	}
	return rly.LastInsertId()
}

func (d *Dao) UpdateNatDynDynamic(c context.Context, id int64, dynamic string) error {
	if _, err := d.db.Exec(c, _updateDynDynamicSQL, dynamic, id); err != nil {
		log.Error("Fail to update native_page_dyn, id=%d dynamic=%s error=%+v", id, dynamic, err)
		return err
	}
	return nil
}

func (d *Dao) UpdateNatDynDynID(c context.Context, id, dynID int64) error {
	if _, err := d.db.Exec(c, _updateDynDynIDSQL, dynID, id); err != nil {
		log.Error("Fail to update native_page_dyn, id=%d dyn_id=%d error=%+v", id, dynID, err)
		return err
	}
	return nil
}
