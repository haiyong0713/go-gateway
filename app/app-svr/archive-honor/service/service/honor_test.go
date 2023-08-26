package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestHonor(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(1)
	)
	convey.Convey("Honor", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			res, err := s.Honor(c, aid)
			fmt.Printf("%+v", res)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestHonorUpdate(t *testing.T) {
	var (
		c    = context.TODO()
		aid  = int64(2)
		typ  = int32(4)
		url  = ""
		desc = ""
	)
	convey.Convey("HonorUpdate", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			s.HonorUpdate(c, aid, typ, url, desc, "")
		})
	})
}

func TestHonorDel(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(1)
		typ = int32(1)
	)
	convey.Convey("HonorDel", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			s.HonorDel(c, aid, typ)
		})
	})
}
