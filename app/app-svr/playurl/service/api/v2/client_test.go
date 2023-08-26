package v2

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
			Platform: "android",
			Mid:      27515255,
			MobiApp:  "android",
			Device:   "android",
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

func TestProject(t *testing.T) {
	var (
		c   = context.TODO()
		req = &ProjectReq{
			Aid:      880078582,
			Cid:      10176578,
			Platform: "android",
			Mid:      27515255,
			MobiApp:  "android",
			Device:   "android",
		}
	)
	convey.Convey("Project", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.Project(c, req)
			ctx.Printf("%+v", reply)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
