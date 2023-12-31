package archive

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestArchiveStat3(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(10097272)
	)
	convey.Convey("Stat3", t, func(ctx convey.C) {
		st, err := d.Stat3(c, aid)
		ctx.Convey("Then err should be nil.st should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(st, convey.ShouldNotBeNil)
		})
	})
}

func TestArchiveStats3(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{10097272}
	)
	convey.Convey("Stats3", t, func(ctx convey.C) {
		stm, err := d.Stats3(c, aids)
		ctx.Convey("Then err should be nil.stm should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(stm, convey.ShouldNotBeNil)
		})
	})
}

func TestArchiveInitStatCache3(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(10097272)
	)
	convey.Convey("InitStatCache3", t, func(ctx convey.C) {
		err := d.InitStatCache3(c, aid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
