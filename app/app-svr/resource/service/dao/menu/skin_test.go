package menu

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

// Test_SideBar test dao side bar
func TestDaoRawSkinExts(t *testing.T) {
	convey.Convey("RawSkinExts", t, func(ctx convey.C) {
		ctx.Convey("When everyting is correct", func(ctx convey.C) {
			rly, err := d.RawSkinExts(context.Background(), time.Now())
			ctx.Convey("Error should be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				s, _ := json.Marshal(rly)
				ctx.Printf("%s", s)
			})
		})

	})
}

// RawSkinLimits
func TestRawSkinLimits(t *testing.T) {
	convey.Convey("RawSkinLimits", t, func(ctx convey.C) {
		ctx.Convey("When everyting is correct", func(ctx convey.C) {
			rly, err := d.RawSkinLimits(context.Background(), []int64{1, 2, 3})
			ctx.Convey("Error should be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				s, _ := json.Marshal(rly)
				ctx.Printf("%s", s)
			})
		})

	})
}
