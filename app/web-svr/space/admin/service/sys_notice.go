package service

import (
	"context"
	"fmt"

	"go-common/library/log"

	"go-gateway/app/web-svr/space/admin/model"
	"go-gateway/app/web-svr/space/admin/util"

	"github.com/jinzhu/gorm"
)

// SysNoticeAdd add system notice
func (s *Service) SysNotice(c context.Context, param *model.SysNoticeList) (pager *model.SysNoticeInfoPager, err error) {
	pager = &model.SysNoticeInfoPager{
		Page: model.Page{
			Num:  param.Pn,
			Size: param.Ps,
		},
	}
	pager.Page.Total, pager.Item, err = s.dao.SysNotice(param)
	if err != nil {
		return
	}
	return
}

// SysNoticeAdd add system notice
func (s *Service) SysNoticeAdd(c context.Context, param *model.SysNoticeAdd) (err error) {
	if err = s.dao.SysNoticeAdd(param); err != nil {
		return
	}
	return
}

// SysNoticeUp update system notice
func (s *Service) SysNoticeUp(c context.Context, param *model.SysNoticeUp) (err error) {
	var uids []int64
	if uids, err = s.dao.SysNoticeUidList(param.ID); err != nil {
		return
	}
	if err = s.validate(param.ID, uids, param.Scopes); err != nil {
		return
	}
	if err = s.dao.SysNoticeUpdate(param); err != nil {
		return
	}
	return
}

// SysNoticeOpt opt system notice
func (s *Service) SysNoticeOpt(c context.Context, param *model.SysNoticeOpt) (err error) {
	if err = s.dao.SysNoticeOpt(param); err != nil {
		return
	}
	return
}

// SysNoticeUidAdd opt system notice
func (s *Service) SysNoticeUidAdd(c context.Context, param *model.SysNotUidAddDel) (err error) {
	var noticeInfo *model.SysNoticeInfo
	if noticeInfo, err = s.dao.SysNoticeInfo(param.ID); err != nil {
		return
	}
	if err = s.validate(0, param.UIDs, noticeInfo.Scopes); err != nil {
		return
	}
	if err = s.dao.SysNoticeUidAdd(param); err != nil {
		return
	}
	return
}

// SysNoticeUidAdd system notice uid list
func (s *Service) SysNoticeUid(c context.Context, param *model.SysNoticeUidParam) (pager *model.SysNoticePager, err error) {
	pager = &model.SysNoticePager{
		Page: model.Page{
			Num:  param.Pn,
			Size: param.Ps,
		},
	}
	value := make([]*model.SysNoticeUid, 0)
	w := map[string]interface{}{
		"system_notice_id": param.ID,
		"is_deleted":       model.NotDeleted,
	}
	query := s.dao.DB.Model(&model.SysNoticeUid{}).Where(w)
	if err = query.Where(w).Count(&pager.Page.Total).Error; err != nil {
		log.Error("SysNoticeUid count error(%v)", err)
		return
	}
	if pager.Page.Total == 0 {
		return
	}
	if err = query.Offset((param.Pn - 1) * param.Ps).Order("uid ASC").Limit(param.Ps).Find(&value).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			return
		}
		log.Error("dao.SysNoticeUidFind error(%v)", err)
		return
	}
	pager.Item = value
	return
}

// SysNoticeUidDel del system notice uid
func (s *Service) SysNoticeUidDel(c context.Context, param *model.SysNotUidAddDel) (err error) {
	return s.dao.SysNoticeUidDel(param)
}

// validate validates if any notice with scopes is enabled for users,
// except specific noticeId
func (s *Service) validate(noticeId int64, uids []int64, scopes string) (err error) {
	if scopes == "" {
		log.Error("validate error, empty scopes")
		return fmt.Errorf("empty scopes")
	}

	scopeList := util.SplitInt(scopes)
	for _, v := range uids {
		if err = s.dao.SysNoticeUidFirst(v, noticeId, scopeList); err != nil {
			return
		}
	}
	return
}
