package dao

import (
	"testing"

	mdlesp "go-gateway/app/web-svr/esports/job/model"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

func TestNoticeUser(t *testing.T) {
	var (
		mids    = []int64{27515232}
		body    = "S8全球总决赛中,你订阅的赛程【2018-10-20 12:00:00 KT VS IG】即将开播，快前去观看比赛吧!"
		contest = &mdlesp.Contest{ID: 1, Stime: 1552961913, LiveRoom: 6}
	)
	convey.Convey("NoticeUser", t, func(ctx convey.C) {
		defer gock.OffAll()
		httpMock("POST", d.pushURL).Reply(200).JSON(`{"code":0,"data":1}`)
		err := d.NoticeUser(mids, body, contest)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestSignature(t *testing.T) {
	var (
		params map[string]string
		secret = ""
	)
	convey.Convey("signature", t, func(ctx convey.C) {
		p1 := d.signature(params, secret)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestGetUUID(t *testing.T) {
	var (
		mids    = "11,22,33"
		contest = &mdlesp.Contest{}
	)
	convey.Convey("getUUID", t, func(ctx convey.C) {
		p1 := d.getUUID(mids, contest)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}
