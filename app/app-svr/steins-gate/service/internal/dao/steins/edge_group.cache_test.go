package steins

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoedgeGroupKey(t *testing.T) {
	convey.Convey("edgeKey", t, func(convCtx convey.C) {
		var (
			edgeID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := edgeGroupKey(edgeID)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoCacheEdgeGroup(t *testing.T) {
	convey.Convey("CacheEdgeGroup", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			ge, err := d.CacheEdgeGroup(c, id)
			convCtx.Convey("Then err should be nil.ge should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(ge, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoCacheEdgeGroups(t *testing.T) {
	convey.Convey("CacheEdgeGroups", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			IDs = []int64{77, 88}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheEdgeGroups(c, IDs)
			fmt.Println(res, err)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddCacheEdgeGroup(t *testing.T) {
	convey.Convey("AddCacheEdgeGroup", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			edgeGroup = &api.EdgeGroup{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheEdgeGroup(c, edgeGroup)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoAddCacheEdgeGroups(t *testing.T) {
	convey.Convey("AddCacheEdgeGroups", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			edgeGroups = make(map[int64]*api.EdgeGroup)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			edgeGroups[1] = &api.EdgeGroup{
				Id:      77,
				GraphId: 333,
			}
			edgeGroups[2] = &api.EdgeGroup{
				Id:      88,
				GraphId: 444,
			}
			err := d.AddCacheEdgeGroups(c, edgeGroups)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
