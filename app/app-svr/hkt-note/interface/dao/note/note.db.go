package note

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/hkt-note/interface/model/note"

	"github.com/pkg/errors"
)

const (
	_noteContentTb = "note_content"

	_updateNoteContent = "INSERT INTO %s(note_id,mid,content,tag) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE content=?,tag=?"
)

func (d *Dao) UpContent(c context.Context, val *note.NtContent) error {
	sql := fmt.Sprintf(_updateNoteContent, tableName(_noteContentTb, val.NoteId))
	if _, err := d.db.Exec(c, sql, val.NoteId, val.Mid, val.Content, val.Tag, val.Content, val.Tag); err != nil {
		return errors.Wrapf(err, "UpContent val(%+v)", val)
	}
	return nil
}

func tableName(table string, id int64) string {
	return fmt.Sprintf("%s_%02d", table, id%50)
}
