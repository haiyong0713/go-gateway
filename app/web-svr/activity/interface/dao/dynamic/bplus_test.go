package dynamic

import (
	"context"
	"testing"

	dymdl "go-gateway/app/web-svr/activity/interface/model/dynamic"

	"github.com/smartystreets/goconvey/convey"
)

func TestDynamicsvrDynamic(t *testing.T) {
	convey.Convey("Dynamic", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			resources = &dymdl.Resources{Array: []*dymdl.RidInfo{{Rid: 10113001, Type: 8}}}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			httpMock("GET", d.dynamicInfoURL).Reply(200).JSON(`{"code":0,"data":{}}`)
			_, err := d.Dynamic(c, resources, 0)
			convCtx.Convey("Then err should be nil.dyResult should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDynamicsvrFetchDynamics(t *testing.T) {
	convey.Convey("FetchDynamics", t, func(convCtx convey.C) {
		var (
			c            = context.Background()
			topicID      = int64(13527)
			frontpageNum = int64(1)
			mid          = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			httpMock("GET", d.feedDynamicURL).Reply(200).JSON(`{"code":0,"data":{"has_more":1}}`)
			reply, err := d.FetchDynamics(c, topicID, mid, frontpageNum, "", "8", 0)
			convCtx.Convey("Then err should be nil.reply should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(reply, convey.ShouldNotBeNil)
			})
		})
	})
}
