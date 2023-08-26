package feature

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/feature"
)

const _tableBuildLimit = "build_limit"

func (d *Dao) SaveBuildLt(ctx context.Context, attrs *feature.BuildLimit) (int, error) {
	if err := d.db.Table(_tableBuildLimit).Save(attrs).Error; err != nil {
		log.Error("d.db.Save(%s, %+v) error(%+v)", _tableBuildLimit, attrs, err)
		return 0, err
	}
	return attrs.ID, nil
}

func (d *Dao) UpdateBuildLt(ctx context.Context, id int, attrs map[string]interface{}) error {
	if err := d.db.Table(_tableBuildLimit).Where("id = ?", id).Update(attrs).Error; err != nil {
		log.Error("d.db.Update(%+v, %+v) error(%+v)", id, attrs, err)
		return err
	}
	return nil
}

func (d *Dao) GetBuildLtByID(ctx context.Context, id int) (*feature.BuildLimit, error) {
	buildLt := new(feature.BuildLimit)
	db := d.db.Table(_tableBuildLimit).Where("id = ?", id)
	if err := db.First(&buildLt).Error; err != nil {
		log.Error("db.First(%s, %+v) error(%+v)", _tableBuildLimit, id, err)
		return nil, err
	}
	return buildLt, nil
}

func (d *Dao) SearchBuildLt(ctx context.Context, req *feature.BuildListReq, needCnt, needFuzzy bool) ([]*feature.BuildLimit, int, error) {
	var (
		cnt      int
		buildLts []*feature.BuildLimit
	)
	db := d.db.Table(_tableBuildLimit).Where("tree_id = ?", req.TreeID).Where("state <> ?", feature.StateDel)
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
			log.Error("db.Count(%s, %+v) error(%+v)", _tableBuildLimit, req, err)
			return nil, 0, err
		}
	}
	if req.Pn > 0 && req.Ps > 0 {
		offset := (req.Pn - 1) * req.Ps
		db = db.Offset(offset).Limit(req.Ps)
	}
	if err := db.Order("`id` DESC").Find(&buildLts).Error; err != nil {
		log.Error("db.Find(%s, %+v) error(%+v)", _tableBuildLimit, req, err)
		return nil, 0, err
	}
	return buildLts, cnt, nil
}

func (d *Dao) BuildLimitServiceCount(ctx context.Context) (map[int]int, error) {
	rows, err := d.db.Table(_tableBuildLimit).Select("DISTINCT(tree_id), count(1)").Group("tree_id").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res = make(map[int]int)
	for rows.Next() {
		var (
			treeID int
			count  int
		)
		if err = rows.Scan(&treeID, &count); err != nil {
			log.Error("TreeList err %v", err)
			return nil, err
		}
		res[treeID] = count
	}
	if err = rows.Err(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}
