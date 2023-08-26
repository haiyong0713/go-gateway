package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	pb "go-gateway/app/app-svr/kvo/interface/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_userDocTaiShan(t *testing.T) {
	var (
		ctx   = context.TODO()
		mid   = int64(1111112645)
		buvid = "Y04C74038475378A45019D205CCF8259B329"
	)
	Convey("mid exist data", t, func() {
		rm, err := svr.userDocTaiShan(ctx, mid, "", 1)
		t.Log(string(rm), err)
		So(rm, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
	Convey("buvid exist data", t, func() {
		rm, err := svr.userDocTaiShan(ctx, 0, buvid, 1)
		t.Log(string(rm), err)
		So(rm, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestService_userDocTaiShan1(t *testing.T) {
	var (
		ctx   = context.TODO()
		mid   = int64(1)
		buvid = "Y04C74038475378A45019D205CCF8259B329s"
	)
	Convey("mid key not found", t, func() {
		rm, err := svr.userDocTaiShan(ctx, mid, "", 1)
		t.Log(err, rm)
		So(rm, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
	Convey("buvid key not found", t, func() {
		rm, err := svr.userDocTaiShan(ctx, 0, buvid, 1)
		t.Log(err, rm)
		So(rm, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestService_addUserDocTaiShan(t *testing.T) {
	var (
		ctx   = context.TODO()
		mid   = int64(1111112645)
		buvid = "Y04C74038475378A45019D205CCF8259B3291"
		cfg   = &pb.DanmuPlayerConfig{}
	)
	cfg.Default()
	cfg.PlayerDanmakuSwitch = true
	Convey("", t, func() {
		err := svr.addUserDocTaiShan(ctx, mid, buvid, 1, cfg)
		So(err, ShouldBeNil)
	})
	Convey("", t, func() {
		err := svr.addUserDocTaiShan(ctx, 0, buvid, 1, cfg)
		So(err, ShouldBeNil)
	})
}

func TestService_UserDocTaiShanN(t *testing.T) {
	var (
		ctx   = context.TODO()
		mid   = int64(1111112645)
		buvid = "Y04C74038475378A45019D205CCF8259B329"
		cfg   = &pb.DanmuPlayerConfig{}
		cfg1  *pb.DanmuPlayerConfig
		n     int = 1000
	)
	cfg.Default()
	Convey("", t, func() {
		for i := 0; i < n; i++ {
			old := cfg.PlayerDanmakuSwitch
			cfg.PlayerDanmakuSwitch = !old
			err := svr.addUserDocTaiShan(ctx, mid, buvid, 1, cfg)
			So(err, ShouldBeNil)
			time.Sleep(100 * time.Millisecond)
			rm, err := svr.userDocTaiShan(ctx, mid, "", 1)
			So(err, ShouldBeNil)
			_ = json.Unmarshal(rm, &cfg1)
			t.Log(cfg1.PlayerDanmakuSwitch, old)
			So(cfg1.PlayerDanmakuSwitch, ShouldEqual, !old)
		}
	})
}
