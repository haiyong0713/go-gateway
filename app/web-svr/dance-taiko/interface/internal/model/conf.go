package model

import (
	"net/http"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
)

type Conf struct {
	Frames        []*KeyFrame
	OK            float64
	Perfect       float64
	Good          float64
	ExampleDevice []string
	Boundary      int64 // 关键帧的左右边界
	Normalizing   bool  // 是否需要归一
	MaxScore      int64 // 最高分
	OttCfg        *OttConfig
	DemoExpire    xtime.Duration
	BwsCfg        *BwsConfig
}

type KeyFrame struct {
	Aid    int64
	Frames []int64
}

type OttConfig struct {
	Qn           int    // 默认清晰度
	ImgUrl       string // 教学图示
	PlayersMax   int    // 最大加入人数
	DefaultScore int    // 默认分数
	QRCodeUrl    string
	QRCodeMsg    string
	QRCodeSize   int
	Expire       *ExpireCfg
}

type ExpireCfg struct {
	RankExpire   xtime.Duration
	FramesExpire xtime.Duration
	GameExpire   xtime.Duration
}

type BwsConfig struct {
	Host       string
	NeedEnergy int
	GameId     int
	EnableBws  bool
}

type HttpConfig struct {
	Server      http.Server
	DanceClient bm.ClientConfig
}

func (c *Conf) Set(text string) error {
	var tmp Conf
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	log.Info("progress-service-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}
