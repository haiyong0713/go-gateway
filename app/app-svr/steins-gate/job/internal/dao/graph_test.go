package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaographKey(t *testing.T) {
	var (
		aid = int64(1)
	)
	convey.Convey("graphKey", t, func(ctx convey.C) {
		p1 := graphKey(aid)
		fmt.Println(p1)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoGraph(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(10113448)
	)
	convey.Convey("Graph", t, func(ctx convey.C) {
		graph, err := d.Graph(c, aid)
		fmt.Println(graph, err)
		ctx.Convey("Then err should be nil.graph should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(graph, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoNodes(t *testing.T) {
	var (
		c       = context.Background()
		graphID = int64(4)
	)
	convey.Convey("Nodes", t, func(ctx convey.C) {
		result, firstCid, err := d.Nodes(c, graphID)
		qq, _ := json.Marshal(result)
		fmt.Println(string(qq), err)
		fmt.Println(firstCid)
		ctx.Convey("Then err should be nil.result should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoreturnGraph(t *testing.T) {
	var (
		c       = context.Background()
		graphID = int64(2)
	)
	convey.Convey("returnGraph", t, func(ctx convey.C) {
		err := d.returnGraph(c, graphID)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoReturnGraph(t *testing.T) {
	var (
		c       = context.Background()
		graphID = int64(0)
	)
	convey.Convey("ReturnGraph", t, func(ctx convey.C) {
		d.ReturnGraph(c, graphID)
		time.Sleep(5 * time.Second)
		ctx.Convey("No return values", func(ctx convey.C) {
		})
	})
}

func TestDaodelGraphCache(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(10113448)
	)
	convey.Convey("delGraphCache", t, func(ctx convey.C) {
		err := d.delGraphCache(c, aid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoDelGraphCache(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(10113448)
	)
	convey.Convey("DelGraphCache", t, func(ctx convey.C) {
		d.DelGraphCache(c, aid)
		time.Sleep(5 * time.Second)
		ctx.Convey("No return values", func(ctx convey.C) {
		})
	})
}
