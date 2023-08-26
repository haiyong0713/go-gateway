package dao

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/conf/paladin.v2"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	grpcmd "google.golang.org/grpc/metadata"
)

var mediaListConf = `
  [MediaListHTTP]
    Host = "https://api.bilibili.com"
    [MediaListHTTP.Config]
      key     = "0e9b9fcce22daaf1"
      secret  = "76aaccc1e756ac1c5b2ec135e6bd6b39"
      dial    = "50ms"
      timeout = "300ms"
`

func buildMediaListDao() *dao {
	d := &dao{}
	daoConf := &paladin.TOML{}
	_ = daoConf.UnmarshalText([]byte(mediaListConf))
	d.mediaListHTTP = newBmClient(d, "mediaListHTTP", daoConf)
	return d
}

func buildMediaListCtx() context.Context {
	md := grpcmd.New(map[string]string{
		"authorization": "identify_v1 be859854289e1b863a6b3bc8aa0cbe11",
	})
	c := grpcmd.NewIncomingContext(context.TODO(), md)
	return device.NewContext(c, device.Device{
		RawPlatform: "android",
		RawMobiApp:  "android",
		Build:       6570000,
		Buvid:       "XX225CE224BD4D86028E403912354191A8615",
	})
}

func TestMediaListReqContext_DoList(t *testing.T) {
	d := buildMediaListDao()
	reqCtx := &MediaListReqContext{
		Ctx: buildMediaListCtx(), PageSize: 99, FetchAll: true, FnDo: d.mediaListHTTP.Do, FnUri: d.mediaListHTTP.composeURI,
		Anchor: &v1.PlayItem{ItemType: model.PlayItemUGC, Oid: 680002266}, MaxWant: 1000,
	}
	start := time.Now()
	list, err := reqCtx.DoList(MediaListDoListOpt{Typ: 2, BizId: 41181058})
	end := time.Now()
	t.Logf("time cost: %d ms", end.Sub(start).Milliseconds())
	if err != nil {
		t.Fatalf("error DoList %v", err)
	}

	t.Logf("list len(%d)", len(list))
	i := 1
	for _, dt := range list {
		fmt.Println(i, dt.Title)
		i++
	}
}
