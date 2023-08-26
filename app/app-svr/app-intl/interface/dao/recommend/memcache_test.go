package recommend

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRcmdCache(t *testing.T) {
	Convey(t.Name(), t, func() {
		res, err := d.RcmdCache(context.Background())
		Convey("Then isAtten should not be nil.", func() {
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
		})
	})
}
