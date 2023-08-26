package like

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLikeCheckTel(t *testing.T) {
	Convey("CheckTel", t, func() {
		var (
			c   = context.Background()
			mid = int64(15555180)
		)
		Convey("When everything goes positive", func() {
			res, err := d.CheckTel(c, mid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}
