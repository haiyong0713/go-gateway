package rank

import (
	"context"
	"errors"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
)

// Source 数据源
type Source interface {
	Get(c context.Context, s *Service, id int64, sid int64) (res []*like.Like, err error)
	GetSourceConfig(c context.Context, s *Service, sid int64) (*like.ActSubject, error)
}

// AidSource 视频数据源
type AidSource struct {
}

// Get 获取视频数据源稿件信息
func (a *AidSource) Get(c context.Context, s *Service, id int64, sid int64) (res []*like.Like, err error) {
	list, err := s.source.GetAidBySid(c, sid)
	res = make([]*like.Like, 0)
	if err != nil {
		log.Errorc(c, "s.source.GetAidBySid(%d) err(%v)", sid, err)
		return
	}
	black, white, err := s.source.AllBlackWhiteArchive(c, id)
	if err != nil {
		log.Errorc(c, "s.source.AllBlackWhiteArchive(%d) err(%v)", id, err)
		return
	}
	blackUp, _, err := s.source.AllBlackWhiteUp(c, id)
	if err != nil {
		log.Errorc(c, "s.source.AllBlackWhiteArchive(%d) err(%v)", id, err)
		return
	}
	blackMap := make(map[int64]struct{})
	blackUpMap := make(map[int64]struct{})
	resMap := make(map[int64]struct{})
	for _, v := range black {
		blackMap[v.OID] = struct{}{}
	}
	for _, v := range blackUp {
		blackUpMap[v.OID] = struct{}{}
	}

	for _, v := range list {
		if _, ok := blackMap[v.Wid]; ok {
			res = append(res, &like.Like{Wid: v.Wid, State: 0, Mid: v.Mid, IsBlack: true})
			resMap[v.Wid] = struct{}{}
			continue
		}
		if _, ok := blackUpMap[v.Mid]; ok {
			res = append(res, &like.Like{Wid: v.Wid, State: 0, Mid: v.Mid, IsBlack: true})
			resMap[v.Wid] = struct{}{}
			continue
		}
		res = append(res, v)
		resMap[v.Wid] = struct{}{}
	}
	for _, v := range white {
		if _, ok := resMap[v.OID]; ok {
			continue
		}
		res = append(res, &like.Like{Wid: v.OID, State: 1})
		resMap[v.OID] = struct{}{}
	}
	return
}

// GetSourceObj 获取稿件
func (s *Service) GetSourceObj(c context.Context, sidSource int) (Source, error) {
	switch sidSource {
	case rankmdl.SIDSourceAid:
		return &AidSource{}, nil
	default:
		return nil, errors.New("can not find source obj")
	}
}

// GetSourceConfig 获取数据源配置信息
func (a *AidSource) GetSourceConfig(c context.Context, s *Service, sid int64) (*like.ActSubject, error) {
	return s.source.GetSubject(c, sid)
}

// GetSourceConfig 获取数据源
func (s *Service) GetSourceConfig(c context.Context, id, sid int64, sidSource int) (*like.ActSubject, []*like.Like, error) {
	// 获取数据源
	sourceObj, err := s.GetSourceObj(c, sidSource)
	if err != nil {
		return nil, nil, err
	}
	aids, err := sourceObj.Get(c, s, id, sid)

	if err != nil {
		return nil, nil, err
	}
	subject, err := sourceObj.GetSourceConfig(c, s, sid)
	if err != nil {
		return nil, nil, err
	}
	return subject, aids, nil
}
