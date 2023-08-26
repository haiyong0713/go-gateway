package show

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

// TestDaoAppSpecialCard
func TestAppSpecialCard(t *testing.T) {
	convey.Convey("AppSpecialCard", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			res, err := d.AppSpecialCard(context.Background())
			ctx.Convey("Error should be nil, res should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestGetAppRcmdOnlineSpecailCardId(t *testing.T) {
	convey.Convey("GetAppRcmdOnlineSpecailCardId", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			ids, err := d.GetAppRcmdOnlineSpecailCardId(context.Background())
			ctx.Convey("Error should be nil, res should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(ids, convey.ShouldNotBeNil)
			})
		})
	},
	)
}

func TestAppRcmdRelatePgc(t *testing.T) {
	convey.Convey("AppRcmdRelatePgc", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppRcmdRelatePgc(context.Background(), time.Now())

			ctx.Convey("Error should be nil, res should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
