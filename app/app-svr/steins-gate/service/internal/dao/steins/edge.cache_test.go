package steins

import (
	"context"
	"fmt"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoedgeKey(t *testing.T) {
	convey.Convey("edgeKey", t, func(convCtx convey.C) {
		var (
			edgeID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := edgeKey(edgeID)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoedgeByNodeKey(t *testing.T) {
	convey.Convey("edgeByNodeKey", t, func(convCtx convey.C) {
		var (
			nodeID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := edgeByNodeKey(nodeID)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoedgeAttrsKey(t *testing.T) {
	convey.Convey("edgeAttrsKey", t, func(convCtx convey.C) {
		var (
			graphID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := edgeAttrsKey(graphID)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoCacheEdge(t *testing.T) {
	convey.Convey("CacheEdge", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			edgeID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			ge, err := d.CacheEdge(c, edgeID)
			convCtx.Convey("Then err should be nil.ge should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(ge, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoCacheEdges(t *testing.T) {
	convey.Convey("CacheEdges", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			edgeIDs = []int64{77, 88}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheEdges(c, edgeIDs)
			fmt.Println(res, err)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddCacheEdge(t *testing.T) {
	convey.Convey("AddCacheEdge", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			eid  = int64(0)
			edge = &api.GraphEdge{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheEdge(c, eid, edge)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoAddCacheEdges(t *testing.T) {
	convey.Convey("AddCacheEdges", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			edges = make(map[int64]*api.GraphEdge)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			edges[1] = &api.GraphEdge{
				Id:      77,
				GraphId: 333,
			}
			edges[2] = &api.GraphEdge{
				Id:      88,
				GraphId: 444,
			}
			err := d.AddCacheEdges(c, edges)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoAddCacheEdgeAttrs(t *testing.T) {
	convey.Convey("AddCacheEdgeAttrs", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			graphID   = int64(4411)
			edgeAttrs = []*model.EdgeAttr{
				{
					FromNID:   1,
					ToNID:     2,
					Attribute: "123",
				},
			}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			edas := new(model.EdgeAttrsCache)
			edas.HasAttrs = true
			edas.EdgeAttrs = edgeAttrs
			err := d.AddCacheEdgeAttrs(c, graphID, edas)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoedgeByNodeCache(t *testing.T) {
	convey.Convey("edgeByNodeCache", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			nodeID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			ge, err := d.edgeByNodeCache(c, nodeID)
			fmt.Println(1111111, ge, err)
			convCtx.Convey("Then err should be nil.ge should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(ge, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaosetEdgeFromNodeCache(t *testing.T) {
	convey.Convey("setEdgeFromNodeCache", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			fromNodeID = int64(1)
			edgeIDs    = &model.EdgeFromCache{
				IsEnd:  false,
				ToEIDs: []int64{2, 3, 4},
			}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.setEdgeFromNodeCache(c, fromNodeID, edgeIDs)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
