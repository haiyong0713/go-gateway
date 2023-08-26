package like

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gopkg.in/h2non/gock.v1"

	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeSubject(t *testing.T) {
	convey.Convey("Subject", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10193)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			n, err := d.Subject(c, sid)
			ctx.Convey("Then err should be nil.n should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(n, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeInOnlinelog(t *testing.T) {
	convey.Convey("InOnlinelog", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10193)
			aid   = int64(1)
			stage = int64(1)
			yes   = int64(1)
			no    = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			rows, err := d.InOnlinelog(c, sid, aid, stage, yes, no)
			ctx.Convey("Then err should be nil.rows should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(rows, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeSubjectList(t *testing.T) {
	convey.Convey("SubjectList", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			types = []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 14, 13, 15, 16, 17, 18, 19}
			ts    = time.Now()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.SubjectList(c, types, ts)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeSubjectTotalStat(t *testing.T) {
	convey.Convey("SubjectTotalStat", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10338)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			rs, err := d.SubjectTotalStat(c, sid)
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(rs, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeAddLotteryTimes(t *testing.T) {
	convey.Convey("AddLotteryTimes", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(315)
			mid = int64(1)
			err error
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.addLotteryTimesURL).Reply(200).SetHeaders(map[string]string{
				"Code": "0",
			})
			d.AddLotteryTimes(c, sid, mid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestListFromEs(t *testing.T) {
	convey.Convey("ListFromEs", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			rs, err := d.ListFromEs(c, &like.EsParams{Sid: 10470, Ps: 50, Pn: 1, Order: "click", Sort: "desc", State: 1})
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(rs, convey.ShouldNotBeNil)
			})
		})
	})
}

// DelLikeState
func TestDelLikeState(t *testing.T) {
	convey.Convey("DelLikeState", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10405)
			lids  = []int64{2614}
			state = 0
			reply = "稿件发布"
			err   error
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("POST", fmt.Sprintf(d.delLikeURL, sid)).Reply(200).SetHeaders(map[string]string{
				"Code": "0",
			})
			err = d.DelLikeState(c, sid, lids, state, reply)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				fmt.Printf("%v", err)
			})
		})
	})
}

func TestDao_UpArcEventRule(t *testing.T) {
	convey.Convey("UpArcEventRule", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10744)
			err error
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err = d.UpArcEventRule(c, sid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				fmt.Printf("%v", err)
			})
		})
	})
}
