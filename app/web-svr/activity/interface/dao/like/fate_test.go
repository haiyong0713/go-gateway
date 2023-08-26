package like

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLikeFateInfoCache(t *testing.T) {
	Convey("FateInfoCache", t, func() {
		var (
			c   = context.Background()
			key = ""
		)
		Convey("When everything goes positive", func() {
			data, err := d.FateInfoCache(c, key)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				So(data, ShouldNotBeNil)
			})
		})
	})
}

func TestLikeFateSwitchCache(t *testing.T) {
	Convey("FateSwitchCache", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			data, err := d.FateSwitchCache(c)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(data)
			})
		})
	})
}

func TestLikeFateConfCache(t *testing.T) {
	Convey("FateConfCache", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			data, err := d.FateConfCache(c)
			Convey("Then err should be nil.data should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(data)
			})
		})
	})
}
