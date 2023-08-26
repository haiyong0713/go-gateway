package dubbing

import (
	"context"
	"strconv"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	mdldubbing "go-gateway/app/web-svr/activity/interface/model/dubbing"
	mdlrank "go-gateway/app/web-svr/activity/interface/model/rank"
	taskmdl "go-gateway/app/web-svr/activity/interface/model/task"
	"go-gateway/pkg/idsafe/bvid"

	"github.com/pkg/errors"
)

const (
	// rankAll 总榜
	rankAll = 1
	// rankWeek 周榜
	rankWeek = 2
	// rankLength 榜单长度
	rankLength = 100
	// childRankLength 子榜单长度
	childRankLength = 20
)

// Personal 用户活动信息
func (s *Service) Personal(c context.Context, mid int64) (res *mdldubbing.MemberActivityInfoReply, err error) {
	res = &mdldubbing.MemberActivityInfoReply{}
	var (
		midScore *mdldubbing.MapMidDubbingScore
		midRules []*taskmdl.MidRule
	)
	eg := errgroup.WithContext(c)
	// 获取用户积分和稿件信息
	eg.Go(func(ctx context.Context) (err error) {
		midScore, err = s.dubbing.GetMidDubbingScore(c, mid)
		if err != nil {
			log.Errorc(c, "s.dubbing.GetMidDubbingScore: error(%v)", err)
			err = errors.Wrapf(err, "s.dubbing.GetMidDubbingScore")
			return err
		}
		return nil
	})
	// 获取任务完成情况
	eg.Go(func(ctx context.Context) (err error) {
		midRules, err = s.task.GetActivityTaskMidStatus(c, s.c.Dubbing.TaskID, mid)
		if err != nil {
			log.Errorc(c, "s.task.GetActivityTaskMidStatus: error(%v)", err)
			err = errors.Wrapf(err, "s.task.GetActivityTaskMidStatus")
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.memberNickname")
		return res, ecode.ActivityRemixMidInfoErr
	}
	res.Rank = &mdldubbing.Rank{}
	res.Score = make(map[int64]*mdldubbing.RedisDubbing)
	if midScore != nil {
		if midScore.Score != nil {
			if sidScore, ok := midScore.Score[s.c.Dubbing.Sid]; ok {
				res.Rank = &mdldubbing.Rank{
					Rank:  sidScore.Rank,
					Score: sidScore.Score,
					Diff:  sidScore.Diff,
				}
			}
			res.Score = midScore.Score
		}

	}
	res.Task = []*taskmdl.MidRule{}
	if midRules != nil {
		res.Task = midRules
	}
	return res, nil
}

// archiveInfoTurnDubbingArcInfo ...
func (s *Service) archiveInfoTurnDubbingArcInfo(c context.Context, aids []int64, arcInfo *api.ArcsReply) []*mdldubbing.Video {
	res := make([]*mdldubbing.Video, 0)
	if arcInfo == nil || arcInfo.Arcs == nil {
		return res
	}
	for _, aid := range aids {
		if arc, ok := arcInfo.Arcs[aid]; ok {
			if arc != nil && arc.IsNormal() {
				var bvidStr string
				var err error
				if bvidStr, err = bvid.AvToBv(arc.Aid); err != nil || bvidStr == "" {
					continue
				}
				res = append(res, &mdldubbing.Video{
					Bvid:     bvidStr,
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

func (s *Service) accountToDubbingAccount(c context.Context, midInfo *accountapi.Info) *mdldubbing.Account {
	return &mdldubbing.Account{
		Mid:  midInfo.Mid,
		Name: midInfo.Name,
		Face: midInfo.Face,
		Sign: midInfo.Sign,
		Sex:  midInfo.Sex,
	}
}

// accountInfoTurnRemixMixInfo ...
func (s *Service) accountInfoTurnDubbingMidInfo(c context.Context, midInfo *accountapi.InfoReply) *mdldubbing.Account {
	if midInfo != nil && midInfo.Info != nil {
		return s.accountToDubbingAccount(c, midInfo.Info)
	}
	return &mdldubbing.Account{}
}

// getRank 获得榜单
func (s *Service) getRank(c context.Context, sid int64) ([]*mdlrank.Redis, error) {
	key := strconv.FormatInt(sid, 10)
	res := make([]*mdlrank.Redis, 0)
	rank, err := s.rank.GetRank(c, key)
	if err != nil {
		log.Errorc(c, "s.rank.GetRank(c, %s) error(%v)", key, err)
		return nil, err
	}
	res = rank
	if sid == s.c.Dubbing.Sid || sid == s.c.Dubbing.DaySid+s.c.Dubbing.Sid {
		if len(rank) > rankLength {
			res = rank[:rankLength]
		}
	} else {
		if len(rank) > childRankLength {
			res = rank[:childRankLength]
		}
	}
	return res, nil
}

func (s *Service) isSidRight(c context.Context, sid int64) bool {
	for _, v := range s.c.Dubbing.SidList {
		if sid == v {
			return true
		}
	}
	return false
}

func (s *Service) getSidByRankType(c context.Context, rankType int, sid int64) (int64, error) {
	switch rankType {
	case rankAll:
		return sid, nil
	case rankWeek:
		return s.c.Dubbing.DaySid + sid, nil
	}
	return 0, ecode.ActivityRemixSidErr
}

func (s *Service) getRankDetail(c context.Context, sid int64) (*mdldubbing.RankReply, error) {
	res := mdldubbing.RankReply{}
	replyRankBatch := make([]*mdldubbing.RankMember, 0)
	rank, err := s.getRank(c, sid)
	if err != nil {
		log.Errorc(c, "s.getRank(c, %d) error(%v)", sid, err)
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
		log.Errorc(c, "eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.getRankDetail")
		return nil, err
	}
	for _, v := range rank {
		replyRank := &mdldubbing.RankMember{}
		if account, ok := memberInfo[v.Mid]; ok {
			replyRank.Account = s.accountToDubbingAccount(c, account)
			video := []*mdldubbing.Video{}
			if v.Aids != nil && len(v.Aids) > 0 {
				for _, v := range v.Aids {
					if arc, ok := archive[v]; ok {
						video = append(video, s.archiveInfoToDubbingVideo(c, arc))
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
func (s *Service) Rank(c context.Context, rankType int, sid int64) (*mdldubbing.RankReply, error) {
	if !s.isSidRight(c, sid) {
		return &mdldubbing.RankReply{}, ecode.ActivitySidError
	}
	sid, err := s.getSidByRankType(c, rankType, sid)

	if err != nil {
		return nil, err
	}
	return s.getRankDetail(c, sid)
}

func (s *Service) archiveInfoToDubbingVideo(c context.Context, arc *api.Arc) *mdldubbing.Video {
	var bvidStr string
	var err error
	if bvidStr, err = bvid.AvToBv(arc.Aid); err != nil || bvidStr == "" {
		return nil
	}
	return &mdldubbing.Video{
		Bvid:     bvidStr,
		TypeName: arc.TypeName,
		Title:    arc.Title,
		Desc:     arc.Desc,
		Duration: arc.Duration,
		Pic:      arc.Pic,
		View:     arc.Stat.View,
	}
}
