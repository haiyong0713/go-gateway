package like

import (
	"context"
	"encoding/json"
	"testing"

	"fmt"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeLotteryIndex(t *testing.T) {
	convey.Convey("LotteryIndex", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			actID    = int64(317)
			platform = int64(1)
			source   = int64(1)
			mid      = int64(27515233)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LotteryIndex(c, actID, platform, source, mid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				fmt.Printf("%+v", err)
				str, _ := json.Marshal(res)
				println(string(str))
				fmt.Printf("%+v", res.Data)
			})
		})
	})
}

func TestAddLotteryTimes(t *testing.T) {
	convey.Convey("LotteryIndex", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(1)
			mid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddLotteryTimes(c, sid, mid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				fmt.Printf("%+v", err)
			})
		})
	})
}
