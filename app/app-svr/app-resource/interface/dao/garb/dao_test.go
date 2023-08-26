package garb

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"testing"

	"github.com/glycerine/goconvey/convey"

	"go-gateway/app/app-svr/app-resource/interface/conf"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

// TestMain dao ut.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-resource")
		flag.Set("conf_token", "z8JNX5MFIyDxyBsqwQyF6pnjWQ5YOA14")
		flag.Set("tree_id", "2722")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-resource-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestSkinList(t *testing.T) {
	convey.Convey("Egg", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			res, err := d.SkinList(context.Background(), []int64{1693})
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(res)
				ctx.Printf("%s", str)
			})
		})
	})
}

func TestSkinUserEquip(t *testing.T) {
	convey.Convey("Egg", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			res, err := d.SkinUserEquip(context.Background(), 15555180)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(res)
				ctx.Printf("%s", str)
			})
		})
	})
}

func TestSkinColorUserList(t *testing.T) {
	convey.Convey("Egg", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			res, err := d.SkinColorUserList(context.Background(), 15555180, 8961, "ios")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(res)
				ctx.Printf("%s", str)
			})
		})
	})
}
