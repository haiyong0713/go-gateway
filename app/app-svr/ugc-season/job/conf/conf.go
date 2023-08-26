package conf

import (
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"
	"go-common/library/railgun"

	"github.com/BurntSushi/toml"
)

// Config is
type Config struct {
	BM                   *bm.ServerConfig
	Log                  *log.Config
	DB                   *DB
	SeasonSub            *databus.Config
	SeasonWithArchivePub *databus.Config
	Redis                *redis.Config
	// season stat databus
	ViewSnSub  *databus.Config
	DmSnSub    *databus.Config
	ReplySnSub *databus.Config
	FavSnSub   *databus.Config
	CoinSnSub  *databus.Config
	ShareSnSub *databus.Config
	LikeSnSub  *databus.Config
	Custom     *Custom
	//rail_gun config
	CoinSnSubV2Config *RailGunConfig
	ArcClient         *warden.ClientConfig
	SeasonClient      *warden.ClientConfig
	VideoUpOpenClient *warden.ClientConfig
}

type Custom struct {
	Flush bool
}

// DB is db config.
type DB struct {
	Archive *sql.Config
	Result  *sql.Config
	Stat    *sql.Config
}

type RailGunConfig struct {
	Cfg          *railgun.Config
	Databus      *railgun.DatabusV1Config
	SingleConfig *railgun.SingleConfig
}

func (c *Config) Set(s string) error {
	var tmp Config
	if _, err := toml.Decode(s, &tmp); err != nil {
		return err
	}
	old, _ := json.Marshal(c)
	nw, _ := json.Marshal(tmp)
	log.Info("service config changed, old=%+v new=%+v", string(old), string(nw))
	*c = tmp
	return nil
}
