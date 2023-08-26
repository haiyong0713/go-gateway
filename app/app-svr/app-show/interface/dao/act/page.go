package act

import (
	"context"
	"sync"
	"time"

	"git.bilibili.co/bapis/bapis-go/activity/service"
	media "git.bilibili.co/bapis/bapis-go/pgc/service/media"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
)

// ActLikes .
func (d *Dao) ActLikes(c context.Context, arg *api.ActLikesReq) (reply *api.LikesReply, err error) {
	return d.actRPC.ActLikes(c, arg)
}

// ActLiked .
func (d *Dao) ActLiked(c context.Context, arg *api.ActLikedReq) (*api.ActLikedReply, error) {
	return d.actRPC.ActLiked(c, arg)
}

// ActRelationInfo .
func (d *Dao) ActRelationInfo(c context.Context, sid, mid int64) (*api.ActRelationInfoReply, error) {
	return d.actRPC.ActRelationInfo(c, &api.ActRelationInfoReq{Mid: mid, Id: sid, Specific: "reserve"})
}

func (d *Dao) RelationReserveCancel(c context.Context, sid, mid int64, from, spmid, buvid, platform, mobiapp string) error {
	ip := metadata.String(c, metadata.RemoteIP)
	req := &api.RelationReserveCancelReq{
		Id:       sid,
		Mid:      mid,
		From:     from,
		Spmid:    spmid,
		Buvid:    buvid,
		Platform: platform,
		Mobiapp:  mobiapp,
		Ip:       ip,
	}
	_, err := d.actRPC.RelationReserveCancel(c, req)
	if err != nil {
		log.Error("d.actRPC.GRPCDoRelation(%v) error(%v)", req, err)
	}
	return err
}

func (d *Dao) GRPCDoRelation(c context.Context, sid, mid int64, from, spmid, buvid, platform, mobiapp string) error {
	ip := metadata.String(c, metadata.RemoteIP)
	req := &api.GRPCDoRelationReq{
		Id:       sid,
		Mid:      mid,
		From:     from,
		Spmid:    spmid,
		Buvid:    buvid,
		Platform: platform,
		Mobiapp:  mobiapp,
		Ip:       ip,
	}
	_, err := d.actRPC.GRPCDoRelation(c, req)
	if err != nil {
		log.Error("d.actRPC.GRPCDoRelation(%v) error(%v)", req, err)
	}
	return err
}

// ReserveFollowings .
func (d *Dao) ReserveFollowings(c context.Context, mid int64, sids []int64) (res map[int64]*api.ReserveFollowingReply, err error) {
	var (
		rly *api.ReserveFollowingsReply
	)
	if rly, err = d.actRPC.ReserveFollowings(c, &api.ReserveFollowingsReq{Sids: sids, Mid: mid}); err != nil {
		log.Error(" d.actRPC.ReserveFollowings(%v,%d) error(%v)", sids, mid, err)
		return
	}
	if rly != nil {
		res = rly.List
	}
	return
}

// AddReserve .
func (d *Dao) AddReserve(c context.Context, sid, mid int64) (err error) {
	if _, err = d.actRPC.AddReserve(c, &api.AddReserveReq{Sid: sid, Mid: mid}); err != nil {
		log.Error(" d.actRPC.AddReserve(%v,%d) error(%v)", sid, mid, err)
	}
	return
}

// DelReserve .
func (d *Dao) DelReserve(c context.Context, sid, mid int64) (err error) {
	if _, err = d.actRPC.DelReserve(c, &api.DelReserveReq{Sid: sid, Mid: mid}); err != nil {
		log.Error(" d.actRPC.DelReserve(%v,%d) error(%v)", sid, mid, err)
	}
	return
}

// ActSubProtocol .
func (d *Dao) ActSubsProtocol(c context.Context, sids []int64) (map[int64]*api.ActSubProtocolReply, error) {
	rly, err := d.actRPC.ActSubsProtocol(c, &api.ActSubsProtocolReq{Sids: sids})
	if err != nil {
		return nil, err
	}
	protos := make(map[int64]*api.ActSubProtocolReply)
	if rly != nil {
		protos = rly.List
	}
	return protos, nil
}

