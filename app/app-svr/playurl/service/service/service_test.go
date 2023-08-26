package service

import (
	"context"
	"flag"
	"fmt"
	"go-common/library/conf/paladin.v2"
	"os"
	"path/filepath"
	"testing"

	"go-common/library/net/trace"

	"git.bilibili.co/bapis/bapis-go/bilibili/app/distribution"
	disbase "git.bilibili.co/bapis/bapis-go/bilibili/app/distribution"
	dp "git.bilibili.co/bapis/bapis-go/bilibili/app/distribution/setting/play"

	pb "go-gateway/app/app-svr/playurl/service/api"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
	"go-gateway/app/app-svr/playurl/service/conf"
	"go-gateway/app/app-svr/playurl/service/model"

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

var (
	s *Service
)

// TestMain is
func TestMain(m *testing.M) {
	dir, _ := filepath.Abs("../cmd/playurl-service-test.toml")
	flag.Set("conf", dir)
	var config = &conf.Config{}
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	if err := paladin.Watch("playurl-service-test.toml", config); err != nil {
		panic(err)
	}
	// taishan缓存必须初始化trace，如果kv拿不到trace将无法校验通过
	trace.Init(config.Tracer)
	m.Run()
	os.Exit(0)
}

func TestServiceloadPasterCID(t *testing.T) {
	convey.Convey("loadPasterCID", t, func(ctx convey.C) {
		s.loadPasterCID()
	})
}

func TestServicePlayURL(t *testing.T) {
	var (
		c   = context.Background()
		req = &pb.PlayURLReq{Aid: 111, Cid: 10162312}
	)
	convey.Convey("PlayURL", t, func(ctx convey.C) {
		reply, err := s.PlayURL(c, req)
		fmt.Printf("%+v", reply)
		ctx.Convey("Then err should be nil.reply should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(reply, convey.ShouldNotBeNil)
		})
	})
}

