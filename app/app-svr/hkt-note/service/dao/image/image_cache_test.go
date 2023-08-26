package image

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/service/model/note"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddCacheImg(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheImg", t, func(ctx convey.C) {
		var (
			mid     = int64(100)
			imageId = int64(1)
			data    = &note.ImgInfo{Location: "test1", ImageId: 1}
		)
		err := d.AddCacheImg(c, mid, imageId, data)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCacheImg(t *testing.T) {
	c := context.Background()
	convey.Convey("CacheImg", t, func(ctx convey.C) {
		var (
			mid     = int64(1)
			imageId = int64(1)
		)
		res, err := d.cacheImg(c, mid, imageId)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestCacheImgs(t *testing.T) {
	c := context.Background()
	convey.Convey("CacheImg", t, func(ctx convey.C) {
		var (
			mid      = int64(100)
			imageIds = []int64{1, 2, 3, 4}
		)
		res, _, err := d.cacheImgs(c, mid, imageIds)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
