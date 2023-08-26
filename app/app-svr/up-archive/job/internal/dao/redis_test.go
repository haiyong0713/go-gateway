package dao

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_BuildArcPassedLock(t *testing.T) {
	var mid int64 = 2089809
	Convey("BuildArcPassedLock", t, func() {
		data, err := d.BuildArcPassedLock(ctx, mid)
		So(err, ShouldBeNil)
		Printf("%+v", data)
	})
}
