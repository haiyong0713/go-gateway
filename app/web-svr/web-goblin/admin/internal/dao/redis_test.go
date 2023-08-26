package dao

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoNewRedis(t *testing.T) {
	Convey("NewRedis", t, func() {
		Convey("When everything goes positive", func() {
			r, err := NewRedis()
			Convey("Then err should be nil.r should not be nil.", func() {
				So(err, ShouldBeNil)
				So(r, ShouldNotBeNil)
			})
		})
	})
}
