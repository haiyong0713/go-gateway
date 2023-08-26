package conf

import (
	"go-common/library/cache/redis"
	"go-common/library/database/orm"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"

	"github.com/BurntSushi/toml"
)

// Config def.
type Config struct {
	HTTPServer         *bm.ServerConfig
	HTTPClient         *bm.ClientConfig
	HTTPGameClient     *bm.ClientConfig
	HTTPActAdminClient *bm.ClientConfig
	ORM                *orm.Config
	Redis              *Redis
	Log                *log.Config
	Tracer             *trace.Config
	Host               *Host
	// grpc
	AccClient     *warden.ClientConfig
	TagGRPC       *warden.ClientConfig
	PlatGRPC      *warden.ClientConfig
	SpaceClient   *warden.ClientConfig
	ActClient     *warden.ClientConfig
	ChaClient     *warden.ClientConfig
	DynvoteClient *warden.ClientConfig
	// config
	Up        *Up
	InfocConf *InfocConf
}

type InfocConf struct {
	CloudInfoc *infoc.Config
	CloudLogID string
}

type Up struct {
	SenderUid             uint64
	PassContent           string
	UnPassContent         string
	ActSenderUid          uint64
	NotifyCodePass        string
	NotifyCodeNotPass     string
	NotifyCodeSpaceOff    string
	NotifyCodePassEdit    string
	NotifyCodeNotPassEdit string
}

// Redis struct
type Redis struct {
	*redis.Config
}

// Host remote host
type Host struct {
	API      string
	SHOW     string
	MNG      string
	Dynamic  string
	ActTmpl  string
	GameCo   string
	ManGaCo  string
	ActAdmin string
}

// MySQL define MySQL config
type MySQL struct {
	Lottery *sql.Config
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
