package note

import (
	"context"
	"testing"

	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestNoteAid(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteAid", t, func(ctx convey.C) {
		res, err := d.NoteAid(c, &notegrpc.NoteListInArcReq{})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestNoteDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("addCacheNoteDetail", t, func(ctx convey.C) {
		noteId := int64(3)
		mid := int64(1)
		res, err := d.NoteDetail(c, noteId, mid)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestNoteContent(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteContent", t, func(ctx convey.C) {
		noteId := int64(52)
		res, err := d.NoteContent(c, noteId)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestNoteUser(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteUser", t, func(ctx convey.C) {
		mid := int64(52)
		res, err := d.NoteUser(c, mid)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestNoteDetails(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteDetails", t, func(ctx convey.C) {
		noteIds := []int64{103, 104, 105, 124343200}
		mid := int64(50)
		res, err := d.NoteDetails(c, noteIds, mid)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestNoteList(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteList", t, func(ctx convey.C) {
		res, err := d.NoteList(c, 50, 1, 3, 12)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
		})
	})
}
