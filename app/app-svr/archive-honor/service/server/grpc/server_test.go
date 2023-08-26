package grpc

import (
	"context"
	"os"
	"testing"

	"go-gateway/app/app-svr/archive-honor/service/api"

	"github.com/smartystreets/goconvey/convey"
)

var (
	client api.ArchiveHonorClient
)

func TestMain(m *testing.M) {
	var err error
	client, err = api.NewClient(nil)
	if err != nil {
		panic(err)
	}
	// if os.Getenv("DEPLOY_ENV") != "" {
	// 	flag.Set("app_id", "main.app-svr.archive-honor-service")
	// 	flag.Set("conf_token", "6a91870821701a2c4e6b49d7fc270af2")
	// 	flag.Set("tree_id", "136937")
	// 	flag.Set("conf_version", "docker-1")
	// 	flag.Set("deploy_env", "uat")
	// 	flag.Set("conf_host", "config.bilibili.co")
	// 	flag.Set("conf_path", "/tmp")
	// 	flag.Set("region", "sh")
	// 	flag.Set("zone", "sh001")
	// } else {
	// flag.Set("conf", "../cmd/archive-honor-service.toml")
	// // }
	// flag.Parse()
	// if err := conf.Init(); err != nil {
	// 	panic(err)
	// }
	// d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestHonor(t *testing.T) {
	convey.Convey("TestHonor", t, func(ctx convey.C) {
		var (
			c          = context.Background()
			aid        = int64(1)
			err        error
			honorReply *api.HonorReply
		)
		ctx.Convey("TestHonor", func(ctx convey.C) {
			honorReply, err = client.Honor(c, &api.HonorRequest{Aid: aid})
			ctx.Println(err)
			if err == nil {
				ctx.Println(honorReply.Honor)
			}
		})
	})

}
