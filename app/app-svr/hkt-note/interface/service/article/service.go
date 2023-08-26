package article

import (
	"go-gateway/app/app-svr/hkt-note/interface/conf"
	"go-gateway/app/app-svr/hkt-note/interface/dao/article"
	"go-gateway/app/app-svr/hkt-note/interface/dao/note"
)

type Service struct {
	c       *conf.Config
	artDao  *article.Dao
	noteDao *note.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		artDao:  article.New(c),
		noteDao: note.New(c),
	}
	return
}

func (s *Service) Close() {
	s.artDao.Close()
}
