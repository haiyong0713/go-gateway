package conf

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"

	"github.com/BurntSushi/toml"
)

// Config struct
type Config struct {
	Log     *log.Config
	BM      *bm.ServerConfig
	HonorDB *sql.Config
	Redis   *redis.Config
	//databus
	ArchiveHonorSub *databus.Config
	StatRankSub     *databus.Config
	//host
	Host *Host
	//grpc
	ArcClient  *warden.ClientConfig
	HttpClient *bm.ClientConfig
	//custom
	Custom *Custom
}

type Custom struct {
	//关闭热门推送
	CloseHot bool
}

// Host
type Host struct {
	Dynamic string
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
