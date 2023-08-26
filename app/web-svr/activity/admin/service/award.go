package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus/report"
	xtime "go-common/library/time"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/admin/model"
	"go-gateway/app/web-svr/activity/admin/model/task"

	"github.com/jinzhu/gorm"
)

func (s *Service) AwardList(_ context.Context, keyword, state string, pn, ps int64) (list []*model.Award, count int64, err error) {
	source := s.dao.DB.Model(&model.Award{})
	if keyword != "" {
		if sid, e := strconv.ParseInt(keyword, 10, 64); e != nil && sid > 0 {
			source = source.Where("sid = ?", sid)
		} else {
			source = source.Where("name LIKE ?", "%"+keyword+"%")
		}
	}
	if state != "" {
		stateData, _ := strconv.Atoi(state)
		source = source.Where("state = ?", stateData)
	}
	if err = source.Count(&count).Error; err != nil {
		log.Error("AwardList count keyword(%s) pn(%d) ps(%d) error (%v)", keyword, pn, ps, err)
		return
	}
	if count == 0 {
		list = make([]*model.Award, 0)
		return
	}
	if err = source.Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&list).Error; err != nil {
		log.Error("AwardList keyword(%s) pn(%d) ps(%d) error (%v)", keyword, pn, ps, err)
		return
	}
	return
}

func (s *Service) AwardDetail(_ context.Context, id int64) (data *model.Award, err error) {
	data = &model.Award{}
	if err = s.DB.Where("id = ?", id).First(data).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("AwardDetail s.DB.Where(sid = %d ) error(%v)", id, err)
	}
	return
}

func (s *Service) AwardAdd(_ context.Context, arg *model.AddAwardArg) (err error) {
	awardAdd := &model.Award{
		Name:         arg.Name,
		Etime:        arg.Etime,
		Sid:          arg.Sid,
		Type:         arg.Type,
		SourceID:     xstr.JoinInts(arg.SourceID),
		SourceExpire: arg.SourceExpire,
		State:        arg.State,
		Author:       arg.Author,
		SidType:      arg.SidType,
		OtherSids:    arg.OtherSids,
	}
	if arg.SidType == 1 {
		//add task
		tx := s.dao.DB.Begin()
		if err = tx.Error; err != nil {
			log.Error("AwardAdd s.dao.DB.Begin error(%v)", err)
			return
		}
		taskAdd := &task.Task{
			Name:        arg.Name,
			BusinessID:  1,
			ForeignID:   arg.Sid,
			FinishCount: 1,
			Attribute:   5,
			AwardType:   2,
			AwardID:     arg.SourceID[0],
			AwardCount:  1,
			State:       1,
			Stime:       xtime.Time(time.Now().Unix()),
			Etime:       arg.Etime,
			AwardExpire: arg.SourceExpire,
		}
		if err = tx.Model(&task.Task{}).Create(taskAdd).Error; err != nil {
			log.Error("AwardAdd s.dao.DB.Model Create(%+v) error(%v)", arg, err)
			err = tx.Rollback().Error
			return
		}
		awardAdd.TaskID = taskAdd.ID
		if err = tx.Model(&model.Award{}).Create(awardAdd).Error; err != nil {
			log.Error("AwardAdd tx.Model Create(%+v) error(%v)", arg, err)
			tx.Rollback()
			if strings.Contains(err.Error(), "Duplicate entry") {
				err = ecode.Error(ecode.Conflict, "该活动id已建立奖励")
			}
			return
		}
		err = tx.Commit().Error
	} else {
		if err = s.dao.DB.Model(&model.Award{}).Create(awardAdd).Error; err != nil {
			log.Error("AwardAdd s.dao.DB.Model Create(%+v) error(%v)", arg, err)
			if strings.Contains(err.Error(), "Duplicate entry") {
				err = ecode.Error(ecode.Conflict, "该活动id已建立奖励")
			}
		}
	}
	return
}

func (s *Service) AwardSave(_ context.Context, arg *model.SaveAwardArg) (err error) {
	preData := new(model.Award)
	if err = s.dao.DB.Where("id = ?", arg.ID).First(&preData).Error; err != nil {
		log.Error("AwardSave s.dao.DB.Where id(%d) error(%d)", arg.ID, err)
		return
	}
	if preData.SidType != arg.SidType {
		err = ecode.Error(ecode.RequestErr, "发放形式不能修改")
		return
	}
	saveArg := map[string]interface{}{
		"name":          arg.Name,
		"etime":         arg.Etime,
		"sid":           arg.Sid,
		"type":          arg.Type,
		"source_id":     xstr.JoinInts(arg.SourceID),
		"source_expire": arg.SourceExpire,
		"state":         arg.State,
		"author":        arg.Author,
		"sid_type":      arg.SidType,
		"other_sids":    arg.OtherSids,
	}
	if arg.SidType == 1 && preData.TaskID > 0 && (preData.SourceExpire != arg.SourceExpire || preData.Etime != arg.Etime) {
		//change task
		tx := s.dao.DB.Begin()
		if err = tx.Error; err != nil {
			log.Error("AwardSave s.dao.DB.Begin error(%v)", err)
			return
		}
		awardUp := map[string]interface{}{"award_expire": arg.SourceExpire, "etime": arg.Etime}
		if err = tx.Model(&task.Task{}).Where("id = ?", preData.TaskID).Updates(awardUp).Error; err != nil {
			log.Error("AwardSave tx save task(%+v) error(%v)", awardUp, err)
			err = tx.Rollback().Error
			return
		}
		if err = tx.Model(&model.Award{}).Where("id = ?", arg.ID).Updates(saveArg).Error; err != nil {
			log.Error("AwardSave tx saveArg(%+v) error(%v)", saveArg, err)
			err = tx.Rollback().Error
			return
		}
		err = tx.Commit().Error
	} else {
		if err = s.dao.DB.Model(&model.Award{}).Where("id = ?", arg.ID).Updates(saveArg).Error; err != nil {
			log.Error("AwardSave saveArg(%+v) error(%v)", saveArg, err)
			if strings.Contains(err.Error(), "Duplicate entry") {
				err = ecode.Error(ecode.Conflict, "该活动id已建立奖励")
				return
			}
		}
	}
	return
}

func (s *Service) AwardLog(c context.Context, oid, mid, pn, ps int64) (list []*report.UserActionLog, count int64, err error) {
	return s.dao.UserAwardLog(c, oid, mid, pn, ps)
}

func (s *Service) AwardLogExport(c context.Context, oid int64) (list []*report.UserActionLog, count int64, err error) {
	return s.dao.UserAwardLog(c, oid, 0, 1, 0)
}
