package mod

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go-common/library/log"

	v1 "git.bilibili.co/bapis/bapis-go/bilibili/app/resource/v1"
	. "github.com/smartystreets/goconvey/convey"
	"go-common/library/conf/paladin.v2"

	"github.com/golang/protobuf/proto"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/mod"
)

var svr *Service

func init() {
	err := paladin.Init()
	if err != nil {
		panic(err)
	}
	cfg := &conf.Config{}
	if err = paladin.Get("app-resource.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	svr = New(cfg)
	time.Sleep(time.Second)
}

func Test_checkCondition(t *testing.T) {

	vc := mod.VersionConfig{
		ID:              1644008,
		VersionID:       2277708,
		Priority:        "high",
		AppVer:          "",
		SysVer:          "",
		Stime:           1659369600,
		Etime:           -62135596800,
		Scale:           "",
		ForbidenDevice:  "",
		Arch:            "",
		AppVers:         nil,
		SysVers:         nil,
		Scales:          nil,
		ForbidenDevices: nil,
		Archs:           nil,
		Mtime:           0,
	}

	configFunc := func(config *mod.VersionConfig) {
		if config == nil {
			return
		}
		if config.AppVer != "" {
			if err := json.Unmarshal([]byte(config.AppVer), &config.AppVers); err != nil {
				log.Error("日志告警 app_ver 值错误,config:%+v,error:%+v", config, err)
			}
			for _, ver := range config.AppVers {
				for cond := range ver {
					if !cond.Valid() {
						log.Error("日志告警 app_ver 值错误,config:%+v", config)
					}
				}
			}
		}
		if config.SysVer != "" {
			if err := json.Unmarshal([]byte(config.SysVer), &config.SysVers); err != nil {
				log.Error("日志告警 sys_ver 值错误,config:%+v,error:%+v", config, err)
			}
			for _, ver := range config.SysVers {
				for cond := range ver {
					if !cond.Valid() {
						log.Error("日志告警 sys_ver 值错误,config:%+v", config)
					}
				}
			}
		}
		if config.Scale != "" {
			vals := strings.Split(config.Scale, ",")
			config.Scales = map[mod.Scale]struct{}{}
			for _, val := range vals {
				scale := mod.Scale(val)
				if !scale.Valid() {
					log.Error("日志告警 scale 值错误,config:%+v", config)
					continue
				}
				config.Scales[scale] = struct{}{}
			}
		}
		if config.ForbidenDevice != "" {
			vals := strings.Split(config.ForbidenDevice, ",")
			config.ForbidenDevices = map[mod.Device]struct{}{}
			for _, val := range vals {
				device := mod.Device(val)
				if !device.Valid() {
					log.Error("日志告警 forbiden_device 值错误,config:%+v", config)
					continue
				}
				config.ForbidenDevices[device] = struct{}{}
			}
		}
		if config.Arch != "" {
			vals := strings.Split(config.Arch, ",")
			config.Archs = map[mod.Arch]struct{}{}
			for _, val := range vals {
				arch := mod.Arch(val)
				if !arch.Valid() {
					log.Error("日志告警 arch 值错误,config:%+v", config)
					continue
				}
				config.Archs[arch] = struct{}{}
			}
		}
	}
	configFunc(&vc)
	checkCondition(1234, 14300, 2, 0, mod.Device("iPhone"), time.Now(), &vc)
}

func Test_checkGray(t *testing.T) {
	checkGray("Y74FD875FC9696B349EC84583EF5EE4D05A9", 16104408, true, &mod.VersionGray{
		ID:             32509,
		VersionID:      2277708,
		Strategy:       2,
		Salt:           "540",
		BucketStart:    -1,
		BucketEnd:      -1,
		ManualDownload: false,
		Mtime:          1658840613,
	})
}

func TestService_GRPCList(t *testing.T) {
	Convey("lite", t, func() {
		now := time.Now()
		req := &v1.ListReq{
			//PoolName:   "pink",
			PoolName:   "",
			ModuleName: "",
			VersionList: []*v1.VersionListReq{
				{
					//PoolName: "pink",
					PoolName: "",
					Versions: []*v1.VersionReq{
						{
							ModuleName: "bnj-loss",
							Version:    4,
						},
						{
							ModuleName: "2021_bnj_theme_pkg",
							Version:    12,
						},
						{
							ModuleName: "TestModName",
							Version:    16,
						},
						{
							ModuleName: "TestModName-test2",
							Version:    2,
						},
						{
							ModuleName: "bnj-app",
							Version:    18,
						},
						{
							ModuleName: "ce_2021_bnj_theme_pkg",
							Version:    7,
						},
						{
							ModuleName: "singleFile",
							Version:    108,
						},
						{
							ModuleName: "999",
							Version:    1,
						},
						{
							ModuleName: "androidhome",
							Version:    3,
						},
						{
							ModuleName: "bnj",
							Version:    4,
						},
						{
							ModuleName: "bnj-app-new",
							Version:    2,
						},
						{
							ModuleName: "KV",
							Version:    1,
						},
						{
							ModuleName: "TestModName-test1",
							Version:    1,
						},
						{
							ModuleName: "bnj-app-8888",
							Version:    1,
						},
						{
							ModuleName: "bnj-loss-like",
							Version:    1,
						},
						{
							ModuleName: "TestModName-test",
							Version:    1,
						},
						{
							ModuleName: "TestModName-test3",
							Version:    1,
						},
						{
							ModuleName: "video_detail_like_animation",
							Version:    17,
						},
					},
				},
			},
			Env:    v1.EnvType_Release,
			SysVer: 0,
			Scale:  0,
			Arch:   0,
			Lite:   0,
		}
		list, err := svr.GRPCListWrap(context.Background(), "android", "dfafkkhuxxxxx", 111234544, 44444, "FDAJFLI", req, now)
		So(err, ShouldBeNil)
		size := proto.Size(list)

		req1 := req
		req1.Lite = 1

		listv1, err := svr.GRPCListWrap(context.Background(), "android", "dfafkkhuxxxxx", 111234544, 44444, "FDAJFLI", req1, now)
		size1 := proto.Size(listv1)
		So(err, ShouldBeNil)
		j, err := json.Marshal(listv1)

		req2 := req
		req2.Lite = 2

		listv2, err := svr.GRPCListWrap(context.Background(), "android", "dfafkkhuxxxxx", 111234544, 44444, "FDAJFLI", req2, now)
		size2 := proto.Size(listv2)
		So(err, ShouldBeNil)

		print(j)
		print(size, size1, size2)
		So(size, ShouldBeGreaterThan, size1)
		So(size1, ShouldBeGreaterThan, size2)
	})

	Convey("specific pool_mod", t, func() {
		now := time.Now()
		req := &v1.ListReq{
			PoolName:   "modmoss",
			ModuleName: "mosstest",
			Env:        v1.EnvType_Release,
			SysVer:     0,
			Scale:      0,
			Arch:       0,
			Lite:       0,
		}
		list, err := svr.GRPCListWrap(context.Background(), "android", "dfafkkhuxxxxx", 111234544, 44444, "FDAJFLI", req, now)
		So(err, ShouldBeNil)
		So(list, ShouldNotBeEmpty)
	})
}
