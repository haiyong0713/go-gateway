package grpc

import (
	"context"
	"go-gateway/app/app-svr/ugc-season/service/api"
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	client api.UGCSeasonClient
)

func TestMain(m *testing.M) {
	var err error
	client, err = api.NewClient(nil)
	if err != nil {
		panic(err)
	}
	// if os.Getenv("DEPLOY_ENV") != "" {
	// 	flag.Set("app_id", "main.app-svr.ugc-season-service")
	// 	flag.Set("conf_token", "82188905f73c6d658c9093f87c27a051")
	// 	flag.Set("tree_id", "117844")
	// 	flag.Set("conf_version", "docker-1")
	// 	flag.Set("deploy_env", "uat")
	// 	flag.Set("conf_host", "config.bilibili.co")
	// 	flag.Set("conf_path", "/tmp")
	// 	flag.Set("region", "sh")
	// 	flag.Set("zone", "sh001")
	// } else {
	// flag.Set("conf", "../cmd/ugc-season-service.toml")
	// // }
	// flag.Parse()
	// if err := conf.Init(); err != nil {
	// 	panic(err)
	// }
	// d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestSeason(t *testing.T) {
	convey.Convey("TestSeason", t, func(ctx convey.C) {
		var (
			c           = context.Background()
			sid         = int64(784)
			err         error
			seasonReply *api.SeasonReply
			viewReply   *api.ViewReply
			statReply   *api.StatReply
			statsReply  *api.StatsReply
		)
		ctx.Convey("TestSeason", func(ctx convey.C) {
			seasonReply, err = client.Season(c, &api.SeasonRequest{SeasonID: sid})
			ctx.Println(err)
			if err == nil {
				ctx.Println(seasonReply.Season)
			}
		})
		ctx.Convey("TestView", func(ctx convey.C) {
			viewReply, err = client.View(c, &api.ViewRequest{SeasonID: sid})
			ctx.Println(err)
			if err == nil {
				ctx.Println(viewReply.View.Season)
				ctx.Println(viewReply.View.Sections)
			}
		})
		ctx.Convey("TestStat", func(ctx convey.C) {
			statReply, err = client.Stat(c, &api.StatRequest{SeasonID: sid})
			ctx.Println(err)
			if err == nil {
				ctx.Println(statReply.Stat)
			}
		})
		ctx.Convey("TestStats", func(ctx convey.C) {
			statsReply, err = client.Stats(c, &api.StatsRequest{SeasonIDs: []int64{sid}})
			ctx.Println(err)
			if err == nil {
				ctx.Println(statsReply.Stats)
			}
		})
		ctx.Convey("TestUpCache", func(ctx convey.C) {
			_, err = client.UpCache(c, &api.UpCacheRequest{SeasonID: sid, Action: "update"})
			ctx.Println(err)
		})
	})

}

func TestUpperList(t *testing.T) {
	convey.Convey("UpperList", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			mid   = int64(27515255)
			err   error
			reply *api.UpperListReply
		)
		ctx.Convey("UpperList", func(ctx convey.C) {
			reply, err = client.UpperList(c, &api.UpperListRequest{Mid: mid, PageNum: 1, PageSize: 10})
			ctx.Println(err)
			if err == nil {
				ctx.Println(reply.Seasons)
			}
		})
	})
}
