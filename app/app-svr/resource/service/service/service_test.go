package service

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"go-common/library/cache/credis"
	"go-common/library/conf/paladin.v2"
	v1 "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/conf"

	"github.com/glycerine/goconvey/convey"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	s *Service
)

func init() {
	dir, _ := filepath.Abs("../cmd/resource-service-test.toml")
	flag.Set("conf", dir)
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("resource-service.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	s = New(cfg)
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		Reset(func() {
			c := context.Background()
			pool := credis.NewRedis(s.c.Redis.Ads.Config)
			pool.Conn(c).Do("FLUSHDB")
		})
		f(s)
	}
}

func TestEntrancesIsHidden(t *testing.T) {
	convey.Convey("EntrancesIsHidden", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			req = &v1.EntrancesIsHiddenRequest{Otype: 0, Oids: []int64{1, 2}, Build: 999999, Plat: 0, Channel: "xiaomi"}
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := s.EntrancesIsHidden(c, req)
			ctx.So(err, convey.ShouldBeNil)
			vv, _ := json.Marshal(reply.Infos)
			fmt.Printf("%s", vv)
		})
	})
}

func TestMngIcon(t *testing.T) {
	convey.Convey("MngIcon", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			req = &v1.MngIconRequest{Oids: []int64{1, 2}, Plat: 0}
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := s.MngIcon(c, req)
			ctx.So(err, convey.ShouldBeNil)
			vv, _ := json.Marshal(reply.Info)
			fmt.Printf("%s", vv)
		})
	})
}
