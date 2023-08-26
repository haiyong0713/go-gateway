package playurl

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	pb "go-gateway/app/app-svr/playurl/service/api"
	"go-gateway/app/app-svr/playurl/service/conf"

	hqgrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurltvproj"
	v1 "git.bilibili.co/bapis/bapis-go/video/vod/playurlugc"

	"github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.playurl-service")
		flag.Set("conf_token", "eec9571409f31d4f8b55a6dfc84d99b8")
		flag.Set("tree_id", "76370")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestPlayurl(t *testing.T) {
	var (
		c      = context.TODO()
		params = &pb.PlayURLReq{
			Aid:      10111162,
			Platform: "android",
			Cid:      10135720,
			Qn:       32,
		}
		reqURL = "http://uat-videodispatch-ugc.bilibili.co/v3/playurl"
	)
	convey.Convey("Playurl", t, func(ctx convey.C) {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", reqURL).Reply(200).JSON(`{"from":"local","result":"suee","quality":32,"format":"flv480","timelength":8990,"accept_format":"flv720,flv480,flv360","accept_description":["720P","480P","360P"],"accept_quality":[64,32,16],"video_codecid":7,"video_project":true,"seek_param":"start","seek_type":"offset"}`)
		p, _, err := d.Playurl(c, params, false, false, reqURL)
		fmt.Printf("%+v", p)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		ctx.Reset(func() {
			gock.OffAll()
			d.client.SetTransport(http.DefaultClient.Transport)
		})
	})
}

func TestPlayurlV2(t *testing.T) {
	var (
		c      = context.TODO()
		params = &v1.RequestMsg{
			Cid:      10154746,
			Platform: "android",
			Qn:       112,
			IsSp:     true,
			Fnver:    0,
			Fnval:    0,
		}
		h5hq = false
		bs   []byte
	)
	convey.Convey("Playurl", t, func(ctx convey.C) {
		p, code, err := d.PlayurlV2(c, params, h5hq, false, false, false)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(p, convey.ShouldNotBeNil)
			ctx.So(code, convey.ShouldEqual, 0)
			if bs, err = json.Marshal(p); err == nil {
				ctx.Printf("response %s", bs)
			}
			params.Platform = _platformHtml5
			params.Fnval = 1
			params.IsSp = false
			p, code, err = d.PlayurlV2(c, params, h5hq, false, false, false)
			ctx.Convey("Then h5 err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(p, convey.ShouldNotBeNil)
				ctx.So(code, convey.ShouldEqual, 0)
				if bs, err = json.Marshal(p); err == nil {
					ctx.Printf("h5 response %s", bs)
				}
				h5hq = true
				p, code, err = d.PlayurlV2(c, params, h5hq, false, false, false)
				ctx.Convey("Then h5 hq err should be nil.", func(ctx convey.C) {
					ctx.So(err, convey.ShouldBeNil)
					ctx.So(p, convey.ShouldNotBeNil)
					ctx.So(code, convey.ShouldEqual, 0)
					if bs, err = json.Marshal(p); err == nil {
						ctx.Printf("h5 hq response %s", bs)
					}
				})
			})
		})
	})
}

func TestProject(t *testing.T) {
	var (
		c      = context.TODO()
		params = &hqgrpc.RequestMsg{
			Cid:      10176578,
			Platform: "android",
			Qn:       176,
			IsSp:     true,
			Fnver:    0,
			Fnval:    0,
		}
		bs []byte
	)
	convey.Convey("Playurl", t, func(ctx convey.C) {
		p, code, err := d.Project(c, params)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(p, convey.ShouldNotBeNil)
			ctx.So(code, convey.ShouldEqual, 0)
			if bs, err = json.Marshal(p); err == nil {
				ctx.Printf("response %s", bs)
			}
		})
	})
}

func TestPlayurl2(t *testing.T) {
	var (
		c      = context.Background()
		params = &v1.RequestMsg{
			Cid:       10210275,
			Platform:  "android",
			Qn:        120,
			IsSp:      true,
			Fnver:     0,
			Fnval:     16,
			Uip:       "172.22.34.51:16100",
			Mid:       42539596,
			FlvProj:   true,
			ForceHost: 1,
			Fourk:     true,
		}
	)
	convey.Convey("Playurl2", t, func(ctx convey.C) {
		p, err := d.playurlGRPC.Playurl2(c, params)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			str, _ := json.Marshal(p)
			fmt.Printf("%v\n", string(str))
			ctx.So(p, convey.ShouldNotBeNil)
		})
	})
	convey.Convey("playurlDisasterGRPC", t, func(ctx convey.C) {
		p, err := d.playurlDisasterGRPC.Playurl2(c, params)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			str, _ := json.Marshal(p)
			fmt.Printf("%v\n", string(str))
			ctx.So(p, convey.ShouldNotBeNil)
		})
	})
}

func TestPlayurlVolume(t *testing.T) {
	var (
		c   = context.TODO()
		cid = uint64(10341397)
		mid = uint64(123)
	)
	convey.Convey("PlayurlVolume", t, func(ctx convey.C) {
		res, err := d.PlayurlVolume(c, cid, mid)
		ress, _ := json.Marshal(res)
		fmt.Printf("%s", ress)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
