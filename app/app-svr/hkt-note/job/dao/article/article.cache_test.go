package article

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/job/model/article"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddCacheArtContent(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheArtContent", t, func(ctx convey.C) {
		val := &article.ArtContCache{
			Cvid:    4420,
			NoteId:  9970684826484742,
			Content: `[{"attributes":{"background":"#f7f8fb","size":"14px","color":"#3f4a53"},"insert":"桃花2"},{"insert":"\n"},{"attributes":{"background":"#f7f8fb","size":"14px","color":"#3f4a53"},"insert":"桃花2桃花2"},{"insert":"\nwww我问问222\n"}]`,
			Tag:     "10234765-1-1-0",
			Deleted: 0,
			Mid:     27515242,
		}
		err := d.AddCacheArtContent(c, 4420, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddCacheArtDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheArtDetail", t, func(ctx convey.C) {
		val := &article.ArtDtlCache{Cvid: -1, NoteId: 0, Mid: 0, PubStatus: 4, PubReason: "", Oid: 0, OidType: 0, Pubtime: 0, Mtime: 0, Title: "", Summary: "", Deleted: 0}
		err := d.AddCacheArtDetail(c, 4437, val, article.TpArtDetailNoteId, 1000)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddCacheArtCntInArc(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheArtCntInArc", t, func(ctx convey.C) {
		err := d.AddCacheArtCntInArc(c, 1, 0, 100)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
