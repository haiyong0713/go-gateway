package dao

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go-common/component/metadata/device"
	"go-gateway/app/app-svr/app-listener/interface/conf"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
	arcMidV1 "go-gateway/app/app-svr/archive/middleware/v1"

	"go-common/library/conf/paladin.v2"

	grpcmd "google.golang.org/grpc/metadata"
)

var musicConf = `
  [Music]
    Host = "http://api.bilibili.com"
    [Music.Config]
      key     = "0e9b9fcce22daaf1"
      secret  = "76aaccc1e756ac1c5b2ec135e6bd6b39"
      dial    = "50ms"
      timeout = "400ms"
`

func buildMusicDao() *dao {
	d := &dao{}
	daoConf := &paladin.TOML{}
	_ = daoConf.UnmarshalText([]byte(musicConf))
	d.music = newBmClient(d, "music", daoConf)
	return d
}

func buildAuthCtx() context.Context {
	md := grpcmd.New(map[string]string{
		"authorization": "identify_v1 5eb5fb4432798e9fd40ebf3baf3d7ab1",
	})
	c := grpcmd.NewIncomingContext(context.TODO(), md)
	return device.NewContext(c, device.Device{
		RawPlatform: "android",
		RawMobiApp:  "android",
		Build:       6520400,
		Buvid:       "XX225CE224BD4D86028E403912354191A8615",
	})
}

func TestMusicPersonlStatus(t *testing.T) {
	d := buildMusicDao()
	conf.C = &conf.AppConfig{
		Switch: conf.SwitchStatus{
			LegacyMusicAPI: true,
		},
	}
	res, err := d.PersonalMenuStatusV1(context.TODO(), PersonalMenuStatusOpt{
		Mid: 532820215,
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%+v", res)
}

func TestMusicMenuList(t *testing.T) {
	d := buildMusicDao()
	res, err := d.MusicMenuListV1(buildAuthCtx(), MusicMenuListOpt{
		Typ: model.MenuCreated, Mid: 5904415,
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(res)
}

func TestMusicMenuDetail(t *testing.T) {
	d := buildMusicDao()
	res, err := d.MusicMenuDetailV1(buildAuthCtx(), MusicMenuDetailOpt{
		MenuId: 28769204,
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(res)
}

func TestSongPlayingDetail(t *testing.T) {
	d := buildMusicDao()
	res, err := d.SongPlayingDetailV1(buildAuthCtx(), SongPlayingDetailOpt{
		SongId: 13526, PlayerArgs: &arcMidV1.PlayerArgs{Qn: 80}, Mid: 5904415,
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%+v\n", res)
}

func TestSpaceSongList(t *testing.T) {
	d := buildMusicDao()
	ts := time.Now()
	res, err := d.SpaceSongList(buildAuthCtx(), SpaceSongListOpt{Mid: 162151104, WithCollaborator: true})
	if err != nil {
		t.Error(err)
		return
	}
	te := time.Now().Sub(ts)
	t.Logf("time cost: %d ms", te.Milliseconds())
	t.Logf("list len %d", len(res))
}
