package manager

import (
	"context"
	"testing"
	"time"

	xtime "go-common/library/time"

	"github.com/smartystreets/goconvey/convey"
)

func TestManagerSpecials(t *testing.T) {
	convey.Convey("Specials", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			sps, _, err := d.Specials(c, 0)
			ctx.Convey("Then err should be nil.sps should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(sps, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestGetSpecialCard(t *testing.T) {
	convey.Convey("TestGetSpecialCard", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			now    = xtime.Time(time.Now().AddDate(-2, 0, 0).Unix())
			offset = int64(1)
			size   = 5000
		)
		specials, nextId, err := d.GetSpecialCard(c, now, offset, size)

		t.Logf("specials(%+v)", specials)
		t.Logf("nextId(%d)", nextId)
		t.Logf("err(%+v)", err)

		ctx.So(specials, convey.ShouldNotBeNil)
		ctx.So(nextId, convey.ShouldNotBeNil)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestGetSpecialCardById(t *testing.T) {
	convey.Convey("TestGetSpecialCardById", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		res, err := d.GetSpecialCardById(c, id)

		t.Logf("res(%+v)", res)
		t.Logf("err(%+v)", err)

		ctx.So(res, convey.ShouldNotBeNil)
		ctx.So(res, convey.ShouldBeNil)
	})
}
