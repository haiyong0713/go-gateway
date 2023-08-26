package dao

import (
	"context"

	"go-gateway/app/web-svr/web/job/internal/model"

	"github.com/pkg/errors"
)

func (d *dao) RegionList(ctx context.Context) ([]*model.Region, error) {
	const _regionListSQL = "SELECT r.id,r.rid,r.reid,r.name,r.logo,r.plat,r.area,r.uri,l.name FROM region_copy AS r,language AS l WHERE r.state=1 AND r.lang_id=l.id ORDER BY r.rank DESC"
	rows, err := d.showDB.Query(ctx, _regionListSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*model.Region
	for rows.Next() {
		r := &model.Region{}
		if err := rows.Scan(&r.ID, &r.Rid, &r.Reid, &r.Name, &r.Logo, &r.Plat, &r.Area, &r.URI, &r.Language); err != nil {
			return nil, errors.Wrap(err, _regionListSQL)
		}
		res = append(res, r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, _regionListSQL)
	}
	return res, nil
}

func (d *dao) RegionConfig(ctx context.Context) (map[int64][]*model.RegionConfig, error) {
	const _regionConfigSQL = "SELECT id,rid,is_rank FROM region_rank_config"
	rows, err := d.showDB.Query(ctx, _regionConfigSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int64][]*model.RegionConfig{}
	for rows.Next() {
		r := &model.RegionConfig{}
		if err = rows.Scan(&r.ID, &r.Rid, &r.ScenesID); err != nil {
			return nil, errors.Wrap(err, _regionConfigSQL)
		}
		r.ConfigChange()
		res[r.Rid] = append(res[r.Rid], r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, _regionConfigSQL)
	}
	return res, nil
}
