package like

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_selContriAwardSQL  = "SELECT id,award_type,current_views,up_archives,`likes`,views,light_videos,bcuts,split_people,split_money,sn_up_archives,sn_likes FROM act_contributions_awards WHERE is_deleted=0 ORDER BY award_type ASC"
	_contributionMisSQL = "SELECT id,mid FROM act_contributions WHERE id > ? ORDER BY id LIMIT 1000"
	_upContributionSQL  = "UPDATE act_contributions SET have_money=? WHERE mid=?"
)

func (d *Dao) ContributionMids(c context.Context, maxID int64) (res []int64, max int64, err error) {
	res = make([]int64, 0, 100)
	rows, err := d.db.Query(c, _contributionMisSQL, maxID)
	if err != nil {
		log.Error("ContributionMids.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, uid int64
		if err = rows.Scan(&id, &uid); err != nil {
			log.Error("ContributionMids row.Scan() error(%v)", err)
			return
		}
		res = append(res, uid)
		if id > max {
			max = id
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("ContributionMids rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) UpUserMoney(ctx context.Context, money float64, mid int64) (int64, error) {
	row, err := d.db.Exec(ctx, _upContributionSQL, money, mid)
	if err != nil {
		return 0, err
	}
	return row.RowsAffected()
}

// RawContriAwards .
func (d *Dao) RawContriAwards(c context.Context) (list []*like.ContriAwards, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _selContriAwardSQL)
	if err != nil {
		log.Error("RawContriAwards:d.db.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		row := new(like.ContriAwards)
		if err = rows.Scan(&row.ID, &row.AwardType, &row.CurrentViews, &row.UpArchives, &row.Likes, &row.Views, &row.LightVideos, &row.Bcuts, &row.SplitPeople, &row.SplitMoney, &row.SnUpArchives, &row.SnLikes); err != nil {
			log.Error("RawContriAwards:rows.Scan() error(%v)", err)
			return
		}
		list = append(list, row)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawContriAwards:rows.Err() error(%v)", err)
	}
	return
}
