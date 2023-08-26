package show

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/app-feed/admin/model/menu"
)

func TestDao_SkinExtSave(t *testing.T) {
	convey.Convey("SkinExtSave", t, func(ctx convey.C) {
		var (
			save   = &menu.SkinExt{SkinID: 3, SkinName: "蓝色妖姬two", Attribute: 2, State: 1, Stime: 1574897442, Etime: 1575897442, Operator: "layan"}
			limits = map[int8]*menu.SkinBuildLimit{0: {Plat: 0, Build: 5460000, Conditions: "gt"}, 1: {Plat: 1, Build: 8390, Conditions: "lt"}}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			id, err := d.SkinExtSave(context.Background(), save, limits)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Print(id)
			})
		})
	})
}

func TestDao_SkinLimits(t *testing.T) {
	convey.Convey("SkinLimits", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			id, err := d.SkinLimits(context.Background(), []int64{1, 2})
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(id)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestDao_SkinModifyState(t *testing.T) {
	convey.Convey("SkinModifyState", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SkinModifyState(context.Background(), 1, 2, 0)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_RawSkinExts(t *testing.T) {
	convey.Convey("RawSkinExts", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			id, err := d.RawSkinExts(context.Background())
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(id)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestDao_SkinExts(t *testing.T) {
	convey.Convey("SkinExts", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			id, err := d.SkinExts(context.Background(), 0, 1, 15)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(id)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestDao_RawSkinExt(t *testing.T) {
	convey.Convey("RawSkinExt", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			id, err := d.RawSkinExt(context.Background(), 1)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(id)
				fmt.Printf("%s", str)
			})
		})
	})
}
