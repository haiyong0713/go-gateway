package service

import (
	"go-common/library/log"
)

type Loader interface {
	Timer() string
	Load() error
}

func loggingError(in func() error) func() {
	return func() {
		if err := in(); err != nil {
			log.Error("%+v", err)
		}
	}
}

func (s *Service) StartLoad(loaders ...Loader) {
	for _, v := range loaders {
		if err := v.Load(); err != nil {
			panic(err)
		}
		if err := s.cron.AddFunc(v.Timer(), loggingError(v.Load)); err != nil {
			panic(err)
		}
	}
	s.cron.Start()
}
