package like

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeUserMatchCheck(t *testing.T) {
	convey.Convey("UserMatchCheck", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			mid  = int64(27515232)
			sids = []int64{10018, 10537, 10532}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			sid, err := d.UserMatchCheck(c, mid, sids)
			convCtx.Convey("Then err should be nil.sid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(sid, convey.ShouldNotBeNil)
				convCtx.Println(sid)
			})
		})
	})
}

func TestRawLikeMidTotal(t *testing.T) {
	convey.Convey("RawLikeMidTotal", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(27515232)
			sid = []int64{10018}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			total, err := d.RawLikeMidTotal(c, mid, sid)
			convCtx.Convey("Then err should be nil.sid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(sid, convey.ShouldNotBeNil)
				convCtx.Println(total)
			})
		})
	})
}

func TestResAudit(t *testing.T) {
	convey.Convey("ResAudit", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.ResAudit(c)
			convCtx.Convey("Then err should be nil.sid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", data)
			})
		})
	})
}

func TestSpecialData(t *testing.T) {
	convey.Convey("SpecialData", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.SpecialData(c, "http://activity.hdslb.com/blackboard/static/jsonlist/81/iLnyOVDMv.json", 1582717837)
			convCtx.Convey("Then err should be nil.sid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Printf("%+v", data)
			})
		})
	})
}
