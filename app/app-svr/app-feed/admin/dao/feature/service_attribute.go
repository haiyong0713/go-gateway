package feature

import (
	"context"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/feature"
)

const (
	_tableSvrAttr = "service_attribute"
)

func (d *Dao) SaveSvrAttr(ctx context.Context, attrs *feature.ServiceAttribute) error {
	if err := d.db.Table(_tableSvrAttr).Save(attrs).Error; err != nil {
		log.Error("d.db.Save(%s, %+v) error(%+v)", _tableSvrAttr, attrs, err)
		return err
	}
	return nil
}

func (d *Dao) GetSvrAttrByTreeID(ctx context.Context, treeID int) (*feature.ServiceAttribute, error) {
	svrAttr := new(feature.ServiceAttribute)
	db := d.db.Table(_tableSvrAttr).Where("tree_id = ?", treeID)
	if err := db.First(&svrAttr).Error; err != nil {
		log.Error("db.First(%s, %+v) error(%+v)", _tableSvrAttr, treeID, err)
		return nil, err
	}
	return svrAttr, nil
}

func (d *Dao) GetSvrAttrPlats(ctx context.Context, treeID int) (map[string]struct{}, error) {
	svrAttr, err := d.GetSvrAttrByTreeID(ctx, treeID)
	if err != nil {
		log.Error("d.GetSvrAttrByTreeID(%+v) error(%+v)", treeID, err)
		return nil, err
	}
	plats := make(map[string]struct{}, len(d.plats))
	if svrAttr == nil {
		return plats, nil
	}
	mobiApps := strings.Split(svrAttr.MobiApps, ",")
	for _, mobiApp := range mobiApps {
		plats[mobiApp] = struct{}{}
	}
	return plats, nil
}

func (d *Dao) TreeList(ctx context.Context) (map[int]int, error) {
	rows, err := d.db.Table(_tableSvrAttr).Select("DISTINCT(tree_id), count(1)").Group("tree_id").Rows()
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
