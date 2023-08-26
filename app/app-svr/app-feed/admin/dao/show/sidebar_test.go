package show

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSideBarsByPlat(t *testing.T) {
	convey.Convey("SideBarsByPlat", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			res, err := d.SideBarsByPlat(context.Background(), 0)
			str, _ := json.Marshal(res)
			fmt.Println(string(str))
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestModulesByPlat(t *testing.T) {
	convey.Convey("ModulesByPlat", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			res, err := d.ModulesByPlat(context.Background(), 1)
			str, _ := json.Marshal(res)
			fmt.Println(string(str))
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
