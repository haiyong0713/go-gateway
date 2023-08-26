package page

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/admin/component"
	"go-gateway/app/web-svr/activity/admin/model/page"
)

func (s *Service) PageList(c context.Context, req *page.ReqPageList) (*page.ResPageList, error) {
	db := component.GlobalOrm.Model(&page.ActPage{})
	if req.Keyword != "" {
		likeKeyword := "%" + req.Keyword + "%"
		db = db.Where("id = ? or name like ? or author = ?", req.Keyword, likeKeyword, req.Keyword)
	}
	if int64(req.SCTime) > 0 {
		db = db.Where("ctime >= ?", req.SCTime)
	}
	if int64(req.ECTime) > 0 {
		db = db.Where("etime <= ?", req.ECTime)
	}
	if req.Creator != "" {
		db = db.Where("creator = ?", req.Creator)
	}
	if req.ReplyID > 0 {
		db = db.Where("reply_id = ?", req.ReplyID)
	}
	if len(req.States) > 0 {
		db = db.Where("state in (?)", req.States)
	}
	if len(req.Mold) > 0 {
		db = db.Where("mold in (?)", req.Mold)
	}
	if len(req.Plat) > 0 {
		db = db.Where("plat in (?)", req.Plat)
	}
	if len(req.Dept) > 0 {
		db = db.Where("dept in (?)", req.Dept)
	}
	res := &page.ResPageList{
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     []*page.ActPage{},
	}
	if err := db.Count(&res.Count).Error; err != nil {
		return nil, err
	}
	return res, db.Order(fmt.Sprintf("%s desc", req.Order)).Offset(req.Page*req.PageSize - req.PageSize).Limit(req.PageSize).Scan(&res.List).Error
}
