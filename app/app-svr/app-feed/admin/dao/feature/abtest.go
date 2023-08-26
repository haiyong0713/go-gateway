package feature

import (
	"context"
	"go-common/library/log"

	featuremdl "go-gateway/app/app-svr/app-feed/admin/model/feature"
)

const (
	_tableABTest = `abtest`
)

func (d *Dao) SearchABTest(c context.Context, req *featuremdl.ABTestReq, needCnt bool) ([]*featuremdl.ABTest, int, error) {
	var (
		cnt     int
		abtests []*featuremdl.ABTest
	)
	db := d.db.Table(_tableABTest).Where("tree_id = ?", req.TreeID)
	if req.KeyName != "" {
		db = db.Where("key_name LIKE ?", "%"+req.KeyName+"%")
	}
	if req.Creator != "" {
		db = db.Where("creator = ?", req.Creator)
	}
	if needCnt {
		if err := db.Count(&cnt).Error; err != nil {
			log.Error("db.Count(%s, %+v) error(%+v)", _tableABTest, req, err)
			return nil, 0, err
		}
	}
	if req.Pn > 0 && req.Ps > 0 {
		offset := (req.Pn - 1) * req.Ps
		db = db.Offset(offset).Limit(req.Ps)
	}
	if err := db.Order("`id` DESC").Find(&abtests).Error; err != nil {
		log.Error("db.Find(%s, %+v) error(%+v)", _tableABTest, req, err)
		return nil, 0, err
	}
	return abtests, cnt, nil
}

func (d *Dao) SaveABTest(c context.Context, attrs *featuremdl.ABTest) (int, error) {
	if err := d.db.Table(_tableABTest).Save(attrs).Error; err != nil {
		log.Error("d.db.Save(%s, %+v) error(%+v)", _tableABTest, attrs, err)
		return 0, err
	}
	return attrs.ID, nil
}

func (d *Dao) GetABTestByID(c context.Context, id int) (*featuremdl.ABTest, error) {
	abtest := new(featuremdl.ABTest)
	db := d.db.Table(_tableABTest).Where("id = ?", id)
	if err := db.First(&abtest).Error; err != nil {
		log.Error("db.First(%s, %+v) error(%+v)", _tableABTest, id, err)
		return nil, err
	}
	return abtest, nil
}

func (d *Dao) UpdateABTest(c context.Context, id int, attrs map[string]interface{}) error {
	if err := d.db.Table(_tableABTest).Where("id = ?", id).Update(attrs).Error; err != nil {
		log.Error("d.db.Update(%+v, %+v) error(%+v)", id, attrs, err)
		return err
	}
	return nil
}

func (d *Dao) ABTestServiceCount(ctx context.Context) (map[int]int, error) {
	rows, err := d.db.Table(_tableABTest).Select("DISTINCT(tree_id), count(1)").Group("tree_id").Rows()
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
