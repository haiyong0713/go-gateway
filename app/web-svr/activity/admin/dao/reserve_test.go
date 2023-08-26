package dao

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoreserveName(t *testing.T) {
	Convey("reserveName", t, func() {
		var (
			sid = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := reserveName(sid)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddReserve(t *testing.T) {
	Convey("AddReserve", t, func() {
		var (
			c   = context.Background()
			sid = int64(0)
			mid = int64(0)
			num = int(0)
		)
		Convey("When everything goes positive", func() {
			err := d.AddReserve(c, sid, mid, num)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoSearchReserve(t *testing.T) {
	Convey("SearchReserve", t, func() {
		var (
			c   = context.Background()
			sid = int64(0)
			mid = int64(0)
			pn  = int(0)
			ps  = int(0)
		)
		Convey("When everything goes positive", func() {
			rly, err := d.SearchReserve(c, sid, mid, pn, ps)
			Convey("Then err should be nil.rly should not be nil.", func() {
				So(err, ShouldBeNil)
				So(rly, ShouldNotBeNil)
			})
		})
	})
}
