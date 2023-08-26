package dao

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_CallPush(t *testing.T) {
	Convey("TestDao_PushTime", t, WithDao(func(d *Dao) {
		err := d.CallPush(context.Background())
		So(err, ShouldBeNil)
	}))
}
