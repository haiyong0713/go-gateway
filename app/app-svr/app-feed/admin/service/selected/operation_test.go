package selected

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_PublishWeekly(t *testing.T) {
	Convey("Test_PublishWeekly", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			err := svf.PublishWeekly(c)
			So(err, ShouldBeNil)
		})
	})
}
