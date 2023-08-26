package record

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoAddCacheRecord(t *testing.T) {
	var (
		c      = context.Background()
		record = &api.GameRecords{}
		params = &model.NodeInfoParam{}
	)
	convey.Convey("AddCacheRecord", t, func(ctx convey.C) {
		err := d.AddCacheRecord(c, record, params)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoCacheRecord(t *testing.T) {
	var (
		c       = context.Background()
		mid     = int64(0)
		graphID = int64(0)
		buvid   = ""
	)
	convey.Convey("CacheRecord", t, func(ctx convey.C) {
		a, err := d.CacheRecord(c, mid, graphID, buvid)
		ctx.Convey("Then err should be nil.a should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(a, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoCacheRecords(t *testing.T) {
	var (
		c        = context.Background()
		mid      = int64(8888)
		graphIDs = []int64{123, 124}
		buvid    = "test2"
	)
	convey.Convey("CacheRecords", t, func(ctx convey.C) {
		res, err := d.CacheRecords(c, mid, graphIDs, buvid)
		convey.Println(res, err)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoAddCacheRecords(t *testing.T) {
	var (
		c       = context.Background()
		records = make(map[int64]*api.GameRecords)
	)
	convey.Convey("AddCacheRecords", t, func(ctx convey.C) {
		records[0] = &api.GameRecords{
			Buvid:       "test1",
			GraphId:     123,
			CurrentNode: 111,
			Mid:         8888,
			Choices:     "test1的choices",
		}
		records[1] = &api.GameRecords{
			Buvid:       "test2",
			GraphId:     124,
			CurrentNode: 222,
			Mid:         8888,
			Choices:     "test2的choices",
		}
		err := d.AddCacheRecords(c, records)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
