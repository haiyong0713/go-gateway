package api

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var client UGCSeasonClient

func init() {
	var err error
	client, err = NewClient(nil)
	if err != nil {
		panic(err)
	}
}

func TestUpperList(t *testing.T) {
	convey.Convey("UpperList", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			req = &UpperListRequest{Mid: 27515255, PageSize: 10, PageNum: 1}
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.UpperList(c, req)
			ctx.So(err, convey.ShouldBeNil)
			for _, v := range reply.Seasons {
				ctx.Printf("%+v", v)
			}
		})
	})
}

func TestView(t *testing.T) {
	convey.Convey("View", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			req = &ViewRequest{SeasonID: 781}
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.View(c, req)
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v", reply.View)
		})
	})
}

func TestViews(t *testing.T) {
	convey.Convey("View", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			req = &ViewsRequest{SeasonIds: []int64{781}}
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.Views(c, req)
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v", reply.Views)
		})
	})
}
