package show

import (
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestShowWebRcmdCardAdd(t *testing.T) {
	convey.Convey("WebRcmdCardAdd", t, func(ctx convey.C) {
		var (
			param = &show.WebRcmdCardAP{
				Type:    1,
				Title:   "title",
				Desc:    "desc",
				Cover:   "cover",
				ReType:  1,
				ReValue: "http://bilibili.com",
				Person:  "quguolin",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.WebRcmdCardAdd(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowWebRcmdCardAddSkipUrl(t *testing.T) {
	convey.Convey("WebRcmdCardAdd", t, func(ctx convey.C) {
		var (
			param = &show.WebRcmdCardAP{
				Type:    1,
				Title:   "title",
				Desc:    "desc",
				Cover:   "cover",
				ReType:  1,
				ReValue: d.config.FeedConfig.SkipCardUrl,
				Person:  "litongyu",
			}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			var (
				err  error
				card = &show.WebRcmdCard{}
			)
			err = d.WebRcmdCardAdd(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			err = d.DB.Model(&show.WebRcmdCard{}).Last(&card).Error
			ctx.Convey("WebRcmdCardFindByID err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(card, convey.ShouldNotResemble, &show.WebRcmdCard{})
				ctx.So(card.ReValue, convey.ShouldBeEmpty)
			})
		})
	})
}

func TestShowWebRcmdCardUpdate(t *testing.T) {
	convey.Convey("WebRcmdCardUpdate", t, func(ctx convey.C) {
		var (
			param = &show.WebRcmdCardUP{
				ID:      1,
				Type:    2,
				Title:   "title2",
				Desc:    "desc2",
				Cover:   "cover2",
				ReType:  2,
				ReValue: "http://bilibili.com",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.WebRcmdCardUpdate(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowWebRcmdCardDelete(t *testing.T) {
	convey.Convey("WebRcmdCardDelete", t, func(ctx convey.C) {
		var (
			id = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.WebRcmdCardDelete(id)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowSWBRcmdCardFindByID(t *testing.T) {
	convey.Convey("WebRcmdCardFindByID", t, func(ctx convey.C) {
		var (
			id = int64(3)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			card, err := d.WebRcmdCardFindByID(id)
			ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(card, convey.ShouldNotBeNil)
			})
		})
	})
}
