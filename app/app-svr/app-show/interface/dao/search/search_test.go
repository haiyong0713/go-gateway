package search

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

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

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestSearchList(t *testing.T) {
	Convey("SearchList", t, func() {
		var (
			rid, build, pn, ps                            int
			mid                                           int64
			ts                                            time.Time
			ip, order, tagName, platform, mobiApp, device string
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.searchURL).Reply(200).JSON(`{
			"code": 0,
			"result": [
				{
					"id": 77734719
				},
				{
					"id": 77304221
				},
				{
					"id": 77290276
				},
				{
					"id": 78405622
				},
				{
					"id": 78389904
				},
				{
					"id": 77779761
				},
				{
					"id": 77998005
				},
				{
					"id": 77080988
				},
				{
					"id": 79285784
				},
				{
					"id": 77642384
				},
				{
					"id": 77188078
				},
				{
					"id": 77216270
				},
				{
					"id": 77684756
				},
				{
					"id": 77516854
				},
				{
					"id": 78144848
				},
				{
					"id": 77514497
				},
				{
					"id": 78429604
				},
				{
					"id": 76940434
				},
				{
					"id": 78707782
				},
				{
					"id": 78794838
				}
			]
		}`)
		res, err := d.SearchList(ctx(), rid, build, pn, ps, mid, ts, ip, order, tagName, platform, mobiApp, device)
		Convey("Then err should be nil.date should not be nil.", func() {
			So(res, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}
