package conf

import (
	"go-common/library/database/orm"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"github.com/BurntSushi/toml"
)

// Config struct
type Config struct {
	// log
	Log *log.Config
	// HttpService
	HttpService *HTTPServers
	// interface XLog
	XLog *log.Config
	// db
	MySQL *MySQL
	// cron
	Cron *Cron
	// redis
	Redis *Redis
}

// HTTPServers Http Servers.
type HTTPServers struct {
	Inner *bm.ServerConfig
	Outer *bm.ServerConfig
}

// MySQL struct
type MySQL struct {
	Show      *orm.Config
	Resource  *sql.Config
	Manager   *sql.Config
	Teenagers *sql.Config
}

type Cron struct {
	LoadTabExt       string
	LoadCustomConfig string
	LoadSkinExt      string
	LoadBWList       string
}

type Redis struct {
	Common *struct {
		*redis.Config
	}
	Show *struct {
		*redis.Config
	}
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
