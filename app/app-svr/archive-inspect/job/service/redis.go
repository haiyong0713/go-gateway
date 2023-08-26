package service

import (
	"context"

	"go-common/library/cache/redis"
	arcmdl "go-gateway/app/app-svr/archive/service/model"

	"github.com/pkg/errors"
)

func (s *Service) getCache(c context.Context, aid int64) (arc [][]byte, view [][]byte, err error) {
	akey := arcmdl.ArcKey(aid)
	pkey := arcmdl.PageKey(aid)
	for k, rds := range s.arcRedises {
		if err := func() error {
			conn := rds.Get(c)
			defer conn.Close()
			abs, err := redis.Bytes(conn.Do("GET", akey))
			if err != nil {
				return err
			}
			arc = append(arc, abs)
			vbs, err := redis.Bytes(conn.Do("GET", pkey))
			if err != nil {
				return err
			}
			view = append(view, vbs)
			return nil
		}(); err != nil {
			return nil, nil, errors.Wrapf(err, "getCache k(%d) ", k)
		}
	}
	return arc, view, nil
}
