package service

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	pb "go-gateway/app/app-svr/kvo/interface/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_AddUserDoc(t *testing.T) {
	var (
		ctx = context.TODO()
		mid = int64(11112645678)
	)
	Convey("add player config ", t, func() {
		err := svr.AddUserDoc(ctx, mid, &pb.DmPlayerConfigReq{
			//UseDefaultConfig: &pb.PlayerDanmakuUseDefaultConfig{Value: false},
			Speed: &pb.PlayerDanmakuSpeed{Value: 60},
		}, "android", "", "")
		So(err, ShouldBeNil)
	})
}

func TestService_Document(t *testing.T) {
	var (
		ctx = context.TODO()
		mid = int64(11112645678)
	)
	Convey("add player config ", t, func() {
		_, _ = svr.DocumentMid(ctx, mid, "player", 1, 1, "ios")
		res, err := svr.DocumentMid(ctx, mid, "player", 1, 1, "ios")
		t.Logf("%s", res.Data)
		So(err, ShouldBeNil)
	})
}

func TestService_AddUserDoc1(t *testing.T) {
	var (
		ctx   = context.TODO()
		buvid = "Y04C74038475378A45019D205CCF8259B329"
	)
	Convey("add player config ", t, func() {
		err := svr.AddUserDoc(ctx, 0, &pb.DmPlayerConfigReq{
			//UseDefaultConfig: &pb.PlayerDanmakuUseDefaultConfig{Value: false},
			Speed: &pb.PlayerDanmakuSpeed{Value: 20},
		}, "android", buvid, "")
		So(err, ShouldBeNil)
		res, err := svr.DocumentBuvid(ctx, buvid, "player", "ios")
		t.Logf("%+v", string(res.Data))
		So(err, ShouldBeNil)
	})
}

func TestService_Document1(t *testing.T) {
	var (
		ctx   = context.TODO()
		buvid = "Y04C74038475378A45019D205CCF8259B329"
	)
	Convey("add player config ", t, func() {
		res, err := svr.DocumentBuvid(ctx, buvid, "player", "ios")
		t.Logf("%+v", string(res.Data))
		So(err, ShouldBeNil)
	})
}

func TestService_MoreDocument(t *testing.T) {
	var (
		ctx    = context.TODO()
		mid    = int64(1111112645)
		waiter = new(sync.WaitGroup)
	)
	Convey("add player config ", t, func() {
		waiter.Add(1)
		go func() {
			defer waiter.Done()
			res, _ := svr.DocumentMid(ctx, mid, "player", 1, 1, "ios")
			t.Logf("%+v", res)
		}()
		waiter.Add(1)
		go func() {
			defer waiter.Done()
			res, _ := svr.DocumentMid(ctx, mid, "player", 1, 1, "ios")
			t.Logf("%+v", res)
		}()
		waiter.Add(1)
		go func() {
			defer waiter.Done()
			res, _ := svr.DocumentMid(ctx, mid, "player", 1, 1, "ios")
			t.Logf("%+v", res)
		}()
		waiter.Add(1)
		go func() {
			defer waiter.Done()
			res, _ := svr.DocumentMid(ctx, mid, "player", 1, 1, "ios")
			t.Logf("%+v", res)
		}()
		waiter.Add(1)
		go func() {
			defer waiter.Done()
			res, _ := svr.DocumentMid(ctx, mid, "player", 1, 1, "ios")
			t.Logf("%+v", res)
		}()
		waiter.Wait()
		res, _ := svr.DocumentMid(ctx, mid, "player", 1, 1, "ios")
		t.Logf("%+v", res)
	})
}

func TestService_LRU(t *testing.T) {
	Convey("lru", t, func() {
		svr.localCache.Add(int64(1), json.RawMessage("ssss"))
		svr.localCache.Add(int64(2), json.RawMessage("ssss"))
		svr.localCache.Add(int64(3), json.RawMessage("ssss"))
		svr.localCache.Add(int64(4), json.RawMessage("ssss"))
		svr.localCache.Add(int64(5), json.RawMessage("ssss"))
		svr.localCache.Add(int64(6), json.RawMessage("ssss"))
		svr.localCache.Add(int64(7), json.RawMessage("ssss"))
		svr.localCache.Add(int64(8), json.RawMessage("ssss"))
		svr.localCache.Add(int64(9), json.RawMessage("ssss"))
		svr.localCache.Add(int64(10), json.RawMessage("ssss"))
	})
}
