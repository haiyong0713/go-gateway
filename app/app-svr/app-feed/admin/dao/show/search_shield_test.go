package show

import (
	"fmt"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestShowSearchShieldAdd(t *testing.T) {
	convey.Convey("SearchShieldAdd", t, func(ctx convey.C) {
		var (
			param = &show.SearchShieldAP{
				CardType:  1,
				CardValue: "10",
				Person:    "quguolin",
				Query:     "[{\"id\":7,\"value\":\"test1\"},{\"id\":8,\"value\":\"test2\"}]",
				Reason:    "test",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SearchShieldAdd(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowSearchShieldUpdate(t *testing.T) {
	convey.Convey("SearchShieldUpdate", t, func(ctx convey.C) {
		var (
			param = &show.SearchShieldUP{
				ID:        1,
				CardType:  1,
				CardValue: "100",
				Person:    "quguolin",
				Query:     "[{\"id\":1,\"value\":\"aaa\"},{\"id\":2,\"value\":\"bbb\"}]",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SearchShieldUpdate(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowSearchShieldOption(t *testing.T) {
	convey.Convey("SearchShieldOption", t, func(ctx convey.C) {
		var (
			up = &show.SearchShieldOption{
				ID:    1,
				Check: 1,
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SearchShieldOption(up)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestSearchShieldTimeValid(t *testing.T) {
	convey.Convey("SWTimeValid", t, func(ctx convey.C) {
		var (
			up = &show.SearchShieldValid{
				Query:     "test11",
				CardType:  1,
				CardValue: "100",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			count, err := d.SearchShieldValid(up)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			fmt.Println("count------")
			fmt.Println(count)
		})
	})
}
