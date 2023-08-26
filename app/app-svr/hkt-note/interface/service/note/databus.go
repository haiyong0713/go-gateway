package note

import (
	"context"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"
	"strconv"
)

//nolint:bilirailguncheck
func (s *Service) sendNoteNotify(c context.Context, cont *note.NtNotifyMsg, mid int64) error {
	msg := &note.NtNotify{
		Topic:   note.TopicNtNotify,
		Content: cont,
	}
	return s.notePub.Send(c, strconv.FormatInt(mid, 10), msg)
}

//nolint:bilirailguncheck
func (s *Service) sendNoteAuditNotify(c context.Context, cont *note.NtAddMsg) error {
	msg := &note.NtAuditNotify{
		Topic:   note.TopicAuditNotify,
		Content: cont,
	}
	return s.noteAuditPub.Send(c, strconv.FormatInt(cont.NoteId, 10), msg)
}