func TestServicePing(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("Ping", t, func(ctx convey.C) {
		err := s.Ping(c)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestServiceClose(t *testing.T) {
	convey.Convey("Close", t, func(ctx convey.C) {
		s.Close()
		ctx.Convey("No return values", func(ctx convey.C) {
		})
	})
}

func TestServiceSteinsPreview(t *testing.T) {
	var (
		c   = context.Background()
		req = &pb.SteinsPreviewReq{
			Aid: 10113421,
			Cid: 10162312,
			Mid: 1,
		}
	)
	convey.Convey("SteinsPreview", t, func(ctx convey.C) {
		reply, err := s.SteinsPreview(c, req)
		ctx.Convey("Then err should be nil.reply should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(reply, convey.ShouldNotBeNil)
		})
	})
}

func TestProject(t *testing.T) {
	var (
		c   = context.Background()
		req = &v2.ProjectReq{Aid: 880078582, Cid: 10176578}
	)
	convey.Convey("Project", t, func(ctx convey.C) {
		reply, err := s.Project(c, req)
		fmt.Printf("%+v", reply)
		ctx.Convey("Then err should be nil.reply should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(reply, convey.ShouldNotBeNil)
		})
	})
}

// PlayConfEdit
func TestPlayConfEdit(t *testing.T) {
	var (
		c   = context.Background()
		req = &v2.PlayConfEditReq{Buvid: "45a31014725e1989",
			PlayConf: []*v2.PlayConfState{{ConfType: v2.ConfType_FLIPCONF}, {ConfType: v2.ConfType_FEEDBACK}}}
	)
	convey.Convey("Project", t, func(ctx convey.C) {
		_, err := s.PlayConfEdit(c, req, "0")
		ctx.Convey("Then err should be nil.reply should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

// PlayView
func TestPlayView(t *testing.T) {
	var (
		c   = context.Background()
		req = &v2.PlayViewReq{Buvid: "45a31014725e1989", Aid: 10318716, Cid: 10211289, Fnval: 16, VerifyVip: 1, TeenagersMode: 1, VoiceBalance: 1}
	)
	convey.Convey("PlayView", t, func(ctx convey.C) {
		rly, err := s.PlayView(c, req)
		ctx.Convey("Then err should be nil.reply should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			fmt.Printf("%v", rly)
		})
	})
}

// PlayView
func TestCheckChronos(t *testing.T) {
	var (
		req = &v2.ChronosPkgReq{
			Aid:      10318718,
			Cid:      10211291,
			Mid:      0,
			MobiApp:  "android_tv_yst",
			Build:    103100,
			Platform: "android",
			Buvid:    "XYA5810A02794E4B75FF812B1448EF0CFCAD7",
		}
	)
	convey.Convey("Project", t, func(ctx convey.C) {
		rly := s.checkChronos(req)
		ctx.Convey("Then err should be nil.reply should not be nil.", func(ctx convey.C) {
			ctx.So(rly, convey.ShouldNotBeNil)
		})
	})
}

func TestTranslateDistributionReply(t *testing.T) {
	playConf := &dp.PlayConfig{
		EnableSubtitle: &disbase.BoolValue{
			Value: true,
		},
		ColorFilter: &disbase.Int64Value{
			Value: 1,
		},
	}
	pany, err := playConf.Marshal()
	assert.NoError(t, err)
	playAny := &types.Any{
		TypeUrl: _distributionPlayConf,
		Value:   pany,
	}
	assert.NoError(t, err)
	cloudConf := &dp.CloudPlayConfig{
		EnablePanorama: &disbase.BoolValue{
			Value: true,
		},
	}
	cany, err := cloudConf.Marshal()
	assert.NoError(t, err)
	cloudAny := &types.Any{
		TypeUrl: _distributionCloudPlayConf,
		Value:   cany,
	}
	assert.NoError(t, err)
	anys := []*types.Any{playAny, cloudAny}
	in := &distribution.GetUserPreferenceReply{
		Value: anys,
	}
	playReply, cloudReply, err := translateDistributionReply(in)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), playReply.ColorFilter.GetValue())
	assert.Equal(t, true, playReply.EnableSubtitle.GetValue())
	assert.Equal(t, true, cloudReply.EnablePanorama.GetValue())

	intConfValue := abilityConfIntValSetter(context.Background(), &ColorFilterWithExp{EnumValue: EnumValue{Value: playReply.ColorFilter.GetValue()}})
	assert.Equal(t, int64(1), intConfValue.GetSelectedVal())
	boolConfValue := abilityConfBoolValSetter(context.Background(), &SubtitleWithExp{BoolValue: BoolValue{Value: playReply.EnableSubtitle.GetValue()}})
	assert.Equal(t, true, boolConfValue.GetSwitchVal())
}

func TestBatchGetAnyByConfValue(t *testing.T) {
	anyMap := map[string]string{
		_distributionCloudPlayConf: _distributionCloudPlayConf,
		_distributionPlayConf:      _distributionPlayConf,
	}
	req := []*model.ConfValueEdit{
		{
			ConfType:  v2.ConfType_BACKGROUNDPLAY,
			ConfValue: &v2.ConfValue{Value: &v2.ConfValue_SwitchVal{SwitchVal: true}},
		},
		{
			ConfType:  v2.ConfType_COLORFILTER,
			ConfValue: &v2.ConfValue{Value: &v2.ConfValue_SelectedVal{SelectedVal: 3}},
		},
	}
	anys, err := convertConfValueToAnys(req)
	assert.NoError(t, err)

	for _, any := range anys {
		assert.NotNil(t, any)
		typeUrl, ok := anyMap[any.TypeUrl]
		assert.Equal(t, true, ok)
		if typeUrl == _distributionPlayConf {
			playConf := &dp.PlayConfig{}
			err := playConf.Unmarshal(any.Value)
			assert.NoError(t, err)
			assert.Equal(t, int64(3), playConf.ColorFilter.GetValue())
		} else {
			cloudConf := &dp.CloudPlayConfig{}
			err := cloudConf.Unmarshal(any.Value)
			assert.NoError(t, err)
			assert.Equal(t, true, cloudConf.EnableBackground.GetValue())
		}
	}
}

func TestPCDN(t *testing.T) {
	var (
		bpcdnInfo = map[string]string{
			"16_1": "http://123/123",
			"16_2": "http://abc/123?a=1&b=1",
			"32_3": "http://!@#$&*^%(**!)(& /x?x=1",
		}
		baseurl     = "http://upos-sz-mirrorcoso1.bilivideo.com/upgcxcode/00/02/735740200/735740200-1-30064.m4s?e=ig8eu"
		playurlInfo = &v2.PlayUrlInfo{Playurl: &v2.ResponseMsg{Dash: &v2.ResponseDash{
			Video: []*v2.DashItem{
				{
					Id:      16,
					Codecid: 1,
					BaseUrl: baseurl,
				},
				{
					Id:      16,
					Codecid: 2,
					BaseUrl: baseurl,
				},
				{
					Id:      32,
					Codecid: 3,
					BaseUrl: baseurl,
				},
			},
		}}}
	)
	joinPCDNToPlayUrlInfo(bpcdnInfo, playurlInfo)
	for _, v := range playurlInfo.Playurl.Dash.Video {
		t.Log(v.Codecid, ":", v.BaseUrl)
	}
}
