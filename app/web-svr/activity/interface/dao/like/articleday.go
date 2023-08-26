package like

import (
	"context"
	"database/sql"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_selArticleDaySQL = "SELECT id,mid,publish,publish_count,finish_task,finish_time,ctime,status FROM act_artile_day2 WHERE mid=?"
	_addArticleDaySQL = "INSERT IGNORE INTO act_artile_day2(mid) VALUES(?)"
	_selAwardSQL      = "SELECT id,activity_uid,condition_min,`condition_max`,split_people,split_money FROM act_artile_award WHERE is_deleted=0 ORDER BY activity_uid ASC,`condition_min` ASC"
)

// RawArticleDay.
func (d *Dao) RawArticleDay(c context.Context, mid int64) (data *like.ArticleDay, err error) {
	data = new(like.ArticleDay)
	row := d.db.QueryRow(c, _selArticleDaySQL, mid)
	if err = row.Scan(&data.ID, &data.Mid, &data.Publish, &data.PublishCount, &data.FinishTask, &data.FinishTime, &data.Ctime, &data.Status); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("RawArticleDay QueryRow mid(%d) error(%v)", mid, err)
		}
	}
	return
}

// JoinArticleDay.
func (d *Dao) JoinArticleDay(ctx context.Context, mid int64) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(ctx, _addArticleDaySQL, mid); err != nil {
		log.Error("JoinArticleDay error d.db.Exec(%d) error(%v)", mid, err)
		return
	}
	return res.LastInsertId()
}

// RawAwards .
func (d *Dao) RawAwards(c context.Context) (res map[string][]*like.ArticleDayAward, err error) {
	var (
		rows *xsql.Rows
		list []*like.ArticleDayAward
	)
	rows, err = d.db.Query(c, _selAwardSQL)
	if err != nil {
		log.Error("RawAwards:d.db.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		row := new(like.ArticleDayAward)
		if err = rows.Scan(&row.ID, &row.ActivityUID, &row.ConditionMin, &row.ConditionMax, &row.SplitPeople, &row.SplitMoney); err != nil {
			log.Error("RawAwards:rows.Scan() error(%v)", err)
			return
		}
		list = append(list, row)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawAwards:rows.Err() error(%v)", err)
		return
	}
	res = make(map[string][]*like.ArticleDayAward, len(list))
	for _, award := range list {
		item := &like.ArticleDayAward{
			ID:           award.ID,
			ActivityUID:  award.ActivityUID,
			ConditionMin: award.ConditionMin,
			ConditionMax: award.ConditionMax,
			SplitPeople:  award.SplitPeople,
			SplitMoney:   award.SplitMoney,
		}
		res[award.ActivityUID] = append(res[award.ActivityUID], item)
	}
	return
}
