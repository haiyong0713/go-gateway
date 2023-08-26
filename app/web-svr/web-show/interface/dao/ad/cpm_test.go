package ad

import (
	"context"
	"testing"

	ad "go-gateway/app/web-svr/web-show/interface/model/resource"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_Cpms(t *testing.T) {
	Convey("test Cpms", t, WithDao(func(d *Dao) {
		mid := int64(5187977)
		ids := []int64{121, 21, 12}
		val := ad.CpmsRequestParam{Mid: mid, Ids: ids}
		data, err := d.Cpms(context.TODO(), val)
		So(err, ShouldBeNil)
		Printf("%+v", data)
	}))
}
