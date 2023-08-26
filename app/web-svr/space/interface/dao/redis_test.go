package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	pb "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaokeyUpArt(t *testing.T) {
	convey.Convey("keyUpArt", t, func(ctx convey.C) {
		var (
			mid = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyUpArt(mid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyUpArc(t *testing.T) {
	convey.Convey("keyUpArc", t, func(ctx convey.C) {
		var (
			mid = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyUpArc(mid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetUpArtCache(t *testing.T) {
	convey.Convey("SetUpArtCache", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			mid  = int64(2222)
			data = &model.UpArtStat{View: 2222, Reply: 2222}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetUpArtCache(c, mid, data)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoUpArtCache(t *testing.T) {
	convey.Convey("UpArtCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(2222)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.UpArtCache(c, mid)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetUpArcCache(t *testing.T) {
	convey.Convey("SetUpArcCache", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			mid  = int64(2222)
			data = &model.UpArcStat{View: 2222, Reply: 2222}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetUpArcCache(c, mid, data)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoUpArcCache(t *testing.T) {
	convey.Convey("UpArcCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(2222)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.UpArcCache(c, mid)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaosetKvCache(t *testing.T) {
	convey.Convey("setKvCache", t, func(ctx convey.C) {
		var (
			conn   = d.redis.Get(context.Background())
			key    = "test_redis"
			value  = []byte("hello")
			expire = d.getCacheExpire(d.redisMinExpire, d.redisMaxExpire)
		)
		defer conn.Close()
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := setKvCache(conn, key, value, expire)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_SetOfficialCache(t *testing.T) {
	convey.Convey("SetOfficialCache", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddCacheOfficial(context.Background(), &pb.OfficialRequest{Mid: 1}, &pb.OfficialReply{Uid: 1})
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_OfficialCache(t *testing.T) {
	convey.Convey("setKvCache", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.CacheOfficial(context.Background(), &pb.OfficialRequest{Mid: 1})
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
				bs, _ := json.Marshal(res)
				fmt.Println(string(bs))
			})
		})
	})
}

func TestDao_CacheUserTab(t *testing.T) {
	convey.Convey("setKvCache", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.CacheUserTab(context.Background(), &pb.UserTabReq{Mid: 27515233})
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
				bs, _ := json.Marshal(res)
				fmt.Println(string(bs))
			})
		})
	})
}
