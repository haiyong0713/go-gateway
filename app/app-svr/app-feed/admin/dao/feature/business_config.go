package feature

import (
	"context"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/feature"
)

const _businessConfig = "business_config"

func (d *Dao) SearchBusinessConfig(_ context.Context, req *feature.BusinessConfigListReq, needCnt, needFuzzy bool) ([]*feature.BusinessConfig, int, error) {
	var (
		cnt             int
		businessConfigs []*feature.BusinessConfig
	)
	db := d.db.Table(_businessConfig).Where("tree_id = ?", req.TreeID).Where("state <> ?", feature.StateDel)
	if req.KeyName != "" {
		if needFuzzy {
			db = db.Where("key_name LIKE ?", "%"+req.KeyName+"%")
		} else {
			db = db.Where("key_name = ?", req.KeyName)
		}
	}
	if req.Creator != "" {
		db = db.Where("creator = ?", req.Creator)
	}
	if needCnt {
		if err := db.Count(&cnt).Error; err != nil {
			log.Error("db.Count(%s, %+v) error(%+v)", _businessConfig, req, err)
			return nil, 0, err
		}
	}
	if req.Pn > 0 && req.Ps > 0 {
		offset := (req.Pn - 1) * req.Ps
		db = db.Offset(offset).Limit(req.Ps)
	}
	if err := db.Order("`id` DESC").Find(&businessConfigs).Error; err != nil {
		log.Error("db.Find(%s, %+v) error(%+v)", _businessConfig, req, err)
		return nil, 0, err
	}
	return businessConfigs, cnt, nil
}

func (d *Dao) BusinessConfigByID(ctx context.Context, id int) (*feature.BusinessConfig, error) {
	businessConfig := new(feature.BusinessConfig)
	db := d.db.Table(_businessConfig).Where("id = ?", id)
	if err := db.First(&businessConfig).Error; err != nil {
		log.Error("db.First(%s, %+v) error(%+v)", _businessConfig, id, err)
		return nil, err
	}
	return businessConfig, nil
}

func (d *Dao) BusinessConfigSave(ctx context.Context, attrs *feature.BusinessConfig) (int, error) {
	if err := d.db.Table(_businessConfig).Save(attrs).Error; err != nil {
		log.Error("d.db.Save(%s, %+v) error(%+v)", _businessConfig, attrs, err)
		return 0, err
	}
	return attrs.ID, nil
}

func (d *Dao) UpdateBusinessConfig(ctx context.Context, id int, attrs map[string]interface{}) error {
	if err := d.db.Table(_businessConfig).Where("id = ?", id).Update(attrs).Error; err != nil {
		log.Error("d.db.Update(%+v, %+v) error(%+v)", id, attrs, err)
		return err
	}
	return nil
}
