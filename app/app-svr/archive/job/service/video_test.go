package service

import (
	"go-gateway/app/app-svr/archive/job/model/archive"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

//func TestServicesteinsGateConsumer(t *testing.T) {
//	convey.Convey("steinsGateConsumer", t, func(ctx convey.C) {
//		s.steinsGateConsumer()
//		ctx.Convey("No return values", func(ctx convey.C) {
//		})
//	})
//}

func TestServicetransStein(t *testing.T) {
	var (
		msg = &archive.SteinsCid{
			Aid: 10098500,
			Cid: 12345,
		}
	)
	convey.Convey("transStein", t, func(ctx convey.C) {
		oldResult, changed, err := s.transStein(msg)
		ctx.Convey("Then err should be nil.oldResult,changed should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(changed, convey.ShouldNotBeNil)
			ctx.So(oldResult, convey.ShouldNotBeNil)
		})
	})
}

func TestServicesteinsHandler(t *testing.T) {
	var (
		msg = &archive.SteinsCid{
			Aid: 10098500,
			Cid: 8940666,
		}
	)
	convey.Convey("steinsHandler", t, func(ctx convey.C) {
		s.steinsHandler(msg)
		ctx.Convey("No return values", func(ctx convey.C) {
		})
	})
}
