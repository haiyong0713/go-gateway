package service

import (
	"context"
	"encoding/json"
	"testing"

	v1 "go-gateway/app/app-svr/kvo/interface/api"
	"go-gateway/app/app-svr/kvo/job/internal/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_addDmConfig(t *testing.T) {
	var (
		ctx               = context.TODO()
		buvid model.Buvid = "Y04C74038475378A45019D205CCF8259B329"
	)
	t.Log(buvid.Crc63())
	Convey("", t, func() {
		err := svr.addConfig(ctx, string(buvid), map[int]v1.ConfigModify{
			v1.DmCfg: &v1.DmPlayerConfigReq{
				SwitchSave: &v1.PlayerDanmakuSwitchSave{
					Value: false,
				},
				Speed: &v1.PlayerDanmakuSpeed{
					Value: 50,
				},
			},
		}, v1.PlatFormAndroid)
		So(err, ShouldBeNil)
	})
}

func TestService_JsonMsg(t *testing.T) {
	var (
		err error
		bm  = `{"Body":{"speed":{"value":20}},"Mid":0,"Platform":"android","Buvid":"Y04C74038475378A45019D205CCF8259B329"}`
	)
	Convey("", t, func() {
		playerConfig := new(model.CfgMessage)
		playerConfig.Body = v1.NewConfigModify(v1.DmCfg)
		if err = json.Unmarshal([]byte(bm), playerConfig); err != nil {
			t.Logf("s.newMessagePlayer() json.Unmarshal() databus action(%s) error(%v)", bm, err)
			return
		}
		t.Logf("%+v", playerConfig)
		So(err, ShouldBeNil)
	})
}
