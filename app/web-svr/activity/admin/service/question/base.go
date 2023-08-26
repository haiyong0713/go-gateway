package question

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/question"
)

// BaseList get question base list.
func (s *Service) BaseList(c context.Context, pn, ps int64) (list []*question.Base, count int64, err error) {
	source := s.dao.DB.Model(&question.Base{})
	if err = source.Count(&count).Error; err != nil {
		log.Error("BaseList count error (%v)", err)
		return
	}
	if count == 0 {
		return
	}
	if err = source.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&list).Error; err != nil {
		log.Error("BaseList list pn(%d) ps(%d) error (%v)", pn, ps, err)
	}
	return
}

// BaseItem get base item.
func (s *Service) BaseItem(c context.Context, id int64) (data *question.Base, err error) {
	data = new(question.Base)
	if err = s.dao.DB.Model(&question.Base{}).Where("id = ?", id).First(data).Error; err != nil {
		log.Error("BaseItem id(%d) error (%v)", id, err)
	}
	return
}

// AddBase add question base data.
func (s *Service) AddBase(c context.Context, arg *question.AddBaseArg) (err error) {
	add := &question.Base{
		Name:           arg.Name,
		Separator:      arg.Separator,
		DistributeType: arg.DistributeType,
		BusinessID:     arg.BusinessID,
		ForeignID:      arg.ForeignID,
		Count:          arg.Count,
		OneTs:          arg.OneTs,
		RetryTs:        arg.RetryTs,
		Stime:          arg.Stime,
		Etime:          arg.Etime,
	}
	if err = s.dao.DB.Model(&question.Base{}).Create(add).Error; err != nil {
		log.Error("AddBase s.dao.DB.Model Create(%+v) error(%v)", arg, err)
	}
	err = s.dao.UserLogCreate(c, add.ID)
	return
}

// SaveBase save base data.
func (s *Service) SaveBase(c context.Context, arg *question.SaveBaseArg) (err error) {
	return s.dao.SaveBase(c, arg)
}
