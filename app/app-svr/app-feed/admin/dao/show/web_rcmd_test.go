package show

import (
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestShowWebRcmdAdd(t *testing.T) {
	convey.Convey("WebRcmdAdd", t, func(ctx convey.C) {
		var (
			param = &show.WebRcmdAP{
				CardType:    2,
				CardValue:   "10111292",
				Stime:       1545701985,
				Etime:       1545711985,
				Priority:    1,
				Person:      "quguolin",
				ApplyReason: "test",
				Partition:   "95,189,190,191",
				Avid:        "10111862",
				Tag:         "23",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.WebRcmdAdd(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowWebRcmdUpdate(t *testing.T) {
	convey.Convey("WebRcmdUpdate", t, func(ctx convey.C) {
		var (
			param = &show.WebRcmdUP{
				ID:          1,
				CardType:    1,
				CardValue:   "11",
				Stime:       1545701985,
				Etime:       1545711985,
				Priority:    1,
				Person:      "quguolin",
				ApplyReason: "test",
				Partition:   "95,189,190,191",
				Avid:        "10111862",
				Tag:         "23",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.WebRcmdUpdate(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowWebRcmdDelete(t *testing.T) {
	convey.Convey("WebRcmdDelete", t, func(ctx convey.C) {
		var (
			id = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.WebRcmdDelete(id)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestShowWebRcmdFindByID(t *testing.T) {
	convey.Convey("WebRcmdFindByID", t, func(ctx convey.C) {
		var (
			id = int64(561)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			card, err := d.WebRcmdFindByID(id)
			ctx.Convey("Then err should be nil.card should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(card, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestShowWebRcmdOption(t *testing.T) {
	convey.Convey("WebRcmdOption", t, func(ctx convey.C) {
		var (
			up = &show.WebRcmdOption{
				ID:    1,
				Check: 2,
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.WebRcmdOption(up)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
