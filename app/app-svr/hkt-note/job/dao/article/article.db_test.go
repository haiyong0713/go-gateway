package article

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/job/model/article"
	"go-gateway/app/app-svr/hkt-note/job/model/note"

	"github.com/smartystreets/goconvey/convey"
)

func TestLatestPass(t *testing.T) {
	c := context.Background()
	convey.Convey("LatestPass", t, func(ctx convey.C) {
		val := &note.ReplyMsg{
			NoteId:  132455,
			Content: "asdffs",
			Oid:     12,
			Mid:     0,
		}
		_, _, _, err := d.LatestArtByNoteId(c, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUpPubStatus(t *testing.T) {
	c := context.Background()
	convey.Convey("UpPubStatus", t, func(ctx convey.C) {
		val := &article.ArtOriginalDB{
			Id:          4420,
			CategoryId:  100,
			State:       4,
			Reason:      "审核通过",
			Mid:         27515242,
			DeletedTime: 0,
			CheckTime:   "2021-04-07 14:29:51",
		}
		err := d.UpPubStatus(c, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDelArtDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("DelArtDetail", t, func(ctx convey.C) {
		err := d.DelArtDetail(c, 4420, 27515242)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestInsertArtContent(t *testing.T) {
	c := context.Background()
	convey.Convey("InsertArtContent", t, func(ctx convey.C) {
		val := &article.ArtContCache{
			Cvid:       4421,
			NoteId:     11241606125977614,
			Content:    `[{"attributes":{"background":"#f7f8fb","size":"14px","color":"#3f4a53"},"insert":"桃花2"},{"insert":"\n"},{"attributes":{"background":"#f7f8fb","size":"14px","color":"#3f4a53"},"insert":"桃花2桃花2"},{"insert":"\nwww我问问222\n"}]`,
			Tag:        "10234765-1-1-0",
			Deleted:    0,
			Mid:        27515242,
			ContLen:    10000,
			PubVersion: 2,
		}
		err := d.InsertArtContent(c, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestInsertArtDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("InsertArtDetail", t, func(ctx convey.C) {
		msg := &note.NtPubMsg{
			Cvid:     4421,
			Mid:      27515242,
			NoteId:   9970684826484742,
			ContLen:  100,
			Title:    "ceshiceshi",
			Summary:  "abcedf阿迪更柔软 sdf",
			Oid:      840016893,
			OidType:  0,
			ArcCover: "http://i1.hdslb.com/bfs/archive/6de59d2f7dc7273663363c1aa9acc2557cbf6597.jpg",
		}
		err := d.InsertArtDetail(c, msg)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArtDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("ArtDetail", t, func(ctx convey.C) {
		res, err := d.ArtDetail(c, 4512, article.TpArtDetailCvid, 0, article.PubStatusLock, false)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestArtCountInArc(t *testing.T) {
	c := context.Background()
	convey.Convey("ArtCountInArc", t, func(ctx convey.C) {
		res, err := d.ArtCountInArc(c, 320009175, 0)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}
