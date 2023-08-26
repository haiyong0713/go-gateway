package service

import (
	"context"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/internal/dao"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

// Service service.
type Service struct {
	dao dao.Dao
	cfg *Config
}

type Config struct {
	ac *paladin.TOML
}

func (c *Config) Set(text string) error {
	if err := c.ac.Set(text); err != nil {
		log.Error("Fail to reload config, config=%s error=%+v", text, err)
		return err
	}
	moduleWl := &ModuleWhitelist{}
	if err := c.ac.Get("ModuleWhitelist").UnmarshalTOML(&moduleWl); err != nil {
		log.Error("Fail to reload ModuleWhitelist, config=%s error=%+v", text, err)
		return err
	}
	pageBl := make(PageModuleBlacklist)
	if err := c.ac.Get("PageModuleBlacklist").UnmarshalTOML(&pageBl); err != nil {
		log.Error("Fail to reload PageModuleBlacklist, config=%s error=%+v", text, err)
		return err
	}
	InitGlobalCardResolver(moduleWl, pageBl)
	return nil
}

type ModuleWhitelist struct {
	AllowAll bool
	Category map[string]bool
}
type PageModuleBlacklist map[string]map[string]bool //native_page.type => category

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		dao: d,
		cfg: &Config{ac: &paladin.TOML{}},
	}
	cf = s.Close
	if err = paladin.Watch("application.toml", s.cfg); err != nil {
		return nil, nil, err
	}
	return s, cf, nil
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}
