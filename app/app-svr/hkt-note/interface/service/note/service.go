package note

import (
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/hkt-note/interface/conf"
	"go-gateway/app/app-svr/hkt-note/interface/dao/article"
	"go-gateway/app/app-svr/hkt-note/interface/dao/note"
)

type Service struct {
	c            *conf.Config
	dao          *note.Dao
	artDao       *article.Dao
	notePub      *databus.Databus
	noteAuditPub *databus.Databus
	infocV2Log   infocV2.Infoc
}

func New(c *conf.Config, infoc infocV2.Infoc) (s *Service) {
	s = &Service{
		c:            c,
		dao:          note.New(c),
		artDao:       article.New(c),
		notePub:      databus.New(c.NotePub),
		noteAuditPub: databus.New(c.NoteAuditPub),
		infocV2Log:   infoc,
	}
	return
}

func (s *Service) Close() {
	s.dao.Close()
}
