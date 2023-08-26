package service

import (
	"context"

	"go-common/library/log"
)

func (s *Service) setWebTop() {
	ctx := context.Background()
	var aids []int64
	if err := retry(func() (err error) {
		aids, err = s.dao.WebTop(ctx)
		return err
	}); err != nil {
		log.Error("日志告警 WebTop error:%+v", err)
		return
	}
	if err := retry(func() (err error) {
		return s.dao.AddCacheWebTop(ctx, aids)
	}); err != nil {
		log.Error("日志告警 AddCacheWebTop error:%+v", err)
		return
	}
}
