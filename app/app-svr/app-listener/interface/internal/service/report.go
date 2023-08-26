package service

import (
	"context"

	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
)

func (s *Service) Event(ctx context.Context, req *v1.EventReq) (reply *v1.EventResp, err error) {
	if err = validatePlayItem(ctx, req.Item, 0); err != nil {
		return
	}
	_, _, auth := DevNetAuthFromCtx(ctx)

	reply = &v1.EventResp{}
	var success bool
	//nolint:exhaustive
	switch req.EventType {
	case v1.EventReq_GUIDE_BAR_SHOW:
		success, err = s.dao.GuideBarShowReport(ctx, dao.GuideBarShowReportOpt{
			Mid:  auth.Mid,
			Type: int64(req.GetItem().GetItemType()),
			Oid:  req.GetItem().GetOid(),
		})
	}

	if !success {
		log.Error("EventReport failed req(%+v) mid(%d)", req, auth.Mid)
	}

	return
}
