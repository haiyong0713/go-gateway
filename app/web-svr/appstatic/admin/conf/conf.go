package conf

import (
	"go-common/library/cache/redis"
	"go-common/library/database/boss"
	"go-common/library/database/orm"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/trace"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
)

// Config def.
type Config struct {
	// http
	BM *bm.ServerConfig
	// db
	ORM *orm.Config
	// gw db
	GWDB *orm.Config
	//peak db
	PeakDB *orm.Config
	// log
	XLog *log.Config
	// tracer
	Tracer *trace.Config
	// cfg
	Cfg *Cfg
	// bfs
	Bfs *Bfs
	// peak bfs
	PeakBfs *Bfs
	// HTTPClient .
	HTTPClient *bm.ClientConfig
	// Redis
	Redis *Redis
	// app-player redis ylf & shjd
	PlayerRedis []*Redis
	// boss cfg
	Boss *boss.Config
	// host
	Host *Host
}

// Host host
type Host struct {
	Boss string
	Cdn  string
}

// PushCfg def.
type PushCfg struct {
	Operation int    // operation number
	QPS       int    // qps limit
	URL       string // push url
	Expire    xtime.Duration
}

// Redis redis
type Redis struct {
	*redis.Config
}

// Bfs reprensents the bfs config
type Bfs struct {
	Key     string
	Secret  string
	Host    string
	Addr    string
	Bucket  string
	Timeout int
	OldURL  string
	NewURL  string
}

// Cfg def.
type Cfg struct {
	HistoryVer     int
	Filetypes      []string // allowed file type to upload
	Push           *PushCfg
	BigfileTimeout xtime.Duration
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
