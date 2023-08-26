package dao

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoListGateway(t *testing.T) {
	Convey("ListGateway", t, func() {
		var (
			ctx = context.Background()
		)
		Convey("When everything goes positive", func() {
			p1, err := d.ListGateway(ctx)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}
