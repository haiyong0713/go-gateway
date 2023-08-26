package show

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoParamList(t *testing.T) {
	convey.Convey("ParamList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			list, err := d.ParamList(context.Background())
			ctx.Convey("Error should be nil, res should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(list, convey.ShouldNotBeNil)
				ctx.Printf("%+v", list)
			})
		})
	})
}
