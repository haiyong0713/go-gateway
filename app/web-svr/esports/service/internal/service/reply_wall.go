package service

import (
	"context"
	"fmt"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

const _maxReplyWall = 6

func (s *Service) GetReplyWallList(ctx context.Context, req *pb.GetReplyWallListReq) (res *pb.GetReplyWallListResponse, err error) {
	res = &pb.GetReplyWallListResponse{}
	replyWallList, err := s.dao.ReplyWallList(ctx)
	if err != nil {
		log.Errorc(ctx, "GetReplyWallInfo s.dao.ReplyWallList() error(%+v)", err)
		return
	}
	if len(replyWallList) == 0 {
		return
	}
	contestInfo, err := s.GetContestInfo(ctx, &pb.GetContestRequest{Cid: replyWallList[0].ContestID})
	if err != nil {
		log.Errorc(ctx, "GetReplyWallInfo s.GetContestInfo() req(%+v) error(%+v)", req, err)
		return
	}
	if contestInfo == nil {
		log.Errorc(ctx, "GetReplyWallInfo s.GetContestInfo() req(%+v) contestInfo is nil", req)
		return
	}
	wallList, err := s.getReplyWallList(ctx, replyWallList)
	if err != nil {
		log.Errorc(ctx, "GetReplyWallInfo s.getReplyWallList() req(%+v) error(%+v)", req, err)
		return
	}
	res = &pb.GetReplyWallListResponse{
		Contest:   contestInfo.Contest,
		ReplyList: wallList,
	}
	return
}

func (s *Service) getReplyWallList(ctx context.Context, replyWallList []*model.ReplyWallModel) (res []*pb.ReplyWallInfo, err error) {
	var replyMids []int64
	for _, wall := range replyWallList {
		replyMids = append(replyMids, wall.Mid)
	}
	userInfoMap, err := s.dao.GetAccountInfos(ctx, replyMids)
	if err != nil {
		log.Errorc(ctx, "ContestReplyWall s.dao.GetAccountInfos() mids(%+v) error(%+v)", replyMids, err)
		return
	}
	for _, wall := range replyWallList {
		userInfo, ok := userInfoMap.Infos[wall.Mid]
		if !ok {
			continue
		}
		res = append(res, &pb.ReplyWallInfo{
			Mid:          userInfo.Mid,
			Name:         userInfo.Name,
			Face:         userInfo.Face,
			Sign:         userInfo.Sign,
			ReplyDetails: wall.ReplyDetails,
		})
	}
	return
}

func (s *Service) GetReplyWallModel(ctx context.Context, req *pb.GetReplyWallModelReq) (res *pb.SaveReplyWallModel, err error) {
	res = &pb.SaveReplyWallModel{}
	replyWallList, err := s.dao.RawReplyWall()
	if err != nil {
		log.Errorc(ctx, "GetReplyWall s.dao.GetReplyWall() req(%+v) error(%+v)", req, err)
		return
	}
	for _, reply := range replyWallList {
		res.ContestID = reply.ContestID
		replyDetails := &pb.ReplyWallModel{
			Mid:          reply.Mid,
			ReplyDetails: reply.ReplyDetails,
		}
		res.ReplyList = append(res.ReplyList, replyDetails)
	}

	return
}

func (s *Service) SaveReplyWall(ctx context.Context, req *pb.SaveReplyWallModel) (res *pb.NoArgsResponse, err error) {
	res = &pb.NoArgsResponse{}
	if err = s.checkParams(ctx, req); err != nil {
		return
	}
	if err = s.dao.ReplyWallUpdateTransaction(ctx, req); err != nil {
		log.Errorc(ctx, "SaveReplyWall s.dao.ReplyWallUpdateTransaction() req(%+v) error(%+v)", req, err)
		return
	}
	s.fanout.Do(ctx, func(ctx context.Context) {
		s.dao.DelCacheReplyWallList(ctx)
	})
	return
}

func (s *Service) checkParams(ctx context.Context, req *pb.SaveReplyWallModel) error {
	var (
		replyMids []int64
		notExists []int64
	)
	if len(req.GetReplyList()) > _maxReplyWall {
		err := xecode.Errorf(xecode.RequestErr, fmt.Sprintf("最大添加%d行评论", _maxReplyWall))
		return err
	}
	for _, wall := range req.ReplyList {
		replyMids = append(replyMids, wall.Mid)
	}
	userinfoMap, err := s.dao.GetAccountInfos(ctx, replyMids)
	if err != nil {
		log.Errorc(ctx, "SaveReplyWall s.dao.GetAccountInfos() req(%+v) error(%+v)", req, err)
		return err
	}
	for _, mid := range replyMids {
		if _, ok := userinfoMap.Infos[mid]; !ok {
			notExists = append(notExists, mid)
		}
	}
	if len(notExists) > 0 {
		err = xecode.Errorf(xecode.RequestErr, fmt.Sprintf("用户uid:(%+v)不正确", notExists))
		return err
	}
	return nil
}
