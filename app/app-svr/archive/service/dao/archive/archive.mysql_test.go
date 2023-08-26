package archive

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRawArc(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(10097272)
	)
	convey.Convey("archive3", t, func(ctx convey.C) {
		_, err := d.RawArc(c, aid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArchivearchives3(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{10097272}
	)
	convey.Convey("archives3", t, func(ctx convey.C) {
		res, err := d.RawArcs(c, aids)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestArchiveAddit(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(800007088)
	)
	convey.Convey("Addit", t, func(ctx convey.C) {
		addr, err := d.RawAddit(c, aid)
		fmt.Println(addr)
		ctx.Convey("Then err should be nil.addit should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArchivestatTbl(t *testing.T) {
	var (
		aid = int64(1)
	)
	convey.Convey("statTbl", t, func(ctx convey.C) {
		p1 := statTbl(aid)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestArchivestat3(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(10097272)
	)
	convey.Convey("stat3", t, func(ctx convey.C) {
		_, err := d.RawStat(c, aid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArchivestats3(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{10097272}
	)
	convey.Convey("stats3", t, func(ctx convey.C) {
		_, err := d.RawStats(c, aids)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArchiveUpperPassed(t *testing.T) {
	var (
		c   = context.TODO()
		mid = int64(1)
	)
	convey.Convey("UpperPassed", t, func(ctx convey.C) {
		_, _, _, err := d.RawUpperPassed(c, mid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArchiveUppersPassed(t *testing.T) {
	var (
		c    = context.TODO()
		mids = []int64{1}
	)
	convey.Convey("UppersPassed", t, func(ctx convey.C) {
		aidm, ptimes, copyrights, err := d.RawUppersPassed(c, mids)
		ctx.Convey("Then err should be nil.aidm,ptimes,copyrights should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(copyrights, convey.ShouldNotBeNil)
			ctx.So(ptimes, convey.ShouldNotBeNil)
			ctx.So(aidm, convey.ShouldNotBeNil)
		})
	})
}

func TestArchiveTypes(t *testing.T) {
	var (
		c = context.TODO()
	)
	convey.Convey("Types", t, func(ctx convey.C) {
		types, err := d.RawTypes(c)
		ctx.Convey("Then err should be nil.types should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(types, convey.ShouldNotBeNil)
		})
	})
}

func TestArchivevideos3(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(10097272)
	)
	convey.Convey("videos3", t, func(ctx convey.C) {
		_, err := d.RawPages(c, aid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestArchivevideosByAids3(t *testing.T) {
	var (
		c    = context.TODO()
		aids = []int64{10097272}
	)
	convey.Convey("videosByAids3", t, func(ctx convey.C) {
		vs, err := d.RawVideosByAids(c, aids)
		ctx.Convey("Then err should be nil.vs should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(vs, convey.ShouldNotBeNil)
		})
	})
}

func TestArchivevideo3(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(10097272)
		cid = int64(10097272)
	)
	convey.Convey("video3", t, func(ctx convey.C) {
		_, err := d.RawPage(c, aid, cid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
