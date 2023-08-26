package note

import (
	"context"
	"fmt"

	"go-common/library/xstr"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/note"

	"github.com/pkg/errors"
)

const (
	_noteDetailTb  = "note_detail"
	_noteContentTb = "note_content"
	_noteUserTb    = "note_user"

	_selectNoteAid     = "SELECT note_id FROM %s WHERE aid=? AND mid=? AND deleted=0 AND oid_type=? ORDER BY ctime DESC LIMIT 1"
	_selectNoteList    = "SELECT note_id,aid,mtime FROM %s WHERE mid=? AND deleted=0 ORDER BY mtime DESC"
	_selectNoteUser    = "SELECT mid,note_size,note_count FROM %s WHERE mid=?"
	_selectNoteContent = "SELECT note_id,tag,content,deleted FROM %s WHERE note_id=? AND deleted=0"
	_selectNoteDetail  = "SELECT note_id,aid,title,summary,note_size,deleted,mtime,audit_status,mid,oid_type FROM %s WHERE note_id=? AND mid=? AND deleted=0"
	_selectNoteDetails = "SELECT note_id,aid,title,summary,note_size,deleted,mtime,audit_status,mid,oid_type FROM %s WHERE note_id IN (%s) AND mid=? AND deleted=0"
)

func (d *Dao) rawNoteAid(c context.Context, req *notegrpc.NoteListInArcReq) ([]int64, error) {
	selSql := fmt.Sprintf(_selectNoteAid, tableName(_noteDetailTb, req.Mid))
	rows, err := d.dbr.Query(c, selSql, req.Oid, req.Mid, req.OidType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]int64, 0)
	for rows.Next() {
		var noteId int64
		if err = rows.Scan(&noteId); err != nil {
			return nil, err
		}
		res = append(res, noteId)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) rawNoteDetail(c context.Context, noteId int64, mid int64) (*note.DtlCache, error) {
	res := &note.DtlCache{}
	selSql := fmt.Sprintf(_selectNoteDetail, tableName(_noteDetailTb, mid))
	row := d.dbr.QueryRow(c, selSql, noteId, mid)
	err := row.Scan(&res.NoteId, &res.Oid, &res.Title, &res.Summary, &res.NoteSize, &res.Deleted, &res.Mtime, &res.AuditStatus, &res.Mid, &res.OidType)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) rawNoteDetails(c context.Context, noteIds []int64, mid int64) (map[int64]*note.DtlCache, error) {
	selSql := fmt.Sprintf(_selectNoteDetails, tableName(_noteDetailTb, mid), xstr.JoinInts(noteIds))
	rows, err := d.dbr.Query(c, selSql, mid)
	if err != nil {
		err = errors.Wrapf(err, "rawNoteDetails noteIds(%v) mid(%d)", noteIds, mid)
		return nil, err
	}
	defer rows.Close()
	res := make(map[int64]*note.DtlCache)
	for rows.Next() {
		tmp := &note.DtlCache{}
		if err = rows.Scan(&tmp.NoteId, &tmp.Oid, &tmp.Title, &tmp.Summary, &tmp.NoteSize, &tmp.Deleted, &tmp.Mtime, &tmp.AuditStatus, &tmp.Mid, &tmp.OidType); err != nil {
			err = errors.Wrapf(err, "rawNoteDetails noteIds(%v) mid(%d)", noteIds, mid)
			return nil, err
		}
		res[tmp.NoteId] = tmp
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrapf(err, "rawNoteDetails noteIds(%v) mid(%d)", noteIds, mid)
		return nil, err
	}
	return res, nil
}

func (d *Dao) rawNoteContent(c context.Context, noteId int64) (*note.ContCache, error) {
	res := &note.ContCache{}
	selSql := fmt.Sprintf(_selectNoteContent, tableName(_noteContentTb, noteId))
	row := d.dbr.QueryRow(c, selSql, noteId)
	if err := row.Scan(&res.NoteId, &res.Tag, &res.Content, &res.Deleted); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) rawNoteUser(c context.Context, mid int64) (*note.UserCache, error) {
	res := &note.UserCache{}
	selSql := fmt.Sprintf(_selectNoteUser, tableName(_noteUserTb, mid))
	row := d.dbr.QueryRow(c, selSql, mid)
	err := row.Scan(&res.Mid, &res.NoteSize, &res.NoteCount)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) rawNoteList(c context.Context, mid, min, max int64) ([]*note.NtList, []string, error) {
	selSql := fmt.Sprintf(_selectNoteList, tableName(_noteDetailTb, mid))
	rows, err := d.dbr.Query(c, selSql, mid)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var (
		res    = make([]*note.NtList, 0)
		strArr = make([]string, 0)
	)
	for rows.Next() {
		r := &note.NtList{}
		if err = rows.Scan(&r.NoteId, &r.Oid, &r.Mtime); err != nil {
			err = errors.Wrapf(err, "rawNoteList mid(%d)", mid)
			return nil, nil, err
		}
		res = append(res, r)
		strArr = append(strArr, fmt.Sprintf("%d-%d", r.NoteId, r.Oid))
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	// 根据min,max截断[]string
	cutStrArr := func() []string {
		if min == 0 && max == -1 {
			return strArr
		}
		if int(min) >= len(strArr) {
			return nil
		}
		if int(max) >= len(strArr) {
			return strArr[min:]
		}
		return strArr[min : max+1]
	}()
	return res, cutStrArr, nil
}

func tableName(table string, id int64) string {
	return fmt.Sprintf("%s_%02d", table, id%50)
}
