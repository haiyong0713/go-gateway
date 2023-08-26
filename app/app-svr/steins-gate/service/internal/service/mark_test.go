package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestService_AddMark(t *testing.T) {
	var (
		c    = context.Background()
		aid  = int64(123123)
		mid  = int64(1)
		mark = int64(5)
	)
	convey.Convey("AddMark", t, func(ctx convey.C) {
		err := s.AddMark(c, aid, mid, mark)
		fmt.Println(err)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestService_GetMark(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(5601)
		mid = int64(123)
	)
	convey.Convey("GetMark", t, func(ctx convey.C) {
		mark, err := s.GetMark(c, aid, mid)
		fmt.Println(err)
		ctx.Convey("Then err should be nil, mark should be greater than 0", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(mark, convey.ShouldBeGreaterThan, 0)
		})
	})
}
