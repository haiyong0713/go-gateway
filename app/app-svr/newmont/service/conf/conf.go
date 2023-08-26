package conf

import (
	"go-common/library/log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	//会员购底tab id
	MallDefaultIDMap map[string]int64
	MallCustomIDMap  map[string]int64
	IconCacheConfig  IconCacheConfig
}

type IconCacheConfig struct {
	PreloadDuration int
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
