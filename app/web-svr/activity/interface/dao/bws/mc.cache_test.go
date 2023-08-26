package bws

import (
	"context"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestBwsCacheUsersMid(t *testing.T) {
	convey.Convey("CacheUsersMid", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(6)
			mid = int64(27515327)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheUsersMid(c, id, mid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAddCacheUsersMid(t *testing.T) {
	convey.Convey("AddCacheUsersMid", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(3)
			val = &bwsmdl.Users{}
			mid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUsersMid(c, id, val, mid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCacheUsersMid(t *testing.T) {
	convey.Convey("DelCacheUsersMid", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(3)
			mid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheUsersMid(c, id, mid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheUsersKey(t *testing.T) {
	convey.Convey("CacheUsersKey", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			id      = int64(6)
			userKey = "4457568c828b253d"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheUsersKey(c, id, userKey)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAddCacheUsersKey(t *testing.T) {
	convey.Convey("AddCacheUsersKey", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			id      = int64(4)
			val     = &bwsmdl.Users{}
			userKey = "123"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUsersKey(c, id, val, userKey)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCacheUsersKey(t *testing.T) {
	convey.Convey("DelCacheUsersKey", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			id      = int64(4)
			userKey = "123"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheUsersKey(c, id, userKey)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCachePoints(t *testing.T) {
	convey.Convey("CachePoints", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(6)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CachePoints(c, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAddCachePoints(t *testing.T) {
	convey.Convey("AddCachePoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(5)
			val = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCachePoints(c, id, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCachePoints(t *testing.T) {
	convey.Convey("DelCachePoints", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(5)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCachePoints(c, id)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheBwsSign(t *testing.T) {
	convey.Convey("CacheBwsSign", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{54}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheBwsSign(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAddCacheBwsSign(t *testing.T) {
	convey.Convey("AddCacheBwsSign", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			values map[int64]*bwsmdl.PointSign
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheBwsSign(c, values)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCacheBwsSign(t *testing.T) {
	convey.Convey("DelCacheBwsSign", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheBwsSign(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheSigns(t *testing.T) {
	convey.Convey("CacheSigns", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(54)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheSigns(c, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAddCacheSigns(t *testing.T) {
	convey.Convey("AddCacheSigns", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(1)
			val = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheSigns(c, id, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCacheSigns(t *testing.T) {
	convey.Convey("DelCacheSigns", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheSigns(c, id)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheBwsPoints(t *testing.T) {
	convey.Convey("CacheBwsPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{5}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheBwsPoints(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAddCacheBwsPoints(t *testing.T) {
	convey.Convey("AddCacheBwsPoints", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			values map[int64]*bwsmdl.Point
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheBwsPoints(c, values)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCacheBwsPoints(t *testing.T) {
	convey.Convey("DelCacheBwsPoints", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{5}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheBwsPoints(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheAchievements(t *testing.T) {
	convey.Convey("CacheAchievements", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(4)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheAchievements(c, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAddCacheAchievements(t *testing.T) {
	convey.Convey("AddCacheAchievements", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(0)
			val = &bwsmdl.Achievements{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheAchievements(c, id, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCacheAchievements(t *testing.T) {
	convey.Convey("DelCacheAchievements", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheAchievements(c, id)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheRechargeLevels(t *testing.T) {
	convey.Convey("CacheRechargeLevels", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheRechargeLevels(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAddCacheRechargeLevels(t *testing.T) {
	convey.Convey("AddCacheRechargeLevels", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			values map[int64]*bwsmdl.PointsLevel
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheRechargeLevels(c, values)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCacheRechargeLevels(t *testing.T) {
	convey.Convey("DelCacheRechargeLevels", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheRechargeLevels(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheRechargeAwards(t *testing.T) {
	convey.Convey("CacheRechargeAwards", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheRechargeAwards(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAddCacheRechargeAwards(t *testing.T) {
	convey.Convey("AddCacheRechargeAwards", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			values map[int64]*bwsmdl.PointsAward
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheRechargeAwards(c, values)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsDelCacheRechargeAwards(t *testing.T) {
	convey.Convey("DelCacheRechargeAwards", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheRechargeAwards(c, ids)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestBwsCacheActFields(t *testing.T) {
	convey.Convey("CacheActFields", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheActFields(c, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestBwsAddCacheActFields(t *testing.T) {
	convey.Convey("AddCacheActFields", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(0)
			val = &bwsmdl.ActFields{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheActFields(c, id, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
