package dao

import (
	"context"
	"fmt"
	"github.com/glycerine/goconvey/convey"
	"testing"
)

func TestDao_PickPlayerStats(t *testing.T) {
	convey.Convey("TestDao_PickPlayerStats", t, func(ctx convey.C) {
		res, err := d.PickPlayerStats(context.Background(), 98, 123, 0, 10)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(res, convey.ShouldNotBeNil)
		for _, a := range res {
			fmt.Println(a)
		}
	})
}
