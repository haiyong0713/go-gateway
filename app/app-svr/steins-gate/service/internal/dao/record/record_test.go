package record

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaogameRecord(t *testing.T) {
	var (
		c              = context.Background()
		mid            = int64(444)
		graphID        = int64(333)
		graphPreviewID = int64(14)
	)
	convey.Convey("recordByGraph", t, func(ctx convey.C) {
		res, err := d.RawRecord(c, mid, graphID, false)
		convey.Println("Not Preview", res)
		res1, err1 := d.RawRecord(c, mid, graphPreviewID, true)
		convey.Println("Preview", res1)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err1, convey.ShouldBeNil)
			ctx.So(res1, convey.ShouldNotBeNil)
		})
	})
}

func TestDaogameRecords(t *testing.T) {
	var (
		c   = context.Background()
		req = &model.RecordReq{
			GraphWithAID: make(map[int64]int64),
			MID:          111006313,
		}
	)
	convey.Convey("recordByGraph", t, func(ctx convey.C) {
		req.GraphWithAID[659] = 10113518
		req.GraphWithAID[631] = 10113631
		req.GraphWithAID[622] = 10113690
		req.GraphWithAID[790] = 10113622
		res, missAids, err := d.Records(c, req)
		str, _ := json.Marshal(res)
		convey.Println(string(str))
		convey.Println("missAids ", missAids) // 第四个graphID无法找到，返回aid
		convey.Println(err)
		time.Sleep(1 * time.Second)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaorecordKey(t *testing.T) {
	var (
		mid     = int64(1)
		graphID = int64(1)
	)
	convey.Convey("recordKey", t, func(ctx convey.C) {
		p1 := recordKey(mid, graphID, "123")
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoRecord(t *testing.T) {
	var (
		c       = context.Background()
		mid     = int64(2)
		graphID = int64(3)
	)
	convey.Convey("Record", t, func(ctx convey.C) {
		a, err := d.Record(c, mid, graphID, "111")
		convey.Println(a)
		// r_2_3_132
		ctx.Convey("Then err should be nil.a should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(a, convey.ShouldNotBeNil)
		})
	})
}

func TestDaosetRecordCache(t *testing.T) {
	var (
		c      = context.Background()
		record = &api.GameRecords{
			Mid:     2,
			Buvid:   "132",
			GraphId: 3,
			Choices: "1,,2,3,,4,4324",
		}
	)
	convey.Convey("setRecordCache", t, func(ctx convey.C) {
		convey.Println(recordKey(record.Mid, record.GraphId, record.Buvid))
		err := d.AddCacheRecord(c, record, &model.NodeInfoParam{
			MobiApp: "123",
			Buvid:   "111",
		})
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaorecordCache(t *testing.T) {
	var (
		c       = context.Background()
		mid     = int64(2)
		graphID = int64(3)
	)
	convey.Convey("recordCache", t, func(ctx convey.C) {
		a, err := d.CacheRecord(c, mid, graphID, "132")
		convey.Println(a, err)
		ctx.Convey("Then err should be nil.a should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(a, convey.ShouldNotBeNil)
		})
	})
}

func TestDaorecordByAid(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(444)
		aid = int64(2223)
	)
	convey.Convey("RecordByAid", t, func(ctx convey.C) {
		res, err := d.RecordByAid(c, mid, aid)
		convey.Println(res)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
