package dao

import (
	"context"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
)

type activityDao struct {
	client activitygrpc.ActivityClient
}

func (d *activityDao) ActSubsProtocol(c context.Context, sids []int64) (map[int64]*activitygrpc.ActSubProtocolReply, error) {
	rly, err := d.client.ActSubsProtocol(c, &activitygrpc.ActSubsProtocolReq{Sids: sids})
	if err != nil {
		return nil, err
	}
	return rly.List, nil
}

func (d *activityDao) ActLikes(c context.Context, req *activitygrpc.ActLikesReq) (*activitygrpc.LikesReply, error) {
	rly, err := d.client.ActLikes(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) RankResult(c context.Context, req *activitygrpc.RankResultReq) (*activitygrpc.RankResultResp, error) {
	rly, err := d.client.RankResult(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) UpList(c context.Context, req *activitygrpc.UpListReq) (*activitygrpc.UpListReply, error) {
	rly, err := d.client.UpList(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) GetVoteActivityRank(c context.Context, req *activitygrpc.GetVoteActivityRankReq) (*activitygrpc.GetVoteActivityRankResp, error) {
	rly, err := d.client.GetVoteActivityRank(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) VoteUserDo(c context.Context, req *activitygrpc.VoteUserDoReq) (*activitygrpc.VoteUserDoResp, error) {
	rly, err := d.client.VoteUserDo(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) VoteUserUndo(c context.Context, req *activitygrpc.VoteUserUndoReq) (*activitygrpc.VoteUserUndoResp, error) {
	rly, err := d.client.VoteUserUndo(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) UpActReserveRelationInfo(c context.Context, req *activitygrpc.UpActReserveRelationInfoReq) (map[int64]*activitygrpc.UpActReserveRelationInfo, error) {
	rly, err := d.client.UpActReserveRelationInfo(c, req)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return map[int64]*activitygrpc.UpActReserveRelationInfo{}, nil
	}
	return rly.List, nil
}

func (d *activityDao) AddReserve(c context.Context, req *activitygrpc.AddReserveReq) error {
	if _, err := d.client.AddReserve(c, req); err != nil {
		return err
	}
	return nil
}

func (d *activityDao) DelReserve(c context.Context, req *activitygrpc.DelReserveReq) error {
	if _, err := d.client.DelReserve(c, req); err != nil {
		return err
	}
	return nil
}

func (d *activityDao) ActSubjects(c context.Context, req *activitygrpc.ActSubjectsReq) (*activitygrpc.ActSubjectsReply, error) {
	rly, err := d.client.ActSubjects(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) ActivityProgress(c context.Context, req *activitygrpc.ActivityProgressReq) (*activitygrpc.ActivityProgressReply, error) {
	rly, err := d.client.ActivityProgress(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) ReserveFollowings(c context.Context, req *activitygrpc.ReserveFollowingsReq) (*activitygrpc.ReserveFollowingsReply, error) {
	rly, err := d.client.ReserveFollowings(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) AwardSubjectState(c context.Context, req *activitygrpc.AwardSubjectStateReq) (*activitygrpc.AwardSubjectStateReply, error) {
	rly, err := d.client.AwardSubjectState(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) ActRelationInfo(c context.Context, req *activitygrpc.ActRelationInfoReq) (*activitygrpc.ActRelationInfoReply, error) {
	rly, err := d.client.ActRelationInfo(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) LotteryUnusedTimes(c context.Context, req *activitygrpc.LotteryUnusedTimesdReq) (*activitygrpc.LotteryUnusedTimesReply, error) {
	rly, err := d.client.LotteryUnusedTimes(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *activityDao) RewardSubject(c context.Context, req *activitygrpc.RewardSubjectReq) error {
	_, err := d.client.RewardSubject(c, req)
	return err
}

func (d *activityDao) GRPCDoRelation(c context.Context, req *activitygrpc.GRPCDoRelationReq) error {
	_, err := d.client.GRPCDoRelation(c, req)
	return err
}

func (d *activityDao) RelationReserveCancel(c context.Context, req *activitygrpc.RelationReserveCancelReq) error {
	_, err := d.client.RelationReserveCancel(c, req)
	return err
}
