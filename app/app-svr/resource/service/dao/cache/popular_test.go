package cache

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCacheAIChannelRes(t *testing.T) {
	convey.Convey("When cpm returns code = 0", t, func(ctx convey.C) {
		_, err := d.CacheAIChannelRes(context.Background(), 20)
		ctx.Convey("Then Error should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
