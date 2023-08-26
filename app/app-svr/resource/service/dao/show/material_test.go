package show

import (
	"context"
	"testing"

	pb2 "go-gateway/app/app-svr/resource/service/api/v2"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetMaterial(t *testing.T) {
	convey.Convey("GetMaterial", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			res, err := d.GetMaterial(context.Background())
			ctx.Convey("Error should be nil, res should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestGetMaterialMap(t *testing.T) {
	convey.Convey("GetMaterialMap", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			res, err := d.GetMaterialMap(context.Background())
			ctx.Convey("Error should be nil, res should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestSetMaterial2Cache(t *testing.T) {
	convey.Convey("SetMaterial2Cache", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			MaterialMap := make(map[int64]*pb2.Material)
			err := d.SetMaterial2Cache(context.Background(), "key", 100, MaterialMap)
			ctx.Convey("Error should be nil, res should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetMaterialFromCache(t *testing.T) {
	convey.Convey("GetMaterialFromCache", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			res, err := d.GetMaterialFromCache(context.Background(), "key")
			ctx.Convey("Error should be nil, res should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
