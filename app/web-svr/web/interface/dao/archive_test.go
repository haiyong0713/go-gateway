package dao

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/web/ecode"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

func TestDaoArcAppeal(t *testing.T) {
	convey.Convey("ArcAppeal", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			mid      = int64(2222)
			business = int(1)
		)
		defer gock.OffAll()
		httpMock("POST", d.arcAppealURL).Reply(200).JSON(`{"code":0}`)
		data := map[string]string{"tid": "27", "oid": "222", "description": "test111"}
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.ArcAppeal(c, mid, data, business)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoAppealTags(t *testing.T) {
	convey.Convey("AppealTags", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			business = int(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			rs, err := d.AppealTags(c, business)
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(rs, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoRelatedAids(t *testing.T) {
	convey.Convey("RelatedAids", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			aid = int64(9912124)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.relatedURL).Reply(200).JSON(`{"data":[{"key":"33817773","value":"14536406,25731794"}]}`)
			aids, err := d.RelatedAids(c, aid)
			ctx.Convey("Then err should be nil.aids should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(aids, convey.ShouldNotBeNil)
				ctx.Printf("%+v", aids)
			})
		})
	})
}

func TestDaokeyArcAppealLimit(t *testing.T) {
	convey.Convey("keyArcAppealLimit", t, func(ctx convey.C) {
		var (
			mid = int64(0)
			aid = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyArcAppealLimit(mid, aid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetArcAppealCache(t *testing.T) {
	convey.Convey("SetArcAppealCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(2222)
			aid = int64(2222)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetArcAppealCache(c, mid, aid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoArcAppealCache(t *testing.T) {
	convey.Convey("ArcAppealCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(2222)
			aid = int64(2222)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.ArcAppealCache(c, mid, aid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldEqual, ecode.ArcAppealLimit)
			})
		})
	})
}

func TestDao_ArcRecommends(t *testing.T) {
	convey.Convey("ArcRecommends", t, func(ctx convey.C) {
		var (
			c             = context.Background()
			aid           = int64(795929775)
			mid           = int64(0)
			needOperation = true
			webRmRepeat   = false
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			aids, err := d.ArcRecommends(c, aid, mid, needOperation, webRmRepeat, "")
			ctx.Convey("Then err should be nil.aids should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(aids, convey.ShouldNotBeNil)
				ctx.Printf("%+v", aids)
			})
		})
	})
}
