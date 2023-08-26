package bwsonline

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
)

func (s *Service) BatchCacheBindRecords(ctx context.Context, startIndex int64, limit int32) (records []*bwsonline.TicketBindRecord, err error) {
	if records, err = s.dao.RawTicketsListByIds(ctx, startIndex, limit, s.c.BwsOnline.BwPark.Year); err != nil {
		return
	}
	err = s.dao.BatchCacheBindRecords(ctx, records)
	return
}

func (s *Service) CheckBind(ctx context.Context, mid int64) (id int64, err error) {
	return s.dao.CheckBindRecord(ctx, mid)
}
