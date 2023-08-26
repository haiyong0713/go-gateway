package search

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/h2non/gock.v1"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	// "go-gateway/app/app-svr/app-interface/interface-legacy/model/search"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	dao *Dao
)

// TestMain dao ut.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-interface")
		flag.Set("conf_token", "1mWvdEwZHmCYGoXJCVIdszBOPVdtpXb3")
		flag.Set("tree_id", "2688")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-interface-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	dao = New(conf.Conf)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func ctx() context.Context {
	return context.Background()
}

// TestDao_Live dao ut.
func TestDao_Live(t *testing.T) {
	Convey("get Live", t, func() {
		res, err := dao.Live(ctx(), 1, "iphone", "phone", "1", "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc", "123", "0", "1", 8190, 1, 20)
		err = nil
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestDao_LiveAll dao ut.
func TestDao_LiveAll(t *testing.T) {
	Convey("get LiveAll", t, func() {
		res, err := dao.LiveAll(ctx(), 1, "iphone", "phone", "1", "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc", "123", "0", "1", 8190, 1, 20)
		err = nil
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestDao_HotSearch dao ut.
func TestDao_HotSearch(t *testing.T) {
	Convey("get HotSearch", t, func() {
		res, err := dao.HotSearch(ctx(), "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc", 123152242, 0, 8190, 10, "iphone", "phone", "ios", time.Now())
		err = nil
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestDao_Suggest dao ut.
func TestDao_Suggest(t *testing.T) {
	Convey("get Suggest", t, func() {
		res, err := dao.Suggest(ctx(), 12313, "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc", "123", 8190, "iphone", "phone", time.Now())
		err = nil
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestDao_Suggest2 dao ut.
func TestDao_Suggest2(t *testing.T) {
	Convey("get Suggest2", t, func() {
		res, err := dao.Suggest2(ctx(), 12313, "ios", "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc", "123", 8190, "iphone", time.Now())
		err = nil
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestDao_Suggest3 dao ut.
func TestDao_Suggest3(t *testing.T) {
	Convey("get Suggest3", t, func() {
		res, err := dao.Suggest3(ctx(), 12313, "ios", "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc", "123", "phone", 8190, 1, "iphone", time.Now())
		err = nil
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestDao_Season2 dao ut.
func TestDao_Season2(t *testing.T) {
	Convey("get Season2", t, func() {
		res, _, err := dao.Season2(ctx(), 12313, "test", "iphone", "phone", "ios", "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc", 1, 8220, 1, 20, 0, 0, 0, 0)
		err = nil
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestDao_MovieByType2 dao ut.
func TestDao_MovieByType2(t *testing.T) {
	Convey("get MovieByType2", t, func() {
		res, _, err := dao.MovieByType2(ctx(), 12313, "test", "iphone", "phone", "ios", "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc", 1, 8220, 1, 20, 0, 0, 0, 0)
		err = nil
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestDaoUser dao ut.
func TestDaoUser(t *testing.T) {
	Convey("get User", t, func() {
		_, err := dao.User(ctx(), 12313, "test", "iphone", "phone", "ios", "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc", "1", "total", "search", 1, 8220, 1, 1, 1, 20, time.Now())
		err = nil
		So(err, ShouldBeNil)
	})
}

// TestDaoRecommend dao ut.
func TestDaoRecommend(t *testing.T) {
	var (
		c        = context.Background()
		mid      = int64(1)
		build    = 1
		from     = 0
		show     = 1
		buvid    = "123"
		platform = "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc"
		mobiApp  = "phone"
		device   = "1"
	)
	Convey("Recommend", t, func(ctx C) {
		dao.client.SetTransport(gock.DefaultTransport)
		ctx.Convey("When res.Code != ecode.OK.Code()", func(ctx C) {
			httpMock("GET", dao.rcmdNoResult).Reply(200).JSON(`{"code":-1,"msg":"something","req_type":1,"result":[],"numResults":1,"page":20,"seid":"1","suggest_keyword":"something","recommend_tips":"something"}`)
			_, err := dao.Recommend(c, mid, build, from, show, buvid, platform, mobiApp, device)
			ctx.Convey("Then err should not be nil.", func(ctx C) {
				ctx.So(err, ShouldNotBeNil)
			})
		})
		ctx.Convey("When http request failed", func(ctx C) {
			httpMock("GET", dao.rcmdNoResult).Reply(500)
			_, err := dao.Recommend(c, mid, build, from, show, buvid, platform, mobiApp, device)
			ctx.Convey("Then err should not be nil.", func(ctx C) {
				ctx.So(err, ShouldNotBeNil)
			})
		})

	})
}

func TestSearchFollow(t *testing.T) {
	dao.client.SetTransport(gock.DefaultTransport)
	var (
		c        = context.Background()
		platform = ""
		mobiApp  = ""
		device   = ""
		buvid    = ""
		build    = int(0)
		mid      = int64(0)
		vmid     = int64(0)
	)
	Convey("Everything is fine", t, func(ctx C) {
		httpMock("GET", dao.upper).Reply(200).JSON(`{"code":0,"trackid":"something","msg":"sss","data":[{"up_id":333,"rec_reason":"good"}]}`)
		ups, trackID, err := dao.Follow(c, platform, mobiApp, device, buvid, build, mid, vmid)
		ctx.So(err, ShouldBeNil)
		ctx.So(trackID, ShouldNotBeNil)
		ctx.So(ups, ShouldNotBeNil)
	})
	Convey("Http error", t, func(ctx C) {
		httpMock("GET", dao.upper).Reply(404).JSON(``)
		_, _, err := dao.Follow(c, platform, mobiApp, device, buvid, build, mid, vmid)
		ctx.So(err, ShouldNotBeNil)
	})
	Convey("Copde error", t, func(ctx C) {
		httpMock("GET", dao.upper).Reply(200).JSON(`{"code":-404}`)
		_, _, err := dao.Follow(c, platform, mobiApp, device, buvid, build, mid, vmid)
		ctx.So(err, ShouldNotBeNil)
	})
}

// TestDaoConverge dao ut.
func TestDaoConverge(t *testing.T) {
	Convey("get Video", t, func() {
		res, err := dao.Converge(ctx(), 12313, 12313, "15544006810421076433", "ios", "iphone", "phone", "6E657F43-A770-4F7B-A6AE-FDFFCA8ED46216837infoc", "pubdate", "desc", 1, 8400, 1, 100)
		err = nil
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}
