package v1

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var client ResourceClient

func init() {
	var err error
	client, err = NewClient(nil)
	if err != nil {
		panic(err)
	}
}

func TestEntrancesIsHidden(t *testing.T) {
	convey.Convey("EntrancesIsHidden", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.EntrancesIsHidden(c, &EntrancesIsHiddenRequest{})
			ctx.So(err, convey.ShouldBeNil)
			for _, v := range reply.Infos {
				ctx.Printf("%+v", v)
			}
		})
	})
}

func TestMineSections(t *testing.T) {
	convey.Convey("MineSections", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.MineSections(c, &MineSectionsRequest{
				Plat:  1,
				Build: 999999,
				Mid:   0,
				Lang:  "hans",
			})
			ctx.So(err, convey.ShouldBeNil)
			for _, v := range reply.Sections {
				ctx.Printf("----%+v", v)
			}
		})
	})
}
