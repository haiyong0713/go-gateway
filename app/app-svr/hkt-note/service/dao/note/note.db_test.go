package note

import (
	"context"
	"testing"

	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawNoteAid(t *testing.T) {
	c := context.Background()
	convey.Convey("rawNoteAid", t, func(ctx convey.C) {
		res, err := d.rawNoteAid(c, &notegrpc.NoteListInArcReq{})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestRawNoteDetail(t *testing.T) {
	c := context.Background()

	convey.Convey("rawNoteDetail", t, func(ctx convey.C) {
		noteId := int64(1)
		mid := int64(1)
		res, err := d.rawNoteDetail(c, noteId, mid)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestRawNoteDetails(t *testing.T) {
	c := context.Background()

	convey.Convey("rawNoteDetail", t, func(ctx convey.C) {
		noteIds := []int64{1, 2}
		mid := int64(1)
		res, err := d.rawNoteDetails(c, noteIds, mid)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestRawNoteContent(t *testing.T) {
	c := context.Background()

	convey.Convey("rawNoteContent", t, func(ctx convey.C) {
		noteId := int64(1)
		res, err := d.rawNoteContent(c, noteId)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestRawNoteUser(t *testing.T) {
	c := context.Background()

	convey.Convey("rawNoteContent", t, func(ctx convey.C) {
		mid := int64(1)
		res, err := d.rawNoteUser(c, mid)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestRawNoteList(t *testing.T) {
	c := context.Background()

	convey.Convey("rawNoteContent", t, func(ctx convey.C) {
		res, str, err := d.rawNoteList(c, 50, 3, 6)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
			ctx.So(len(str), convey.ShouldBeGreaterThan, 0)

		})
	})
}
