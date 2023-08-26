package like

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"time"
)

func (s *Service) TicketSign(c context.Context, mid int64, ticket string) (interface{}, error) {
	find := false
	for _, id := range s.c.Ticket.Mid {
		if id == mid {
			find = true
			break
		}
	}
	if !find {
		log.Errorc(c, "white mid list %v", s.c.Ticket.Mid)
		return nil, ecode.Error(ecode.RequestErr, "账号无操作权限，请更换账号")
	}
	ti, err := s.dao.GetTicketByCode(c, ticket)
	if err != nil {
		return nil, err
	}
	if ti == nil || ti.ID == 0 {
		return nil, ecode.Error(ecode.RequestErr, "电子门票码不存在")
	}
	if ti.State != 0 {
		return ti, ecode.Error(ecode.RequestErr, "重复签到，该电子门票码已完成签到")
	}
	_, err = s.dao.UpdateTicketState(c, ti.ID, 1)
	ti.State = 1
	ti.Mtime = xtime.Time(time.Now().Unix())
	return ti, err
}
