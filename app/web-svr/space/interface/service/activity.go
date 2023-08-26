package service

import (
	"context"
	"sort"

	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/space/interface/model"

	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
)

func (s *Service) Reservation(ctx context.Context, mid, vmid int64) ([]*model.UpActReserveRelationInfo, error) {
	arg := &actgrpc.UpActUserSpaceCardReq{
		Upmid: vmid,
		Mid:   mid,
		From:  actgrpc.UpCreateActReserveFrom_FromSpace,
	}
	reply, err := s.actGRPC.UpActUserSpaceCard(ctx, arg)
	if err != nil {
		return nil, err
	}
	if len(reply.List) == 0 {
		return nil, nil
	}
	sort.SliceStable(reply.List, func(i, j int) bool {
		// 赛事预约（组内开始时间升序）>直播预约（组内原定开播时间升序）>稿件预约
		if reply.List[i].Type != reply.List[j].Type {
			return reply.List[i].Type > reply.List[j].Type
		}
		if reply.List[i].StartShowTime != reply.List[j].StartShowTime {
			return reply.List[i].StartShowTime < reply.List[j].StartShowTime
		}
		if reply.List[i].LivePlanStartTime != reply.List[j].LivePlanStartTime {
			return reply.List[i].LivePlanStartTime < reply.List[j].LivePlanStartTime
		}
		return reply.List[i].Sid < reply.List[j].Sid
	})
	return asSpaceReservationCardList(reply.List, mid == vmid), nil
}

func asSpaceReservationCardList(data []*actgrpc.UpActReserveRelationInfo, isSpaceOwner bool) []*model.UpActReserveRelationInfo {
	var res []*model.UpActReserveRelationInfo
	for _, v := range data {
		// 预约先审后发：主态可见，客态不可见
		if !isSpaceOwner && v.UpActVisible == actgrpc.UpActVisible_OnlyUpVisible {
			continue
		}
		if v.State >= actgrpc.UpActReserveRelationState_UpReserveRelatedWaitCallBack && v.State <= actgrpc.UpActReserveRelationState_UpReserveRelatedCallBackDone {
			continue
		}
		if _, ok := v.Hide[int64(actgrpc.UpCreateActReserveFrom_FromSpace)]; ok {
			continue
		}
		info := &model.UpActReserveRelationInfo{}
		info.FromUpActReserveRelationInfo(v, isSpaceOwner)
		info.FromUpActReserveLotteryInfo(v)
		if isSpaceOwner && v.UpActVisible == actgrpc.UpActVisible_OnlyUpVisible {
			info.Name = "[审核中]" + info.Name
		}
		res = append(res, info)
	}
	return res
}

func (s *Service) Reserve(ctx context.Context, req *model.AddReserveReq) error {
	arg := &actgrpc.AddReserveReq{
		Sid:      req.Sid,
		Mid:      req.Mid,
		From:     req.From,
		Typ:      req.Type,
		Oid:      req.Oid,
		Ip:       metadata.String(ctx, metadata.RemoteIP),
		Platform: req.Platform,
		Mobiapp:  req.Mobiapp,
		Buvid:    req.Buvid,
		Spmid:    req.Spmid,
	}
	_, err := s.actGRPC.AddReserve(ctx, arg)
	return err
}

func (s *Service) ReserveCancel(ctx context.Context, mid int64, sid int64) error {
	arg := &actgrpc.DelReserveReq{
		Sid: sid,
		Mid: mid,
	}
	_, err := s.actGRPC.DelReserve(ctx, arg)
	return err
}

func (s *Service) UpReserveCancel(ctx context.Context, mid int64, sid int64) error {
	arg := &actgrpc.CancelUpActReserveReq{
		Mid:  mid,
		Sid:  sid,
		From: actgrpc.UpCreateActReserveFrom_FromSpace,
	}
	_, err := s.actGRPC.CancelUpActReserve(ctx, arg)
	return err
}
