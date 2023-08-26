package mark

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaomarkKey(t *testing.T) {
	var (
		aid = int64(0)
		mid = int64(0)
	)
	convey.Convey("markKey", t, func(ctx convey.C) {
		p1 := markKey(aid, mid)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoevaluationKey(t *testing.T) {
	var (
		aid = int64(0)
	)
	convey.Convey("evaluationKey", t, func(ctx convey.C) {
		p1 := evaluationKey(aid)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoCacheMark(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(0)
		mid = int64(0)
	)
	convey.Convey("CacheMark", t, func(ctx convey.C) {
		res, err := d.CacheMark(c, aid, mid)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoAddCacheMark(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(0)
		mid = int64(0)
		val = int64(0)
	)
	convey.Convey("AddCacheMark", t, func(ctx convey.C) {
		err := d.AddCacheMark(c, aid, mid, val)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoCacheEvaluation(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(234)
	)
	convey.Convey("CacheEvaluation", t, func(ctx convey.C) {
		res, err := d.CacheEvaluation(c, aid)
		fmt.Println(res, err)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoAddCacheEvaluation(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(123)
		val = int64(444)
	)
	convey.Convey("AddCacheEvaluation", t, func(ctx convey.C) {
		err := d.AddCacheEvaluation(c, aid, val)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
