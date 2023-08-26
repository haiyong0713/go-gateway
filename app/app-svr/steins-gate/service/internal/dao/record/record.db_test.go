package record

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoRawRecord(t *testing.T) {
	var (
		c       = context.Background()
		mid     = int64(0)
		graphID = int64(0)
		preview bool
	)
	convey.Convey("RawRecord", t, func(ctx convey.C) {
		res, err := d.RawRecord(c, mid, graphID, preview)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoRecordByAid(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(0)
		aid = int64(0)
	)
	convey.Convey("RecordByAid", t, func(ctx convey.C) {
		res, err := d.RecordByAid(c, mid, aid)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoAddRecord(t *testing.T) {
	var (
		c    = context.Background()
		rec1 = &api.GameRecords{
			CurrentNode: 1,
			Choices:     "1,2,3,4",
			GraphId:     1,
		}
		preview1 = false
		rec2     = &api.GameRecords{
			CurrentNode: 1,
			Choices:     "1,2,3,4,5,6",
			GraphId:     14,
		}
		preview2 = true
	)
	convey.Convey("AddRecord", t, func(ctx convey.C) {
		err1 := d.AddRecord(c, rec1, preview1)
		err2 := d.AddRecord(c, rec2, preview2)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err1, convey.ShouldBeNil)
			ctx.So(err2, convey.ShouldBeNil)
		})
	})
}

func TestDaoRecordByAids(t *testing.T) {
	var c = context.Background()
	convey.Convey("RecordByAid", t, func(ctx convey.C) {
		res, err := d.RecordByAids(c, 111006313, []int64{10113518, 10113523})
		fmt.Println(res, err)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoRawRecords(t *testing.T) {
	var c = context.Background()
	convey.Convey("TestDaoRawRecords", t, func(ctx convey.C) {
		res, err := d.RawRecords(c, 111006313, []int64{659, 631, 622}, "haha")
		fmt.Println(res, err)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
