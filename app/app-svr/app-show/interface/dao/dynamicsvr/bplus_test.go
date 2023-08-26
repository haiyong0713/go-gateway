package dynamicsvr

import (
	"context"
	"testing"

	dymdl "go-gateway/app/app-svr/app-show/interface/model/dynamic"

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
			_, err := d.Dynamic(c, resources, "ios", "", "", 0, nil)
			convCtx.Convey("Then err should be nil.dyResult should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDynamicsvrActiveUsers(t *testing.T) {
	convey.Convey("ActiveUsers", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			topicID = int64(914)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			httpMock("GET", d.activeUserURL).Reply(200).JSON(`{"code":0,"data":{"view_count":100,"discuss_count":3}}`)
			date, err := d.ActiveUsers(c, topicID, 1)
			convCtx.Convey("Then err should be nil.date should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(date, convey.ShouldNotBeNil)
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
			deviceID     = ""
			types        = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			httpMock("GET", d.feedDynamicURL).Reply(200).JSON(`{"code":0,"data":{"has_more":1}}`)
			reply, err := d.FetchDynamics(c, topicID, mid, frontpageNum, 0, deviceID, types, "iphone", "", "", "", "")
			convCtx.Convey("Then err should be nil.reply should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(reply, convey.ShouldNotBeNil)
			})
		})
	})
}

// BriefDynamics
func TestBriefDynamics(t *testing.T) {
	convey.Convey("BriefDynamics", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			topicID = int64(13527)
			mid     = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			httpMock("GET", d.feedDynamicURL).Reply(200).JSON(`{"code":0,"data":{"has_more":1}}`)
			reply, err := d.BriefDynamics(c, topicID, 6, mid, "8", "", 0)
			convCtx.Convey("Then err should be nil.reply should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(reply, convey.ShouldNotBeNil)
			})
		})
	})
}

// HasFeed
func TestHasFeed(t *testing.T) {
	convey.Convey("HasFeed", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			topicID = int64(13527)
			types   = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			httpMock("GET", d.feedDynamicURL).Reply(200).JSON(`{"code":0,"data":{"has_dyns":1}}`)
			reply, err := d.HasFeed(c, topicID, 0, types)
			convCtx.Convey("Then err should be nil.reply should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(reply, convey.ShouldNotBeNil)
			})
		})
	})
}
