package source

import (
	"context"
	"go-common/library/log"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
	"time"
)

// AllBlackWhiteArchive 黑白名单稿件列表
func (s *Service) AllBlackWhiteArchive(c context.Context, id int64) (black []*rankmdl.Intervention, white []*rankmdl.Intervention, err error) {
	var (
		batch int
	)
	list := make([]*rankmdl.Intervention, 0)
	for {
		likeList, err := s.rankDao.AllIntervention(c, id, rankmdl.InterventionObjectArchive, s.mysqlOffset(batch), maxArcBatchLikeLimit)
		if err != nil {
			log.Errorc(c, "s.rankDao.AllIntervention: error(%v)", err)
			return nil, nil, err
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
	black, white = s.blackandWhite(c, list)
	return black, white, nil
}

// AllBlackWhiteUp 黑白名单Up列表
func (s *Service) AllBlackWhiteUp(c context.Context, id int64) (black []*rankmdl.Intervention, white []*rankmdl.Intervention, err error) {
	var (
		batch int
	)
	list := make([]*rankmdl.Intervention, 0)
	for {
		likeList, err := s.rankDao.AllIntervention(c, id, rankmdl.InterventionObjectUp, s.mysqlOffset(batch), maxArcBatchLikeLimit)
		if err != nil {
			log.Errorc(c, "s.rankDao.AllIntervention: error(%v)", err)
			return nil, nil, err
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
	black, white = s.blackandWhite(c, list)
	return black, white, nil
}

// blackandWhite 黑白名单列表
func (s *Service) blackandWhite(c context.Context, list []*rankmdl.Intervention) (black []*rankmdl.Intervention, white []*rankmdl.Intervention) {
	black = make([]*rankmdl.Intervention, 0)
	white = make([]*rankmdl.Intervention, 0)
	if list != nil {
		for _, v := range list {
			if v.InterventionType == rankmdl.InterventionTypeBlack {
				black = append(black, v)
				continue
			}
			if v.InterventionType == rankmdl.InterventionTypeWhite {
				white = append(white, v)
				continue
			}
		}
	}
	return black, white
}
