package feature

import (
	"context"

	"go-common/library/log"

	featureMdl "go-gateway/app/app-svr/app-feed/admin/model/feature"
)

const (
	_tableSwitchTV = "switch_tv"
)

func (d *Dao) SwitchTV(ctx context.Context, req *featureMdl.SwitchTvListReq, needCnt bool) ([]*featureMdl.SwitchTV, int, error) {
	var (
		cnt       int
		tvSwitchs []*featureMdl.SwitchTV
	)
	db := d.db.Table(_tableSwitchTV).Where("deleted = 0")
	if needCnt {
		if err := db.Count(&cnt).Error; err != nil {
			log.Error("db.Count(%s, %+v) error(%+v)", _tableSwitchTV, req, err)
			return nil, 0, err
		}
	}
	if req.Pn > 0 && req.Ps > 0 {
		offset := (req.Pn - 1) * req.Ps
		db = db.Offset(offset).Limit(req.Ps)
	}
	if err := db.Order("`id` DESC").Find(&tvSwitchs).Error; err != nil {
		log.Error("db.Find(%s, %+v) error(%+v)", _tableSwitchTV, req, err)
		return nil, 0, err
	}
	return tvSwitchs, cnt, nil
}

func (d *Dao) SwitchTVAll(ctx context.Context) ([]*featureMdl.SwitchTV, error) {
	var tvSwitchs []*featureMdl.SwitchTV
	db := d.db.Table(_tableSwitchTV).Where("deleted = 0")
	if err := db.Order("`id` DESC").Find(&tvSwitchs).Error; err != nil {
		log.Error("db.Find(%s) error(%+v)", _tableSwitchTV, err)
		return nil, err
	}
	return tvSwitchs, nil
}

func (d *Dao) SaveSwitchTv(c context.Context, attrs *featureMdl.SwitchTV) (int, error) {
	if err := d.db.Table(_tableSwitchTV).Save(attrs).Error; err != nil {
		log.Error("d.db.Save(%s, %+v) error(%+v)", _tableSwitchTV, attrs, err)
		return 0, err
	}
	return attrs.ID, nil
}

func (d *Dao) UpdateSwitchTv(c context.Context, id int, attrs map[string]interface{}) error {
	if err := d.db.Table(_tableSwitchTV).Where("id = ?", id).Update(attrs).Error; err != nil {
		log.Error("d.db.UpdateSwitchTv(%s, %d, %+v) error(%+v)", _tableSwitchTV, id, attrs, err)
		return err
	}
	return nil
}
