package conf

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
)

// Config is
type Config struct {
	BM             *bm.ServerConfig
	Log            *log.Config
	DB             *DB
	Cron           *Cron
	Custom         *Custom
	ArcRedises     []*redis.Config
	Taishan        *Taishan
	TaiShanClient  *warden.ClientConfig
	CreativeClient *warden.ClientConfig
	LocationClient *warden.ClientConfig
}

type Taishan struct {
	Table string
	Token string
}

type Custom struct {
	Internal xtime.Duration
}

type Cron struct {
	// 每10分钟检查稿件数据
	CheckModifyAids string
}

// DB is db config.
type DB struct {
	Result *sql.Config
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	log.Info("progress-service-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}
