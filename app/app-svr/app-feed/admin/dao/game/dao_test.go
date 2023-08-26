package game

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/conf"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app", "feed-admin")
		flag.Set("app_id", "main.web-svr.feed-admin")
		flag.Set("conf_token", "e0d2b216a460c8f8492473a2e3cdd218")
		flag.Set("tree_id", "45266")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/feed-admin-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	d.client.SetTransport(gock.DefaultTransport)
	os.Exit(m.Run())
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestDao_GameInfoApp(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("GameInfoApp", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.gameURL+_gameInfoApp).Reply(200).JSON(`{"code": 0,"data":{"ID":12,"Title":"测试游戏","Image":"https://i0.hdslb.com/bfs/game/fc49cfd3b8732c6ed939b39e48c9d39601b47e2d.png"}}`)
			_, err := d.GameInfoApp(c, 151, 1)
			ctx.Convey("Then err should be nil.p should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_GamePC(t *testing.T) {
	var (
		id = int64(12)
		c  = context.Background()
	)
	convey.Convey("GamePC", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.gameURL+_rcmdPc).Reply(200).JSON(`{"code": 0,"data":{"ID":12,"Title":"测试游戏","Image":"https://i0.hdslb.com/bfs/game/fc49cfd3b8732c6ed939b39e48c9d39601b47e2d.png"}}`)
			p, err := d.GamePC(c, _rcmdPc, id)
			ctx.Convey("Then err should be nil.p should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(p, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_SearchGame(t *testing.T) {
	var (
		id = int64(28)
		c  = context.Background()
	)
	convey.Convey("SearchGame", t, func(ctx convey.C) {
		defer gock.OffAll()
		httpMock("GET", d.gameURL+_searchPc).Reply(200).JSON(`{"code": 0,"data":{"ID":12,"Title":"测试游戏","Image":"https://i0.hdslb.com/bfs/game/fc49cfd3b8732c6ed939b39e48c9d39601b47e2d.png"}}`)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p, err := d.SearchGame(c, id)
			ctx.Convey("Then err should be nil.p should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(p, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_WebRcmdGame(t *testing.T) {
	var (
		id = int64(12)
		c  = context.Background()
	)
	convey.Convey("WebRcmdGame", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.gameURL+_rcmdPc).Reply(200).JSON(`{"code": 0,"data":{"ID":12,"Title":"测试游戏","Image":"https://i0.hdslb.com/bfs/game/fc49cfd3b8732c6ed939b39e48c9d39601b47e2d.png"}}`)
			p, err := d.WebRcmdGame(c, id)
			ctx.Convey("Then err should be nil.p should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(p, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_GamesPC(t *testing.T) {
	var (
		id = int64(12)
		c  = context.Background()
	)
	convey.Convey("GamesPC", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			defer gock.OffAll()
			httpMock("GET", d.gameURL+_rcmdPc).Reply(200).JSON(`{"code": 0,"data":{"ID":12,"Title":"测试游戏","Image":"https://i0.hdslb.com/bfs/game/fc49cfd3b8732c6ed939b39e48c9d39601b47e2d.png"}}`)
			p, err := d.GamesPCInfo(c, id)
			ctx.Convey("Then err should be nil.p should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(p, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(p)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_GameInfo(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("PassportQueryByMids", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			httpMock("GET", d.gameURL+_gameInfoURI).Reply(200).JSON(`{"code":0,"message":"","data":{"is_online":true,"game_base_id":49,"game_name":"111命运-冠位指定（Fate/GO）","game_icon":"https://uat-i0.hdslb.com/bfs/game/a94e53322f6ea2d57ff85382f0329750ad1d427a.png","game_link":"bilibili://game_center/detail?id=49\u0026sourceType=adPut","game_status":0,"begin_date":0}}`)
			p, err := d.GameInfo(c, 49, 1)
			ctx.Convey("Then err should be nil.p should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(p, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(p)
			fmt.Println(string(bs))
		})
	})
}
