package selected

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
)

func (d *Dao) GetAllEntrances(ctx context.Context) (res []*selected.PopTopEntrance, err error) {
	err = d.DB.Table(new(selected.PopTopEntrance).TableName()).Order("rank, mtime ASC").Find(&res).Error
	return
}

func (d *Dao) UpdateEntrancesRank(ctx context.Context, res []*selected.PopTopEntrance) (err error) {
	if len(res) == 0 {
		return
	}
	tx := d.DB.Begin().Table(new(selected.PopTopEntrance).TableName())
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()
	for _, r := range res {
		err = tx.Where("id = ?", r.ID).Update("rank", r.Rank).Error
		if err != nil {
			log.Error("UpdateEntrancesRank res(%+v) error(%v)", res, err)
			return
		}
	}
	return
}
