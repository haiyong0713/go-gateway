package search

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/search"
	"go-gateway/app/app-svr/app-feed/admin/util"

	filterGRPC "git.bilibili.co/bapis/bapis-go/filter/service"
)

// DySeachList channel DySeach list
func (s *Service) DySeachList(lp *search.DySeachLP) (pager *search.DySeaPager, err error) {
	pager = &search.DySeaPager{
		Page: common.Page{
			Num:  lp.Pn,
			Size: lp.Ps,
		},
	}
	w := map[string]interface{}{
		"is_deleted": common.NotDeleted,
	}
	query := s.dao.DB.Model(&search.DySeach{})
	if lp.Word != "" {
		query = query.Where("word like ?", "%"+lp.Word+"%")
	}
	if err = query.Where(w).Count(&pager.Page.Total).Error; err != nil {
		log.Error("DySeachList count error(%v)", err)
		return
	}
	DySeas := make([]*search.DySeach, 0)
	if err = query.Where(w).Order("`position` ASC").Offset((lp.Pn - 1) * lp.Ps).Limit(lp.Ps).Find(&DySeas).Error; err != nil {
		log.Error("DySeachList Find error(%v)", err)
		return
	}
	pager.Item = DySeas
	return
}

func (s *Service) validate(c context.Context, id int64, word string, position int64) (err error) {
	var (
		reply *filterGRPC.FilterReply
	)
	//验证顺位是否重复
	if position != 0 {
		if err = s.dao.DySeachValidat(position, id, ""); err != nil {
			return
		}
	}
	//验证敏感词
	if word != "" {
		//验证热词是否重复
		if err = s.dao.DySeachValidat(0, id, word); err != nil {
			return
		}
		arg := &filterGRPC.FilterReq{
			Area:    common.DySearFilArea,
			Message: word,
		}
		if reply, err = s.dao.FilterGRPC.Filter(c, arg); err != nil {
			return
		}
		if reply.Level >= common.DySearFilLevel {
			err = fmt.Errorf("包含敏感词，请检查后发布")
			return
		}
	}
	return
}

// AddDySeach add channel DySeach
func (s *Service) AddDySeach(c context.Context, param *search.DySeachAP) (err error) {
	if err = s.validate(c, 0, param.Word, param.Position); err != nil {
		return
	}
	if err = s.dao.DySeachAdd(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogDynSear, param.Uname, param.Uid, param.ID, common.ActionAdd, param); err != nil {
		log.Error("AddDySeach AddLog error(%v)", err)
		return
	}
	return
}

// UpdateDySeach update channel DySeach
func (s *Service) UpdateDySeach(c context.Context, param *search.DySeachUP, name string, uid int64) (err error) {
	if err = s.validate(c, param.ID, "", param.Position); err != nil {
		return
	}
	if err = s.dao.DySeachUpdate(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogDynSear, name, uid, param.ID, common.ActionUpdate, param); err != nil {
		log.Error("UpdateDySeach AddLog error(%v)", err)
		return
	}
	return
}

// DeleteDySeach delete channel DySeach
func (s *Service) DeleteDySeach(c context.Context, id int64, name string, uid int64) (err error) {
	if err = s.dao.DySeachDelete(id); err != nil {
		return
	}
	if err = util.AddLogs(common.LogDynSear, name, uid, id, common.ActionDelete, id); err != nil {
		log.Error("DeleteDySeach AddLog error(%v)", err)
		return
	}
	return
}
