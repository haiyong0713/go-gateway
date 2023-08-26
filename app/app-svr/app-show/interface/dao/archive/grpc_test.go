package archive

import (
	"context"
	"testing"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestArchiveArcsWithPlayurl(t *testing.T) {
	var (
		c    = context.Background()
		aids = []*arcgrpc.PlayAv{{Aid: 10114610}}
	)
	convey.Convey("ArcsWithPlayurl", t, func(ctx convey.C) {
		res, err := dao.ArcsPlayer(c, aids)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
