package dao

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoSearchUpList(t *testing.T) {
	Convey("SearchUpList", t, func() {
		var (
			c     = context.Background()
			uid   = int64(0)
			state = int64(0)
			pn    = int64(0)
			ps    = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.SearchUpList(c, uid, state, pn, ps)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoUpActEdit(t *testing.T) {
	Convey("UpActEdit", t, func() {
		var (
			c     = context.Background()
			id    = int64(1)
			state = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.UpActEdit(c, id, state, 0)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
