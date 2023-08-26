package like

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	gnk "gopkg.in/h2non/gock.v1"

	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeLikeListState(t *testing.T) {
	convey.Convey("Like", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10541)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			ns, err := d.LikeListState(c, sid)
			str, _ := json.Marshal(ns)
			fmt.Println(string(str), err)
			ctx.Convey("Then err should be nil.ns should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(ns, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeList(t *testing.T) {
	convey.Convey("LikeList", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			sid    = int64(10297)
			offset = int(1)
			limit  = int(10)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			list, err := d.LikeList(c, sid, offset, limit)
			ctx.Convey("Then err should be nil.list should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(list, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeCnt(t *testing.T) {
	convey.Convey("LikeCnt", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10297)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			count, err := d.LikeCnt(c, sid)
			ctx.Convey("Then err should be nil.count should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(count, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeSetObjectStat(t *testing.T) {
	convey.Convey("SetObjectStat", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			sid   = int64(10297)
			stat  = &like.SubjectTotalStat{}
			count = int(10)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			defer gnk.OffAll()
			httpMock("GET", d.setObjStatURL).Reply(200).JSON(`{"code":0}`)
			err := d.SetObjectStat(c, sid, stat, count)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeSetViewRank(t *testing.T) {
	convey.Convey("SetViewRank", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(10297)
			aids = []int64{1, 2}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			defer gnk.OffAll()
			httpMock("GET", d.setViewRankURL).Reply(200).JSON(`{"code":0}`)
			err := d.SetViewRank(c, sid, aids, "")
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeSetLikeContent(t *testing.T) {
	convey.Convey("SetLikeContent", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			defer gnk.OffAll()
			httpMock("GET", d.setLikeContentURL).Reply(200).JSON(`{"code":0}`)
			err := d.SetLikeContent(c, id)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestLikeOidsInfo(t *testing.T) {
	convey.Convey("LikeOidsInfo", t, func(ctx convey.C) {
		var (
			c          = context.Background()
			typ  int64 = likemdl.ARTICLE
			sid        = int64(10462)
			oids       = []int64{2291}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			data, err := d.LikeOidsInfo(c, typ, oids)
			ctx.Convey("Then err should be nil.count should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
				if v, ok := data[oids[0]]; ok {
					ctx.Printf("%+v", data[oids[0]])
					ctx.Print(v.Sid == sid)
				}
			})
		})
	})
}

// LikeAidItem
func TestLikeAidItem(t *testing.T) {
	convey.Convey("LikeAidItem", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(1)
			aid = int64(1)
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			_, err := d.LikeAidItem(c, sid, aid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// LikeWidItem
func TestLikeWidItem(t *testing.T) {
	convey.Convey("LikeAidItem", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(1)
			aid = int64(1)
			typ = 1
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			_, err := d.LikeWidItem(c, sid, aid, typ)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
