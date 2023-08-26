package remix

import (
	"context"
	"fmt"
	"strconv"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	mdlrank "go-gateway/app/web-svr/activity/interface/model/rank"
	remixmdl "go-gateway/app/web-svr/activity/interface/model/remix"
	taskmdl "go-gateway/app/web-svr/activity/interface/model/task"

	"github.com/pkg/errors"
)

const (
	// rankAll 总榜
	rankAll = 1
	// rankDay 日榜
	rankDay = 2
	// rankLength 榜单长度
	rankLength = 100
)

// MemberCount ...
func (s *Service) MemberCount(c context.Context) (res *remixmdl.MoneyCountReply, err error) {
	res = &remixmdl.MoneyCountReply{}
	moneyCount := &remixmdl.MoneyCount{}
	count, err := s.task.GetActityTaskCount(c, s.c.Remix.TaskID)
	if err != nil {
		log.Error("s.task.GetActityTaskCount: error(%v)", err)
		err = errors.Wrapf(err, "s.task.GetActityTaskCount")
		res.MoneyCount = moneyCount
		return res, ecode.ActivityRemixCountErr
	}
	moneyCount.Count = count
	moneyCount.Money = s.c.Remix.AllMoney
	res.MoneyCount = moneyCount
	return res, nil

}

// Personal 用户活动信息
func (s *Service) Personal(c context.Context, mid int64) (res *remixmdl.MemberActivityInfoReply, err error) {
	res = &remixmdl.MemberActivityInfoReply{}
	var (
		midInfo  *accountapi.InfoReply
		arcInfo  *api.ArcsReply
		midScore *mdlrank.Redis
		aids     []int64
		midRules []*taskmdl.MidRule
	)
	eg := errgroup.WithContext(c)
	// 获取账号信息
	eg.Go(func(ctx context.Context) error {
		midInfo, err = s.accClient.Info3(ctx, &accountapi.MidReq{Mid: mid})
		if err != nil {
			log.Error("s.accClient.Info3: error(%v)", err)
			err = errors.Wrapf(err, "s.accClient.Info3")
			return err
		}
		if midInfo == nil || midInfo.Info == nil {
			return ecode.ActivityRemixMidInfoErr
		}
		return nil
	})
	// 获取用户积分和稿件信息
	eg.Go(func(ctx context.Context) (err error) {
		key := fmt.Sprintf("%d", s.c.Remix.Sid)
		midScore, err = s.rank.GetMidRank(c, key, mid)
		if err != nil {
			log.Error("s.remix.GetMidActivityScore: error(%v)", err)
			err = errors.Wrapf(err, "s.remix.GetMidActivityScore")
			return err
		}
		return nil
	})
	// 获取任务完成情况
	eg.Go(func(ctx context.Context) (err error) {
		midRules, err = s.task.GetActivityTaskMidStatus(c, s.c.Remix.TaskID, mid)
		if err != nil {
			log.Error("s.task.GetActivityTaskMidStatus: error(%v)", err)
			err = errors.Wrapf(err, "s.task.GetActivityTaskMidStatus")
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.memberNickname")
		return res, ecode.ActivityRemixMidInfoErr
	}
	if midScore != nil && midScore.Aids != nil && len(midScore.Aids) > 0 {
		aids = midScore.Aids
		arcInfo, err = client.ArchiveClient.Arcs(c, &api.ArcsRequest{Aids: aids})
		if err != nil {
			log.Error("s.arcClient.Arcs: error(%v)", err)
		}
	}
	res.Account = s.accountToRemixAccount(c, midInfo.Info)
	res.Rank = &remixmdl.Rank{
		Video: []*remixmdl.Video{},
	}
	if midScore != nil {
		res.Rank = &remixmdl.Rank{
			Rank:  midScore.Rank,
			Video: s.archiveInfoTurnRemixArcInfo(c, aids, arcInfo),
			Score: midScore.Score,
		}
	}
	res.Task = []*taskmdl.MidRule{}
	if midRules != nil {
		res.Task = midRules
	}
	return res, nil
}

// archiveInfoTurnRemixArcInfo ...
func (s *Service) archiveInfoTurnRemixArcInfo(c context.Context, aids []int64, arcInfo *api.ArcsReply) []*remixmdl.Video {
	res := make([]*remixmdl.Video, 0)
	if arcInfo == nil || arcInfo.Arcs == nil {
		return res
	}
	for _, aid := range aids {
		if arc, ok := arcInfo.Arcs[aid]; ok {
			if arc != nil && arc.IsNormal() {
				res = append(res, &remixmdl.Video{
					Aid:      arc.Aid,
					TypeName: arc.TypeName,
					Title:    arc.Title,
					Desc:     arc.Desc,
					Duration: arc.Duration,
					Pic:      arc.Pic,
					View:     arc.Stat.View,
				})
			}
		}
	}
	return res
}

func (s *Service) accountToRemixAccount(c context.Context, midInfo *accountapi.Info) *remixmdl.Account {
	return &remixmdl.Account{
		Mid:  midInfo.Mid,
		Name: midInfo.Name,
		Face: midInfo.Face,
		Sign: midInfo.Sign,
		Sex:  midInfo.Sex,
	}
}

// accountInfoTurnRemixMixInfo ...
func (s *Service) accountInfoTurnRemixMidInfo(c context.Context, midInfo *accountapi.InfoReply) *remixmdl.Account {
	if midInfo != nil && midInfo.Info != nil {
		return s.accountToRemixAccount(c, midInfo.Info)
	}
	return &remixmdl.Account{}
}

// getRank 获得榜单
func (s *Service) getRank(c context.Context, sid int64) ([]*mdlrank.Redis, error) {
	key := strconv.FormatInt(sid, 10)
	res := make([]*mdlrank.Redis, 0)
	rank, err := s.rank.GetRank(c, key)
	if err != nil {
		log.Error("s.rank.GetRank(c, %s) error(%v)", key, err)
		return nil, err
	}
	res = rank
	if len(rank) > rankLength {
		res = rank[:rankLength]
	}
	return res, nil
}

func (s *Service) getSidByRankType(c context.Context, rankType int) (int64, error) {
	switch rankType {
	case rankAll:
		return s.c.Remix.Sid, nil
	case rankDay:
		return s.c.Remix.DaySid, nil
	}
	return 0, ecode.ActivityRemixSidErr
}

func (s *Service) getRankDetail(c context.Context, sid int64) (*remixmdl.RankReply, error) {
	res := remixmdl.RankReply{}
	replyRankBatch := make([]*remixmdl.RankMember, 0)
	rank, err := s.getRank(c, sid)
	if err != nil {
		log.Error("s.getRank(c, %d) error(%v)", sid, err)
		return nil, err
	}
	var (
		memberInfo map[int64]*accountapi.Info
		archive    map[int64]*api.Arc
	)
	mids := make([]int64, 0)
	aids := make([]int64, 0)
	for _, v := range rank {
		mids = append(mids, v.Mid)
		aids = append(aids, v.Aids...)
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {

		memberInfo, err = s.account.MemberInfo(c, mids)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		archive, err = s.archive.ArchiveInfo(c, aids)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.getRankDetail")
		return nil, err
	}
	for _, v := range rank {
		replyRank := &remixmdl.RankMember{}
		if account, ok := memberInfo[v.Mid]; ok {
			replyRank.Account = s.accountToRemixAccount(c, account)
			video := []*remixmdl.Video{}
			if v.Aids != nil && len(v.Aids) > 0 {
				for _, v := range v.Aids {
					if arc, ok := archive[v]; ok {
						video = append(video, s.archiveInfoToRemixVideo(c, arc))
					}
				}
				replyRank.Videos = video
			}
		}
		replyRank.Score = v.Score
		replyRankBatch = append(replyRankBatch, replyRank)
	}
	res.Rank = replyRankBatch
	return &res, nil
}

// Rank 排行榜获取
func (s *Service) Rank(c context.Context, rankType int) (*remixmdl.RankReply, error) {
	sid, err := s.getSidByRankType(c, rankType)
	if err != nil {
		return nil, err
	}
	return s.getRankDetail(c, sid)
}

// ChildRank child rank
func (s *Service) ChildRank(c context.Context, sid int64) (*remixmdl.RankReply, error) {
	return s.getRankDetail(c, sid)
}

func (s *Service) archiveInfoToRemixVideo(c context.Context, arc *api.Arc) *remixmdl.Video {
	return &remixmdl.Video{
		Aid:      arc.Aid,
		TypeName: arc.TypeName,
		Title:    arc.Title,
		Desc:     arc.Desc,
		Duration: arc.Duration,
		Pic:      arc.Pic,
		View:     arc.Stat.View,
	}
}
