package service

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/space/interface/model"

	pgcapi "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	progapi "git.bilibili.co/bapis/bapis-go/pgc/service/progress"
	seasonapi "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"

	"go-common/library/sync/errgroup.v2"
)

const (
	_isNewSecond = 259200
	_isNew       = 1
)

var (
	_emptyBangumiList = make([]*model.Bangumi, 0)
	_emptyFollowList  = make([]*model.FollowCard, 0)
)

// BangumiList get bangumi list by mid.
func (s *Service) BangumiList(c context.Context, mid, vmid int64, pn, ps int) ([]*model.Bangumi, int, error) {
	if mid != vmid {
		if err := s.privacyCheck(c, vmid, model.PcyBangumi); err != nil {
			return nil, 0, err
		}
	}
	reply, err := s.dao.BangumiList(c, mid, vmid, pn, ps)
	if err != nil {
		return nil, 0, err
	}
	if len(reply.Seasons) == 0 {
		return _emptyBangumiList, 0, err
	}
	var res []*model.Bangumi
	for _, v := range reply.Seasons {
		res = append(res, &model.Bangumi{
			SeasonID:      strconv.FormatInt(int64(v.SeasonId), 10),
			Title:         v.Title,
			IsFinish:      strconv.FormatInt(int64(v.IsFinish), 10),
			NewestEpIndex: v.NewEp.Title,
			TotalCount:    strconv.FormatInt(int64(v.TotalCount), 10),
			Cover:         v.Cover,
		})
	}
	return res, int(reply.Total), nil
}

// BangumiConcern bangumi concern.
func (s *Service) BangumiConcern(c context.Context, mid, seasonID int64) (err error) {
	return s.dao.BangumiConcern(c, mid, seasonID)
}

// BangumiUnConcern bangumi unconcern.
func (s *Service) BangumiUnConcern(c context.Context, mid, seasonID int64) (err error) {
	return s.dao.BangumiUnConcern(c, mid, seasonID)
}

// FollowList get ogv data follow list.
// nolint:gocognit
func (s *Service) FollowList(c context.Context, mid, vmid int64, typ, pn, ps, followStatus int32) (list []*model.FollowCard, count int32, err error) {
	if mid != vmid {
		if err = s.privacyCheck(c, vmid, model.PcyBangumi); err != nil {
			return
		}
	}
	var (
		reply         *pgcapi.MyFollowsReply
		seasonReply   *seasonapi.CardsInfoReply
		followReply   *pgcapi.FollowStatusByMidReply
		progressReply *progapi.ProgressProfileReply
		seasonIDs     []int32
	)
	if reply, err = s.pgcFollowClient.MyFollows(c, &pgcapi.MyFollowsReq{Mid: vmid, FollowType: typ, Pn: pn, Ps: ps, FollowStatus: followStatus}); err != nil {
		log.Error("FollowList s.pgcSeasonClient.MyFollows(%d,%d,%d,%d,%d) error(%v)", vmid, typ, pn, ps, followStatus, err)
		err = nil
		list = _emptyFollowList
		return
	}
	count = reply.Total
	if len(reply.Follows) == 0 {
		list = _emptyFollowList
		return
	}
	for _, v := range reply.Follows {
		if v.SeasonId > 0 {
			seasonIDs = append(seasonIDs, v.SeasonId)
		}
	}
	if len(seasonIDs) == 0 {
		list = _emptyFollowList
		return
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (e error) {
		if seasonReply, e = s.pgcSeasonClient.Cards(ctx, &seasonapi.SeasonInfoReq{SeasonIds: seasonIDs}); e != nil {
			log.Error("FollowList s.pgcSeasonClient.Cards(%v) error(%v)", seasonIDs, e)
			return e
		}
		return nil
	})
	if mid > 0 {
		if mid != vmid {
			// both follow
			group.Go(func(ctx context.Context) (e error) {
				if followReply, e = s.pgcFollowClient.StatusByMid(ctx, &pgcapi.FollowStatusByMidReq{Mid: mid, SeasonId: seasonIDs}); e != nil {
					log.Error("FollowList s.pgcFollowClient.StatusByMid(%d,%v) error(%v)", mid, seasonIDs, e)
				}
				return nil
			})
		} else {
			// progress
			group.Go(func(ctx context.Context) (e error) {
				if progressReply, e = s.pgcProgressClient.ProfileByMid(ctx, &progapi.ProgressInfoReq{Mid: mid, SeasonId: seasonIDs}); e != nil {
					log.Error("FollowList s.pgcProgressClient.ProfileByMid(%d,%v) error(%v)", mid, seasonIDs, e)
				}
				return nil
			})
		}
	}
	if err = group.Wait(); err != nil {
		log.Error("FollowList mid(%d) pn(%d),ps(%d) error(%v)", vmid, pn, ps, err)
		err = nil
		list = _emptyFollowList
		return
	}
	if seasonReply == nil || len(seasonReply.Cards) == 0 {
		list = _emptyFollowList
		return
	}
	for _, v := range reply.Follows {
		if season, ok := seasonReply.Cards[v.SeasonId]; ok {
			card := &model.FollowCard{CardInfoProto: season, FollowStatus: v.FollowStatus}
			if mid > 0 {
				if mid == vmid {
					var isPubNew int32
					card.BothFollow = true
					nowTs := time.Now().Unix()
					if newPub, e := time.Parse("2006-01-02 15:04:05", season.NewEp.PubTime); e == nil {
						if nowTs-newPub.Unix() <= _isNewSecond {
							isPubNew = _isNew
						}
					}
					if progressReply != nil && len(progressReply.Progresses) > 0 {
						if progress, ok := progressReply.Progresses[v.SeasonId]; ok {
							// progress.ts >= today -3 && newEp.pub_time>= today -3
							if isPubNew == _isNew && nowTs-progress.Time <= _isNewSecond {
								card.IsNew = _isNew
							}
							card.Progress = progress.IndexShow
						} else {
							card.IsNew = isPubNew
						}
					} else {
						card.IsNew = isPubNew
					}
				} else if followReply != nil && len(followReply.Result) > 0 {
					if follow, ok := followReply.Result[v.SeasonId]; ok {
						card.BothFollow = follow.Follow
					}
				}
			}
			list = append(list, card)
		}
	}
	return
}
