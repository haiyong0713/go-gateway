package elastic

import (
	"context"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestElasticSelResES(t *testing.T) {
	var (
		c   = context.Background()
		req = &selected.ReqSelES{
			Ps: 20,
			Pn: 1,
		}
	)
	convey.Convey("SelResES", t, func(ctx convey.C) {
		data, err := d.SelResES(c, req)
		ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(data, convey.ShouldNotBeNil)
		})
	})
}
