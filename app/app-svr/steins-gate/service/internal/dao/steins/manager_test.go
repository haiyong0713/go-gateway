package steins

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

func TestDaoManagerList(t *testing.T) {
	convey.Convey("ManagerList", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.ManagerList(c, aid)
			convCtx.Convey("Then err should be nil.data should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoRecentArcs(t *testing.T) {
	convey.Convey("RecentArcs", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			param = &model.RecentArcReq{
				Stime: 1563206400,
				Etime: 1563261315,
			}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RecentArcs(c, param)
			fmt.Println(res, err)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
