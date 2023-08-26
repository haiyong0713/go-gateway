package dao

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoSearchMallItems(t *testing.T) {
	Convey("SearchMallItems", t, func() {
		var (
			ctx      = context.Background()
			mid      = uint64(14504142)
			word     = "手办"
			page     = 0
			pageSize = 20
		)
		Convey("When everything goes positive", func() {
			items, hasMore, err := d.SearchMallItems(ctx, mid, word, page, pageSize)
			Convey("Then err should be nil.items should not be nil.", func() {
				So(err, ShouldBeNil)
				So(items, ShouldNotBeNil)
				So(hasMore, ShouldBeTrue)
			})
		})
	})
}
