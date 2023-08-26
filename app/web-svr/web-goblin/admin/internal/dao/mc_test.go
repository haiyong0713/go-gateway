package dao

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoNewMC(t *testing.T) {
	Convey("NewMC", t, func() {
		Convey("When everything goes positive", func() {
			mc, err := NewMC()
			Convey("Then err should be nil.mc should not be nil.", func() {
				So(err, ShouldBeNil)
				So(mc, ShouldNotBeNil)
			})
		})
	})
}
