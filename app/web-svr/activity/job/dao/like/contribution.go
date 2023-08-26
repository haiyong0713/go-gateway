package like

import (
	"context"
	"database/sql"
	xsql "database/sql"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/job/model/like"
)

const (
	_selContributionSQL    = "SELECT id,mid,up_archives,likes,views,light_videos,bcuts,sn_up_archives,sn_likes FROM act_contributions WHERE mid=?"
	_updateContributionSQL = "UPDATE act_contributions SET up_archives=?,likes=?,views=?,light_videos=?,bcuts=?,sn_up_archives=?,sn_likes=?  WHERE mid=?"
	_updateBcutLikesSQL    = "UPDATE act_contributions SET bcut_likes=? WHERE mid=?"
	_AddContributionSQL    = "INSERT INTO act_contributions(mid,up_archives,likes,views,light_videos,bcuts,sn_up_archives,sn_likes) VALUES(?,?,?,?,?,?,?,?)"
	_selContriAwardSQL     = "SELECT id,award_type,up_archives,`likes`,views,light_videos,bcuts,sn_up_archives,sn_likes FROM act_contributions_awards WHERE is_deleted=0 ORDER BY award_type ASC"
	_upContriAward         = "UPDATE act_contributions_awards SET split_people = CASE %s END WHERE id IN (%s)"
	_upTotalViewsAwardSQL  = "UPDATE act_contributions_awards SET current_views=? WHERE id=?"
)

// RawContribution.
func (d *Dao) RawContribution(c context.Context, mid int64) (data *like.ActContributions, err error) {
	data = new(like.ActContributions)
	row := d.db.QueryRow(c, _selContributionSQL, mid)
	if err = row.Scan(&data.ID, &data.Mid, &data.UpArchives, &data.Likes, &data.Views, &data.LightVideos, &data.Bcuts, &data.SnUpArchives, &data.SnLikes); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("RawContribution QueryRow mid(%d) error(%v)", mid, err)
		}
	}
	return
}

// SelContriAward.
func (d *Dao) SelContriAward(c context.Context) (res []*like.ContriAward, err error) {
	rows, err := d.db.Query(c, _selContriAwardSQL)
	if err != nil {
		log.Error("SelContriAward.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		row := new(like.ContriAward)
		if err = rows.Scan(&row.ID, &row.AwardType, &row.UpArchives, &row.Likes, &row.Views, &row.LightVideos, &row.Bcuts, &row.SnUpArchives, &row.SnLikes); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		res = append(res, row)
	}
	err = rows.Err()
	return
}

// UpUserContribution .
func (d *Dao) UpUserContribution(c context.Context, mid int64, data *like.ContributionUser) (err error) {
	if _, err = d.db.Exec(c, _updateContributionSQL, data.UpArchives, data.Likes, data.Views, data.LightVideos, data.Bcuts, data.SnUpArchives, data.SnLikes, mid); err != nil {
		log.Error("UpUserContribution:d.db.Exec(%s,%+v,%d) error(%v)", _updateContributionSQL, data, mid, err)
	}
	return
}

// UpUserBcutLikes .
func (d *Dao) UpUserBcutLikes(c context.Context, mid int64, bcutLikes int32) (err error) {
	if _, err = d.db.Exec(c, _updateBcutLikesSQL, bcutLikes, mid); err != nil {
		log.Error("UpUserBcutLikes:d.db.Exec(%s,%d,%d) error(%v)", _updateBcutLikesSQL, bcutLikes, mid, err)
	}
	return
}

// AddUserContribution .
func (d *Dao) AddUserContribution(ctx context.Context, mid int64, data *like.ContributionUser) (int64, error) {
	res, err := d.db.Exec(ctx, _AddContributionSQL, mid, data.UpArchives, data.Likes, data.Views, data.LightVideos, data.Bcuts, data.SnUpArchives, data.SnLikes)
	if err != nil {
		log.Error("d.AddUserContribution(%d,%d,%d,%d,%d,%d) error(%+v)", mid, data.UpArchives, data.Likes, data.Views, data.LightVideos, data.Bcuts, err)
		return 0, err
	}
	return res.LastInsertId()
}

// UpContriAwardPeople .
func (d *Dao) UpContriAwardPeople(c context.Context, SplitPepleCalc map[int64]int64) (affected int64, err error) {
	var (
		caseStr string
		ids     []int64
		res     xsql.Result
	)
	if len(SplitPepleCalc) == 0 {
		return
	}
	for id, peoples := range SplitPepleCalc {
		caseStr = fmt.Sprintf("%s WHEN id = %d THEN %d", caseStr, id, peoples)
		ids = append(ids, id)
	}
	if res, err = d.db.Exec(c, fmt.Sprintf(_upContriAward, caseStr, xstr.JoinInts(ids))); err != nil {
		log.Error("UpContriAwardPeople:d.db.Exec() error(%v)", _upContriAward, err)
		return
	}
	return res.RowsAffected()
}

// UpTotalVeiwAward .
func (d *Dao) UpTotalVeiwAward(c context.Context, id, currentTotal int64) (err error) {
	if _, err = d.db.Exec(c, _upTotalViewsAwardSQL, currentTotal, id); err != nil {
		log.Error("UpTotalVeiwAward:d.db.Exec(%d,%d) error(%v)", _upTotalViewsAwardSQL, currentTotal, id, err)
	}
	return
}
