package comic

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestComicComic(t *testing.T) {
	convey.Convey("Comic", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(65)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			data, err := d.ComicTitle(c, id)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestComicComicInfo(t *testing.T) {
	convey.Convey("Comic", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(65)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			data, err := d.ComicInfo(c, id)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(data)
			fmt.Println(string(bs))
		})
	})
}
