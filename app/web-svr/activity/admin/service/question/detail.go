package question

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/question"
	"go-gateway/app/web-svr/activity/tools/lib/function"
	"strings"
)

// DetailList get question detail list.
func (s *Service) DetailList(c context.Context, baseID, pn, ps int64) (list []*question.Detail, count int64, err error) {
	source := s.dao.DB.Model(&question.Detail{}).Where(" base_id = ? and state != ? ", baseID, question.StateOffline)
	if err = source.Count(&count).Error; err != nil {
		log.Errorc(c, "DetailList count baseID(%d) error (%v)", baseID, err)
		return
	}
	if count == 0 {
		return
	}
	if err = source.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&list).Error; err != nil {
		log.Errorc(c, "DetailList list baseID(%d) pn(%d) ps(%d) error (%v)", baseID, pn, ps, err)
	}
	return
}

// AddDetail add question detail data.
func (s *Service) AddDetail(c context.Context, arg *question.AddDetailArg) (err error) {
	add := &question.Detail{
		BaseID:      arg.BaseID,
		Name:        arg.Name,
		RightAnswer: arg.RightAnswer,
		WrongAnswer: arg.WrongAnswer,
		Attribute:   arg.Attribute,
		State:       arg.State,
		Pic:         arg.Pic,
	}
	if err = s.dao.DB.Model(&question.Detail{}).Create(add).Error; err != nil {
		log.Error("AddDetail s.dao.DB.Model Create(%+v) error(%v)", arg, err)
	}
	return
}

// BatchAddDetail question batch add.
func (s *Service) BatchAddDetail(c context.Context, baseID int64, arg []*question.AddDetailArg) (err error) {
	var count int64

	for k, v := range arg {
		if s.checkLetax(c, v) {
			arg[k].State = question.State4Process
		}
	}

	if err = s.dao.DB.Model(&question.Base{}).Where("id = ?", baseID).Count(&count).Error; err != nil {
		log.Error("BatchAddDetail baseID(%d) error(%v)", baseID, err)
		return
	}
	if count == 0 {
		err = ecode.NothingFound
		return
	}
	return s.dao.BatchAddDetail(c, arg)
}

func (s *Service) checkLetax(ctx context.Context, arg *question.AddDetailArg) bool {

	if function.InInt64Slice(arg.BaseID, s.c.GaoKaoAnswer.BaseID) {
		log.Infoc(ctx, "BatchAddDetail find:%v , SpitTag:%v", arg.BaseID, s.c.GaoKaoAnswer.SpitTag)
		if strings.Contains(arg.Name, s.c.GaoKaoAnswer.SpitTag) ||
			strings.Contains(arg.RightAnswer, s.c.GaoKaoAnswer.SpitTag) ||
			strings.Contains(arg.WrongAnswer, s.c.GaoKaoAnswer.SpitTag) {
			//arg[k].State =  question.State4Process
			return true
		}
	}

	return false
}

// SaveDetail save detail data.
func (s *Service) SaveDetail(c context.Context, arg *question.SaveDetailArg) (err error) {
	if s.checkLetax(c, &question.AddDetailArg{
		BaseID:      arg.BaseID,
		Name:        arg.Name,
		RightAnswer: arg.RightAnswer,
		WrongAnswer: arg.WrongAnswer,
	}) {
		arg.State = question.State4Process
	}
	return s.dao.SaveDetail(c, arg)
}

// UpdateDetailState ...
func (s *Service) UpdateDetailState(c context.Context, id int64, state int) (err error) {
	return s.dao.UpdateDetailState(c, id, state)
}
