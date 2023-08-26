package card

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/article"
	"go-gateway/app/app-svr/app-feed/admin/dao/card"
)

// Service is resource card service
type Service struct {
	dao        *card.Dao
	articleDao *article.Dao
	c          *conf.Config
}

// New new a resource card service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao:        card.New(c),
		articleDao: article.New(c),
		c:          c,
	}
	return
}
