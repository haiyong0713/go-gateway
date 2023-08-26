package result

import (
	"context"
	"testing"

	"go-common/library/time"

	"go-gateway/app/app-svr/ugc-season/job/model/archive"

	"github.com/smartystreets/goconvey/convey"
)

func TestTxAddSeason(t *testing.T) {
	var (
		c        = context.TODO()
		s        = &archive.Season{SeasonID: 1, Title: "test", Desc: "test desc", Cover: "xxx", Mid: 111, Attribute: 0, SignState: 0}
		maxPtime = time.Time(1560841403)
		//maxPtime = "2019-01-01 00:00:00"
		firstAid = int64(1)
	)
	convey.Convey("TxAddSeason", t, func(ctx convey.C) {
		tx, err := d.BeginTran(c)
		err = d.TxAddSeason(c, tx, s, maxPtime, firstAid, 1)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		tx.Commit()
	})
}

func TestTxAddSection(t *testing.T) {
	var (
		c    = context.TODO()
		secs []*archive.SeasonSection
	)
	secs = append(secs, &archive.SeasonSection{SeasonID: 1, Title: "te'st", SectionID: 1})
	convey.Convey("TxAddSection", t, func(ctx convey.C) {
		tx, err := d.BeginTran(c)
		err = d.TxAddSection(c, tx, secs)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		tx.Commit()
	})
}

func TestTxAddEp(t *testing.T) {
	var (
		c     = context.TODO()
		eps   []*archive.SeasonEp
		title = "te'st"
	)
	eps = append(eps, &archive.SeasonEp{SeasonID: 1, Title: title, SectionID: 1, EpID: 1})
	convey.Convey("TxAddEp", t, func(ctx convey.C) {
		tx, err := d.BeginTran(c)
		err = d.TxAddEp(c, tx, eps)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		tx.Rollback()
	})
}

func TestTxDelSeasonByID(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("TxDelSeasonByID", t, func(ctx convey.C) {
		tx, err := d.BeginTran(c)
		err = d.TxDelSeasonByID(c, tx, sid)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestTxDelSecByID(t *testing.T) {
	var (
		c   = context.TODO()
		ids = []int64{1}
	)
	convey.Convey("TxDelSecByID", t, func(ctx convey.C) {
		tx, err := d.BeginTran(c)
		err = d.TxDelSecByID(c, tx, ids)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestTxDelEpByID(t *testing.T) {
	var (
		c   = context.TODO()
		ids = []int64{1}
	)
	convey.Convey("TxDelEpByID", t, func(ctx convey.C) {
		tx, err := d.BeginTran(c)
		err = d.TxDelEpByID(c, tx, ids)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestTxDelSecBySID(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("TxDelSecBySID", t, func(ctx convey.C) {
		tx, err := d.BeginTran(c)
		err = d.TxDelSecBySID(c, tx, sid)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestTxDelEpBySID(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("TxDelEpBySID", t, func(ctx convey.C) {
		tx, err := d.BeginTran(c)
		err = d.TxDelEpBySID(c, tx, sid)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestTxUpSeasonMtime(t *testing.T) {
	var (
		c   = context.TODO()
		sid = int64(1)
	)
	convey.Convey("TxUpSeasonMtime", t, func(ctx convey.C) {
		tx, err := d.BeginTran(c)
		err = d.TxUpSeasonMtime(c, tx, sid)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
