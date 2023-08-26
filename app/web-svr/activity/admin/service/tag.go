package service

import (
	"context"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/admin/model"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"time"
)

// TagIsActivity tag是否活动tag
func (s *Service) TagIsActivity(c context.Context, tagName string) (res *model.TagIsActivityRes, err error) {
	res = &model.TagIsActivityRes{}
	list, err := s.GetSubjectByTagName(c, tagName)
	if err != nil {
		return res, err
	}
	if len(list) > 0 {
		res.Status = model.TagIsActivity
		return
	}
	res.Status = model.TagIsNotActivity
	return
}

// TagToActivity 升级tag为活动tag
func (s *Service) TagToActivity(c context.Context, tagName string, startTime, endTime int64, userName string) (res *model.TagToActivityRes, err error) {
	list, err := s.GetSubjectByTagName(c, tagName)
	if err != nil {
		return res, err
	}
	if len(list) > 0 {
		return res, ecode.Error(ecode.RequestErr, fmt.Sprintf("已有同名活动tag"))
	}
	// 创建活动
	_, err = s.createTagActivity(c, tagName, startTime, endTime, userName)
	if err != nil {
		log.Errorc(c, "s.createActivity(%s,%d,%d)", tagName, startTime, endTime)
		return nil, err
	}
	return nil, nil
}

// TagToNormal 降级tag为普通tag
func (s *Service) TagToNormal(c context.Context, tagName string, userName string) (res *model.TagToActivityRes, err error) {
	list, err := s.GetSubjectByTagName(c, tagName)
	if err != nil {
		return res, err
	}
	if list != nil && len(list) > 0 {
		sid := make([]int64, 0)
		for _, v := range list {
			sid = append(sid, v.ID)
		}
		err = s.OffileSubject(c, sid, userName)
		if err != nil {
			log.Errorc(c, "s.OffileSubject(%v)", sid)
			return nil, err
		}
	}
	err = s.tagToNormal(c, tagName)
	return nil, err
}

func (s *Service) createTagActivity(c context.Context, tagName string, startTime, endTime int64, userName string) (int64, error) {
	stime := time.Unix(startTime, 0).Format("2006-01-02 15:04:05")
	etime := time.Unix(endTime, 0).Format("2006-01-02 15:04:05")
	params := &model.AddList{
		Type:        model.VIDEO2,
		Tags:        tagName,
		Name:        tagName,
		Author:      userName,
		State:       model.SubOnLine,
		Stime:       stime,
		Etime:       etime,
		WeightStime: stime,
		WeightEtime: etime,
	}
	return s.AddActSubject(c, params, tagrpc.TagType_TypeUper)
}

// GetSubjectByTagName .
func (s *Service) GetSubjectByTagName(c context.Context, tagName string) (res []*model.ActSubject, err error) {
	subjectTagList := make([]*model.ActSubject, 0)
	subjectList := make([]*model.ActSubject, 0)
	res = make([]*model.ActSubject, 0)

	eg := errgroup.WithContext(c)
	// 根据配置的tag名获取sid
	eg.Go(func(ctx context.Context) (e error) {
		if subjectTagList, e = s.getSubjectTagNameByTagName(ctx, tagName); e != nil {
			log.Errorc(c, " s.getSubjectTagNameByTagName(%s) error(%v)", tagName, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if subjectList, e = s.getSubjectByName(ctx, tagName); e != nil {
			log.Errorc(c, " s.getSubjectTagNameByTagName(%s) error(%v)", tagName, e)
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "eg.Wait error(%v)", err)
		return nil, err
	}

	subjectListMap := make(map[int64]*model.ActSubject)
	subjectTagList = append(subjectTagList, subjectList...)
	for _, v := range subjectTagList {
		subjectListMap[v.ID] = v
	}
	for _, v := range subjectListMap {
		res = append(res, v)
	}
	return
}

// getSubjectTagNameByTagName 根据配置的tag名获取sid
func (s *Service) getSubjectTagNameByTagName(c context.Context, tagName string) (res []*model.ActSubject, err error) {
	subjectList, err := s.SubProtocolByTagName(c, tagName)
	if err != nil {
		log.Errorc(c, "SubProtocolByTagName tagName (%s) err(%v)", tagName, err)
		return nil, err
	}
	if len(subjectList) > 0 {
		subjectIDList := make([]int64, 0)
		now := time.Now().Unix()
		for _, v := range subjectList {
			subjectIDList = append(subjectIDList, v.Sid)
		}
		listParams := &model.ListSub{
			IDs:    subjectIDList,
			Sctime: now,
			Ectime: now,
			States: []int{model.SubOnLine},
			Types:  []int{model.VIDEO, model.VIDEOLIKE, model.VIDEO2, model.PHONEVIDEO, model.SMALLVIDEO}, // 1,4,13,16,17
		}
		return s.SubjectListAll(c, listParams)
	}
	return []*model.ActSubject{}, nil
}

// getSubjectByName 根据名称获取
func (s *Service) getSubjectByName(c context.Context, name string) (res []*model.ActSubject, err error) {
	if name == "" {
		return []*model.ActSubject{}, nil
	}
	now := time.Now().Unix()
	listParams := &model.ListSub{
		Sctime: now,
		Ectime: now,
		States: []int{model.SubOnLine},
		Types:  []int{model.VIDEO, model.VIDEOLIKE, model.VIDEO2, model.PHONEVIDEO, model.SMALLVIDEO}, // 1,4,13,16,17
		Name:   name,
	}
	return s.SubjectListAll(c, listParams)
}
