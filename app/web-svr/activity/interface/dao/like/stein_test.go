package like

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLikeSteinList(t *testing.T) {
	Convey("SteinList", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			data, err := d.SteinList(c)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(data)
			})
		})
	})
}
