package search

import (
	"context"
	"fmt"

	"go-common/library/log"

	model "go-gateway/app/app-svr/app-feed/admin/model/search"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	ACTION_BLACKLIST_ADD    = "blacklist_add"
	ACTION_BLACKLIST_EDIT   = "blacklist_edit"
	ACTION_BLACKLIST_OPTION = "blacklist_option"
)

func (s *Service) BrandBlacklistAdd(c context.Context, req *model.BrandBlacklistAddReq) (resp *model.BrandBlacklistAddResp, err error) {
	resp = &model.BrandBlacklistAddResp{}
	if resp.BlacklistId, resp.EnabledQuery, err = s.dao.BrandBlacklistAdd(c, req); err != nil {
		log.Errorc(c, "BrandBlacklistAddReq req(%+v) error(%v)", req, err)
		return
	}
	if err1 := util.AddBrandBlacklistLogs(req.Username, req.Uid, resp.BlacklistId, ACTION_BLACKLIST_ADD, nil, req, nil); err1 != nil {
		log.Errorc(c, "AddBrandBlacklistLogs error(%v)", err1)
	}
	return
}

func (s *Service) BrandBlacklistEdit(c context.Context, req *model.BrandBlacklistEditReq) (resp *model.BrandBlacklistEditResp, err error) {
	var oldItem *model.BrandBlacklistItem
	resp = &model.BrandBlacklistEditResp{}
	if resp.EnabledQuery, oldItem, err = s.dao.BrandBlacklistEdit(c, req); err != nil {
		log.Errorc(c, "BrandBlacklistEditReq req(%+v) error(%v)", req, err)
		return
	}
	if err1 := util.AddBrandBlacklistLogs(req.Username, req.Uid, req.BlacklistId, ACTION_BLACKLIST_EDIT, oldItem, req, nil); err1 != nil {
		log.Errorc(c, "AddBrandBlacklistLogs error(%v)", err1)
	}
	return
}

func (s *Service) BrandBlacklistOption(c context.Context, req *model.BrandBlacklistOptionReq) (resp *model.BrandBlacklistOptionResp, err error) {
	var (
		affectedQuery []*model.BrandBlacklistQuery
	)
	resp = &model.BrandBlacklistOptionResp{}
	if resp.EnabledQuery, affectedQuery, err = s.dao.BrandBlacklistOption(c, req); err != nil {
		log.Errorc(c, "BrandBlacklistOptionReq req(%+v) error(%v)", req, err)
		return
	}
	if len(affectedQuery) > 0 {
		action := fmt.Sprintf("%s_%d", ACTION_BLACKLIST_OPTION, req.Option)
		if err1 := util.AddBrandBlacklistLogs(req.Username, req.Uid, req.BlacklistId, action, nil, nil, affectedQuery); err1 != nil {
			log.Errorc(c, "AddBrandBlacklistLogs error(%v)", err1)
		}
	}
	return
}

func (s *Service) BrandBlacklistList(c context.Context, req *model.BrandBlacklistListReq) (resp *model.BrandBlackListListResp, err error) {
	resp = &model.BrandBlackListListResp{}
	resp.Page = &model.Page{Pn: req.Pn, Ps: req.Ps}

	if resp.Page.Total, resp.List, err = s.dao.BrandBlacklistList(c, req); err != nil {
		log.Errorc(c, "BrandBlacklistListReq req(%+v) error(%v)", req, err)
		return
	}
	return
}

func (s *Service) OpenBrandBlacklistList(c context.Context, req *model.BrandBlacklistListReq) (resp *model.BrandBlackListListResp, err error) {
	resp = &model.BrandBlackListListResp{
		Page: &model.Page{
			Pn: req.Pn,
			Ps: req.Ps,
		},
	}

	list := s.BrandBlacklistCache
	sliceStart, sliceEnd := util.PaginateSlice(req.Pn, req.Ps, len(list))
	resp.List = list[sliceStart:sliceEnd]
	resp.Page.Total = len(list)
	return
}
