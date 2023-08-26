package note

import (
	"context"
	"fmt"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/job/model/note"

	"github.com/pkg/errors"
)

const (
	_noteUserTb    = "note_user"
	_noteDetailTb  = "note_detail"
	_noteContentTb = "note_content"

	_updateNoteContent = "UPDATE %s SET content=? WHERE note_id=?"
	_updateNoteDetail  = "INSERT INTO %s(note_id,mid,aid,title,summary,note_size,audit_status,oid_type) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE title=?,summary=?,note_size=?,audit_status=?"
	_updateNoteUser    = "INSERT INTO %s(mid,note_size,note_count) VALUES (?,?,?) ON DUPLICATE KEY UPDATE note_size=?,note_count=?"
	_selectNoteUser    = "SELECT mid,note_size,note_count FROM %s WHERE mid=?"
	_selectNoteDetail  = "SELECT aid,title,summary,note_size,deleted,mtime,audit_status,mid FROM %s WHERE note_id=? AND deleted=0"
	_selectNoteAid     = "SELECT note_id FROM %s WHERE aid=? AND mid=? AND oid_type=? AND deleted=0 ORDER BY ctime DESC LIMIT 1"
	_selectNoteContent = "SELECT note_id,tag,content,deleted FROM %s WHERE note_id=? AND deleted=0"
	_selectNoteSize    = "SELECT note_size FROM %s WHERE mid=? AND deleted=0"
	_delNoteCont       = "UPDATE %s SET deleted=1 WHERE note_id=?"
	_delNoteDetail     = "UPDATE %s SET deleted=1 WHERE note_id IN (%s) AND mid=?"
)

func (d *Dao) UpContent(c context.Context, content string, noteId int64) error {
	sql := fmt.Sprintf(_updateNoteContent, tableName(_noteContentTb, noteId))
	if _, err := d.db.Exec(c, sql, content, noteId); err != nil {
		return errors.Wrapf(err, "UpContent content(%s) notId(%d)", content, noteId)
	}
	return nil
}

func (d *Dao) NoteUserData(c context.Context, mid int64) (size int64, count int64, err error) {
	selSql := fmt.Sprintf(_selectNoteSize, tableName(_noteDetailTb, mid))
	rows, err := d.db.Query(c, selSql, mid)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var oneSize int64
		if err = rows.Scan(&oneSize); err != nil {
			return 0, 0, err
		}
		size += oneSize
		count += 1
	}
	if err = rows.Err(); err != nil {
		return 0, 0, err
	}
	return size, count, nil
}

func (d *Dao) DelNoteCont(c context.Context, noteId int64) error {
	sql := fmt.Sprintf(_delNoteCont, tableName(_noteContentTb, noteId%50))
	_, err := d.db.Exec(c, sql, noteId)
	if err != nil {
		return errors.Wrapf(err, "DelNoteCont noteId(%d)", noteId)
	}
	return nil
}

func (d *Dao) DelNoteDetail(c context.Context, noteIdsStr string, mid int64) error {
	sql := fmt.Sprintf(_delNoteDetail, tableName(_noteDetailTb, mid%50), noteIdsStr)
	res, err := d.db.Exec(c, sql, mid)
	if err != nil {
		return errors.Wrapf(err, "DelNoteDetail noteIds(%s) mid(%d)", noteIdsStr, mid)
	}
	if rows, e := res.RowsAffected(); e != nil || rows == 0 {
		log.Warn("noteInfo DelNoteDetail sql(%s) mid(%d) rowsAffected(%d) err(%+v)", sql, mid, rows, e)
	}
	return nil
}

func (d *Dao) UpNoteDetail(c context.Context, val *note.NtAddMsg) error {
	sql := fmt.Sprintf(_updateNoteDetail, tableName(_noteDetailTb, val.Mid%50))
	res, err := d.db.Exec(c, sql, val.NoteId, val.Mid, val.Oid, val.Title, val.Summary, val.NoteSize, val.AuditStatus, val.OidType, val.Title, val.Summary, val.NoteSize, val.AuditStatus)
	if err != nil {
		return errors.Wrapf(err, "UpNoteDetail val(%+v)", val)
	}
	if rows, e := res.RowsAffected(); e != nil || rows == 0 {
		log.Warn("noteInfo UpNoteDetail val(%+v) rows(%d) err(%+v)", val, rows, e)
	}
	return nil
}

func (d *Dao) NoteContent(c context.Context, noteId int64) (*note.ContCache, error) {
	res := &note.ContCache{}
	selSql := fmt.Sprintf(_selectNoteContent, tableName(_noteContentTb, noteId))
	row := d.db.QueryRow(c, selSql, noteId)
	if err := row.Scan(&res.NoteId, &res.Tag, &res.Content, &res.Deleted); err != nil {
		if err == sql.ErrNoRows {
			res.NoteId = -1
			return res, nil
		}
		return nil, errors.Wrapf(err, "rawNoteContent noteId(%d)", noteId)
	}
	return res, nil
}

func (d *Dao) NoteDetail(c context.Context, noteId int64, mid int64) (*note.DtlCache, error) {
	res := &note.DtlCache{}
	selSql := fmt.Sprintf(_selectNoteDetail, tableName(_noteDetailTb, mid))
	row := d.db.QueryRow(c, selSql, noteId)
	err := row.Scan(&res.Aid, &res.Title, &res.Summary, &res.NoteSize, &res.Deleted, &res.Mtime, &res.AuditStatus, &res.Mid)
	if err != nil {
		if err == sql.ErrNoRows {
			res.NoteId = -1
			return res, nil
		}
		err = errors.Wrapf(err, "rawNoteDetail noteId(%d)", noteId)
		return nil, err
	}
	return res, nil
}

func (d *Dao) NoteAid(c context.Context, mid int64, aid int64, oidType int) (int64, error) {
	selSql := fmt.Sprintf(_selectNoteAid, tableName(_noteDetailTb, mid))
	row := d.db.QueryRow(c, selSql, aid, mid, oidType)
	var noteId int64
	err := row.Scan(&noteId)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		err = errors.Wrapf(err, "rawNoteDetail mid(%d) aid(%d)", mid, aid)
		return 0, err
	}
	return noteId, nil
}

func (d *Dao) UpNoteUser(c context.Context, val *note.UserCache) error {
	sql := fmt.Sprintf(_updateNoteUser, tableName(_noteUserTb, val.Mid%50))
	if _, err := d.db.Exec(c, sql, val.Mid, val.NoteSize, 1, val.NoteSize, val.NoteCount); err != nil {
		return errors.Wrapf(err, "UpNoteUser val(%+v)", val)
	}
	return nil
}

func (d *Dao) RawNoteUser(c context.Context, mid int64) (*note.UserCache, error) {
	res := &note.UserCache{}
	selSql := fmt.Sprintf(_selectNoteUser, tableName(_noteUserTb, mid%50))
	row := d.db.QueryRow(c, selSql, mid)
	err := row.Scan(&res.Mid, &res.NoteSize, &res.NoteCount)
	if err != nil {
		if err == sql.ErrNoRows {
			res.Mid = -1
			return res, nil
		}
		err = errors.Wrapf(err, "RawNoteUser mid(%d)", mid)
		return nil, err
	}
	return res, nil
}

func tableName(table string, id int64) string {
	return fmt.Sprintf("%s_%02d", table, id%50)
}
