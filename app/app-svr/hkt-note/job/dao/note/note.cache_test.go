package note

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/job/model/note"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddCacheNoteUser(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheNoteUser", t, func(ctx convey.C) {
		req := &note.UserCache{
			Mid:       52,
			NoteSize:  1,
			NoteCount: 1,
		}
		err := d.AddCacheNoteUser(c, 52, req)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestRemCacheNoteList(t *testing.T) {
	c := context.Background()
	convey.Convey("RemCacheNoteList", t, func(ctx convey.C) {
		var (
			mid = int64(50)
			val = "100-1"
		)
		err := d.RemCacheNoteList(c, mid, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddCacheNoteAid(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheNoteAid", t, func(ctx convey.C) {
		var (
			mid    = int64(50)
			aid    = int64(1)
			noteId = int64(1234567)
		)
		err := d.AddCacheNoteAid(c, mid, aid, noteId, 0)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
