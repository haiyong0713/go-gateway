package note

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestFilterV3(t *testing.T) {
	c := context.Background()
	convey.Convey("FilterV3", t, func(ctx convey.C) {
		_, err := d.FilterV3(c, "抖腿,heiheihie抖音sf44refdd嘎嘎嘎dd", 1, 27515244)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
