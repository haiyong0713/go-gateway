package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

func TestDao_Live(t *testing.T) {
	convey.Convey("test live", t, func(ctx convey.C) {
		defer gock.OffAll()
		httpMock("GET", d.liveURL).Reply(200).JSON(`{"code": 0}`)
		mid := int64(28272030)
		platform := "ios"
		data, err := d.Live(context.Background(), mid, platform)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%v", data)
	})
}
