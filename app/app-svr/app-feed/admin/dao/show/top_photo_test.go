package show

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestPopTopPhotoAdd(t *testing.T) {
	convey.Convey("PopTopPhotoAdd", t, func(ctx convey.C) {
		//var (
		//	param = &show.PopTopPhotoAD{
		//		TopPhoto:   "http://i0.hdslb.com/bfs/app/b077f3c38759da7a5d9eb5007bcc07221b53f0d2.png",
		//		LocationId: 1,
		//		Deleted:    0,
		//	}
		//)
		//ctx.Convey("When everything gose positive", func(ctx convey.C) {
		//	err := d.PopTopPhotoAdd(context.Background(), param)
		//	ctx.Convey("Then err should be nil.", func(ctx convey.C) {
		//		ctx.So(err, convey.ShouldBeNil)
		//	})
		//})
	})
}

func TestPopTopPhotoDeleted(t *testing.T) {
	convey.Convey("PopTopPhotoDeleted", t, func(ctx convey.C) {
		var (
			ID = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopTopPhotoDeleted(context.Background(), ID)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopTopPhotoNotDeleted(t *testing.T) {
	convey.Convey("PopTopPhotoNotDeleted", t, func(ctx convey.C) {
		var (
			ID = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopTopPhotoNotDeleted(context.Background(), ID)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopTPFind(t *testing.T) {
	convey.Convey("PopTPFind", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.PopTPFind(context.Background(), 1, 10)
			ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestPopTopPhotoUpdate(t *testing.T) {
	//var (
	//	//	ID       = int64(1)
	//	//	topPhoto = ""
	//)
	//convey.Convey("PopTPFind", t, func(ctx convey.C) {
	//	ctx.Convey("When everything gose positive", func(ctx convey.C) {
	//		err := d.PopTopPhotoUpdate(context.Background(), ID, topPhoto)
	//		ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
	//			ctx.So(err, convey.ShouldBeNil)
	//		})
	//	})
	//})
}
