package show

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

func (d *Dao) ArticleCardAdd(ctx context.Context, param *show.ArticleCardAD) (id int64, err error) {
	if err = d.DB.Create(param).Error; err != nil {
		log.Error("dao.show.ArticleCardAdd error(%v) param(%v)", err, param)
		return
	}
	return
}

func (d *Dao) ArticleCardUpdate(ctx context.Context, param *show.ArticleCardUP) (err error) {
	up := map[string]interface{}{
		"cover":      param.Cover,
		"article_id": param.ArticleID,
	}
	if err = d.DB.Model(&show.ArticleCardUP{}).Where("id = ?", param.ID).Update(up).Error; err != nil {
		log.Error("dao.show.ArticleCardUpdate error(%v) param(%v)", err, param)
		return
	}
	return
}

func (d *Dao) ArticleCardOperate(ctx context.Context, id int64, state int) (err error) {
	up := map[string]interface{}{
		"state": state,
	}
	if err = d.DB.Model(&show.ArticleCardUP{}).Where("id = ?", id).Update(up).Error; err != nil {
		log.Error("dao.show.ArticleCardOperate error(%v) id(%d)", err, id)
		return
	}
	return
}

func (d *Dao) ArticleCardList(ctx context.Context, id int64, state int, createby string, pn, ps int) (res *show.ArticleCardRes, err error) {
	const (
		// -1代表全部 3代表删除
		_stateAll = -1
		_stateDel = 3
	)
	res = &show.ArticleCardRes{
		Pager: show.PagerCfg{
			Num:  pn,
			Size: ps,
		},
	}
	w := map[string]interface{}{}
	if id > 0 {
		w["id"] = id
	}
	if createby != "" {
		w["create_by"] = createby
	}
	if state != _stateAll {
		w["state"] = state
	}
	if err = d.DB.Model(&show.ArticleCard{}).Where(w).Where("state <> ?", _stateDel).Count(&res.Pager.Total).Error; err != nil {
		log.Error("dao.ArticleCardList Index count error(%v)", err)
		return
	}
	if res.Pager.Total == 0 {
		return
	}
	if err = d.DB.Model(&show.ArticleCard{}).Offset((pn-1)*ps).Limit(ps).Where(w).Where("state <> ?", _stateDel).Order("mtime DESC").Find(&res.Items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			res = nil
			err = nil
		} else {
			log.Error("dao.ArticleCardList error(%v) id(%d) create_by(%s) pn(%d) ps(%d)", err, id, createby, pn, ps)
		}
	}
	return
}