// RewardSubject 领奖组件领奖接口
func (d *Dao) RewardSubject(c context.Context, id, mid int64) (err error) {
	_, err = d.actRPC.RewardSubject(c, &api.RewardSubjectReq{Id: id, Mid: mid})
	return
}

// AwardSubjectState 领奖组件获取奖励状态接口
func (d *Dao) AwardSubjectState(c context.Context, id, mid int64) (state int, err error) {
	reply, err := d.actRPC.AwardSubjectState(c, &api.AwardSubjectStateReq{Id: id, Mid: mid})
	if err != nil {
		return
	}
	if reply == nil {
		return
	}
	state = int(reply.State)
	return
}

// AwardSubjectStates 批量领奖组件获取奖励状态接口
func (d *Dao) AwardSubjectStates(c context.Context, ids []int64, mid int64) (states map[int64]int, err error) {
	var (
		g     = errgroup.WithContext(c)
		mutex = sync.Mutex{}
	)
	states = map[int64]int{}
	for _, id := range ids {
		tmpid := id
		g.Go(func(c context.Context) (err error) {
			reply, err := d.actRPC.AwardSubjectState(c, &api.AwardSubjectStateReq{Id: tmpid, Mid: mid})
			if err != nil {
				log.Error("%+v", err)
				return
			}
			if reply != nil {
				mutex.Lock()
				states[tmpid] = int(reply.State)
				mutex.Unlock()
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

func (d *Dao) GetCharacterEps(c context.Context, charID, seasonID int32) ([]int64, error) {
	req := &media.CharacterIdsOidsReq{
		CharacterIdOpusIds: map[int32]*media.OpusIdsReq{charID: {Ids: []int32{seasonID}}},
		Otype:              100,
	}
	reply, err := d.charGRPC.RelInfos(c, req)
	if err != nil {
		log.Error("Fail to get RelInfos, req=%+v error=%+v", req, err)
		return nil, err
	}
	relInfo, ok := reply.GetInfos()[charID]
	if !ok || relInfo.GetCharacterEp() == nil {
		return []int64{}, nil
	}
	epList, ok := relInfo.GetCharacterEp()[seasonID]
	if !ok || epList.GetCharacterEp() == nil {
		return []int64{}, nil
	}
	epIDs := make([]int64, 0, len(epList.GetCharacterEp()))
	for _, ep := range epList.GetCharacterEp() {
		if ep == nil {
			continue
		}
		epIDs = append(epIDs, int64(ep.GetEpId()))
	}
	return epIDs, nil
}

func (d *Dao) ReserveProgress(c context.Context, sid, mid, ruleID, typ, dataType int64, dimension api.GetReserveProgressDimension) (int64, error) {
	if sid == 0 {
		return 0, nil
	}
	req := &api.GetReserveProgressReq{
		Sid: sid,
		Mid: mid,
		Rules: []*api.ReserveProgressRule{
			{Dimension: dimension, RuleId: ruleID, Type: typ, DataType: dataType},
		},
	}
	rly, err := d.actRPC.GetReserveProgress(c, req)
	if err != nil {
		log.Error("Fail to request actRPC.GetReserveProgress, req=%+v error=%+v", req, err)
		return 0, err
	}
	for _, v := range rly.Data {
		if v == nil || v.Rule == nil {
			continue
		}
		if v.Rule.Dimension == dimension && v.Rule.RuleId == ruleID && v.Rule.Type == typ && v.Rule.DataType == dataType {
			return v.Progress, nil
		}
	}
	return 0, nil
}

func (d *Dao) AppJumpUrl(c context.Context, bizType api.AppJumpBizType, memory int64, ua string) (string, error) {
	req := &api.AppJumpReq{
		BizType:   bizType,
		Memory:    memory,
		UserAgent: ua,
	}
	rly, err := d.actRPC.AppJumpUrl(c, req)
	if err != nil {
		log.Error("Fail to get appJumpUrl, req=%+v error=%+v", req, err)
		return "", err
	}
	return rly.JumpUrl, nil
}

func (d *Dao) LotteryUnusedTimes(c context.Context, mid int64, lotteryID string) (*api.LotteryUnusedTimesReply, error) {
	return d.actRPC.LotteryUnusedTimes(c, &api.LotteryUnusedTimesdReq{Sid: lotteryID, Mid: mid})
}

func (d *Dao) UpList(c context.Context, sid, pn, ps, mid int64, typ string) (*api.UpListReply, error) {
	req := &api.UpListReq{Sid: sid, Type: typ, Pn: pn, Ps: ps, Mid: mid}
	rly, err := d.actRPC.UpList(c, req)
	if err != nil {
		log.Error("Fail to get upList, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

func (d *Dao) ActivityProgress(c context.Context, sid, typ, mid int64, gids []int64) (*api.ActivityProgressReply, error) {
	req := &api.ActivityProgressReq{Sid: sid, Gids: gids, Type: typ, Mid: mid, Time: time.Now().Unix()}
	rly, err := d.actRPC.ActivityProgress(c, req)
	if err != nil {
		log.Error("Fail to request ActivityProgress, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

// UpActReserveRelationInfo.
func (d *Dao) UpActReserveRelationInfo(c context.Context, mid int64, sids []int64) (map[int64]*api.UpActReserveRelationInfo, error) {
	req := &api.UpActReserveRelationInfoReq{Sids: sids, Mid: mid}
	rly, err := d.actRPC.UpActReserveRelationInfo(c, req)
	if err != nil {
		log.Error("Fail to request UpActReserveRelationInfo, req=%+v error=%+v", req, err)
		return nil, err
	}
	if rly == nil {
		return make(map[int64]*api.UpActReserveRelationInfo), nil
	}
	return rly.List, nil
}

func (d *Dao) GetVoteActivityRank(c context.Context, actID, groupID, pn, ps, sort, mid int64) (*api.GetVoteActivityRankResp, error) {
	req := &api.GetVoteActivityRankReq{ActivityId: actID, SourceGroupId: groupID, Pn: pn, Ps: ps, Sort: sort, Mid: mid}
	rly, err := d.actRPC.GetVoteActivityRank(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

// VoteUserDo
func (d *Dao) VoteUserDo(c context.Context, actID, groupID, itemID, mid, voteCount int64, risk *api.Risk) (int64, int64, error) {
	req := &api.VoteUserDoReq{ActivityId: actID, SourceGroupId: groupID, SourceItemId: itemID, Mid: mid, VoteCount: voteCount, Risk: risk}
	rly, err := d.actRPC.VoteUserDo(c, req)
	if err != nil {
		return 0, 0, err
	}
	if rly != nil {
		return rly.UserAvailVoteCount, rly.UserCanVoteCountForItem, nil
	}
	return 0, 0, ecode.NothingFound
}

// VoteUserUndo
func (d *Dao) VoteUserUndo(c context.Context, actID, groupID, itemID, mid int64) (int64, int64, error) {
	req := &api.VoteUserUndoReq{ActivityId: actID, SourceGroupId: groupID, SourceItemId: itemID, Mid: mid}
	rly, err := d.actRPC.VoteUserUndo(c, req)
	if err != nil {
		return 0, 0, err
	}
	if rly != nil {
		return rly.UserAvailVoteCount, rly.UserCanVoteCountForItem, nil
	}
	return 0, 0, ecode.NothingFound
}

func (d *Dao) RankResult(c context.Context, id, pn, ps int64) (*api.RankResultResp, error) {
	return d.actRPC.RankResult(c, &api.RankResultReq{RankID: id, Pn: pn, Ps: ps})
}

func (d *Dao) ActSubject(c context.Context, sid int64) (*api.ActSubjectReply, error) {
	rly, err := d.actRPC.ActSubject(c, &api.ActSubjectReq{Sid: sid})
	if err != nil {
		log.Error("Fail to reqeust actGRPC.ActSubject, sid=%d error=%+v", sid, err)
		return nil, err
	}
	return rly, nil
}
