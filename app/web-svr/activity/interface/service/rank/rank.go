package rank

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	l "go-gateway/app/web-svr/activity/interface/model/like"
	rankmdl "go-gateway/app/web-svr/activity/interface/model/rank"
	"go-gateway/pkg/idsafe/bvid"

	accountapi "git.bilibili.co/bapis/bapis-go/account/service"

	"github.com/pkg/errors"

	"go-gateway/app/app-svr/archive/service/api"
)

func (s *Service) getRankConfig(c context.Context, sid int64, attributeType int) (config *rankmdl.Rank, err error) {
	config, err = s.rank.GetRankConfig(c, sid, attributeType)
	if err == nil && config != nil && config.SID != 0 {
		return config, nil
	}
	config, err = s.rank.GetRankConfigBySid(c, sid, attributeType)
	if err != nil {
		return nil, err
	}
	if config != nil {
		err = s.rank.SetRankConfig(c, sid, attributeType, config)
	}
	return

}

// Personal 用户活动信息
func (s *Service) Personal(c context.Context, sid int64, attributeType int, mid int64) (res *rankmdl.MidRankReply, err error) {
	res = &rankmdl.MidRankReply{}
	var (
		aids []int64
	)

	config, midScore, err := s.getMidRank(c, mid, sid, attributeType)
	if err != nil {
		log.Error("s.getMidRank: error(%v)", err)
		err = errors.Wrapf(err, "s.getMidRank")
		return &rankmdl.MidRankReply{}, nil
	}
	if midScore == nil {
		return &rankmdl.MidRankReply{}, nil
	}

	if config.IsShowScore == rankmdl.IsNotShowScore {
		midScore.Score = 0
	}

	res.Rank = midScore.Rank
	res.Score = midScore.Score
	res.Video = make([]*rankmdl.Video, 0)
	if midScore != nil {
		if midScore.Aids != nil {
			for _, v := range midScore.Aids {
				aids = append(aids, v.Aid)
			}
			archive, err := s.archive.ArchiveInfo(c, aids)
			if err != nil {
				log.Errorc(c, "s.archive.ArchiveInfo")
				return res, nil
			}
			for _, v := range midScore.Aids {
				if archive != nil {
					if arc, ok := archive[v.Aid]; ok {
						if config.IsShowScore == rankmdl.IsNotShowScore {
							v.Score = 0
						}
						res.Video = append(res.Video, s.archiveInfoToRankVideo(c, arc, v.Score))
					}
				}

			}
		}

	}

	return res, nil
}

// getSubject 。。。
func (s *Service) getSubject(c context.Context, sid int64) (subject *l.SubjectItem, err error) {
	if subject, err = s.like.ActSubject(c, sid); err != nil {
		return
	}
	return
}

func (s *Service) getMidRank(c context.Context, mid int64, sid int64, attributeType int) (rank *rankmdl.Rank, res *rankmdl.MidRank, err error) {
	rank, err = s.getRankConfig(c, sid, rankmdl.AidSource)
	if err != nil || rank == nil {
		return nil, nil, err
	}
	rankName := fmt.Sprintf("%d_%d", rank.ID, attributeType)
	res, err = s.rank.GetMidRank(c, rankName, mid)
	return
}

func (s *Service) getRank(c context.Context, sid int64, attributeType int) (rank *rankmdl.Rank, res []*rankmdl.MidRank, err error) {
	rank, err = s.getRankConfig(c, sid, rankmdl.AidSource)
	if err != nil || rank == nil {
		return nil, nil, err
	}
	rankName := fmt.Sprintf("%d_%d", rank.ID, attributeType)
	res, err = s.rank.GetRank(c, rankName)
	return
}

