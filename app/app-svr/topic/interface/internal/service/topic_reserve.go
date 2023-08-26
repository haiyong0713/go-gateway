package service

import (
	"context"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
	"go-common/library/log"

	api "go-gateway/app/app-svr/topic/interface/api"
)

func (s *Service) TopicReserveButtonClick(ctx context.Context, req *api.TopicReserveButtonClickReq) (resp *api.TopicReserveButtonClickReply, err error) {
	args := &topicsvc.ReserveButtonClickReq{
		Uid:          req.Uid,
		ReserveId:    req.ReserveId,
		ReserveTotal: req.ReserveTotal,
		CurBtnStatus: topicsvc.ReserveButtonStatus(req.CurBtnStatus),
	}
	res, err := s.topicGRPC.ReserveButtonClick(ctx, args)
	if err != nil {
		log.Error("s.TopicReserveButtonClick req:%+v, err:%+v", req, err)
		return nil, err
	}
	resp = &api.TopicReserveButtonClickReply{
		FinalBtnStatus: api.ReserveButtonStatus(res.FinalBtnStatus),
		BtnMode:        api.ReserveButtonMode(res.BtnMode),
		ReserveUpdate:  res.ReserveUpdate,
		DescUpdate:     res.DescUpdate,
		HasActivity:    res.HasActivity,
		ActivityUrl:    res.ActivityUrl,
		Toast:          res.Toast,
		ReserveCalendarInfo: &api.ReserveCalendarInfo{
			Title:       res.GetReserveCalendarInfo().GetTitle(),
			StartTs:     res.GetReserveCalendarInfo().GetStartTs(),
			EndTs:       res.GetReserveCalendarInfo().GetEndTs(),
			Description: res.GetReserveCalendarInfo().GetDescription(),
			BusinessId:  res.GetReserveCalendarInfo().GetBusinessId(),
		},
	}
	if resp.ReserveUpdate < 50 {
		resp.ReserveUpdate = 0
		resp.DescUpdate = ""
	}
	return resp, nil
}
