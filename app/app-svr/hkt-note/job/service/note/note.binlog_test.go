package note

import (
	"context"
	"github.com/glycerine/goconvey/convey"
	"go-gateway/app/app-svr/hkt-note/job/model/note"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTreatDetailBinlog(t *testing.T) {
	Convey("treatDetailBinlog", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &note.NtDetailDB{
				NoteId:   3,
				Mid:      1,
				Aid:      10111784,
				NoteIdx:  0,
				NoteSize: 800,
				Title:    "aaaitle",
				Summary:  "aaaummary",
				Deleted:  0,
			}
		)
		s.treatNoteDetailBinlog(c, req)
	}))
}

func TestUpdateUser(t *testing.T) {
	Convey("updateUser", t, WithService(func(s *Service) {
		var (
			c = context.Background()
		)
		convey.Convey("DelNoteDetail", t, func(ctx convey.C) {
			err := s.updateUser(c, 1)
			ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	}))
}
