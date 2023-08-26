package note

import (
	"context"
	"testing"

	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestNoteSize(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteSize", t, func(ctx convey.C) {
		mid := int64(1)
		res, err := d.NoteSize(c, 1, mid)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestSeqId(t *testing.T) {
	c := context.Background()
	convey.Convey("SeqId", t, func(ctx convey.C) {
		id, err := d.SeqId(c)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(id, convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestSimpleNotes(t *testing.T) {
	c := context.Background()
	convey.Convey("SimpleNotes", t, func(ctx convey.C) {
		res, err := d.SimpleNotes(c, []int64{8393830995853322, 8471210397007886, 8477311637127174}, 111004103, notegrpc.SimpleNoteType_PUBLISH)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestSeason(t *testing.T) {
	c := context.Background()
	convey.Convey("Season", t, func(ctx convey.C) {
		res, err := d.cheeseSeason(c, 81)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestNoteList(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteList", t, func(ctx convey.C) {
		res, err := d.NoteList(c, 27515232, 1, 10, 0, 0, 0, notegrpc.NoteListType_USER_PUBLISHED)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestArcsForbid(t *testing.T) {
	c := context.Background()
	convey.Convey("ArcsForbid", t, func(ctx convey.C) {
		res, err := d.ArcsForbid(c, []int64{760056315, 840086693})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
