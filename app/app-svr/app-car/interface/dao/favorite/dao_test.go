package favorite

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-car/interface/conf"

	"github.com/glycerine/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-car")
		flag.Set("conf_token", "2c36153a9c62b282e740ae1ba31cd8ad")
		flag.Set("tree_id", "275976")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		dir, _ := filepath.Abs("../../cmd/app-car.toml")
		flag.Set("conf", dir)
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func ctx() context.Context {
	return context.Background()
}

func TestUserToViews(t *testing.T) {
	var (
		mid int64
		c   = context.TODO()
		pn  = 1
		ps  = 20
	)
	convey.Convey("UserToViews", t, func(ctx convey.C) {
		res, err := d.UserToViews(c, mid, pn, ps)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestFavoritesAll(t *testing.T) {
	var (
		mid, vmid, favid int64
		favtype          int
		c                = context.TODO()
		pn               = 1
		ps               = 20
	)
	convey.Convey("FavoritesAll", t, func(ctx convey.C) {
		res, err := d.FavoritesAll(c, mid, vmid, favid, favtype, pn, ps)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUserFolders(t *testing.T) {
	var (
		mid, oid int64
		favtype  int
		c        = context.TODO()
	)
	convey.Convey("UserFolders", t, func(ctx convey.C) {
		res, err := d.UserFolders(c, mid, oid, favtype)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestFolders(t *testing.T) {
	var (
		ids     []int64
		mid     int64
		favtype int
		c       = context.TODO()
	)
	convey.Convey("Folders", t, func(ctx convey.C) {
		res, err := d.Folders(c, ids, favtype, mid)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestFolderSpace(t *testing.T) {
	var (
		accessKey string
		mid       int64
		c         = context.TODO()
	)
	convey.Convey("FolderSpace", t, func(ctx convey.C) {
		res, err := d.FolderSpace(c, accessKey, mid)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
