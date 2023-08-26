package steins

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoLatestGraphList(t *testing.T) {
	var (
		c     = context.Background()
		aid   = int64(0)
		limit = int(0)
	)
	convey.Convey("LatestGraphList", t, func(ctx convey.C) {
		list, err := d.LatestGraphList(c, aid, limit)
		ctx.Convey("Then err should be nil.list should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(list, convey.ShouldNotBeNil)
		})
	})
}

func TestDaographByID(t *testing.T) {
	var (
		c       = context.Background()
		graphID = int64(0)
	)
	convey.Convey("graphByID", t, func(ctx convey.C) {
		data, err := d.graphByID(c, graphID)
		ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(data, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_graphWithStarting(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("graphByID", t, func(ctx convey.C) {
		data, err := d.graphWithStarting(c, 10114386, false)
		fmt.Println(data, err)
		ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(data, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoSaveGraph(t *testing.T) {
	var (
		c     = context.Background()
		param = &model.SaveGraphParam{}
	)
	convey.Convey("SaveGraph", t, func(ctx convey.C) {
		graphID, err := d.SaveGraph(c, 1, false, param, nil, 0)
		ctx.Convey("Then err should be nil.graphID should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(graphID, convey.ShouldNotBeNil)
		})
	})
}

func TestDaograph(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(10114549)
	)
	convey.Convey("graph", t, func(ctx convey.C) {
		a, err := d.graph(c, aid, false)
		fmt.Println("not preview", a)
		a1, err1 := d.graph(c, aid, true)
		fmt.Println("preview", a1)
		ctx.So(err, convey.ShouldBeNil)
		ctx.So(a, convey.ShouldNotBeNil)
		ctx.So(err1, convey.ShouldBeNil)
		ctx.So(a1, convey.ShouldNotBeNil)
	})
}
