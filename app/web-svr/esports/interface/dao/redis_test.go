package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	arcmdl "go-gateway/app/app-svr/archive/service/api"
	pb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaokeyCale(t *testing.T) {
	convey.Convey("keyCale", t, func(ctx convey.C) {
		var (
			stime = ""
			etime = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyCale(stime, etime)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyCont(t *testing.T) {
	convey.Convey("keyCont", t, func(ctx convey.C) {
		var (
			ps = int(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyCont(ps)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyVideo(t *testing.T) {
	convey.Convey("keyVideo", t, func(ctx convey.C) {
		var (
			ps = int(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyVideo(ps)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyContID(t *testing.T) {
	convey.Convey("keyContID", t, func(ctx convey.C) {
		var (
			cid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyContID(cid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyCoFav(t *testing.T) {
	convey.Convey("keyCoFav", t, func(ctx convey.C) {
		var (
			mid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyCoFav(mid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyCoAppFav(t *testing.T) {
	convey.Convey("keyCoAppFav", t, func(ctx convey.C) {
		var (
			mid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyCoAppFav(mid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeySID(t *testing.T) {
	convey.Convey("keySID", t, func(ctx convey.C) {
		var (
			sid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keySID(sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyMatchAct(t *testing.T) {
	convey.Convey("keyMatchAct", t, func(ctx convey.C) {
		var (
			aid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyMatchAct(aid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyMatchModule(t *testing.T) {
	convey.Convey("keyMatchModule", t, func(ctx convey.C) {
		var (
			mmid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyMatchModule(mmid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyKnock(t *testing.T) {
	convey.Convey("keyKnock", t, func(ctx convey.C) {
		var (
			mdID = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyKnock(mdID)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyTop(t *testing.T) {
	convey.Convey("keyTop", t, func(ctx convey.C) {
		var (
			aid = int64(1)
			ps  = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyTop(aid, ps)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaokeyPoint(t *testing.T) {
	convey.Convey("keyPoint", t, func(ctx convey.C) {
		var (
			aid  = int64(1)
			mdID = int64(1)
			ps   = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := keyPoint(aid, mdID, ps)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoFMatCache(t *testing.T) {
	convey.Convey("FMatCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.FMatCache(c)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoFVideoCache(t *testing.T) {
	convey.Convey("FVideoCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.FVideoCache(c)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaofilterCache(t *testing.T) {
	convey.Convey("filterCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = "1"
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.filterCache(c, key)
			ctx.Convey("Then err should be nil.rs should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSetFMatCache(t *testing.T) {
	convey.Convey("SetFMatCache", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			fs map[string][]*model.Filter
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetFMatCache(c, fs)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSetFVideoCache(t *testing.T) {
	convey.Convey("SetFVideoCache", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			fs map[string][]*model.Filter
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetFVideoCache(c, fs)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaosetFilterCache(t *testing.T) {
	convey.Convey("setFilterCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = ""
			fs  map[string][]*model.Filter
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.setFilterCache(c, key, fs)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoContestCache(t *testing.T) {
	convey.Convey("ContestCache", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			ps = int(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, total, err := d.ContestCache(c, ps)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoFavCoCache(t *testing.T) {
	convey.Convey("FavCoCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, total, err := d.FavCoCache(c, mid)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoFavCoAppCache(t *testing.T) {
	convey.Convey("FavCoAppCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, total, err := d.FavCoAppCache(c, mid)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaocosCache(t *testing.T) {
	convey.Convey("cosCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			key = "1"
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, total, err := d.cosCache(c, key)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetContestCache(t *testing.T) {
	convey.Convey("SetContestCache", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			ps       = int(1)
			con      = &model.Contest{}
			contests = []*model.Contest{con}
			total    = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetContestCache(c, ps, contests, total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSetFavCoCache(t *testing.T) {
	convey.Convey("SetFavCoCache", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			mid      = int64(0)
			con      = &model.Contest{}
			contests = []*model.Contest{con}
			total    = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetFavCoCache(c, mid, contests, total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSetAppFavCoCache(t *testing.T) {
	convey.Convey("SetAppFavCoCache", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			mid      = int64(0)
			con      = &model.Contest{}
			contests = []*model.Contest{con}
			total    = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetAppFavCoCache(c, mid, contests, total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoDelFavCoCache(t *testing.T) {
	convey.Convey("DelFavCoCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.DelFavCoCache(c, mid)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaosetCosCache(t *testing.T) {
	convey.Convey("setCosCache", t, func(ctx convey.C) {
		var (
			c        = context.Background()
			key      = ""
			con      = &model.Contest{}
			contests = []*model.Contest{con}
			total    = int(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.setCosCache(c, key, contests, total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoCalendarCache(t *testing.T) {
	convey.Convey("CalendarCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
			p = &model.ParamFilter{}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.CalendarCache(c, p)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSetCalendarCache(t *testing.T) {
	convey.Convey("SetCalendarCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			p     = &model.ParamFilter{}
			cal   = &model.Calendar{}
			cales = []*model.Calendar{cal}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetCalendarCache(c, p, cales)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoVideoCache(t *testing.T) {
	convey.Convey("VideoCache", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			ps = int(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, total, err := d.VideoCache(c, ps)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetVideoCache(t *testing.T) {
	convey.Convey("SetVideoCache", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			ps     = int(0)
			arc    = &arcmdl.Arc{}
			videos = []*arcmdl.Arc{arc}
			total  = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetVideoCache(c, ps, videos, total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoseasonsCache(t *testing.T) {
	convey.Convey("seasonsCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			key   = "1"
			start = int(0)
			end   = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, total, err := d.seasonsCache(c, key, start, end)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaosetSeasonsCache(t *testing.T) {
	convey.Convey("setSeasonsCache", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			key     = "1"
			seasons = []*model.Season{}
			total   = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.setSeasonsCache(c, key, seasons, total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSeasonCache(t *testing.T) {
	convey.Convey("SeasonCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			start = int(0)
			end   = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, total, err := d.SeasonCache(c, start, end)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetSeasonCache(t *testing.T) {
	convey.Convey("SetSeasonCache", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			seasons = []*model.Season{}
			total   = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetSeasonCache(c, seasons, total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSeasonMCache(t *testing.T) {
	convey.Convey("SeasonMCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			start = int(0)
			end   = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, total, err := d.SeasonMCache(c, start, end)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSetSeasonMCache(t *testing.T) {
	convey.Convey("SetSeasonMCache", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			seasons = []*model.Season{}
			total   = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SetSeasonMCache(c, seasons, total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaofrom(t *testing.T) {
	convey.Convey("from", t, func(ctx convey.C) {
		var (
			i = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := from(i)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaocombine(t *testing.T) {
	convey.Convey("combine", t, func(ctx convey.C) {
		var (
			sort  = int64(0)
			count = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := combine(sort, count)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoCacheEpContests(t *testing.T) {
	convey.Convey("CacheEpContests", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.CacheEpContests(c, ids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddCacheEpContests(t *testing.T) {
	convey.Convey("AddCacheEpContests", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			data map[int64]*model.Contest
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddCacheEpContests(c, data)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoCacheEpSeasons(t *testing.T) {
	convey.Convey("CacheEpSeasons", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.CacheEpSeasons(c, ids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddCacheEpSeasons(t *testing.T) {
	convey.Convey("AddCacheEpSeasons", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			data map[int64]*model.Season
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddCacheEpSeasons(c, data)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoGetActPageCache(t *testing.T) {
	convey.Convey("GetActPageCache", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			id = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.GetActPageCache(c, id)
			ctx.Convey("Then err should be nil.act should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoAddActPageCache(t *testing.T) {
	convey.Convey("AddActPageCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			aid = int64(0)
			act = &model.ActivePage{}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddActPageCache(c, aid, act)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoGetActModuleCache(t *testing.T) {
	convey.Convey("GetActModuleCache", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			mmid = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.GetActModuleCache(c, mmid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoAddActModuleCache(t *testing.T) {
	convey.Convey("AddActModuleCache", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			mmid   = int64(0)
			module []*model.Video
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddActModuleCache(c, mmid, module)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoGetActTopCache(t *testing.T) {
	convey.Convey("GetActTopCache", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			aid = int64(0)
			ps  = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, total, err := d.GetActTopCache(c, aid, ps)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddActTopCache(t *testing.T) {
	convey.Convey("AddActTopCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			aid   = int64(0)
			ps    = int64(0)
			con   = &model.Contest{}
			tops  = []*model.Contest{con}
			total = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddActTopCache(c, aid, ps, tops, total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoGetActPointsCache(t *testing.T) {
	convey.Convey("GetActPointsCache", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			aid  = int64(0)
			mdID = int64(0)
			ps   = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, total, err := d.GetActPointsCache(c, aid, mdID, ps)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(total, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddActPointsCache(t *testing.T) {
	convey.Convey("AddActPointsCache", t, func(ctx convey.C) {
		var (
			c      = context.Background()
			aid    = int64(0)
			mdID   = int64(0)
			ps     = int64(0)
			con    = &model.Contest{}
			points = []*model.Contest{con}
			total  = int(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddActPointsCache(c, aid, mdID, ps, points, total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoGetActKnockoutCache(t *testing.T) {
	convey.Convey("GetActKnockoutCache", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			mdID = int64(0)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			_, err := d.GetActKnockoutCache(c, mdID)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoAddActKnockoutCache(t *testing.T) {
	convey.Convey("AddActKnockoutCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mdID  = int64(0)
			knock = [][]*model.TreeList{}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddActKnockoutCache(c, mdID, knock)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
func TestDaoAddActKnockCacheTime(t *testing.T) {
	convey.Convey("AddActKnockCacheTime", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			mdID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddActKnockCacheTime(c, mdID)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoGetMActCache(t *testing.T) {
	convey.Convey("GetMActCache", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.GetMActCache(c, aid)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				println(res)
			})
		})
	})
}

func TestDaoAddMActCache(t *testing.T) {
	convey.Convey("AddMActCache", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			aid = int64(1)
			act = &model.Active{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddMActCache(c, aid, act)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoGetLiveCache(t *testing.T) {
	convey.Convey("GetLiveCache", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			liveID = int64(45)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetLiveCache(c, liveID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoAddLiveCache(t *testing.T) {
	convey.Convey("AddLiveCache", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			liveID = int64(45)
			live   = &model.ActiveLive{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddLiveCache(c, liveID, live)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoGameSeasonCache(t *testing.T) {
	convey.Convey("GameSeasonCache", t, func(ctx convey.C) {
		var (
			c  = context.Background()
			tp = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			seasons, err := d.GameSeasonCache(c, tp)
			ctx.Convey("Then err should be nil.res,total should not be nil.", func(ctx convey.C) {
				ctx.So(len(seasons), convey.ShouldBeGreaterThanOrEqualTo, 0)
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSetGameSeasonCache(t *testing.T) {
	convey.Convey("SetGameSeasonCache", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			tp      = int64(0)
			seasons []*model.Season
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			season1 := &model.Season{ID: 11, Title: "aaaaa"}
			seasons = append(seasons, season1)
			season2 := &model.Season{ID: 22, Title: "bbbbb"}
			seasons = append(seasons, season2)
			err := d.SetGameSeasonCache(c, tp, seasons)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoAddGuessCCache(t *testing.T) {
	convey.Convey("AddGuessCollecCache", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			value = &model.GuessCollection{
				Game:   []*model.Filter{{ID: 1, Title: "game title test"}},
				Season: []*model.Season{{ID: 2, Title: "season title test"}},
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.AddGuessCollecCache(c, 1, value)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoGetGuessCCache(t *testing.T) {
	convey.Convey("GuessCollecCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.GuessCollecCache(c, 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestAddGuessDetail(t *testing.T) {
	convey.Convey("AddGuessDetailCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			data := &model.GuessDetail{
				Contest: &model.Contest{ID: 1, GameStage: "test"},
			}
			err := d.AddGuessDetailCache(c, 1, data)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGuessDetailCache(t *testing.T) {
	convey.Convey("GuessDetailCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.GuessDetailCache(c, 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDelGuessDetailCache(t *testing.T) {
	convey.Convey("GuessDetailCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.DelGuessDetailCache(c, 1)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddGuessRecent(t *testing.T) {
	convey.Convey("AddGuessDetailCache", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			param := &model.ParamEsGuess{
				HomeID: 1,
				AwayID: 2,
			}
			contest := &model.Contest{
				ID:        1,
				GameStage: "test",
				Stime:     100,
				Etime:     200,
			}
			err := d.AddGuessRecent(c, param, []*model.Contest{contest})
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetGuessRecent(t *testing.T) {
	convey.Convey("TestGetGuessRecent", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			param := &model.ParamEsGuess{
				HomeID: 1,
				AwayID: 2,
			}
			data, err := d.GuessRecent(c, param)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			bs, _ := json.Marshal(data)
			fmt.Println(string(bs))
		})
	})
}

func TestDaoCacheSearchMainIDs(t *testing.T) {
	convey.Convey("CacheSearchMainIDs", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.CacheSearchMainIDs(c)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddCacheSearchMainIDs(t *testing.T) {
	convey.Convey("AddCacheSearchMainIDs", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			data := []int64{11, 22, 33}
			err := d.AddCacheSearchMainIDs(c, data)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoCacheSearchMD(t *testing.T) {
	convey.Convey("CacheSearchMD", t, func(ctx convey.C) {
		var (
			c       = context.Background()
			mainIDs = []int64{11, 22, 33}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.CacheSearchMD(c, mainIDs)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoAddCacheSearchMD(t *testing.T) {
	convey.Convey("AddCacheSearchMD", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			data map[int64]*model.SearchRes
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			data = make(map[int64]*model.SearchRes, 1)
			data[11] = &model.SearchRes{
				SearchMain: &model.SearchMain{ID: 1, QueryName: "aaa", Stime: 1570761949, Etime: 1571625950},
				ContestIDs: []int64{11, 22, 33},
			}
			err := d.AddCacheSearchMD(c, data)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoCacheEpGames(t *testing.T) {
	convey.Convey("CacheEpGames", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			gids = []int64{1, 3}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.CacheEpGames(c, gids)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestDaoAddCacheEpGames(t *testing.T) {
	convey.Convey("AddCacheEpGames", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			value map[int64]*pb.Game
		)
		value = make(map[int64]*pb.Game, 1)
		value[1] = &pb.Game{
			ID:       1,
			Title:    "lol one",
			SubTitle: "one",
			GameType: 1,
		}
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.AddCacheEpGames(c, value)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoCacheEpGameMap(t *testing.T) {
	convey.Convey("CacheEpGames", t, func(ctx convey.C) {
		var (
			c          = context.Background()
			cids       = []int64{1, 5}
			tp   int64 = 3
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.CacheEpGameMap(c, cids, tp)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestDaoAddCacheEpGameMap(t *testing.T) {
	convey.Convey("AddCacheEpGameMap", t, func(ctx convey.C) {
		var (
			c           = context.Background()
			value       = map[int64]int64{1: 2, 5: 3}
			tp    int64 = 3
		)

		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.AddCacheEpGameMap(c, value, tp)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
