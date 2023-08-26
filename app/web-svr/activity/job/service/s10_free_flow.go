package service

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/s10"
	"time"
)

func (s *Service) freeFlowProc() {
	if s.freeFlowSub == nil {
		return
	}
	for {
		msg, ok := <-s.freeFlowSub.Messages()
		if !ok {
			break
		}
		msg.Commit()
		m := new(s10.FreeFlow)
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			log.Error("freeFlowProc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		s.freeFlow(m)
	}
}

func (s *Service) freeFlow(m *s10.FreeFlow) {
	ctx := context.Background()
	switch m.Type {
	case 0:
		_, err := s.retryFreeFlow(func() (int64, error) {
			return s.s10Dao.AddRawFreeFlow(ctx, m.Message)
		})
		if err != nil {
			log.Error("s10retry AddRawFreeFlow(tel:%v) error:%v", m.Message, err)
		}
		return
	case 1:
		mid, err := s.retryFreeFlow(func() (int64, error) {
			return s.s10Dao.MidByTel(ctx, m.Message)
		})
		if mid == 0 || err != nil {
			if err != nil {
				log.Error("s10retry MidByTel(tel:%v) error:%v", m.Message, err)
			}
			return
		}
		id, err := s.retryFreeFlow(func() (int64, error) {
			return s.s10Dao.RawFreeFlow(ctx, m.Message)
		})
		if err != nil || id == 0 {
			if err != nil {
				log.Error("s10retry RawFreeFlow(tel:%v) error:%v", m.Message, err)
			}
			return
		}
		_, err = s.retryFreeFlow(func() (int64, error) {
			return s.s10Dao.AddFreeFlowUser(ctx, mid, m.Source)
		})
		if err != nil {
			log.Error("s10retry AddFreeFlowUser(tel:%v) error:%v", mid, err)
			return
		}
		//db:source: 0-联通；1-移动；缓存值：1-哨兵；2-联通；3-移动
		s.s10Dao.AddUserFlowCache(ctx, mid, int64(2+m.Source))
	}
}

func (s *Service) retryFreeFlow(f func() (int64, error)) (id int64, err error) {
	for i := 0; i < 3; i++ {
		if id, err = f(); err == nil {
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
	return
}
