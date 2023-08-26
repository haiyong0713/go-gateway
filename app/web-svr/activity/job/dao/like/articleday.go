package like

import (
	"context"
	"database/sql"
	xsql "database/sql"
	"fmt"

	"go-common/library/log"
	"go-common/library/xstr"
)

const (
	_selArticleByMid       = "SELECT publish FROM act_artile_day2 WHERE mid=?"
	_selArticleRiskMid     = "SELECT mid FROM act_artile_day2 WHERE status=1 limit 10000"
	_upArticlePublishByMid = "UPDATE  act_artile_day2 SET  publish=? WHERE mid=?"
	_upArticleCountByMid   = "UPDATE  act_artile_day2 SET  publish=?,publish_count=? WHERE mid=?"
	_upArticleRiskByMid    = "UPDATE  act_artile_day2 SET  status=? WHERE mid=?"
	_sleArticleAwardSQL    = "SELECT id,condition_min FROM act_artile_award WHERE is_deleted=0 AND activity_uid=? ORDER BY `condition_min` ASC"
	_upArticleAward        = "UPDATE act_artile_award SET split_people = CASE %s END WHERE id IN (%s)"
)

// RawArticlePublish .
func (d *Dao) RawArticlePublish(c context.Context, mid int64) (res string, err error) {
	row := d.db.QueryRow(c, _selArticleByMid, mid)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}

// UpArticlePublish .
func (d *Dao) UpArticlePublish(c context.Context, publish string, mid int64) (err error) {
	if _, err = d.db.Exec(c, _upArticlePublishByMid, publish, mid); err != nil {
		log.Error("DayClockIn UpArticlePublish:d.db.Exec(%s,%d) error(%v)", _upArticlePublishByMid, publish, mid, err)
	}
	return
}

// UpArticlePublishCount .
func (d *Dao) UpArticlePublishCount(c context.Context, publish string, publishCount int, mid int64) (err error) {
	if _, err = d.db.Exec(c, _upArticleCountByMid, publish, publishCount, mid); err != nil {
		log.Error("UpArticlePublishCount:d.db.Exec(%s,%s,%d,%d) error(%v)", _upArticleCountByMid, publish, publishCount, mid, err)
	}
	return
}

// UpArticleRisk .
func (d *Dao) UpArticleRisk(c context.Context, status, mid int64) (err error) {
	if _, err = d.db.Exec(c, _upArticleRiskByMid, status, mid); err != nil {
		log.Error("UpArticleRisk:d.db.Exec(%d,%d) error(%v)", _upArticleRiskByMid, status, mid, err)
	}
	return
}

// SelArticleAward.
func (d *Dao) SelArticleAward(c context.Context, activityUID string) (res map[int64]int64) {
	rows, err := d.db.Query(c, _sleArticleAwardSQL, activityUID)
	if err != nil {
		log.Error("notice.Query error(%v)", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]int64)
	for rows.Next() {
		var id, min int64
		if err = rows.Scan(&id, &min); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		res[id] = min
	}
	if err = rows.Err(); err != nil {
		log.Error("SelArticleAward rows.Err error(%v)", err)
		return
	}
	return
}

// SelArticleRiskMids.
func (d *Dao) SelArticleRiskMids(c context.Context) (res map[int64]struct{}, err error) {
	rows, err := d.db.Query(c, _selArticleRiskMid)
	if err != nil {
		log.Error("SelArticleRiskMids db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]struct{})
	for rows.Next() {
		var mid int64
		if err = rows.Scan(&mid); err != nil {
			log.Error("SelArticleRiskMids row.Scan error(%v)", err)
			return
		}
		res[mid] = struct{}{}
	}
	err = rows.Err()
	return
}

// UpAwardPeople .
func (d *Dao) UpAwardPeople(c context.Context, yesterdayCalc map[int64]int64) (affected int64, err error) {
	var (
		caseStr string
		ids     []int64
		res     xsql.Result
	)
	if len(yesterdayCalc) == 0 {
		return
	}
	for id, peoples := range yesterdayCalc {
		caseStr = fmt.Sprintf("%s WHEN id = %d THEN %d", caseStr, id, peoples)
		ids = append(ids, id)
	}
	if res, err = d.db.Exec(c, fmt.Sprintf(_upArticleAward, caseStr, xstr.JoinInts(ids))); err != nil {
		log.Error("UpAwardPeople:d.db.Exec() error(%v)", _upArticleAward, err)
		return
	}
	return res.RowsAffected()
}
