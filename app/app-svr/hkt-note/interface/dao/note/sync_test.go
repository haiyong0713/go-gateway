package note

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestBroadcastSync(t *testing.T) {
	c := context.Background()
	convey.Convey("BroadcastSync", t, func(ctx convey.C) {
		err := d.BroadcastSync(c, 1, "hash1")
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
