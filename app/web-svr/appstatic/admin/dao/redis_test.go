package dao

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_ZAddPush(t *testing.T) {
	Convey("TestDao_ZAddPush", t, WithDao(func(d *Dao) {
		err := d.AddPushTime(context.Background(), 123456789)
		So(err, ShouldBeNil)
	}))
}

func TestDao_PushTime(t *testing.T) {
	Convey("TestDao_PushTime", t, WithDao(func(d *Dao) {
		res, err := d.PushTime(context.Background())
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		fmt.Println(res)
	}))
}
