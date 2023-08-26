package conf

import (
	"context"
	"flag"
	"go-common/library/conf/paladin"
	"go-common/library/log"
	"strings"
)

var (
	devConf       string
	devConfClient paladin.Client
)

func init() {
	flag.StringVar(&devConf, "dev_conf", "", "development environment config")
}

// Init conf.
func Init(load func() error, watchKeys ...string) (err error) {
	if err = paladin.Init(); err != nil {
		return
	}
	if devConf != "" {
		if devConfClient, err = paladin.NewFile(devConf); err != nil {
			return
		}
	}
	if err = load(); err != nil {
		return
	}
	if len(watchKeys) == 0 {
		watchKeys = []string{"application.toml"}
	}
	ctx := context.Background()
	go func() {
		for {
			event := paladin.WatchEvent(ctx, watchKeys...)
			_, ok := <-event
			if !ok {
				log.Infoc(ctx, "config event closed!")
				continue
			}
			log.Infoc(ctx, "config reload!")
			load()
		}
	}()
	if devConfClient != nil {
		go func() {
			for {
				event := devConfClient.WatchEvent(ctx, watchKeys...)
				_, ok := <-event
				if !ok {
					log.Infoc(ctx, "config event closed!")
					continue
				}
				log.Infoc(ctx, "config reload!")
				load()
			}
		}()
	}
	return
}

func LoadInto(tmpConf interface{}) (err error) {
	cmap := paladin.GetAll().Load()
	for k, v := range cmap {
		if strings.HasSuffix(k, ".toml") {
			if err := v.UnmarshalTOML(tmpConf); err != nil {
				if err != paladin.ErrNotExist {
					return err
				}
			}
		}
	}
	if devConfClient != nil {
		cmap = devConfClient.GetAll().Load()
		for k, v := range cmap {
			if strings.HasSuffix(k, ".toml") {
				if err := v.UnmarshalTOML(tmpConf); err != nil {
					if err != paladin.ErrNotExist {
						return err
					}
				}
			}
		}
	}
	return nil
}
