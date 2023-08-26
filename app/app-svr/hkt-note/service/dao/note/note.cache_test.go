package note

import (
	"context"
	"testing"

	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"
	"go-gateway/app/app-svr/hkt-note/service/model/note"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddCacheNoteAid(t *testing.T) {
	c := context.Background()
	convey.Convey("addCacheNoteAid", t, func(ctx convey.C) {
		noteId := []int64{12345, 67890}
		err := d.addCacheNoteAid(c, &notegrpc.NoteListInArcReq{}, noteId)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddCacheNoteDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("addCacheNoteDetail", t, func(ctx convey.C) {
		noteId := int64(1)
		val := &note.DtlCache{
			NoteId:   2,
			Oid:      12,
			Title:    "222222",
			Summary:  "shhuhummer test",
			NoteSize: 300,
			Deleted:  0,
			Mtime:    243234343,
		}
		err := d.AddCacheNoteDetail(c, noteId, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddCacheNoteContent(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheNoteContent", t, func(ctx convey.C) {
		noteId := int64(1)
		val := &note.ContCache{
			NoteId:  1,
			Content: "2435324532453",
			Tag:     "1-1-200,2-2-300,3-3-400",
			Deleted: 0,
		}
		err := d.AddCacheNoteContent(c, noteId, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCacheNoteDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheNoteContent", t, func(ctx convey.C) {
		noteId := int64(1)
		res, err := d.cacheNoteDetail(c, noteId)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestCacheNoteDetails(t *testing.T) {
	c := context.Background()
	convey.Convey("cacheNoteDetails", t, func(ctx convey.C) {
		noteIds := []int64{1, 2, 3}
		res, _, err := d.cacheNoteDetails(c, noteIds, 50)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestCacheNoteContent(t *testing.T) {
	c := context.Background()
	convey.Convey("cacheNoteContent", t, func(ctx convey.C) {
		noteId := int64(1)
		res, err := d.cacheNoteContent(c, noteId)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestAddCacheNoteUser(t *testing.T) {
	c := context.Background()
	convey.Convey("addCacheNoteUser", t, func(ctx convey.C) {
		mid := int64(1)
		req := &note.UserCache{
			Mid:       1,
			NoteSize:  1234532,
			NoteCount: 3,
		}
		err := d.addCacheNoteUser(c, mid, req)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCacheNoteUser(t *testing.T) {
	c := context.Background()
	convey.Convey("cacheNoteUser", t, func(ctx convey.C) {
		mid := int64(1)
		res, err := d.cacheNoteUser(c, mid)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestAddCacheAllNoteList(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheAllNoteList", t, func(ctx convey.C) {
		mid := int64(1)
		val := []*note.NtList{
			{Oid: 1, NoteId: 1, Mtime: 3},
			{Oid: 2, NoteId: 2, Mtime: 2},
			{Oid: 3, NoteId: 3, Mtime: 1},
		}
		err := d.AddCacheAllNoteList(c, mid, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCacheNoteList(t *testing.T) {
	c := context.Background()
	convey.Convey("CacheNoteList", t, func(ctx convey.C) {
		res, err := d.cacheNoteList(c, 12345, 1, 2, 12)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
