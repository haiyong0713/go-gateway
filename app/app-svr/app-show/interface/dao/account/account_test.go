package account

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-show")
		flag.Set("conf_token", "Pae4IDOeht4cHXCdOkay7sKeQwHxKOLA")
		flag.Set("tree_id", "2687")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-show-test.toml")
	}
	flag.Parse()
	cfg, err := confInit()
	if err != nil {
		panic(err)
	}
	d = New(cfg)
	os.Exit(m.Run())
}

func confInit() (*conf.Config, error) {
	err := paladin.Init()
	if err != nil {
		return nil, err
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("app-show.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func ctx() context.Context {
	return context.Background()
}

func TestCards3GRPC(t *testing.T) {
	Convey("Cards3GRPC", t, func() {
		_, err := d.Cards3GRPC(ctx(), []int64{1})
		Convey("Then err should be nil.", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestRelations3GRPC(t *testing.T) {
	Convey("Relations3GRPC", t, func() {
		_, err := d.Relations3GRPC(ctx(), 1, []int64{1})
		So(err, ShouldBeNil)
		Convey("Then err should be nil.", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestIsAttentionGRPC(t *testing.T) {
	Convey("get IsAttentionGRPC all", t, func() {
		res := d.IsAttentionGRPC(ctx(), []int64{1}, 1)
		Convey("Then err should be nil.", func() {
			So(res, ShouldBeEmpty)
		})
	})
}

func TestInfo3GRPC(t *testing.T) {
	Convey("get Info3GRPC all", t, func() {
		_, err := d.Info3GRPC(ctx(), int64(1))
		Convey("Then err should be nil.", func() {
			So(err, ShouldBeNil)
		})
	})
}
