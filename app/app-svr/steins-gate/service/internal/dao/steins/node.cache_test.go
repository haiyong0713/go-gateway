package steins

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestSteinsnodeKey(t *testing.T) {
	var (
		nodeID = int64(0)
	)
	convey.Convey("nodeKey", t, func(ctx convey.C) {
		p1 := nodeKey(nodeID)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestSteinsCacheNode(t *testing.T) {
	var (
		c      = context.Background()
		nodeID = int64(0)
	)
	convey.Convey("CacheNode", t, func(ctx convey.C) {
		res, err := d.CacheNode(c, nodeID)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestSteinsCacheNodes(t *testing.T) {
	var (
		c       = context.Background()
		nodeIDs = []int64{}
	)
	convey.Convey("CacheNodes", t, func(ctx convey.C) {
		res, err := d.CacheNodes(c, nodeIDs)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestSteinsAddCacheNode(t *testing.T) {
	var (
		c    = context.Background()
		nid  = int64(0)
		node = &api.GraphNode{}
	)
	convey.Convey("AddCacheNode", t, func(ctx convey.C) {
		err := d.AddCacheNode(c, nid, node)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestSteinsAddCacheNodes(t *testing.T) {
	var (
		c     = context.Background()
		nodes map[int64]*api.GraphNode
	)
	convey.Convey("AddCacheNodes", t, func(ctx convey.C) {
		err := d.AddCacheNodes(c, nodes)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
