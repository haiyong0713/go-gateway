package native

import (
	"context"
	"database/sql"
	"fmt"
	xsql "go-common/library/database/sql"
	"go-common/library/xstr"
	v1 "go-gateway/app/web-svr/native-page/interface/api"

	"github.com/pkg/errors"
)

var (
	_modulesSQL = "SELECT id,category,f_id,native_id,state,rank,meta,width,length,num,title,ctime,mtime,dy_sort,`ukey`,`attribute`,`bg_color`,`title_color`,`more_color`,`t_name`,`card_style`,`av_sort`,`font_color`,`p_type`,`caption`,`remark`,`bar`,`stime`,`etime`,`live_type`,`colors`,`conf_sort` FROM native_module WHERE id in (%s)"
	_natIDsSQL  = "SELECT `id`,`rank` FROM `native_module` WHERE `native_id` = ? AND `p_type` = ? AND `state` = 1"
	_natUkeySQL = "SELECT `id` FROM `native_module` WHERE `native_id` = ? AND `state` = 1 AND `ukey` = ?"
)

// Modules .
func (d *Dao) RawNativeModules(c context.Context, ids []int64) (list map[int64]*v1.NativeModule, err error) {
	if len(ids) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_modulesSQL, xstr.JoinInts(ids)))
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	list = make(map[int64]*v1.NativeModule)
	for rows.Next() {
		t := &v1.NativeModule{}
		if err = rows.Scan(&t.ID, &t.Category, &t.Fid, &t.NativeID, &t.State, &t.Rank, &t.Meta, &t.Width, &t.Length, &t.Num, &t.Title, &t.Ctime, &t.Mtime, &t.DySort, &t.Ukey, &t.Attribute, &t.BgColor, &t.TitleColor, &t.MoreColor, &t.TName, &t.CardStyle, &t.AvSort, &t.FontColor, &t.PType, &t.Caption, &t.Remark, &t.Bar, &t.Stime, &t.Etime, &t.LiveType, &t.Colors, &t.ConfSort); err != nil {
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

// RawSortModules .
func (d *Dao) RawSortModules(c context.Context, nat int64, pType int32) (res map[int64]int64, err error) {
	rows, err := d.db.Query(c, _natIDsSQL, nat, pType)
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	res = make(map[int64]int64)
	for rows.Next() {
		t := &v1.NativeModule{}
		if err = rows.Scan(&t.ID, &t.Rank); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		res[t.ID] = t.Rank
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
	}
	return
}

// RawNativeUkey .
func (d *Dao) RawNativeUkey(c context.Context, pid int64, ukey string) (moduleID int64, err error) {
	row := d.db.QueryRow(c, _natUkeySQL, pid, ukey)
	var t sql.NullInt64
	if err = row.Scan(&t); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return
	}
	moduleID = t.Int64
	return
}
