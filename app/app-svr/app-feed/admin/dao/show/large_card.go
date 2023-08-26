package show

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

// PopLargeCardAdd add event topic
func (d *Dao) PopLargeCardAdd(ctx context.Context, param *show.PopLargeCardAD) (err error) {
	if err = d.DB.Create(param).Error; err != nil {
		log.Error("dao.show.PopLargeCardAdd error(%v) param(%v)", err, param)
		return
	}
	return
}

// PopLargeCardUpdate update event topic
func (d *Dao) PopLargeCardUpdate(ctx context.Context, param *show.PopLargeCardUP) (err error) {
	up := map[string]interface{}{
		"title":      param.Title,
		"rid":        param.RID,
		"white_list": param.WhiteList,
		"auto":       param.Auto,
	}
	if err = d.DB.Model(&show.PopLargeCard{}).Where("id = ?", param.ID).Update(up).Error; err != nil {
		log.Error("dao.show.PopLargeCardUpdate error(%v) param(%v)", err, param)
		return
	}
	return
}

// PopLargeCardDelete delete event topic
func (d *Dao) PopLargeCardDelete(ctx context.Context, id int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.Deleted,
	}
	if err = d.DB.Model(&show.PopLargeCard{}).Where("id=?", id).Update(up).Error; err != nil {
		log.Error("dao.show.PopLargeCardDelete error(%v) id(%d)", err, id)
		return
	}
	return
}

// PopLargeCardNotDelete not delete event topic
func (d *Dao) PopLargeCardNotDelete(ctx context.Context, id int64) (err error) {
	up := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if err = d.DB.Model(&show.PopLargeCard{}).Where("id=?", id).Update(up).Error; err != nil {
		log.Error("dao.show.PopLargeCardNotDelete error(%v) id(%d)", err, id)
		return
	}
	return
}

// PopLargeCardList search
func (d *Dao) PopLargeCardList(ctx context.Context, id int64, createby string, rid int64, pn, ps int) (res *show.PopLargeCardRes, err error) {
	res = &show.PopLargeCardRes{
		Pager: show.PagerCfg{
			Num:  pn,
			Size: ps,
		},
	}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	if id > 0 {
		w["id"] = id
	}
	if createby != "" {
		w["create_by"] = createby
	}
	if rid > 0 {
		w["rid"] = rid
	}
	if err = d.DB.Model(&show.PopLargeCard{}).Where(w).Count(&res.Pager.Total).Error; err != nil {
		log.Error("dao.PopLargeCard Index count error(%v)", err)
		return
	}
	if res.Pager.Total == 0 {
		return
	}
	if err = d.DB.Model(&show.PopLargeCard{}).Offset((pn - 1) * ps).Limit(ps).Where(w).Order("mtime DESC").Find(&res.Items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			res = nil
			err = nil
		} else {
			log.Error("dao.PopLargeCardList error(%v) id(%d) create_by(%s) rid(%d) pn(%d) ps(%d)", err, id, createby, rid, pn, ps)
		}
	}
	return
}
