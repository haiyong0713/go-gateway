package search

import (
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/search"

	"github.com/smartystreets/goconvey/convey"
)

func TestSearchDySeaAdd(t *testing.T) {
	convey.Convey("DySeachAdd", t, func(ctx convey.C) {
		var (
			param = &search.DySeachAP{}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.DySeachAdd(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestSearchDySeaUpdate(t *testing.T) {
	convey.Convey("DySeachUpdate", t, func(ctx convey.C) {
		var (
			param = &search.DySeachUP{}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.DySeachUpdate(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestSearchDySeaDelete(t *testing.T) {
	convey.Convey("DySeachDelete", t, func(ctx convey.C) {
		var (
			id = int64(0)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.DySeachDelete(id)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestSearchDySeaFindByID(t *testing.T) {
	convey.Convey("DySeaFindByID", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.DySeachValidat(1000, 10000, "")
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
