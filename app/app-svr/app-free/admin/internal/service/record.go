package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-free/admin/internal/model"
)

func (s *Service) AllRecords(ctx context.Context, isStateSuccess bool) (rm map[model.ISP][]*model.FreeRecord, err error) {
	res, err := s.rcdDao.AllFreeRecords(ctx)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	rm = make(map[model.ISP][]*model.FreeRecord, len(res))
	for _, r := range res {
		//0000-00-00 00:00:00
		if r.SuccessTime < 0 {
			r.SuccessTime = 0
		}
		if r.CancelTime < 0 {
			r.CancelTime = 0
		}
		if isStateSuccess && r.State != model.StateSucess {
			continue
		}
		rm[r.ISP] = append(rm[r.ISP], r)
	}
	return
}

func (s *Service) Records(ctx context.Context, ips []string) (rm map[model.ISP][]*model.FreeRecord, err error) {
	var res []*model.FreeRecord
	if len(ips) == 0 {
		res, err = s.rcdDao.AllFreeRecords(ctx)
		if err != nil {
			log.Error("%+v", err)
			return
		}
	} else {
		ipInts := make([]int64, 0, len(ips))
		for _, ip := range ips {
			ipInts = append(ipInts, model.InetAtoN(ip))
		}
		res, err = s.rcdDao.FreeRecords(ctx, ipInts)
		if err != nil {
			log.Error("%+v", err)
			return
		}
	}
	rm = make(map[model.ISP][]*model.FreeRecord, len(res))
	for _, r := range res {
		//0000-00-00 00:00:00
		if r.SuccessTime < 0 {
			r.SuccessTime = 0
		}
		if r.CancelTime < 0 {
			r.CancelTime = 0
		}
		r.CtimeHuman = r.Ctime.Time().String()
		rm[r.ISP] = append(rm[r.ISP], r)
	}
	return
}

func (s *Service) InsertFreeRecords(ctx context.Context, rs []*model.FreeRecord) error {
	for _, r := range rs {
		r.IPStartInt = model.InetAtoN(r.IPStart)
		r.IPEndInt = model.InetAtoN(r.IPEnd)
	}
	return s.rcdDao.InsertFreeRecords(ctx, rs)
}
