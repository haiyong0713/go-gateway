package steins

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoRawEdge(t *testing.T) {
	var (
		c  = context.Background()
		id = int64(0)
	)
	convey.Convey("RawEdge", t, func(ctx convey.C) {
		edge, err := d.RawEdge(c, id)
		ctx.Convey("Then err should be nil.edge should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(edge, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoRawEdges(t *testing.T) {
	var (
		c   = context.Background()
		ids = []int64{}
	)
	convey.Convey("RawEdges", t, func(ctx convey.C) {
		res, err := d.RawEdges(c, ids)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoedgeByNode(t *testing.T) {
	var (
		c        = context.Background()
		fromNode = int64(0)
	)
	convey.Convey("edgeByNode", t, func(ctx convey.C) {
		res, err := d.edgeByNode(c, fromNode)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoGraphEdgeList(t *testing.T) {
	var (
		c       = context.Background()
		graphID = int64(0)
	)
	convey.Convey("GraphEdgeList", t, func(ctx convey.C) {
		edges, err := d.GraphEdgeList(c, graphID, true)
		ctx.Convey("Then err should be nil.edges should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(edges, convey.ShouldNotBeNil)
		})
	})
}
