package dao

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoGetNearbyLocationsTopK(t *testing.T) {
	Convey("GetNearbyLocationsTopK", t, func() {
		var (
			ctx = context.Background()
			k   = int(10)
			lat = float64(30.67807)
			lng = float64(104.151805)
		)
		Convey("When everything goes positive", func() {
			locations, err := d.GetNearbyLocationsTopK(ctx, k, lat, lng)
			Convey("Then err should be nil.locations should not be nil.", func() {
				So(err, ShouldBeNil)
				So(locations, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchLabs(t *testing.T) {
	Convey("SearchLabs", t, func() {
		var (
			ctx      = context.Background()
			word     = "成都"
			lat      = float64(30.67807)
			lng      = float64(104.151805)
			page     = int(0)
			pageSize = int(20)
		)
		Convey("When everything goes positive", func() {
			locations, _, err := d.SearchLabs(ctx, word, lat, lng, page, pageSize)
			Convey("Then err should be nil.locations should not be nil.", func() {
				So(err, ShouldBeNil)
				So(locations, ShouldNotBeNil)
			})
		})
	})
}
