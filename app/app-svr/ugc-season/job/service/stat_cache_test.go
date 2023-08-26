package service

import (
	"context"
	"go-gateway/app/app-svr/ugc-season/service/api"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestServicesnStatPBKey(t *testing.T) {
	convey.Convey("snStatPBKey", t, func(convCtx convey.C) {
		var (
			sid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := snStatPBKey(sid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestServiceupdateSnCache(t *testing.T) {
	convey.Convey("updateSnCache", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			st = &api.Stat{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := s.updateSnCache(c, st)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
