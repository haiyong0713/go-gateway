package image

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestUpload(t *testing.T) {
	c := context.Background()
	convey.Convey("Upload", t, func(ctx convey.C) {
		mid := int64(2)
		res, err := d.NoteImgUpload(c, mid, "", nil)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDownload(t *testing.T) {
	c := context.Background()
	convey.Convey("Upload", t, func(ctx convey.C) {
		mid := int64(27515242)
		location := "/bfs/note/372f4ad9cb5e3bcd0361618ad1ef16663a255a72.jpg"
		res, fileType, err := d.NoteImgDownload(c, mid, location)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
			ctx.So(len(fileType), convey.ShouldNotBeEmpty)
		})
	})
}
