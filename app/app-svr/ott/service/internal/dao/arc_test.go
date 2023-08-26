package dao

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoInsertAcc(t *testing.T) {
	convey.Convey("InsertAcc", t, func(ctx convey.C) {
		res, err := d.SimpleArchives(context.Background(), []int64{10318755, 10318753})
		ctx.So(res, convey.ShouldNotBeNil)
		ctx.So(err, convey.ShouldBeNil)
	})
}
