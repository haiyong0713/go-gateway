package show

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/jinzhu/gorm"
)

const _stateDel = 3

func (d *Dao) PopLiveCardAdd(ctx context.Context, param *show.PopLiveCardAD) (id int64, err error) {
	if err = d.DB.Create(param).Error; err != nil {
		log.Error("dao.show.PopLiveCardAdd error(%v) param(%v)", err, param)
		return
	}
	id = param.ID
	return
}

func (d *Dao) PopLiveCardUpdate(ctx context.Context, param *show.PopLiveCardUP) (err error) {
	up := map[string]interface{}{
		"cover": param.Cover,
		"rid":   param.RID,
	}
	if err = d.DB.Model(&show.PopLiveCard{}).Where("id = ?", param.ID).Update(up).Error; err != nil {
		log.Error("dao.show.PopLiveCardUpdate error(%v) param(%v)", err, param)
		return
	}
	return
}

func (d *Dao) PopLargeCardOperate(ctx context.Context, id int64, state int) (err error) {
	up := map[string]interface{}{
		"state": state,
	}
	if err = d.DB.Model(&show.PopLiveCard{}).Where("id=?", id).Update(up).Error; err != nil {
		log.Error("dao.show.PopLargeCardOperate error(%v) id(%d)", err, id)
		return
	}
	return
}

// PopLiveCardList search
func (d *Dao) PopLiveCardList(ctx context.Context, id int64, state int, createby string, pn, ps int) (res *show.PopLiveCardRes, err error) {
	res = &show.PopLiveCardRes{
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
	if state != -1 { // -1代表全部 3代表删除
		w["state"] = state
	}
	if err = d.DB.Model(&show.PopLiveCard{}).Where(w).Where("state <> ?", _stateDel).Count(&res.Pager.Total).Error; err != nil {
		log.Error("dao.PopLiveCardList Index count error(%v)", err)
		return
	}
	if res.Pager.Total == 0 {
		return
	}
	if err = d.DB.Model(&show.PopLiveCard{}).Offset((pn-1)*ps).Limit(ps).Where(w).Where("state <> ?", _stateDel).Order("mtime DESC").Find(&res.Items).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			res = nil
			err = nil
		} else {
			log.Error("dao.PopLiveCardList error(%v) id(%d) create_by(%s) pn(%d) ps(%d)", err, id, createby, pn, ps)
		}
	}
	return
}
