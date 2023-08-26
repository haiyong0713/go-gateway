package oss

import (
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/fawkes/service/conf"
)

var (
	d *Dao
)

// TestMain init ut main.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.fawkes-admin")
		flag.Set("conf_token", "f773b096629252c98ec2270f2d938a47")
		flag.Set("tree_id", "92020")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../cmd/fawkes-admin-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

//// TestFormLike dao ut.
//func TestFormLike(t *testing.T) {
//	convey.Convey("keyContribute", t, func(ctx convey.C) {
//		key := d.darkness("")
//		ctx.Convey("key should not be equal to xxxx", func(ctx convey.C) {
//			ctx.So(key, convey.ShouldNotEqual, "secret")
//		})
//	})
//}
