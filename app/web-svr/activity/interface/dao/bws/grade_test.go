package bws

import (
	"context"
	"testing"

	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"github.com/smartystreets/goconvey/convey"
)

func TestbwsUserGrade(t *testing.T) {
	convey.Convey("CacheUserGrade", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res := bwsUserGrade(1)
			convCtx.Convey("Then res should not be nil.", func(convCtx convey.C) {
				convCtx.Print(res)
			})
		})
	})
}

func TestbwsAchieveGrade(t *testing.T) {
	convey.Convey("CacheUserGrade", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res := bwsAchieveGrade(1, "test")
			convCtx.Convey("Then res should not be nil.", func(convCtx convey.C) {
				convCtx.Print(res)
			})
		})
	})
}

func TestbuildUserGrade(t *testing.T) {
	convey.Convey("CacheUserGrade", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res := buildUserGrade(10, 1576479652)
			convCtx.Convey("Then res should not be nil.", func(convCtx convey.C) {
				convCtx.Print(res)
			})
		})
	})
}

func TestCacheUserGrade(t *testing.T) {
	convey.Convey("CacheUserGrade", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, _, err := d.CacheUserGrade(c, 1, 1)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddCacheUserGrade(t *testing.T) {
	convey.Convey("AddCacheUserGrade", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUserGrade(c, 1, map[int64]*bwsmdl.UserGrade{1: {Amount: 1, Mtime: 1576143450}})
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddUserGrade(t *testing.T) {
	convey.Convey("AddUserGrade", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.AddUserGrade(c, 1, 1, "test")
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestCacheUsersRank(t *testing.T) {
	convey.Convey("CacheUsersRank", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.CacheUsersRank(c, 1, 0, 100)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Println(data)
			})
		})
	})
}

func TestCacheAchievesGrade(t *testing.T) {
	convey.Convey("CacheAchievesGrade", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.CacheAchievesGrade(c, 1, []string{})
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Println(data)
			})
		})
	})
}

func TestAddCacheAchievesGrade(t *testing.T) {
	convey.Convey("AddCacheAchievesGrade", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheAchievesGrade(c, 1, map[string]int64{"test": 1})
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAchievesGrade(t *testing.T) {
	convey.Convey("AchievesGrade", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.AchievesGrade(c, 1, []string{})
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Println(data)
			})
		})
	})
}

func TestRawUsersAchievesGrade(t *testing.T) {
	convey.Convey("RawUsersAchievesGrade", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.RawUsersAchievesGrade(c, 1, []string{})
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Println(data)
			})
		})
	})
}

func TestRawGradeInfo(t *testing.T) {
	convey.Convey("RawGradeInfo", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.RawGradeInfo(c, "8af2ec0295e86c33")
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Println(data)
			})
		})
	})
}

func TestRawUsersGrade(t *testing.T) {
	convey.Convey("RawUsersGrade", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			data, err := d.RawUsersGrade(c, 1)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Println(data)
			})
		})
	})
}

func TestDelUserGrade(t *testing.T) {
	convey.Convey("DelUserGrade", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelUserGrade(c, 1)
			convCtx.Convey("Then err should be nil.mid should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
