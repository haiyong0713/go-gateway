package dao

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/web/interface/model/search"

	"github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

func TestDaoSearchAll(t *testing.T) {
	convey.Convey("SearchAll", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchAllArg{Pn: 1, Keyword: "test", Rid: 1}
			buvid = ""
			ua    = ""
		)
		typ := search.WxSearchType
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchAll(c, mid, arg, buvid, ua, typ)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchVideo(t *testing.T) {
	convey.Convey("SearchVideo", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypeVideo, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchVideo(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchBangumi(t *testing.T) {
	convey.Convey("SearchBangumi", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypeBangumi, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchBangumi(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchPGC(t *testing.T) {
	convey.Convey("SearchPGC", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypeMovie, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchMovie(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchLive(t *testing.T) {
	convey.Convey("SearchLive", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypeLive, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchLive(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchLiveRoom(t *testing.T) {
	convey.Convey("SearchLiveRoom", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypeLiveRoom, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchLiveRoom(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchLiveUser(t *testing.T) {
	convey.Convey("SearchLiveUser", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypeLiveUser, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchLiveUser(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchArticle(t *testing.T) {
	convey.Convey("SearchArticle", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypeArticle, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchArticle(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchSpecial(t *testing.T) {
	convey.Convey("SearchSpecial", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypeSpecial, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchSpecial(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchTopic(t *testing.T) {
	convey.Convey("SearchTopic", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypeTopic, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchTopic(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchUser(t *testing.T) {
	convey.Convey("SearchUser", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypeUser, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchUser(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchPhoto(t *testing.T) {
	convey.Convey("SearchPhoto", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(0)
			arg   = &search.SearchTypeArg{Pn: 1, SearchType: search.SearchTypePhoto, Keyword: "test"}
			buvid = ""
			ua    = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchPhoto(c, mid, arg, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchRec(t *testing.T) {
	convey.Convey("SearchRec", t, func(ctx convey.C) {
		var (
			c          = context.Background()
			mid        = int64(0)
			pn         = int(1)
			ps         = int(10)
			keyword    = "test"
			fromSource = ""
			buvid      = ""
			ua         = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SearchRec(c, mid, pn, ps, keyword, fromSource, buvid, ua)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchDefault(t *testing.T) {
	convey.Convey("SearchDefault", t, func(ctx convey.C) {
		var (
			c          = context.Background()
			mid        = int64(0)
			fromSource = ""
			buvid      = ""
			ua         = ""
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.SearchDefault(c, mid, fromSource, buvid, ua, true)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
				ctx.Printf("%+v", data)
			})
		})
	})
}

func TestDaoUpRecommend(t *testing.T) {
	convey.Convey("UpRecommend", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(2089809)
			arg = &search.SearchUpRecArg{ServiceArea: "reg_ok", Platform: "web"}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.searchUpRecURL).Reply(200).JSON(`{"msg":"success.","trackid":"1504440269061579120","code":0}`)
			rs, trackID, err := d.UpRecommend(c, mid, arg)
			ctx.Convey("Then err should be nil.rs,trackID should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(trackID, convey.ShouldNotBeNil)
				ctx.Println(trackID)
				ctx.Printf("%+v", rs)
			})
		})
	})
}

func TestDaoSearchEgg(t *testing.T) {
	convey.Convey("SearchEgg", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.SearchEgg(c)
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
			})
		})
	})
}
