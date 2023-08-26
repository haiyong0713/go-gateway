package live

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"

	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	"github.com/golang/mock/gomock"
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

func TestFeed(t *testing.T) {
	Convey("Feed", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.live).Reply(200).JSON(`{"code":0,"count":1,"lives":[{"owner":{"face":"xxx","mid":1,"name":"xxxx"}}]}`)
		res, err := d.Feed(ctx(), 1, "", "", time.Now())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestRecommend(t *testing.T) {
	Convey("Recommend", t, func() {
		d.clientAsyn.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.rec).Reply(200).JSON(`{
			"code": 0,
			"data": {
				"count": 1,
				"lives": {
					"subject": [{
						"owner": {
							"face": "xxx",
							"mid": 1,
							"name": "xxxx"
						}
					}],
					"hot": [{
						"owner": {
							"face": "xxx",
							"mid": 1,
							"name": "xxxx"
						}
					}]
				}
			}
		}`)
		res, err := d.Recommend(time.Now())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestGetMultiple(t *testing.T) {
	Convey("GetMultiple", t, func() {
		var (
			roomIds  []int64
			mockCtrl = gomock.NewController(t)
			res      map[int64]*livexroom.Infos
			err      error
		)
		defer mockCtrl.Finish()
		mockArc := livexroom.NewMockRoomClient(mockCtrl)
		d.roomRPCClient = mockArc
		mockArc.EXPECT().GetMultiple(context.TODO(), gomock.Any()).Return(res, nil)
		res, err = d.GetMultiple(ctx(), roomIds)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
