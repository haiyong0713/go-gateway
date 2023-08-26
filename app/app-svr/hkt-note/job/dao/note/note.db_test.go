package note

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/job/model/note"

	"github.com/smartystreets/goconvey/convey"
)

func TestDelNoteCont(t *testing.T) {
	c := context.Background()
	convey.Convey("DelNoteDetail", t, func(ctx convey.C) {
		err := d.DelNoteCont(c, 100)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDelNoteDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("DelNoteDetail", t, func(ctx convey.C) {
		err := d.DelNoteDetail(c, "100,101,102,103", 50)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestNoteUserData(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteUserData", t, func(ctx convey.C) {
		size, cnt, err := d.NoteUserData(c, 50)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(size, convey.ShouldBeGreaterThan, 0)
			ctx.So(cnt, convey.ShouldBeGreaterThan, 0)

		})
	})
}

func TestUpNoteUser(t *testing.T) {
	c := context.Background()
	convey.Convey("UpNoteUser", t, func(ctx convey.C) {
		val := &note.UserCache{
			Mid:       50,
			NoteSize:  20,
			NoteCount: 6,
		}
		err := d.UpNoteUser(c, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestNoteAid(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteAid", t, func(ctx convey.C) {
		noteId, err := d.NoteAid(c, 50, 2, 0)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(noteId, convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestUpContent(t *testing.T) {
	c := context.Background()
	convey.Convey("UpContent", t, func(ctx convey.C) {
		err := d.UpContent(c, "test", 100)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestNoteContent(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteContent", t, func(ctx convey.C) {
		res, err := d.NoteContent(c, 124343200)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestNoteDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteDetail", t, func(ctx convey.C) {
		res, err := d.NoteDetail(c, 101, 50)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
