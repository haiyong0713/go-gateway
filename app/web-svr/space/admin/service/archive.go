package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/admin/model"
)

func (s *Service) TopArc(_ context.Context, mid int64) (data *model.TopArc, err error) {
	data = &model.TopArc{Mid: mid}
	if err = s.dao.DB.Table(data.TableName()).Where("mid=?", mid).First(&data).Error; err != nil {
		log.Error("TopArc mid:%d error(%v)", mid, err)
	}
	return
}

func (s *Service) TopArcClear(c context.Context, mid int64) (oldData *model.TopArc, err error) {
	if oldData, err = s.TopArc(c, mid); err != nil {
		return
	}
	if err = s.dao.DB.Table(oldData.TableName()).Where("mid=?", mid).Where("mid=?", mid).Update("recommend_reason", "").Error; err != nil {
		log.Error("TopArcClear mid:%d error(%v)", mid, err)
	}
	return
}

func (s *Service) Masterpiece(_ context.Context, mid int64) (list []*model.Masterpiece, err error) {
	data := &model.Masterpiece{Mid: mid}
	if err = s.dao.DB.Table(data.TableName()).Where("mid=?", mid).Find(&list).Error; err != nil {
		log.Error("Masterpiece mid:%d error (%v)", mid, err)
		return
	}
	if len(list) == 0 {
		err = ecode.NothingFound
	}
	return
}

func (s *Service) MasterpieceClear(_ context.Context, mid, aid int64) (oldData *model.Masterpiece, err error) {
	oldData = &model.Masterpiece{Mid: mid}
	if err = s.dao.DB.Table(oldData.TableName()).Where("mid=?", mid).Where("aid=?", aid).First(&oldData).Error; err != nil {
		log.Error("MasterpieceClear mid:%d error (%v)", mid, err)
		return
	}
	if err = s.dao.DB.Table(oldData.TableName()).Where("mid=?", mid).Where("mid=?", mid).Where("aid=?", aid).Update("recommend_reason", "").Error; err != nil {
		log.Error("MasterpieceClear mid:%d aid:%d error(%v)", mid, aid, err)
	}
	return
}
