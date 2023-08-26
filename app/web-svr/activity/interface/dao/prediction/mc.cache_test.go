package prediction

import (
	"context"
	"fmt"
	"testing"

	premdl "go-gateway/app/web-svr/activity/interface/model/prediction"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddCachePredictions(t *testing.T) {
	convey.Convey("AddCachePredictions", t, func(ctx convey.C) {
		var (
			pre = make(map[int64]*premdl.Prediction)
			c   = context.Background()
		)
		pre[2] = &premdl.Prediction{ID: 2, Sid: 10293, Name: "更新测试", State: 1, Ctime: 1548057296, Mtime: 1548057296}
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddCachePredictions(c, pre)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCachePredictions(t *testing.T) {
	convey.Convey("CachePredictions", t, func(ctx convey.C) {
		var (
			ids = []int64{2}
			c   = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			list, err := d.CachePredictions(c, ids)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Print(list)
			})
		})
	})
}

func TestAddCachePredItems(t *testing.T) {
	convey.Convey("AddCachePredItems", t, func(ctx convey.C) {
		var (
			pre = make(map[int64]*premdl.PredictionItem)
			c   = context.Background()
		)
		pre[1] = &premdl.PredictionItem{ID: 2, Sid: 10293, Pid: 1, State: 1, Ctime: 1548057296, Mtime: 1548057296}
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddCachePredItems(c, pre)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCachePredItems(t *testing.T) {
	convey.Convey("CachePredItems", t, func(ctx convey.C) {
		var (
			ids = []int64{1}
			c   = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			list, err := d.CachePredItems(c, ids)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Print(list)
			})
		})
	})
}
