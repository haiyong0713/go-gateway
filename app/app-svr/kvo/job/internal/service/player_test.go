package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-common/library/conf/paladin"
	"go-common/library/queue/databus"
	v1 "go-gateway/app/app-svr/kvo/interface/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_addPlayerConfig(t *testing.T) {
	var (
		ctx = context.TODO()
		mid = int64(1111112645)
	)
	Convey("add player config ", t, func() {
		err := svr.addPlayerConfig(ctx, mid, &v1.DmPlayerConfigReq{
			SwitchSave: &v1.PlayerDanmakuSwitchSave{
				Value: false,
			},
			Blockspecial: &v1.PlayerDanmakuBlockspecial{
				Value: false,
			},
			Speed: &v1.PlayerDanmakuSpeed{
				Value: 40,
			},
		}, v1.PlatFormAndroid)
		So(err, ShouldBeNil)
		t.Logf("err:%v", err)
	})
}

func TestService_PlayerSend(t *testing.T) {
	var (
		databusCfg struct {
			PlayerPub *databus.Config
		}
		ctx = context.TODO()
		mid = int64(111004263)
		msg = &struct {
			Body     *v1.DmPlayerConfigReq
			Mid      int64
			Platform string
		}{
			Body: &v1.DmPlayerConfigReq{
				Blocktop: &v1.PlayerDanmakuBlocktop{
					Value: true,
				},
			},
			Mid:      mid,
			Platform: v1.PlatFormAndroid,
		}
		data []byte
		err  error
	)
	if err = paladin.Get("databus.toml").UnmarshalTOML(&databusCfg); err != nil {
		panic(err)
	}
	if data, err = json.Marshal(msg); err != nil {
		panic(err)
	}
	act := &struct {
		Action string          `json:"action"`
		Data   json.RawMessage `json:"data"`
	}{
		Action: "dm_player_config",
		Data:   data,
	}

	playerPub := databus.New(databusCfg.PlayerPub)
	Convey("send msg", t, func() {
		err := playerPub.Send(ctx, fmt.Sprint(mid), act)
		So(err, ShouldBeNil)
	})
}
