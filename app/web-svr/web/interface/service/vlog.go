package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/xstr"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	tagSvrgrpc "git.bilibili.co/bapis/bapis-go/community/service/tag"
)

const (
	_channelView = 9
	_vlogRankMax = 100
)

func (s *Service) abnormalAID(ctx context.Context, aids []int64) (arcs []*model.BvArc, err error) {
	var (
		aidRes    []int64
		arcsReply *arcgrpc.ArcsReply
	)
	// 过滤下返回的aid
	for _, aid := range aids {
		if aid > 0 {
			aidRes = append(aidRes, aid)
		}
	}
	if len(aidRes) == 0 {
		return
	}
	if arcsReply, err = s.arcGRPC.Arcs(ctx, &arcgrpc.ArcsRequest{Aids: aidRes}); err != nil {
		log.Error("[abnormalAID] s.arcGRPC.Arcs() aids(%s) error(%v)", xstr.JoinInts(aids), err)
		return
	}
	for _, aid := range aids {
		if arc, ok := arcsReply.Arcs[aid]; ok && arc.IsNormal() {
			arcs = append(arcs, model.CopyFromArcToBvArc(arc, s.avToBv(arc.Aid)))
		}
	}
	return
}

// Vlog .
func (s *Service) Vlog(ctx context.Context, param *model.VlogParam) (arcs []*model.BvArc, err error) {
	var (
		channelReply *taggrpc.ChannelResourcesReply
		channelReq   = &taggrpc.ChannelResourcesReq{
			Tid:           param.TID,
			Mid:           param.MID,
			Plat:          param.Plat,
			Channel:       param.ChnID,
			LoginEvent:    param.LoginEnvent,
			RequestCnt:    param.Ps,
			DisplayId:     param.Pn,
			From:          _channelView,
			Type:          0,
			Buvid:         param.Buvid,
			Build:         param.Build,
			DisplayMethod: param.Rank,
		}
	)
	arcs = make([]*model.BvArc, 0)
	if channelReply, err = s.tagGRPC.ChannelPartitionResources(ctx, channelReq); err != nil {
		log.Error("[Vlog] s.tagGRPC.ChannelPartitionResources() chanelID(%d) tid(%d) error(%v)", param.ChnID, param.TID, err)
		return
	}
	if len(channelReply.Oids) == 0 {
		log.Info("[Vlog] tagGRPC return aids length(%d)", len(channelReply.Oids))
		return
	}
	if arcs, err = s.abnormalAID(ctx, channelReply.Oids); err != nil {
		log.Error("[Vlog] abnormalAID error(%v)", err)
		return
	}
	if cnt := len(arcs); cnt%2 != 0 {
		arcs = arcs[:cnt-1]
	}
	return
}

// VlogRank vlog 排行榜
func (s *Service) VlogRank(ctx context.Context, param *model.VlogRankParam) (arcs []*model.BvArc) {
	var (
		err         error
		tagSvrReply *tagSvrgrpc.PartitionRankReply
	)
	if param.Pn*param.Ps > _vlogRankMax {
		return
	}
	if tagSvrReply, err = s.tagSvrGRPC.PartitionRank(ctx, &tagSvrgrpc.PartitionRankReq{Tid: param.TID, Pn: param.Pn, Ps: param.Ps}); err != nil {
		log.Error("[VlogRank] s.tagSvrGRPC.PartitionRank() tid(%d) PN(%d) PS(%d) error(%v)", param.TID, param.Pn, param.Ps, err)
		return
	}
	if tagSvrReply == nil || len(tagSvrReply.Oids) == 0 {
		log.Warn("[VlogRank] tagSvrGRPC return aids length(%d) tid(%d)", len(tagSvrReply.Oids), param.TID)
		return
	}
	if arcs, err = s.abnormalAID(ctx, tagSvrReply.Oids); err != nil {
		log.Error("[VlogRank] abnormalAID() tid(%d) error(%v)", param.TID, err)
	}
	return
}
