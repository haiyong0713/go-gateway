package external

import (
	"go-gateway/app/app-svr/app-feed/interface/conf"
	"go-gateway/app/app-svr/app-feed/interface/dao/dynamic"
)

// Service .
type Service struct {
	dynamic *dynamic.Dao
}

// New .
func New(c *conf.Config) *Service {
	return &Service{
		dynamic: dynamic.New(c),
	}
}
