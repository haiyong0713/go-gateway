package service

import (
	"context"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/ugc-season/job/model/stat"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestServiceconsumerSnproc(t *testing.T) {
	convey.Convey("consumerSnproc", t, func(convCtx convey.C) {
		var (
			k = ""
			d = &databus.Databus{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			s.consumerSnproc(k, d)
			convCtx.Convey("No return values", func(convCtx convey.C) {
			})
		})
	})
}

func TestServicestatSnDealproc(t *testing.T) {
	convey.Convey("statSnDealproc", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			s.statSnDealproc()
			convCtx.Convey("No return values", func(convCtx convey.C) {
			})
		})
	})
}

func TestServicesnMsgUpdate(t *testing.T) {
	convey.Convey("snMsgUpdate", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			ms = &stat.Msg{
				Aid:   600098173,
				Click: 10,
				Type:  "view",
			}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			s.snMsgUpdate(c, ms)
			convCtx.Convey("No return values", func(convCtx convey.C) {
			})
		})
	})
}

func TestServiceseasonResDel(t *testing.T) {
	convey.Convey("seasonResDel", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = int64(4512)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			s.seasonResDel(c, sid)
			convCtx.Convey("No return values", func(convCtx convey.C) {
			})
		})
	})
}

func TestServiceseasonResUpdate(t *testing.T) {
	convey.Convey("seasonResUpdate", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = int64(4512)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			s.seasonResUpdate(c, sid)
			//rt := &retry.Info{Action: retry.FailUpSeasonStat}
			//rt.Data.SeasonID = sid
			//_ = s.PushToRetryList(c, rt)
			//convCtx.Convey("No return values", func(convCtx convey.C) {
			//})
		})
	})
}

func TestServicestatUpdate(t *testing.T) {
	convey.Convey("statUpdate", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			sid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := s.statUpdate(c, sid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestServicesnStatReSum(t *testing.T) {
	convey.Convey("snStatReSum", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			sid  = int64(0)
			aids = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			snStat, err := s.snStatReSum(c, sid, aids)
			convCtx.Convey("Then err should be nil.snStat should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(snStat, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestServicegetStats(t *testing.T) {
	convey.Convey("getStats", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			aids = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			stats, err := s.getStats(c, aids)
			convCtx.Convey("Then err should be nil.stats should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(stats, convey.ShouldNotBeNil)
			})
		})
	})
}
