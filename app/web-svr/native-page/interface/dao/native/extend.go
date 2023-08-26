package native

import (
	"context"
	"database/sql"
	"go-gateway/app/web-svr/native-page/interface/api"
)

var (
	_extByPidSQL = "SELECT `id`,`pid`,`white_value`,`ctime`,`mtime` FROM `native_page_ext` WHERE `pid` = ?"
)

func (d *Dao) RawNativeExtend(c context.Context, pid int64) (*api.NativePageExtend, error) {
	row := d.db.QueryRow(c, _extByPidSQL, pid)
	t := &api.NativePageExtend{}
	err := row.Scan(&t.Id, &t.Pid, &t.WhiteValue, &t.Ctime, &t.Mtime)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return nil, err
	}
	return t, nil
}
