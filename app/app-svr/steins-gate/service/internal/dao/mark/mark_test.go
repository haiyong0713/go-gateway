package mark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoMark(t *testing.T) {
	convey.Convey("Mark", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(5601)
			mid = int64(123)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Mark(c, aid, mid)
			fmt.Println(res, err)
			convCtx.Convey("Then err should be nil. res should be greater than 0", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestDaoEval(t *testing.T) {
	convey.Convey("Mark", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(10200126)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Evaluation(c, aid)
			fmt.Println(res, err)
			convCtx.Convey("Then err should be nil. res should be greater than 0", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestDaoAddMark(t *testing.T) {
	convey.Convey("AddMark", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			aid  = int64(5601)
			mid  = int64(123)
			mark = int64(8)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddMark(c, aid, mid, mark)
			convCtx.Convey("Then err should be nil", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		time.Sleep(3 * time.Second)
	})
}
