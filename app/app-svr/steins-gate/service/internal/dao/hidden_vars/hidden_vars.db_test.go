package hidden_vars

import (
	"context"
	"fmt"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaohvarsInfo(t *testing.T) {
	var (
		c         = context.Background()
		mid       = int64(27515257)
		graphInfo = &api.GraphInfo{}
		hvarReq   = &model.HvarReq{
			CurrentID: 8,
			Choices:   "4,6,7,8",
		}
	)
	convey.Convey("hvarsInfo", t, func(ctx convey.C) {
		a, err := d.hvarsInfo(c, mid, "", graphInfo, hvarReq)
		fmt.Println(a)
		ctx.Convey("Then err should be nil.a should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(a, convey.ShouldNotBeNil)
		})
	})
}
