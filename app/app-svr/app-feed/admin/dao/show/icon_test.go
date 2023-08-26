package show

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/icon"

	"github.com/smartystreets/goconvey/convey"
)

func TestIconSave(t *testing.T) {
	var (
		icArg = &icon.Icon{
			//ID:      1,
			Module:      "",
			Icon:        "",
			EffectGroup: 1,
			EffectURL:   "",
			Stime:       time.Time(1574666727),
			Etime:       time.Time(1574666740),
		}
	)
	convey.Convey("IconSave", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			id, err := d.IconSave(context.Background(), icArg)
			fmt.Println(id)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestUpdateIconState(t *testing.T) {
	convey.Convey("UpdateIconState", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			_, err := d.UpdateIconState(context.Background(), 1, 0)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestIcon(t *testing.T) {
	convey.Convey("Icon", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			res, err := d.Icon(context.Background(), 2)
			str, _ := json.Marshal(res)
			fmt.Println(string(str))
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestIcons(t *testing.T) {
	convey.Convey("Icons", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			res, total, err := d.Icons(context.Background(), 1, 15)
			str, _ := json.Marshal(res)
			fmt.Println(string(str))
			fmt.Println(total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
