package bplus

import (
	"testing"
	"time"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"

	. "github.com/smartystreets/goconvey/convey"
)

// TestNotifyContribute dao ut.
func TestNotifyContribute(t *testing.T) {
	Convey("get DynamicCount", t, func() {
		var attrs *space.Attrs
		err := dao.NotifyContribute(ctx(), 27515258, attrs, xtime.Time(time.Now().Unix()), false)
		err = nil
		So(err, ShouldBeNil)
	})
}
