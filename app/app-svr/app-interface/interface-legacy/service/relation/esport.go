package relation

import (
	"context"

	"go-common/library/log"
)

func (s *Service) EsportAdd(c context.Context, mid, matchID int64) (err error) {
	if err = s.matchDao.AddFav(c, mid, matchID); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) EsportCancel(c context.Context, mid, matchID int64) (err error) {
	if err = s.matchDao.DelFav(c, mid, matchID); err != nil {
		log.Error("%v", err)
	}
	return
}
