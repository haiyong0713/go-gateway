package service

import (
	"context"
	"time"

	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	likemdl "go-gateway/app/web-svr/activity/job/model/like"
)

func (s *Service) upSubjectLikeStickTop(ctx context.Context, preSubject, subject *likemdl.ActSubject) error {
	if subject == nil {
		return nil
	}
	needUpStickTop := func() bool {
		if preSubject == nil {
			if subject.IsForbidListSearch() || subject.IsForbidListOversea() || subject.IsForbidListRcmd() || subject.IsForbidListOther() {
				return true
			}
			return false
		}
		if preSubject.IsForbidListSearch() != subject.IsForbidListSearch() ||
			preSubject.IsForbidListOversea() != subject.IsForbidListOversea() ||
			preSubject.IsForbidListRcmd() != subject.IsForbidListRcmd() ||
			preSubject.IsForbidListOther() != subject.IsForbidListOther() {
			return true
		}
		return false
	}()
	if !needUpStickTop {
		return nil
	}
	likeList, err := s.likeListWithOffset(ctx, subject.ID)
	if err != nil {
		return err
	}
	var aids []int64
	for _, v := range likeList {
		if v != nil && v.Wid > 0 {
			aids = append(aids, v.Wid)
		}
	}
	aidsLen := len(aids)
	if aidsLen == 0 {
		log.Warn("upLikeStickTop sid(%d) len(aids) == 0", subject.ID)
		return nil
	}
	archives := make(map[int64]*arcmdl.Arc, aidsLen)
	for i := 0; i < aidsLen; i += _aidBulkSize {
		time.Sleep(10 * time.Millisecond)
		var partAids []int64
		if i+_aidBulkSize > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_aidBulkSize]
		}
		partArcs, err := s.arcClient.Arcs(ctx, &arcmdl.ArcsRequest{Aids: partAids})
		if err != nil {
			log.Error("upLikeStickTop s.arcClient.Arcs partAids(%v) error(%v)", partAids, err)
			continue
		}
		for _, v := range partArcs.GetArcs() {
			if v != nil && v.IsNormal() {
				blocked := func() bool {
					for _, blockMid := range s.c.Faction.BlockMids {
						if v.Author.Mid == blockMid {
							return true
						}
					}
					return false
				}()
				if blocked {
					continue
				}
				archives[v.Aid] = v
			}
		}
	}
	var stickTopOnIDs, stickTopOffIDs []int64
	for _, v := range likeList {
		if v == nil || v.Wid == 0 {
			continue
		}
		arc, ok := archives[v.Wid]
		if !ok || arc == nil || !arc.IsNormal() {
			continue
		}
		if v.StickTop == 0 && checkStickTopOn(subject, arc) {
			stickTopOnIDs = append(stickTopOnIDs, v.ID)
		}
		if v.StickTop == 1 && !checkStickTopOn(subject, arc) {
			stickTopOffIDs = append(stickTopOffIDs, v.ID)
		}
	}
	if len(stickTopOnIDs) > 0 {
		if err = s.upLikeStickTop(ctx, stickTopOnIDs, 1); err != nil {
			log.Error("upLikeStickTop s.upLikeStickTop stickTopOnIDs(%v) error(%v)", stickTopOnIDs, err)
		}
	}
	if len(stickTopOffIDs) > 0 {
		if err = s.upLikeStickTop(ctx, stickTopOffIDs, 0); err != nil {
			log.Error("upLikeStickTop s.upLikeStickTop stickTopOffIDs(%v) error(%v)", stickTopOffIDs, err)
		}
	}
	return nil
}

func (s *Service) upLikeStickTop(ctx context.Context, ids []int64, stickTop int) error {
	if len(ids) == 0 {
		return nil
	}
	idsLen := len(ids)
	for i := 0; i < idsLen; i += _aidBulkSize {
		time.Sleep(10 * time.Millisecond)
		var partIDs []int64
		if i+_aidBulkSize > idsLen {
			partIDs = ids[i:]
		} else {
			partIDs = ids[i : i+_aidBulkSize]
		}
		if _, err := s.dao.UpLikeStickTop(ctx, partIDs, stickTop); err != nil {
			return err
		}
	}
	return nil
}

func checkStickTopOn(subject *likemdl.ActSubject, arc *arcmdl.Arc) bool {
	if (subject.IsForbidListSearch() && arc.AttrVal(likemdl.AttrBitNoSearch) == arcmdl.AttrYes) ||
		(subject.IsForbidListOversea() && arc.AttrVal(arcmdl.AttrBitOverseaLock) == arcmdl.AttrYes) ||
		(subject.IsForbidListRcmd() && arc.AttrVal(arcmdl.AttrBitNoRecommend) == arcmdl.AttrYes) ||
		(subject.IsForbidListOther() && ((arc.AttrVal(arcmdl.AttrBitNoDynamic) == arcmdl.AttrYes) || arc.AttrVal(arcmdl.AttrBitNoWeb) == arcmdl.AttrYes) || (arc.AttrVal(arcmdl.AttrBitNoMobile) == arcmdl.AttrYes)) {
		return true
	}
	return false
}

func (s *Service) likeListWithOffset(ctx context.Context, sid int64) ([]*likemdl.Like, error) {
	var (
		batch int
		list  []*likemdl.Like
	)
	for {
		likeList, err := s.likeList(ctx, sid, s.mysqlOffset(batch), maxArcBatchLikeLimit, _retryTimes)
		if err != nil {
			log.Error("s.dao.LikeList: error(%v)", err)
			return nil, err
		}
		if len(likeList) > 0 {
			list = append(list, likeList...)
		}
		if len(likeList) < maxArcBatchLikeLimit {
			break
		}
		time.Sleep(100 * time.Microsecond)
		batch++
	}
	return list, nil
}

func (s *Service) upSubjectStickTop(ctx context.Context, sid int64, arc *likemdl.Archive) {
	if arc == nil {
		return
	}
	subInfo, err := s.dao.ActSubject(ctx, sid)
	if err != nil {
		log.Error("upSubjectStickTop ActSubject sid:%d error:%v", sid, err)
		return
	}
	if subInfo == nil {
		return
	}
	stickTop := 0
	if checkStickTopOn(subInfo, &arcmdl.Arc{Aid: arc.Aid, MissionID: arc.MissionID, Attribute: arc.Attribute}) {
		stickTop = 1
	}
	if _, err = s.dao.UpLikeStickTopByWid(ctx, sid, arc.Aid, stickTop); err != nil {
		log.Error("upSubjectStickTop UpLikeStickTopByWid sid:%d aid:%d stickTop:%d error:%v", sid, arc.Aid, stickTop, err)
	}
}
