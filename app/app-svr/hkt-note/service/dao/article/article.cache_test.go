package article

import (
	"context"
	"testing"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/hkt-note/service/model/article"

	"github.com/smartystreets/goconvey/convey"
)

func TestCacheArtCntInArc(t *testing.T) {
	c := context.Background()
	convey.Convey("CacheArtCntInArc", t, func(ctx convey.C) {
		res, err := d.cacheArtCntInArc(c, 222, 1)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestAddCacheArtCntInArc(t *testing.T) {
	c := context.Background()
	convey.Convey("addCacheArtCntInArc", t, func(ctx convey.C) {
		err := d.addCacheArtCntInArc(c, 222, 1, 0)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCacheArtDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("cacheArtDetail", t, func(ctx convey.C) {
		res, err := d.cacheArtDetail(c, 4420, "note_id")
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestAddCacheArtDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("addCacheArtDetail", t, func(ctx convey.C) {
		val := &article.ArtDtlCache{
			Cvid:      4423,
			NoteId:    3,
			Deleted:   0,
			Mid:       27515242,
			Oid:       843608664,
			PubStatus: 1,
		}
		err := d.addCacheArtDetail(c, 3, "note_id", val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddCacheArtContent(t *testing.T) {
	c := context.Background()
	convey.Convey("addCacheArtContent", t, func(ctx convey.C) {
		val := &article.ArtContCache{
			Cvid:    4421,
			NoteId:  9970684826484742,
			Content: `[{"attributes":{"background":"#f7f8fb","size":"14px","color":"#3f4a53"},"insert":"桃花2"},{"insert":"\n"},{"attributes":{"background":"#f7f8fb","size":"14px","color":"#3f4a53"},"insert":"桃花2桃花2"},{"insert":"\nwww我问问222\n"}]`,
			Tag:     "10234765-1-1-0",
			Deleted: 0,
			Mid:     27515242,
		}
		err := d.addCacheArtContent(c, 4420, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCacheArtContent(t *testing.T) {
	c := context.Background()
	convey.Convey("cacheArtContent", t, func(ctx convey.C) {
		res, err := d.cacheArtContent(c, 4420)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestCacheArtList(t *testing.T) {
	c := context.Background()
	convey.Convey("CacheArtList", t, func(ctx convey.C) {
		key := d.arcListKey(100, 0)
		res, err := d.cacheArtList(c, key, 0, 2)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestAddCacheArtList(t *testing.T) {
	c := context.Background()
	convey.Convey("addCacheArtList", t, func(ctx convey.C) {
		key := d.arcListKey(1, 0)
		val := []*article.ArtList{{Cvid: 4421, NoteId: 9970684826484742, Pubtime: xtime.Time(12345)}, {Cvid: 4420, NoteId: 10066914834907146, Pubtime: xtime.Time(123456)}}
		err := d.addCacheArtList(c, key, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCacheArtDetails(t *testing.T) {
	c := context.Background()
	convey.Convey("addCacheArtList", t, func(ctx convey.C) {
		res, _, err := d.cacheArtDetails(c, []int64{4421, 4422, 4423, 4424}, article.TpArtDetailCvid)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCacheArtListCount(t *testing.T) {
	c := context.Background()
	convey.Convey("cacheArtListCount", t, func(ctx convey.C) {
		res, err := d.cacheArtListCount(c, d.userListKey(1))
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldBeGreaterThanOrEqualTo, 0)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
