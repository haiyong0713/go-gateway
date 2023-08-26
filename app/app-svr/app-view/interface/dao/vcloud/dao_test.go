package vcloud

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-view/interface/conf"

	vcloud "git.bilibili.co/bapis/bapis-go/video/vod/playurlstory"
	"github.com/stretchr/testify/assert"
)

var (
	dao *Dao
)

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-view")
		flag.Set("conf_token", "3a4CNLBhdFbRQPs7B4QftGvXHtJo92xw")
		flag.Set("tree_id", "4575")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-view-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	dao = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestShortFormVideoInfo(t *testing.T) {
	ctx := context.Background()
	reply, err := dao.ShortFormVideoInfo(ctx, &vcloud.RequestMsg{
		Cids:      []uint64{222648222},
		Uip:       metadata.String(ctx, metadata.RemoteIP),
		Mid:       0,
		TfType:    vcloud.TFType(1),
		BackupNum: 1,
	}, 222648222)
	assert.Equal(t, err, nil)
	assert.NotEqual(t, reply, nil)
}
