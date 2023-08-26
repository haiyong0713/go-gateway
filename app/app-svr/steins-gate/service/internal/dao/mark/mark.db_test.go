package mark

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaorawMark(t *testing.T) {
	convey.Convey("rawMark", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(1)
			mid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.rawMark(c, aid, mid)
			fmt.Println(res, err)
			convCtx.Convey("Then err should be nil. res should be greater than 0", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestDao_addMark(t *testing.T) {
	convey.Convey("addMark", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			aid  = int64(5601)
			mid  = int64(123)
			mark = int64(3)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.addMark(c, aid, mid, mark)
			convCtx.Convey("Then err should be nil", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaorawEvaluation(t *testing.T) {
	convey.Convey("rawMark", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(123)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.rawEvaluation(c, aid)
			fmt.Println(res, err)
			convCtx.Convey("Then err should be nil. res should be greater than 0", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestDao_addEvaluation(t *testing.T) {
	convey.Convey("addMark", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			aid        = int64(5601)
			evaluation = int64(343)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheEvaluation(c, aid, evaluation)
			convCtx.Convey("Then err should be nil", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
