package feed

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-card/interface/model/card/ai"

	"github.com/smartystreets/goconvey/convey"
)

func TestFeedkeyRcmd(t *testing.T) {
	convey.Convey("keyRcmd", t, func(ctx convey.C) {
		p1 := keyRcmd()
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestFeedAddRcmdCache(t *testing.T) {
	var (
		c  = context.Background()
		is = []*ai.Item{}
	)
	convey.Convey("AddRcmdCache", t, func(ctx convey.C) {
		err := d.AddRcmdCache(c, is)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
