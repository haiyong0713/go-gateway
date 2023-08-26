package dao

import (
	"testing"

	"go-gateway/app/app-svr/up-archive/service/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_CacheArcPassedExists(t *testing.T) {
	var (
		mid int64 = 15555180
	)
	Convey("RawArcPassed", t, func() {
		data, err := d.CacheArcPassedExists(ctx, mid, api.Without_none)
		So(err, ShouldBeNil)
		Printf("%v", data)
	})

}
