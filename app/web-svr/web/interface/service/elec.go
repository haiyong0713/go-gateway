package service

import (
	"context"

	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/ecode"
	"go-gateway/app/web-svr/web/interface/model"

	payrank "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
)

// ElecShow elec show.
func (s *Service) ElecShow(c context.Context, upMid, aid, mid int64, arc *arcmdl.Arc) (*model.ElecShow, error) {
	if arc == nil {
		arcReply, err := s.arcGRPC.Arc(c, &arcmdl.ArcRequest{Aid: aid})
		if err != nil {
			log.Error("ElecShow s.arcGRPC.Arc(%d) error(%v)", aid, err)
			return nil, err
		}
		arc = arcReply.Arc
	}
	if arc == nil || arc.Copyright != int32(arcmdl.CopyrightOriginal) {
		return nil, ecode.ElecDenied
	}
	if err := func() error {
		if arc.IsNormalPremiere() {
			// 首映稿件
			return nil
		}
		if !arc.IsNormal() {
			return ecode.ElecDenied
		}
		return nil
	}(); err != nil {
		return nil, err
	}
	var ok bool
	for _, val := range s.c.Rule.ElecShowTypeIDs {
		if arc.TypeID == val {
			ok = true
			break
		}
	}
	if !ok {
		return nil, ecode.ElecDenied
	}
	reply, err := s.payRankGRPC.UPRankWithPanelByUPMid(c, &payrank.RankElecUPReq{UPMID: upMid, Mid: mid})
	if err != nil {
		log.Error("ElecShow s.payRankGRPC.UPRankWithPanelByUPMid upMid:%d error:%v", upMid, err)
		return nil, err
	}
	if reply == nil || reply.RankElecUPProto == nil {
		log.Error("ElecShow s.payRankGRPC.UPRankWithPanelByUPMid upMid:%d reply nil", upMid)
		return nil, ecode.ElecDenied
	}
	var list []*model.ElecUserList
	for _, v := range reply.RankElecUPProto.List {
		if v == nil {
			continue
		}
		tmp := &model.ElecUserList{
			Mid:       v.UpMID,
			PayMid:    v.MID,
			Rank:      v.Rank,
			Uname:     v.Nickname,
			Avatar:    v.Avatar,
			Message:   v.Message,
			TrendType: v.TrendType,
		}
		if v.VIP != nil {
			tmp.VipInfo = model.ElecVipInfo{
				VipType:    v.VIP.Type,
				VipDueMsec: v.VIP.DueDate,
				VipStatus:  v.VIP.Status,
			}
		}
		if v.Hidden {
			tmp.MsgDeleted = 1
		}
		list = append(list, tmp)
	}
	return &model.ElecShow{
		ShowInfo: &model.ShowInfo{
			Show:    reply.Show,
			State:   int8(reply.State),
			Title:   reply.UpowerTitle,
			JumpUrl: reply.UpowerJumpUrl,
			Icon:    reply.UpowerIconUrl,
		},
		TotalCount: reply.RankElecUPProto.CountUPTotalElec,
		List:       list,
	}, nil
}
