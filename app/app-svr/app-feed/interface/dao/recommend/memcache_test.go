package recommend

import (
	"context"

	"go-common/library/cache/memcache"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRecommendkeyRcmd(t *testing.T) {
	convey.Convey("keyRcmd", t, func(ctx convey.C) {
		p1 := keyRcmd()
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestRecommendkeyFollowModeList(t *testing.T) {
	convey.Convey("keyFollowModeList", t, func(ctx convey.C) {
		p1 := keyFollowModeList()
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestRecommendRcmdCache(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("RcmdCache", t, func(ctx convey.C) {
		is, err := d.RcmdCache(c)
		if len(is) == 0 {
			conn := d.mc.Get(c)
			key := keyRcmd()
			item := &memcache.Item{Key: key, Object: []*ai.Item{{ID: 1}}, Flags: memcache.FlagJSON, Expiration: d.expireMc}
			conn.Set(item)
			conn.Close()
		}
		is, err = d.RcmdCache(c)
		ctx.Convey("Then err should be nil.is should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(is, convey.ShouldNotBeNil)
		})
	})
}

func TestRecommendAddFollowModeListCache(t *testing.T) {
	var (
		c    = context.Background()
		list map[int64]struct{}
	)
	convey.Convey("AddFollowModeListCache", t, func(ctx convey.C) {
		err := d.AddFollowModeListCache(c, list)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestRecommendFollowModeListCache(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("FollowModeListCache", t, func(ctx convey.C) {
		list, err := d.FollowModeListCache(c)
		if len(list) == 0 {
			amap := make(map[int64]struct{})
			amap[1] = struct{}{}
			d.AddFollowModeListCache(context.Background(), amap)
			list, err = d.FollowModeListCache(c)
		}
		ctx.Convey("Then err should be nil.list should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(list, convey.ShouldNotBeNil)
		})
	})
}
