package dwtime

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	eggdao "go-gateway/app/app-svr/app-resource/interface/dao/egg"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

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

var (
	d   *Dao
	egg *eggdao.Dao
)

func ctx() context.Context {
	return context.Background()
}

func TestResolveDeeplinkMetaAbIdOnline(t *testing.T) {
	assert.Equal(t, "yuz_7", "yuz_7")
}

func TestCdnPeakHours(t *testing.T) {
	Convey("CdnPeakHours", t, func() {
		_, err := egg.Egg(context.Background(), time.Now())

		domain := "i0.hdslb.com"
		day := "20220523"
		value, err := d.CdnPeakHours(ctx(), domain, day)
		bs, _ := json.Marshal(value)
		fmt.Printf("%s", string(bs))
		Convey("Then err should not be nil", func() {
			So(err, ShouldBeNil)
		})
	})
}
