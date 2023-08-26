package frontpage

import (
	"time"

	"github.com/pkg/errors"

	xecode "go-common/library/ecode"
	xtime "go-common/library/time"

	model "go-gateway/app/app-svr/app-feed/admin/model/frontpage"
	"go-gateway/app/app-svr/app-feed/ecode"
)

func (d *Dao) GetConfig(id int64) (res *model.Config, err error) {
	res = &model.Config{}
	if err = d.ORMManager.Model(model.Config{}).
		Where("id = ?", id).
		First(res).Error; err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			res = nil
			err = ecode.FrontPageConfigNotFound
		}
		return
	}
	res.Position = 1

	return
}

func (d *Dao) GetBaseDefaultConfig() (res *model.Config, err error) {
	res = &model.Config{}
	// 无生效中默认图或在默认区，直接返回兜底图
	if err = d.ORMManager.Model(model.Config{}).Where("id = ?", model.DefaultConfigID).Find(res).Error; err != nil {
		return nil, err
	}
	res.Position = 1
	res.LocPolicyGroupID = 0
	return
}

func (d *Dao) GetOnlineConfigs(resourceID int64) (res []*model.Config, err error) {
	now := time.Now()
	res = make([]*model.Config, 0)
	if err = d.ORMManager.Model(model.Config{}).
		Where("resource_id = ? AND stime <= ? AND etime >= ? AND state = ? AND id != ?", resourceID, now, now, 0, model.DefaultConfigID).
		Order("stime").Find(&res).Error; err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			res = nil
			err = nil
		}
		return
	}
	for i := range res {
		res[i].Position = 1
	}

	return
}

func (d *Dao) GetHiddenConfigs(resourceID int64, pn int64, ps int64) (res []*model.Config, total int64, err error) {
	if pn == 0 {
		pn = 1
	}
	if ps == 0 {
		ps = 5
	}
	now := time.Now()
	query := d.ORMManager.Model(model.Config{}).
		Where("resource_id = ? AND stime > ? AND state = ?", resourceID, now, 0)
	if err = query.Count(&total).Error; err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			err = nil
		}
		return
	}
	if err = query.Order("stime").Offset((pn - 1) * ps).Limit(ps).Find(&res).Error; err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			err = nil
		}
		return
	}

	return
}

func (d *Dao) GetConfigByTimeAndLoc(resourceID int64, stime time.Time, etime time.Time, locationPolicyGroupID int64) (res *model.Config, err error) {
	if resourceID < 0 {
		return nil, xecode.RequestErr
	}
	res = &model.Config{}
	if err = d.ORMManager.Model(model.Config{}).
		Where("resource_id = ? AND stime <= ? AND etime > ? AND state = ? AND loc_policy_group_id = ? AND id != ?", resourceID, etime, stime, 0, locationPolicyGroupID, model.DefaultConfigID).
		Order("stime").
		First(&res).Error; err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			err = nil
		} else {
			err = errors.Wrapf(err, "Dao: GetConfigBySTimeAndLoc")
		}
		return
	}
	res.Position = 1

	return
}

func (d *Dao) AddConfig(toAddConfig *model.Config) (res *model.Config, err error) {
	if toAddConfig == nil {
		return nil, xecode.Error(xecode.RequestErr, "AddConfig参数缺失")
	}
	now := time.Now()
	toAddConfig.CTime = xtime.Time(now.Unix())
	toAddConfig.MTime = xtime.Time(now.Unix())
	if err = d.ORMManager.Model(model.Config{}).Create(toAddConfig).Error; err != nil {
		return nil, errors.Wrapf(err, "Dao: AddConfig")
	}
	res = toAddConfig

	return
}

func (d *Dao) UpdateConfig(configID int64, updateMap map[string]interface{}) (err error) {
	if configID == 0 {
		return xecode.RequestErr
	}
	if len(updateMap) == 0 {
		return
	}
	now := time.Now()
	updateMap["mtime"] = now
	err = d.ORMManager.Model(model.Config{}).Where("id = ?", configID).UpdateColumns(updateMap).Error

	return
}

func (d *Dao) GetConfigHistories(resourceID int64, pn int64, ps int64) (res []*model.Config, total int64, err error) {
	if pn == 0 {
		pn = 1
	}
	if ps == 0 {
		ps = 20
	}

	now := time.Now()
	query := d.ORMManager.Model(model.Config{}).Where("resource_id = ? AND stime <= ? AND state != ?", resourceID, now, -1)
	if err = query.Count(&total).Error; err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			err = nil
		}
		return
	}
	if err = query.Order("stime").Offset((pn - 1) * ps).Limit(ps).Find(&res).Error; err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			err = nil
		}
		return
	}

	return
}