// GetRankDetail 排行榜结果
func (s *Service) GetRankDetail(c context.Context, sid int64, attributeType int, ps, pn int) (*rankmdl.ResultReply, error) {
	res := rankmdl.ResultReply{}
	replyRankBatch := make([]*rankmdl.Result, 0)
	var rank []*rankmdl.MidRank
	var config *rankmdl.Rank
	var (
		start = ((pn - 1) * ps)
		end   = start + ps
	)
	config, rank, err := s.getRank(c, sid, attributeType)
	if err != nil {
		log.Errorc(c, "s.getRank(c, %d) error(%v)", sid, err)
		return nil, err
	}
	page := &rankmdl.Page{}
	page.Total = len(rank)
	page.Num = pn
	page.Size = ps
	res.Page = page

	if len(rank)-1 < start {
		page.HasMore = false
		return &res, nil
	}
	if end >= len(rank)-1 {
		end = len(rank)
		page.HasMore = false
	} else {
		page.HasMore = true
	}
	rank = rank[start:end]
	var (
		memberInfo map[int64]*accountapi.Info
		archive    map[int64]*api.Arc
	)
	mids := make([]int64, 0)
	aids := make([]int64, 0)
	mapAidScore := make(map[int64]*rankmdl.AidScore)
	for _, v := range rank {
		mids = append(mids, v.Mid)
		if v.Aids != nil {
			for _, v := range v.Aids {
				aids = append(aids, v.Aid)
				mapAidScore[v.Aid] = v
			}

		}
	}
	newAids := aids
	// newAids, err := s.filterList(c, aids, subject)
	eg2 := errgroup.WithContext(c)
	eg2.Go(func(ctx context.Context) (err error) {

		memberInfo, err = s.account.MemberInfo(c, mids)
		return err
	})
	eg2.Go(func(ctx context.Context) (err error) {
		archive, err = s.archive.ArchiveInfo(c, newAids)
		return err
	})
	if err := eg2.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		err = errors.Wrapf(err, "s.getRankDetail")
		return nil, err
	}
	var count int64
	for _, v := range rank {
		if count >= config.Top {
			break
		}
		replyRank := &rankmdl.Result{}
		if account, ok := memberInfo[v.Mid]; ok {
			video := []*rankmdl.Video{}
			if v.Aids != nil && len(v.Aids) > 0 {
				for _, v := range v.Aids {
					if arc, ok := archive[v.Aid]; ok {
						if score, ok1 := mapAidScore[v.Aid]; ok1 {
							if arc.IsNormal() {
								if config.IsShowScore == rankmdl.IsNotShowScore {
									score.Score = 0
								}
								video = append(video, s.archiveInfoToRankVideo(c, arc, score.Score))
							}

						}
					}
				}
				if len(video) > 0 {
					replyRank.Videos = video
					replyRank.Account = s.accountToRankAccount(c, account)
					if config.IsShowScore == rankmdl.IsNotShowScore {
						v.Score = 0
					}
					replyRank.Score = v.Score
					count++
					replyRankBatch = append(replyRankBatch, replyRank)

				}
			}
		}
	}
	res.Rank = replyRankBatch
	return &res, nil
}

func (s *Service) archiveInfoToRankVideo(c context.Context, arc *api.Arc, score int64) *rankmdl.Video {
	var bvidStr string
	var err error
	if bvidStr, err = bvid.AvToBv(arc.Aid); err != nil || bvidStr == "" {
		return nil
	}
	return &rankmdl.Video{
		Bvid:     bvidStr,
		TypeName: arc.TypeName,
		Title:    arc.Title,
		Desc:     arc.Desc,
		Duration: arc.Duration,
		Pic:      arc.Pic,
		View:     arc.Stat.View,
		Like:     arc.Stat.Like,
		Danmaku:  arc.Stat.Danmaku,
		Reply:    arc.Stat.Reply,
		Fav:      arc.Stat.Fav,
		Coin:     arc.Stat.Coin,
		Share:    arc.Stat.Share,
		Score:    score,
		PubDate:  arc.PubDate,
		Mid:      arc.Author.Mid,
	}
}

func (s *Service) filterList(c context.Context, aids []int64, subject *l.SubjectItem) (res []int64, err error) {
	res = make([]int64, 0)
	if len(aids) > 0 {
		aidsMap, err := s.filterAid(c, aids, subject)
		if err != nil {
			return nil, err
		}
		for _, v := range aids {
			if _, ok := aidsMap[v]; ok {
				res = append(res, v)
			}
		}
	}
	return res, nil

}

func (s *Service) filterAid(c context.Context, aids []int64, subject *l.SubjectItem) (list map[int64]struct{}, err error) {
	aidFlowControlMap, err := s.archive.ArchiveFlowControl(c, aids)
	if err != nil {
		return
	}
	list = make(map[int64]struct{})
	for _, aid := range aids {
		if flowControl, ok := aidFlowControlMap[aid]; ok {
			if flowControl != nil && flowControl.ForbiddenItems != nil {
				for _, control := range flowControl.ForbiddenItems {
					if control != nil && control.Value == l.FlowControlYes {
						switch control.Key {
						case l.ArchiveNoRank:
							if subject.IsShieldRank() {
								continue
							}
						case l.ArchiveNoDynamic:
							if subject.IsShieldDynamic() {
								continue
							}
						case l.ArchiveNoRecommend:
							if subject.IsShieldRecommend() {
								continue
							}
						case l.ArchiveNoHot:
							if subject.IsShieldHot() {
								continue
							}
						case l.ArchiveNoFansDynamic:
							if subject.IsShieldFansDynamic() {
								continue
							}
						case l.ArchiveNoSearch:
							if subject.IsShieldSearch() {
								continue
							}
						case l.ArchiveNoOversea:
							if subject.IsShieldOversea() {
								continue
							}
						}
					}
				}
			}
		}
		list[aid] = struct{}{}
	}
	return
}

func (s *Service) accountToRankAccount(c context.Context, midInfo *accountapi.Info) *rankmdl.Account {
	return &rankmdl.Account{
		Mid:  midInfo.Mid,
		Name: midInfo.Name,
		Face: midInfo.Face,
		Sign: midInfo.Sign,
		Sex:  midInfo.Sex,
	}
}
