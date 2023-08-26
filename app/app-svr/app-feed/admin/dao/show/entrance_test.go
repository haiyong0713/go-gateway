package show

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestPopEntranceAdd(t *testing.T) {
	convey.Convey("PopEntranceAdd", t, func(ctx convey.C) {
		param := &show.EntranceSave{
			EntranceCore: show.EntranceCore{
				Title:       "testOnly",
				Icon:        "t1",
				RedirectUri: "www.bilibili.com",
				Rank:        0,
				ModuleID:    "hot-channel",
				Grey:        20,
				RedDot:      1,
				RedDotText:  "有更新",
				BuildLimit:  "",
				WhiteList:   "1,2,3",
				BlackList:   "4,5,6",
				State:       1,
			},
			Version: 1,
		}
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			id, err := d.PopEntranceAdd(context.Background(), param)
			fmt.Println(id, err)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(id, convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestPopEntranceEdit(t *testing.T) {
	convey.Convey("PopEntranceEdit", t, func(ctx convey.C) {
		param := &show.EntranceSave{
			EntranceCore: show.EntranceCore{
				Title:       "t2",
				Icon:        "t2",
				RedirectUri: "www.bilibili.com",
				Rank:        0,
				ModuleID:    "hot-channel",
				Grey:        20,
				RedDot:      1,
				RedDotText:  "有更新",
				BuildLimit:  "",
				WhiteList:   "1,2,3",
				BlackList:   "4,5,6",
				State:       1,
				ID:          306,
			},
			Version: 1,
		}
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopEntranceEdit(context.Background(), param)
			fmt.Println("123 ", err)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPopEntrance(t *testing.T) {
	convey.Convey("PopEntrance", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.PopEntrance(context.Background(), 0, 1, 10)
			str, _ := json.Marshal(res)
			fmt.Println(string(str))
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestPopEntranceOperate(t *testing.T) {
	convey.Convey("TestPopEntranceOperate", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.PopEntranceOperate(context.Background(), 16, 0)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestRedDotUpdate(t *testing.T) {
	convey.Convey("TestRedDotUpdate", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.RedDotUpdate(context.Background(), "good-history", 0)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestEntranceTopPhotoUpdate(t *testing.T) {
	convey.Convey("EntranceTopPhotoUpdate", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			var (
				id       = int64(19)
				topPhoto = "test"
			)
			err := d.EntranceTopPhotoUpdate(context.Background(), id, topPhoto)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestEntranceGetTopPhoto(t *testing.T) {
	convey.Convey("EntranceGetTopPhoto", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			var (
				id = int64(19)
			)
			res, err := d.EntranceGetTopPhoto(context.Background(), id)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
