package source

import (
	"context"
	"go-gateway/app/app-svr/archive/service/api"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/job/model/like"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
	"go-gateway/app/web-svr/activity/job/model/source"

	"github.com/pkg/errors"

	flowcontrolapi "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

const (
	// maxArcBatchLikeLimit 一次从表中获取稿件数量
	maxArcBatchLikeLimit = 1000
	// dbChannelLength db channel 长度
	dbChannelLength = 100
	// arcChannelLength 稿件channel长度
	arcChannelLength = 100
)

// GetSubject ...
func (s *Service) GetSubject(c context.Context, sid int64) (*like.ActSubject, error) {
	subjectList, err := s.dao.SubjectDetailByIds(c, []int64{sid})
	if err != nil {
		return nil, err
	}
	if len(subjectList) != 1 {
		err = errors.New("s.dao.SubjectDetailByIds subject len error (%d)")
		return nil, err
	}
	subject := subjectList[0]
	return subject, nil
}

// GetSubjectAllArchive 获取数据源中所有稿件数据
func (s *Service) GetSubjectAllArchive(c context.Context, sid int64) (list []*source.Archive, err error) {
	sids, err := s.GetAllSid(c, sid)
	list = make([]*source.Archive, 0)
	if err != nil {
		return
	}
	subject, err := s.GetSubject(c, sid)
	if err != nil {
		return
	}
	archiveList, err := s.GetArchiveBySids(c, sids)
	if err != nil {
		return
	}
	return s.FilterArchive(c, subject, archiveList)
}

// FilterArchive 过滤可展示稿件
func (s *Service) FilterArchive(c context.Context, subject *like.ActSubject, archive []*source.Archive) (list []*source.Archive, err error) {
	list = make([]*source.Archive, 0)
	if archive == nil {
		return
	}
	for _, v := range archive {
		if subject.IsShieldDynamic() && v.NoDynamic {
			v.State = source.ArchiveStateNotNormal
		}
		if subject.IsShieldRank() && v.NoRank {
			v.State = source.ArchiveStateNotNormal
		}
		if subject.IsShieldRecommend() && v.NoRecommend {
			v.State = source.ArchiveStateNotNormal
		}
		if subject.IsShieldHot() && v.NoHot {
			v.State = source.ArchiveStateNotNormal
		}
		if subject.IsShieldFansDynamic() && v.NoFansDynamic {
			v.State = source.ArchiveStateNotNormal
		}
		if subject.IsShieldSearch() && v.NoSearch {
			v.State = source.ArchiveStateNotNormal
		}
		if subject.IsShieldOversea() && v.NoOversea {
			v.State = source.ArchiveStateNotNormal
		}
		list = append(list, v)
	}
	return
}

// GetAllSid 获取数据源
func (s *Service) GetAllSid(c context.Context, sid int64) ([]int64, error) {
	sids := make([]int64, 0)
	sids = append(sids, sid)
	subjectChild, err := s.dao.SubjectChild(c, sid)
	if err != nil {
		log.Errorc(c, "s.dao.SubjectChild sid:%d error:%v", sid, err)
		return nil, err
	}
	if subjectChild != nil {
		sids = append(sids, subjectChild.ChildIdsList...)
	}
	return sids, nil
}

// GetAidBySid 根据sid获取子母数据源的全部稿件id
func (s *Service) GetAidBySid(c context.Context, sid int64) ([]*like.Like, error) {
	sids, err := s.GetAllSid(c, sid)
	if err != nil {
		return nil, err
	}
	return s.LikesAll(c, sids)
}

// GetArchiveBySids 根据数据源id获取所有稿件
func (s *Service) GetArchiveBySids(c context.Context, sids []int64) (list []*source.Archive, err error) {
	archiveList, err := s.LikesAll(c, sids)
	if len(archiveList) > 0 {
		return s.ArchiveInfoDetailFilter(c, archiveList, true)
	}
	return

}

// LikesAllAids ...
func (s *Service) LikesAllAids(c context.Context, sids []int64) ([]int64, error) {
	archiveList, err := s.LikesAll(c, sids)
	if err != nil {
		return nil, err
	}
	aids := make([]int64, 0)
	if archiveList != nil {
		for _, v := range archiveList {
			aids = append(aids, v.Wid)
		}
		return aids, nil
	}
	return []int64{}, nil
}

// LikesNormalAids ...
func (s *Service) LikesNormalAids(c context.Context, sids []int64) ([]int64, error) {
	archiveList, err := s.LikesNormal(c, sids)
	if err != nil {
		return nil, err
	}
	aids := make([]int64, 0)
	if archiveList != nil {
		for _, v := range archiveList {
			aids = append(aids, v.Wid)
		}
		return aids, nil
	}
	return []int64{}, nil
}

// LikesAll 稿件信息获取,包括被删除的
func (s *Service) LikesAll(c context.Context, sid []int64) ([]*like.Like, error) {
	var (
		batch int
	)
	list := make([]*like.Like, 0)
	for {
		likeList, err := s.dao.LikesAllList(c, sid, s.mysqlOffset(batch), maxArcBatchLikeLimit)
		if err != nil {
			log.Errorc(c, "s.dao.LikesAllList: error(%v)", err)
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

// LikesNormal 稿件信息获取,不包括删除
func (s *Service) LikesNormal(c context.Context, sid []int64) ([]*like.Like, error) {
	var (
		batch int
	)
	list := make([]*like.Like, 0)
	for {
		likeList, err := s.dao.LikesNormalList(c, sid, s.mysqlOffset(batch), maxArcBatchLikeLimit)
		if err != nil {
			log.Errorc(c, "s.dao.LikesNormalList: error(%v)", err)
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

// mysqlOffset count mysql offset
func (s *Service) mysqlOffset(batch int) int {
	return batch * maxArcBatchLikeLimit
}

// ArchiveInfoDetailFilter 稿件点赞信息详情并过滤
func (s *Service) ArchiveInfoDetailFilter(c context.Context, archives []*like.Like, needFilter bool) ([]*source.Archive, error) {
	var (
		archive     map[int64]*api.Arc
		flowControl map[int64]*flowcontrolapi.FlowCtlInfoReply
	)
	var list = make([]*source.Archive, 0)
	eg := errgroup.WithContext(c)
	aids := make([]int64, 0)
	notNormal := make([]*like.Like, 0)
	for _, v := range archives {
		if v.IsNormal() {
			aids = append(aids, v.Wid)
		} else {
			notNormal = append(notNormal, v)
		}
	}
	eg.Go(func(ctx context.Context) (err error) {
		archive, err = s.ArchiveInfo(c, aids)
		if err != nil {
			log.Errorc(c, "s.ArchiveInfo err(%v)", err)
		}
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		if needFilter {
			flowControl, err = s.ArchiveFlowControl(c, aids)
			if err != nil {
				log.Errorc(c, "s.ArchiveFlowControl err(%v)", err)
			}
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return list, err
	}
	for _, v := range aids {
		if arc, ok := archive[v]; ok {
			listArc := &source.Archive{}
			if arc.IsNormal() {
				listArc = &source.Archive{
					Mid:     arc.Author.Mid,
					Aid:     arc.Aid,
					View:    int64(arc.Stat.View),
					Danmaku: int64(arc.Stat.Danmaku),
					Reply:   int64(arc.Stat.Reply),
					Fav:     int64(arc.Stat.Fav),
					Coin:    int64(arc.Stat.Coin),
					Share:   int64(arc.Stat.Share),
					Like:    int64(arc.Stat.Like),
					Videos:  int64(arc.Videos),
					TypeID:  int64(arc.TypeID),
					State:   source.ArchiveStateNormal,
					Ctime:   int64(arc.Ctime),
					PubTime: int64(arc.PubDate),
				}
				if flow, ok := flowControl[v]; ok {
					if flow != nil {
						for _, control := range flow.ForbiddenItems {
							if control != nil && control.Value == like.FlowControlYes {
								switch control.Key {
								case like.ArchiveNoRank:
									listArc.NoRank = true
								case like.ArchiveNoDynamic:
									listArc.NoDynamic = true
								case like.ArchiveNoRecommend:
									listArc.NoRecommend = true
								case like.ArchiveNoHot:
									listArc.NoHot = true
								case like.ArchiveNoFansDynamic:
									listArc.NoFansDynamic = true
								case like.ArchiveNoSearch:
									listArc.NoSearch = true
								case like.ArchiveNoOversea:
									listArc.NoOversea = true
								}
							}

						}
					}
				}
			} else {
				listArc = &source.Archive{
					Mid:   arc.Author.Mid,
					Aid:   arc.Aid,
					State: source.ArchiveStateNotNormal,
				}
			}

			list = append(list, listArc)
		}
	}
	for _, v := range notNormal {
		list = append(list, &source.Archive{
			Mid:   v.Mid,
			Aid:   v.Wid,
			State: source.ArchiveStateNotNormal,
		})
	}
	return list, nil
}

// ArchiveInfoDetailFromSnapshotFilter 稿件信息来源于快照
func (s *Service) ArchiveInfoDetailFromSnapshotFilter(c context.Context, id, batch int64, attributeType int, archives []*like.Like, needFilter bool) ([]*source.Archive, error) {
	var (
		archive     map[int64]*rankmdl.Snapshot
		flowControl map[int64]*flowcontrolapi.FlowCtlInfoReply
	)
	var list = make([]*source.Archive, 0)
	eg := errgroup.WithContext(c)
	aids := make([]int64, 0)
	notNormal := make([]*like.Like, 0)
	for _, v := range archives {
		if v.IsBlack {
			notNormal = append(notNormal, v)
		} else {
			aids = append(aids, v.Wid)
		}
	}
	eg.Go(func(ctx context.Context) (err error) {
		archive, err = s.ArchiveSnapshotInfoByAids(c, id, batch, attributeType, aids)
		if err != nil {
			log.Errorc(c, "s.ArchiveInfo err(%v)", err)
		}
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		if needFilter {
			flowControl, err = s.ArchiveFlowControl(c, aids)
			if err != nil {
				log.Errorc(c, "s.ArchiveFlowControl err(%v)", err)
			}
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return list, err
	}
	for _, v := range aids {
		if arc, ok := archive[v]; ok {
			listArc := &source.Archive{}
			if arc.State == rankmdl.SnapshotStateNormal {
				listArc = &source.Archive{
					Mid:     arc.MID,
					Aid:     arc.AID,
					View:    arc.View,
					Danmaku: arc.Danmaku,
					Reply:   arc.Reply,
					Fav:     arc.Fav,
					Coin:    arc.Coin,
					Share:   arc.Share,
					Like:    arc.Like,
					Videos:  arc.Videos,
					State:   source.ArchiveStateNormal,
					Ctime:   int64(arc.ArcCtime),
					PubTime: int64(arc.PubTime),
				}
				if flow, ok := flowControl[v]; ok {
					if flow != nil {
						for _, control := range flow.ForbiddenItems {
							if control != nil && control.Value == like.FlowControlYes {
								switch control.Key {
								case like.ArchiveNoRank:
									listArc.NoRank = true
								case like.ArchiveNoDynamic:
									listArc.NoDynamic = true
								case like.ArchiveNoRecommend:
									listArc.NoRecommend = true
								case like.ArchiveNoHot:
									listArc.NoHot = true
								case like.ArchiveNoFansDynamic:
									listArc.NoFansDynamic = true
								case like.ArchiveNoSearch:
									listArc.NoSearch = true
								case like.ArchiveNoOversea:
									listArc.NoOversea = true
								}
							}

						}
					}
				}
			} else {
				listArc = &source.Archive{
					Mid:   arc.MID,
					Aid:   arc.AID,
					State: source.ArchiveStateNotNormal,
				}
			}

			list = append(list, listArc)
		}
	}
	for _, v := range notNormal {
		list = append(list, &source.Archive{
			Mid:   v.Mid,
			Aid:   v.Wid,
			State: source.ArchiveStateNotNormal,
		})
	}
	return list, nil
}
