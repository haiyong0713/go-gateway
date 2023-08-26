package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_DynamicNum(t *testing.T) {
	convey.Convey("test dynamic cnt", t, func(ctx convey.C) {
		vmid := int64(88895254)
		data, err := d.DynamicNum(context.Background(), vmid)
		convey.So(err, convey.ShouldBeNil)
		convey.Printf("%d", data)
	})
}
