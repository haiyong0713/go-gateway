package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoVideoUpView(t *testing.T) {
	convey.Convey("VideoUpView", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			aid = int64(10113518)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			view, err := d.VideoUpView(c, aid)
			str, _ := json.Marshal(view)
			fmt.Println(string(str), err)
			ctx.Convey("Then err should be nil.view should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(view, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_VideoDimension(t *testing.T) {
	convey.Convey("VideoUpView", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			cid = int64(10165019)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			view, err := d.BvcDimension(c, cid)
			fmt.Println(view, err)
			ctx.Convey("Then err should be nil.view should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(view, convey.ShouldNotBeNil)
			})
		})
	})
}
