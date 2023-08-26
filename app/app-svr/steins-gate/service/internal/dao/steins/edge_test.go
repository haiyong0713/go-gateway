package steins

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoEdges(t *testing.T) {
	convey.Convey("Edges", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			keys = []int64{1, 2, 3, 4, 5, 6, 7}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Edges(c, keys, nil)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoedges(t *testing.T) {
	convey.Convey("Edges", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			keys = []int64{1, 2, 3, 4, 5, 6, 7}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.edges(c, keys)
			fmt.Println(res)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoEdge(t *testing.T) {
	convey.Convey("Edge", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			key = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Edge(c, key, nil)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoEdgesByFromNode(t *testing.T) {
	convey.Convey("EdgesByFromNode", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			fromNodeID = int64(115)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			edges, err := d.EdgesByFromNode(c, fromNodeID)
			convCtx.Convey("Then err should be nil.edges should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(edges, convey.ShouldNotBeNil)
			})
			fmt.Println(edges)
			time.Sleep(2 * time.Second)
		})
	})
}
