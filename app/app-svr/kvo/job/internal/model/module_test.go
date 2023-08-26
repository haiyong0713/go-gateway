package model

import (
	"encoding/json"
	"testing"

	v1 "go-gateway/app/app-svr/kvo/interface/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPlayerSha1(t *testing.T) {
	player := v1.DanmuPlayerConfig{}
	player.Default()
	player.PlayerDanmakuSwitchSave = true
	player.PlayerDanmakuEnableblocklist = true
	player.PlayerDanmakuOpacity = 0.32
	playerSha1 := player.ToPlayerSha1()

	Convey("eq", t, func() {
		bmsha1, _ := json.Marshal(playerSha1)
		var player1 v1.DanmuPlayerConfig
		_ = json.Unmarshal(bmsha1, &player1)
		srcbm, _ := json.Marshal(player)
		dstbm, _ := json.Marshal(player1)
		So(string(srcbm) == string(dstbm), ShouldBeTrue)
	})

}
