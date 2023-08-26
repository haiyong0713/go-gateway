package resource

import (
	"context"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFrontpagekeyDefaultPage(t *testing.T) {
	Convey("keyDefaultPage", t, func() {
		var (
			resourceId = int64(100)
		)
		Convey("When everything goes positive", func() {
			p1 := keyDefaultPage(resourceId)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
				log.Info("defaultKey : %+v", p1)
			})
		})
	})
}

func TestFrontpagekeyOnlinePage(t *testing.T) {
	Convey("keyOnlinePage", t, func() {
		var (
			resource_id = int64(100)
		)
		Convey("When everything goes positive", func() {
			p1 := keyOnlinePage(resource_id)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
				log.Info("onlineKey : %+v", p1)
			})
		})
	})
}

func TestFrontpagekeyHiddenPage(t *testing.T) {
	Convey("keyHiddenPage", t, func() {
		var (
			resource_id = int64(100)
		)
		Convey("When everything goes positive", func() {
			p1 := keyHiddenPage(resource_id)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
				log.Info("hiddenKey : %+v", p1)
			})
		})
	})
}

func TestFrontpageCacheDefaultPage(t *testing.T) {
	Convey("CacheDefaultPage", t, func() {
		var (
			c   = context.Background()
			req = &pb.FrontPageReq{
				ResourceId: 100,
			}
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheDefaultPage(c, req)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
				log.Info("DefaultPage : %+v", res)
			})
		})
	})
}

func TestFrontpageAddCacheDefaultPage(t *testing.T) {
	Convey("AddCacheDefaultPage", t, func() {
		var (
			c   = context.Background()
			req = &pb.FrontPageReq{
				ResourceId: 100,
			}
			res = &pb.FrontPage{
				Id:           1,
				Title:        "test1",
				Pos:          1,
				Logo:         "http://logtest1",
				Litpic:       "http://litpictest1",
				JumpUrl:      "http://jumpurltest1",
				IsSplitLayer: 1,
				SplitLayer:   "[{\"images\":[{\"src\":\"https://uat-i0.hdslb.com/bfs/vc/79e96640d55f220a40ae1026171f3f8ddae1afe4.jpg\"}],\"initial\":{\"scale\":1},\"offset\":{},\"offsetCurve\":{}}]",
				Style:        1,
			}
		)
		Convey("When everything goes positive", func() {
			err := d.AddCacheDefaultPage(c, req, res)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestFrontpageCacheHiddenPage(t *testing.T) {
	Convey("CacheHiddenPage", t, func() {
		var (
			c   = context.Background()
			req = &pb.FrontPageReq{
				ResourceId: 100,
			}
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheHiddenPage(c, req)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
				log.Info("HiddenPage : %+v", res)
			})
		})
	})
}

func TestFrontpageAddCacheHiddenPage(t *testing.T) {
	Convey("AddCacheHiddenPage", t, func() {
		var (
			c   = context.Background()
			req = &pb.FrontPageReq{
				ResourceId: 100,
			}
			frontPages []*pb.FrontPage
			res        = &pb.FrontPageResp{}
		)
		Convey("When everything goes positive", func() {
			item1 := &pb.FrontPage{
				Id:           2,
				Title:        "hidden2",
				Pos:          2,
				Logo:         "http://logtest2",
				Litpic:       "http://litpictest2",
				JumpUrl:      "http://jumpurltest2",
				IsSplitLayer: 1,
				SplitLayer:   "[{\"images\":[{\"src\":\"https://uat-i0.hdslb.com/bfs/vc/79e96640d55f220a40ae1026171f3f8ddae1afe4.jpg\"}],\"initial\":{\"scale\":1},\"offset\":{},\"offsetCurve\":{}}]",
				Style:        1,
			}
			item2 := &pb.FrontPage{
				Id:           3,
				Title:        "hidden3",
				Pos:          2,
				Logo:         "http://logtest2",
				Litpic:       "http://litpictest2",
				JumpUrl:      "http://jumpurltest2",
				IsSplitLayer: 1,
				SplitLayer:   "[{\"images\":[{\"src\":\"https://uat-i0.hdslb.com/bfs/vc/79e96640d55f220a40ae1026171f3f8ddae1afe4.jpg\"}],\"initial\":{\"scale\":1},\"offset\":{},\"offsetCurve\":{}}]",
				Style:        1,
			}
			frontPages = append(frontPages, item1)
			frontPages = append(frontPages, item2)
			res.Hidden = frontPages
			err := d.AddCacheHiddenPage(c, req, res.Hidden)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestFrontpageCacheOnlinePage(t *testing.T) {
	Convey("CacheOnlinePage", t, func() {
		var (
			c   = context.Background()
			req = &pb.FrontPageReq{
				ResourceId: 100,
			}
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheOnlinePage(c, req)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
				log.Info("OnlinePage : %+v", res)
			})
		})
	})
}

func TestFrontpageAddCacheOnlinePage(t *testing.T) {
	Convey("AddCacheOnlinePage", t, func() {
		var (
			c   = context.Background()
			req = &pb.FrontPageReq{
				ResourceId: 100,
			}
			res        = &pb.FrontPageResp{}
			frontpages []*pb.FrontPage
		)
		Convey("When everything goes positive", func() {
			item1 := &pb.FrontPage{
				Id:           2,
				Title:        "online2",
				Pos:          2,
				Logo:         "http://logtest2",
				Litpic:       "http://litpictest2",
				JumpUrl:      "http://jumpurltest2",
				IsSplitLayer: 1,
				SplitLayer:   "[{\"images\":[{\"src\":\"https://uat-i0.hdslb.com/bfs/vc/79e96640d55f220a40ae1026171f3f8ddae1afe4.jpg\"}],\"initial\":{\"scale\":1},\"offset\":{},\"offsetCurve\":{}}]",
				Style:        1,
			}
			item2 := &pb.FrontPage{
				Id:           3,
				Title:        "online3",
				Pos:          2,
				Logo:         "http://logtest2",
				Litpic:       "http://litpictest2",
				JumpUrl:      "http://jumpurltest2",
				IsSplitLayer: 1,
				SplitLayer:   "[{\"images\":[{\"src\":\"https://uat-i0.hdslb.com/bfs/vc/79e96640d55f220a40ae1026171f3f8ddae1afe4.jpg\"}],\"initial\":{\"scale\":1},\"offset\":{},\"offsetCurve\":{}}]",
				Style:        1,
			}
			frontpages = append(frontpages, item1)
			frontpages = append(frontpages, item2)
			res.Online = frontpages
			err := d.AddCacheOnlinePage(c, req, res.Online)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
