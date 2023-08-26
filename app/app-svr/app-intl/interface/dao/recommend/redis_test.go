package recommend

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPositionCache(t *testing.T) {
	Convey(t.Name(), t, func() {
		res, err := d.PositionCache(context.Background(), 12)
		Convey("Then isAtten should not be nil.", func() {
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
		})
	})
}

func TestAddPositionCache(t *testing.T) {
	Convey(t.Name(), t, func() {
		err := d.AddPositionCache(context.Background(), 12, 1)
		Convey("Then isAtten should not be nil.", func() {
			So(err, ShouldBeNil)
		})
	})
}
