package native

import (
	"context"
	"database/sql"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"

	v1 "go-gateway/app/web-svr/native-page/interface/api"

	"github.com/pkg/errors"
)

var (
	_tsPagesSQL   = "SELECT `id`,`ctime`,`mtime`,`state`,`pid`,`title`,`foreign_id`,`video_display`,`audit_type`,`audit_time`,`share_image`,`template` FROM `native_ts_page` WHERE id in (%s)"
	_tsLastIDSQL  = "SELECT `id` FROM `native_ts_page` WHERE `pid` = ? ORDER BY `id` DESC LIMIT 1"
	_tsLastIDsSQL = "SELECT `pid`, max(`id`) FROM `native_ts_page` WHERE `pid` in (%s) group by `pid`;"
	_tsAddPageSQL = "INSERT INTO `native_ts_page` (`title`,`foreign_id`,`state`,`pid`,`video_display`,`audit_type`,`audit_time`,`share_image`,`template`) VALUES (?,?,?,?,?,?,?,?,?)"
	_tsDisplaySQL = "UPDATE `native_ts_page` set `video_display`=?, `audit_time`=?, `share_image`=? where id=?"
	_tsUpStateSQL = "UPDATE `native_ts_page` set `state`=?,`msg`=? where id=?"
)

// RawNativePages .
func (d *Dao) RawNtTsPages(c context.Context, ids []int64) (list map[int64]*v1.NativeTsPage, err error) {
	if len(ids) == 0 {
		return
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_tsPagesSQL, xstr.JoinInts(ids)))
	if err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	list = make(map[int64]*v1.NativeTsPage)
	for rows.Next() {
		t := &v1.NativeTsPage{}
		if err = rows.Scan(&t.Id, &t.Ctime, &t.Mtime, &t.State, &t.Pid, &t.Title, &t.ForeignID, &t.VideoDisplay, &t.AuditType, &t.AuditTime, &t.ShareImage, &t.Template); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return
		}
		list[t.Id] = t
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
	}
	return
}

// RawNtPidToTsID .
func (d *Dao) RawNtPidToTsID(c context.Context, pid int64) (int64, error) {
	row := d.db.QueryRow(c, _tsLastIDSQL, pid)
	t := &v1.NativeTsPage{}
	if err := row.Scan(&t.Id); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		if err != nil {
			return 0, err
		}
	}
	return t.Id, nil
}

// RawNtPidToTsIDs
func (d *Dao) RawNtPidToTsIDs(c context.Context, pids []int64) (map[int64]int64, error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_tsLastIDsSQL, xstr.JoinInts(pids)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make(map[int64]int64)
	for rows.Next() {
		var pid, tsID sql.NullInt64
		if err = rows.Scan(&pid, &tsID); err != nil {
			err = errors.Wrap(err, "rows.Scan")
			return nil, err
		}
		list[pid.Int64] = tsID.Int64
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err")
		return nil, err
	}
	return list, nil
}

// TsPageSave .
func (d *Dao) TsPageSave(c context.Context, p *v1.NativeTsPage) (int64, error) {
	res, err := d.db.Exec(c, _tsAddPageSQL, p.Title, p.ForeignID, p.State, p.Pid, p.VideoDisplay, p.AuditType, p.AuditTime, p.ShareImage, p.Template)
	if err != nil {
		log.Error("TsPageSave arg:%v error(%v)", p, err)
		return 0, err
	}
	return res.LastInsertId()
}

func (d *Dao) UpdateVideoDisplay(c context.Context, id, auditTime int64, videoDisplay, shareImage string) error {
	if _, err := d.db.Exec(c, _tsDisplaySQL, videoDisplay, auditTime, shareImage, id); err != nil {
		log.Error("Fail to update video_display, id=%+v video_display=%+v audit_time=%+v share_image=%+v error=%+v", id, videoDisplay, auditTime, shareImage, err)
		return err
	}
	return nil
}

func (d *Dao) UpdateTsState(c context.Context, id, state int64, msg string) error {
	if _, err := d.db.Exec(c, _tsUpStateSQL, state, msg, id); err != nil {
		log.Error("Fail to update native_ts_page state, id=%+v state=%+v msg=%+v error=%+v", id, state, msg, err)
		return err
	}
	return nil
}
