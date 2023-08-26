package tag

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestTagsByIDs(t *testing.T) {
	convey.Convey("TagsInfoByIDs", t, func(ctx convey.C) {
		var (
			c    = context.TODO()
			mid  int64
			tids []int64
		)
		_, err := d.TagsInfoByIDs(c, mid, tids)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
