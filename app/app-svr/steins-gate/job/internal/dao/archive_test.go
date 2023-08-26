package dao

import (
	"context"
	"fmt"
	"go-gateway/app/app-svr/steins-gate/job/internal/model"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoArcView(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(10113448)
	)
	convey.Convey("ArcView", t, func(ctx convey.C) {
		res, err := d.ArcView(c, aid)
		fmt.Println(res, err)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoSendCid(t *testing.T) {
	var (
		ctx  = context.Background()
		acid = &model.SteinsCid{
			Aid: 123,
			Cid: 345,
		}
	)
	convey.Convey("SendCid", t, func(c convey.C) {
		d.UpArcFirstCid(ctx, acid.Aid, acid.Cid)
		time.Sleep(5 * time.Second)
		c.Convey("No return values", func(ctx convey.C) {
		})
	})
}
