package archive

import (
	"context"
	"testing"

	"fmt"
	"github.com/smartystreets/goconvey/convey"
)

func TestVipInfo(t *testing.T) {
	var (
		c = context.TODO()
	)
	convey.Convey("VipInfo", t, func(ctx convey.C) {
		rly, err := d.VipInfo(c, 27515399, "", false)
		ctx.Convey("Then err should be nil.addit should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			fmt.Printf("%v", rly.Res.IsValid())
		})
	})
}
