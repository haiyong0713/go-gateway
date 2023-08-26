package service

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	api "go-gateway/app/app-svr/app-player/interface/api/playurl"
	"go-gateway/app/app-svr/app-player/interface/conf"
	"go-gateway/app/app-svr/app-player/interface/model"
	playurlV2Api "go-gateway/app/app-svr/playurl/service/api/v2"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	s *Service
)

func init() {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-player")
		flag.Set("conf_token", "e477d98a7c5689623eca4f32f6af735c")
		flag.Set("tree_id", "52581")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
		flag.Parse()
		if err := conf.Init(); err != nil {
			panic(err)
		}
	} else {
		dir, _ := filepath.Abs("../cmd/app-player-test.toml")
		flag.Set("conf", dir)
		conf.Init()
	}
	s = New(conf.Conf)
}

func TestPlayURLV2(t *testing.T) {
	var params = &model.Param{
		AID:     1,
		MobiApp: "android",
		CID:     1,
		Qn:      32,
	}
	Convey("TestPlayURLV2", t, func() {
		s.PlayURLV2(context.TODO(), 1, params, 1)
	})
}

func TestPlayURLGRPC(t *testing.T) {
	var params = &model.Param{
		AID:           480043011,
		MobiApp:       "iphone",
		CID:           10206989,
		Qn:            32,
		Device:        "ios",
		Platform:      "iphone",
		Fnver:         0,
		Fnval:         16,
		Buvid:         "4493e276814ce9dbdb9b06214227c271",
		Build:         8450,
		TeenagersMode: 0,
	}
	Convey("TestPlayURLGRPC", t, func() {
		res, _ := s.PlayURLGRPC(context.TODO(), 27515255, params, 1)
		str, _ := json.Marshal(res)
		fmt.Printf("%s", str)
	})
}

// PlayView
func TestPlayView(t *testing.T) {
	var params = &model.Param{
		AID:           10113243,
		MobiApp:       "iphone",
		CID:           10162001,
		Qn:            112,
		Device:        "ios",
		Platform:      "iphone",
		Fnver:         0,
		Fnval:         16,
		Buvid:         "4493e276814ce9dbdb9b06214227c215",
		Build:         9120,
		TeenagersMode: 0,
		PreferCodecID: 7,
	}
	Convey("TestPlayURLGRPC", t, func() {
		res, _ := s.PlayView(context.Background(), 27515255, params, 1)
		str, _ := json.Marshal(res)
		fmt.Printf("%s", str)
	})
}

func TestPlayConfEdit(t *testing.T) {
	var params = &api.PlayConfEditReq{}
	params.PlayConf = append(params.PlayConf, &api.PlayConfState{ConfType: api.ConfType_FEEDBACK, Show: true})
	params.PlayConf = append(params.PlayConf, &api.PlayConfState{ConfType: api.ConfType_PLAYBACKMODE, Show: true})
	params.PlayConf = append(params.PlayConf, &api.PlayConfState{ConfType: api.ConfType_SCALEMODE, Show: true})
	Convey("PlayConfEdit", t, func() {
		res, err := s.PlayConfEdit(context.Background(), nil, params)
		str, _ := json.Marshal(res)
		fmt.Printf("%s", str)
		fmt.Printf("error(%v)", err)
	})
}

func TestProject(t *testing.T) {
	var params = &model.Param{
		AID:           480043011,
		MobiApp:       "iphone",
		CID:           10206989,
		Qn:            32,
		Device:        "ios",
		Platform:      "iphone",
		Fnver:         0,
		Fnval:         16,
		Buvid:         "4493e276814ce9dbdb9b06214227c271",
		Build:         8450,
		TeenagersMode: 0,
	}
	Convey("TestProject", t, func() {
		res, _ := s.Project(context.TODO(), 27515255, params, 1)
		//str, _ := json.Marshal(res)
		fmt.Printf("%+v", res)
	})
}

func TestCdn(t *testing.T) {
	var bk []string
	bk = append(bk, "https://upos-sz-mirrorcoso1.bilivideo.com/upgcxcode/51/66/207646651/207646651-1-30032.m4s?e=ig8euxZMXg8gNEV4NC")
	bk = append(bk, "https://sz-mirrorcoso1.bilivideo.com/upgcxcode/51/66/207646651/207646651-1-30032.m4s?e=ig8euxZMXg8gNEV4NC")
	dash := &playurlV2Api.DashItem{
		BaseUrl:   "https://upos-mirrorcoso1.bilivideo.com/upgcxcode/51/66/207646651/207646651-1-30032.m4s?e=ig8euxZMXg8gNEV4NC",
		BackupUrl: bk,
	}
	var video []*playurlV2Api.DashItem
	video = append(video, dash)
	p := &playurlV2Api.ResponseMsg{
		Dash: &playurlV2Api.ResponseDash{
			Video: video,
		},
	}
	Convey("TestCdn", t, func() {
		res := s.calCdnScore(context.TODO(), p, "", nil, 0)
		ss, _ := json.Marshal(res)
		fmt.Printf("%s", ss)
	})
}
