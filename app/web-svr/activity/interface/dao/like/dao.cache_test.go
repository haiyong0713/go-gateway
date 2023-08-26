package like

import (
	"context"
	"testing"

	"fmt"

	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/smartystreets/goconvey/convey"
)

func TestLikeLike(t *testing.T) {
	convey.Convey("Like", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(77)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.Like(c, id)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeActSubject(t *testing.T) {
	convey.Convey("ActSubject", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(10256)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ActSubject(c, id)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeMissionBuff(t *testing.T) {
	convey.Convey("LikeMissionBuff", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			id  = int64(10256)
			mid = int64(77)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LikeMissionBuff(c, id, mid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeMissionGroupItems(t *testing.T) {
	convey.Convey("MissionGroupItems", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			keys = []int64{1, 2}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.MissionGroupItems(c, keys)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeActMission(t *testing.T) {
	convey.Convey("ActMission", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			id  = int64(10256)
			lid = int64(7)
			mid = int64(77)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ActMission(c, id, lid, mid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeActLikeAchieves(t *testing.T) {
	convey.Convey("ActLikeAchieves", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ActLikeAchieves(c, id)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeActMissionFriends(t *testing.T) {
	convey.Convey("ActMissionFriends", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			id  = int64(10256)
			lid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ActMissionFriends(c, id, lid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeActUserAchieve(t *testing.T) {
	convey.Convey("ActUserAchieve", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ActUserAchieve(c, id)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeMatchSubjects(t *testing.T) {
	convey.Convey("MatchSubjects", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			keys = []int64{10256}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.MatchSubjects(c, keys)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestLikeLikeContent(t *testing.T) {
	convey.Convey("LikeContent", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			keys = []int64{1, 2}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.LikeContent(c, keys)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestActSubjectProtocol(t *testing.T) {
	convey.Convey("LikeContent", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10298)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ActSubjectProtocol(c, sid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

func TestCacheActSubjectProtocol(t *testing.T) {
	convey.Convey("LikeContent", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10298)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.CacheActSubjectProtocol(c, sid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

func TestAddCacheActSubjectProtocol(t *testing.T) {
	convey.Convey("LikeContent", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			sid      = int64(10256)
			protocol = &like.ActSubjectProtocol{ID: 1, Sid: 10256}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddCacheActSubjectProtocol(c, sid, protocol)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestActSubjectProtocols(t *testing.T) {
	convey.Convey("LikeContent", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			sids = []int64{10554, 10553}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ActSubjectProtocols(c, sids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

func TestCacheActSubjectProtocols(t *testing.T) {
	convey.Convey("LikeContent", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			sids = []int64{10554, 10553}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.CacheActSubjectProtocols(c, sids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

func TestAddCacheActSubjectProtocols(t *testing.T) {
	convey.Convey("LikeContent", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			protocol = make(map[int64]*like.ActSubjectProtocol)
		)
		protocol[1] = &like.ActSubjectProtocol{ID: 1, Sid: 10256}
		protocol[2] = &like.ActSubjectProtocol{ID: 2, Sid: 10257}
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddCacheActSubjectProtocols(c, protocol)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestReserveOnly(t *testing.T) {
	convey.Convey("ReserveOnly", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ReserveOnly(c, 10529, 1548785)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

func TestCacheReserveOnly(t *testing.T) {
	convey.Convey("CacheReserveOnly", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.CacheReserveOnly(c, 10629, 1548785)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

// AddCacheReserveOnly
func TestAddCacheReserveOnly(t *testing.T) {
	convey.Convey("CacheReserveOnly", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddCacheReserveOnly(c, 10529, &like.HasReserve{ID: 1, State: 1, Num: 1}, 1548785)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// ReservesTotal
func TestReservesTotal(t *testing.T) {
	convey.Convey("ReservesTotal", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ReservesTotal(c, []int64{10529, 10629})
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%+v", res)
			})
		})
	})
}

// ActStochastic
func TestActStochastic(t *testing.T) {
	convey.Convey("ActStochastic", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ActStochastic(c, 10733, 0)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}

// ActRandom
func TestActRandom(t *testing.T) {
	convey.Convey("ActRandom", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ActRandom(c, 10733, 0, 0, -1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}

func TestActEsLikesIDs(t *testing.T) {
	convey.Convey("ActEsLikesIDs", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ActEsLikesIDs(c, 10533, 0, 2, 4)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}
