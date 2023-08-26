package service

import (
	"context"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestServicecheckBuilds(t *testing.T) {
	Convey("checkBuilds", t, func() {
		var (
			req    = &pb.PopUpsReq{}
			builds = ""
		)
		Convey("When everything goes positive", func() {
			valid, err := checkBuilds(req, builds)
			Convey("Then err should be nil.valid should not be nil.", func() {
				So(err, ShouldBeNil)
				So(valid, ShouldNotBeNil)
			})
		})
	})
}

func TestServicePopUps(t *testing.T) {
	Convey("PopUps", t, func() {
		var (
			c   = context.Background()
			req = &pb.PopUpsReq{}
		)
		Convey("When everything goes positive", func() {
			reply, err := s.PopUps(c, req)
			Convey("Then err should be nil.reply should not be nil.", func() {
				So(err, ShouldBeNil)
				So(reply, ShouldNotBeNil)
			})
		})
	})
}
