package note

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/interface/model/note"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestNoteAidSetNX(t *testing.T) {
	c := context.Background()
	convey.Convey("NoteAidSetNX", t, func(ctx convey.C) {
		_, err := d.NoteAidSetNX(c, 1, &note.NoteAddReq{Oid: 222, OidType: 1})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestRemCacheNoteList(t *testing.T) {
	c := context.Background()
	val := make(map[int64]*notegrpc.SimpleNoteCard)
	val[1] = &notegrpc.SimpleNoteCard{
		NoteId: 1,
		Oid:    11,
		Mid:    12345,
	}
	val[2] = &notegrpc.SimpleNoteCard{
		NoteId: 2,
		Oid:    22,
		Mid:    12345,
	}
	val[3] = &notegrpc.SimpleNoteCard{
		NoteId: 3,
		Oid:    33,
		Mid:    12345,
	}
	convey.Convey("RemCacheNoteList", t, func(ctx convey.C) {
		err := d.RemCacheNoteList(c, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
