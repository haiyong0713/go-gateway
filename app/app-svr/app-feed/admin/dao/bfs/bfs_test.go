package bfs

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/admin/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func init() {
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
	time.Sleep(time.Second)
}

func TestUpload(t *testing.T) {
	Convey("pull file bfs", t, func() {
		res, err := d.Upload(ctx(), "image/jpeg", bytes.NewReader([]byte("ddd")))
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestValidGif(t *testing.T) {
	Convey("pull file bfs", t, func() {
		file, err := os.Open("./gif.gif")
		if err != nil {
			return
		}
		defer file.Close()
		fInfo, err := file.Stat()
		if err != nil {
			panic(err)
		}
		buf := make([]byte, fInfo.Size())
		reader := bufio.NewReader(file)
		_, err = reader.Read(buf)
		err = d.ValidGif(ctx(), "1111", buf)
		So(err, ShouldBeNil)
	})
}
