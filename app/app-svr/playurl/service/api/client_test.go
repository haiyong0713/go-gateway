package api

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var client PlayURLClient

func init() {
	var err error
	client, err = NewClient(nil)
	if err != nil {
		panic(err)
	}
}

func TestPlayURL(t *testing.T) {
	var (
		c   = context.TODO()
		req = &PlayURLReq{
			Aid:      10112614,
			Cid:      10154746,
			Platform: "pc",
			Mid:      27515255,
		}
	)
	convey.Convey("PlayURL", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.PlayURL(c, req)
			ctx.Printf("%+v", reply)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSteinsPreview(t *testing.T) {
	var (
		c   = context.TODO()
		req = &SteinsPreviewReq{
			Aid:      10112614,
			Cid:      10154746,
			Platform: "pc",
			Mid:      27515255,
		}
	)
	convey.Convey("SteinsPreview", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.SteinsPreview(c, req)
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v\n", reply)
		})
	})
}
