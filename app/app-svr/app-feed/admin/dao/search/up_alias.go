package search

import (
	model "go-gateway/app/app-svr/app-feed/admin/model/search"
	"time"

	"go-common/library/log"
)

func (d *Dao) AddUpAlias(alias *model.UpAlias) (err error) {
	if err = d.DB.Model(&model.UpAlias{}).Create(alias).Error; err != nil {
		log.Error("AddUpAlias error(%v)", err)
	}
	return err
}

func (d *Dao) EditUpAlias(alias *model.UpAlias) (err error) {
	columns := map[string]interface{}{
		"id":           alias.Id,
		"search_words": alias.SearchWords,
		"stime":        alias.Stime,
		"etime":        alias.Etime,
		"is_forever":   alias.IsForever,
		"applier":      alias.Applier,
	}
	if err = d.DB.Model(&model.UpAlias{}).Where("id = ?", alias.Id).Update(columns).Error; err != nil {
		log.Error("EditUpAlias error(%v)", err)
	}
	return err
}

func (d *Dao) ToggleAlias(id int64, state int32) (err error) {
	columns := map[string]interface{}{
		"state": state,
	}
	if err = d.DB.Model(&model.UpAlias{}).Where("id = ?", id).Update(columns).Error; err != nil {
		log.Error("ToggleAlias error(%v)", err)
	}
	return err
}

func (d *Dao) FindAliasByParam(mid int64, nickname, searchWords, applier string, pn, ps int32) (ret []*model.UpAlias, total int32, err error) {
	ret = make([]*model.UpAlias, 0)
	action := d.DB.Model(&model.UpAlias{})
	if mid > 0 {
		action = action.Where("mid = ?", mid)
	}
	if nickname != "" {
		action = action.Where("nickname = ?", nickname)
	}
	if searchWords != "" {
		action = action.Where("search_words like ?", "%"+searchWords+"%")
	}
	if applier != "" {
		action = action.Where("applier = ?", applier)
	}

	if err = action.Count(&total).Error; err != nil {
		log.Error("FindAliasByParam total error(%v)", err)
		return ret, total, err
	}

	if pn > 0 && ps > 0 {
		action = action.Offset((pn - 1) * ps).Limit(ps)
	}

	if err = action.Order("id desc").Find(&ret).Error; err != nil {
		log.Error("FindAliasByParam find error(%v)", err)
		return ret, total, err
	}

	return ret, total, err
}

func (d *Dao) FindAliasForSync(effectTime int64) (ret []*model.UpAlias, err error) {
	ret = make([]*model.UpAlias, 0)
	formattedTime := time.Unix(effectTime, 0).Format("2006-01-02 15:04:05")
	err = d.DB.Model(&model.UpAlias{}).
		Where("state = 1").
		Where("is_forever = 1 OR (is_forever = 0 AND stime <= ? AND etime >= ?)", formattedTime, formattedTime).
		Find(&ret).Error
	if err != nil {
		log.Error("FindAliasByParam find error(%v)", err)
		return nil, err
	}

	return ret, err
}

func (d *Dao) FindAllAlias() (ret []*model.UpAlias, err error) {
	ret = make([]*model.UpAlias, 0)
	if err = d.DB.Model(&model.UpAlias{}).Order("id desc").Find(&ret).Error; err != nil {
		log.Error("FindAliasByParam find error(%v)", err)
		return nil, err
	}
	return ret, err
}
