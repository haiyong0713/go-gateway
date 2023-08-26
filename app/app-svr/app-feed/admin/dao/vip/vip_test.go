package vip

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestLiveAppMRoom(t *testing.T) {
	convey.Convey("LiveRoom", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			rs, err := d.Show(c, []int64{10000683})
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(rs, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(rs)
			fmt.Println(string(bs))
		})
	})
}
