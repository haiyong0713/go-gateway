package conf

import (
	"encoding/json"

	"github.com/BurntSushi/toml"

	"go-common/library/cache/redis"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/railgun"
)

// Config struct
type Config struct {
	BM                     *bm.ServerConfig
	Log                    *log.Config
	Redis                  *redis.Config
	ArchiveExtraBizRailgun *SingleRailgun
}

type SingleRailgun struct {
	Cfg     *railgun.Config
	Databus *railgun.DatabusV1Config
	Single  *railgun.SingleConfig
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	old, _ := json.Marshal(c)
	nw, _ := json.Marshal(tmp)
	log.Info("service config changed, old=%+v new=%+v", string(old), string(nw))
	*c = tmp
	return nil
}
