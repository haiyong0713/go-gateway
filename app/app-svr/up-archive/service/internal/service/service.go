package service

import (
	"context"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/up-archive/service/api"
	"go-gateway/app/app-svr/up-archive/service/internal/dao"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"

	"github.com/BurntSushi/toml"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
)

type defaultHttp interface {
}

var Provider = wire.NewSet(New, wire.Bind(new(defaultHttp), new(*Service)), wire.Bind(new(api.UpArchiveServer), new(*Service)))

type Config struct {
	ArchiveClient *warden.ClientConfig
	Cron          struct {
		TypesSpec string
	}
	Search struct {
		OneWordLens []int
		Degrade     bool
	}
	SearchGray struct {
		Bucket int64
	}
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	*c = tmp
	return nil
}

// Service service.
type Service struct {
	ac          *Config
	dao         dao.Dao
	archiveGRPC arcapi.ArchiveClient
	cache       *fanout.Fanout
	cron        *cron.Cron
	types       map[int32]*arcapi.Tp
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:    &Config{},
		dao:   d,
		cache: fanout.New("cache"),
		cron:  cron.New(),
	}
	cf = s.Close
	if err = paladin.Watch("application.toml", s.ac); err != nil {
		err = errors.WithStack(err)
		return
	}
	if s.archiveGRPC, err = arcapi.NewClient(s.ac.ArchiveClient); err != nil {
		err = errors.WithStack(err)
		return
	}
	if err = s.initCron(); err != nil {
		return
	}
	return
}

func (s *Service) initCron() error {
	if err := s.loadTypes(); err != nil {
		return err
	}
	if err := s.cron.AddFunc(s.ac.Cron.TypesSpec, func() { _ = s.loadTypes() }); err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

func (s *Service) loadTypes() error {
	types, err := s.Types(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	s.types = types
	return nil
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.cron.Stop()
	s.dao.Close()
}
