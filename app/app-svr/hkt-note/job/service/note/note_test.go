package note

import (
	"context"
	"go-gateway/app/app-svr/hkt-note/job/model/note"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTreatNoteNotifyMsg(t *testing.T) {
	Convey("treatNoteNotifyMsg", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &note.NtAddMsg{
				Aid:      600024226,
				Mid:      27515242,
				NoteId:   11,
				Title:    "test title",
				Summary:  "test summary",
				NoteSize: 11,
				Mtime:    1595919076,
			}
		)
		s.treatNoteAddNotifyMsg(c, req)
	}))
}
