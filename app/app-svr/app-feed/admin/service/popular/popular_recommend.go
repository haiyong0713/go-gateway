package popular

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// PopRecommendList channel PopRecommend list
func (s *Service) PopRecommendList(lp *show.PopRecommendLP) (pager *show.PopRecommendPager, err error) {
	pager = &show.PopRecommendPager{
		Page: common.Page{
			Num:  lp.Pn,
			Size: lp.Ps,
		},
	}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	query := s.showDao.DB.Model(&show.PopRecommend{})
	if lp.AID > 0 {
		w["card_value"] = lp.AID
	}
	if lp.Person != "" {
		query = query.Where("person like ?", "%"+lp.Person+"%")
	}
	if err = query.Where(w).Count(&pager.Page.Total).Error; err != nil {
		log.Error("popularSvc.PopRecommendList count error(%v)", err)
		return
	}
	PopRecommends := make([]*show.PopRecommend, 0)
	if err = query.Where(w).Order("`id` DESC").Offset((lp.Pn - 1) * lp.Ps).Limit(lp.Ps).Find(&PopRecommends).Error; err != nil {
		log.Error("popularSvc.PopRecommendList Find error(%v)", err)
		return
	}
	pager.Item = PopRecommends
	return
}

// IsDup is add duplicate
func (s *Service) IsDup(value string, id int64) (err error) {
	var card *show.PopRecommend
	if card, err = s.showDao.PopRFindByID(value); err != nil {
		return
	}
	if id == 0 {
		if card != nil && card.ID != 0 {
			err = fmt.Errorf(fmt.Sprintf("id为%s的稿件已经添加过", value))
			return
		}
	} else {
		if id != card.ID {
			err = fmt.Errorf(fmt.Sprintf("id为%s的稿件已经添加过", value))
			return
		}
	}
	return
}

// AddPopRecommend add channel PopRecommend
func (s *Service) AddPopRecommend(c context.Context, param *show.PopRecommendAP, name string, uid int64) (err error) {
	if err = s.IsDup(param.CardValue, 0); err != nil {
		return
	}
	if err = s.showDao.PopRecommendAdd(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogPopRcmd, name, uid, 0, common.ActionAdd, param); err != nil {
		log.Error("popularSvc.AddPopRecommend AddLog error(%v)", err)
		return
	}
	return
}

// UpdatePopRecommend update channel PopRecommend
func (s *Service) UpdatePopRecommend(c context.Context, param *show.PopRecommendUP, name string, uid int64) (err error) {
	if err = s.IsDup(param.CardValue, param.ID); err != nil {
		return
	}
	if err = s.showDao.PopRecommendUpdate(param); err != nil {
		return
	}
	if err = util.AddLogs(common.LogPopRcmd, name, uid, 0, common.ActionUpdate, param); err != nil {
		log.Error("popularSvc.UpdatePopRecommend AddLog error(%v)", err)
		return
	}
	return
}

// DeletePopRecommend delete channel PopRecommend
func (s *Service) DeletePopRecommend(id int64, name string, uid int64) (err error) {
	if err = s.showDao.PopRecommendDelete(id); err != nil {
		return
	}
	if err = util.AddLogs(common.LogPopRcmd, name, uid, id, common.ActionDelete, id); err != nil {
		log.Error("popularSvc.DeletePopRecommend AddLog error(%v)", err)
		return
	}
	return
}
