package conf

import (
	"github.com/BurntSushi/toml"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

// Config struct
type Config struct {
	Log     *log.Config
	BM      *bm.ServerConfig
	ExtraDB *sql.Config
	Redis   *redis.Config
	//grpc
	HttpClient    *bm.ClientConfig
	Custom        *Custom
	DemotionExtra *DemotionExtra
}

// Host
type Host struct {
	Dynamic string
}

// Custom 服务调用限制
type Custom struct {
	ExtraCallers map[string]string
	// 开关，测试用
	ExtraCallersSwitch bool
}

type DemotionExtra struct {
	AidKeys map[string][]string
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
