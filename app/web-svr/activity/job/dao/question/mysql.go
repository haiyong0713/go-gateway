package question

import (
	"context"
	"database/sql"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/time"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/question"
	quesmdl "go-gateway/app/web-svr/activity/job/model/question"

	"github.com/pkg/errors"
)

const (
	_basesSQL              = "SELECT id,business_id,foreign_id,count,one_ts,retry_ts,stime,etime FROM question_base WHERE stime <= ? AND etime >= ?"
	_detailListSQL         = "SELECT id , base_id , name , right_answer , wrong_answer , attribute ,pic FROM question_detail WHERE base_id = ? AND state = ?"
	_answerUserSQL         = "SELECT id,mid FROM act_s10_answer WHERE id>? ORDER BY id ASC limit 100"
	_updateUserRankSQL     = "UPDATE act_s10_answer SET user_rank = CASE %s END WHERE mid IN (%s)"
	_updateUserRankZeroSQL = "UPDATE act_s10_answer SET user_rank = 0"
	_updateUserDataSQL     = "UPDATE act_s10_answer SET user_score=?,answer_times=?,user_rank=? WHERE id=?"
	_updateDetailSQL       = "UPDATE question_detail SET name = ? , right_answer = ? , wrong_answer = ? , state = ? WHERE id = ?"
)

// RawBases get question base.
func (d *Dao) RawBases(c context.Context, stime, etime time.Time) (data []*question.Base, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _basesSQL, stime, etime)
	if err != nil {
		log.Errorc(c, "RawBases:d.db.Query(%d,%d) error(%v)", stime, etime, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(question.Base)
		if err = rows.Scan(&n.ID, &n.BusinessID, &n.ForeignID, &n.Count, &n.OneTs, &n.RetryTs, &n.Stime, &n.Etime); err != nil {
			log.Errorc(c, "RawBases:rows.Scan() error(%v)", err)
			return
		}
		data = append(data, n)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(c, "RawBases:rows.Err() error(%v)", err)
	}
	return
}

// RawDetailList raw detail list.
func (d *Dao) RawDetailList(c context.Context, baseID int64, state int) (details []*question.Detail, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _detailListSQL, baseID, state)
	if err != nil {
		log.Errorc(c, "RawDetailList:d.db.Query(%d) error(%v)", baseID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		item := &question.Detail{}
		if err = rows.Scan(&item.ID, &item.BaseID, &item.Name, &item.RightAnswer, &item.WrongAnswer, &item.Attribute, &item.Pic); err != nil {
			log.Errorc(c, "RawDetailList:rows.Scan() error(%v)", err)
			return
		}
		details = append(details, item)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(c, "RawBases:rows.Err() error(%v)", err)
	}
	return
}

// AnswerUsers
func (d *Dao) AnswerUsers(ctx context.Context, id int64) (rs []*quesmdl.AnswerUser, err error) {
	rs = []*quesmdl.AnswerUser{}
	rows, err := d.db.Query(ctx, _answerUserSQL, id)
	if err != nil {
		err = errors.Wrap(err, "AnswerUsers:d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &quesmdl.AnswerUser{}
		err = rows.Scan(&r.ID, &r.Mid)
		if err != nil {
			err = errors.Wrap(err, "AnswerUsers:rows.Scan error")
			return
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "AnswerUsers:rows.Err")
	}
	return
}

// UpdateUserRank
func (d *Dao) UpdateUserRank(ctx context.Context, userMap map[int64]int) (affected int64, err error) {
	var (
		caseStr string
		mids    []int64
		res     sql.Result
	)
	if len(userMap) == 0 {
		return
	}
	for mid, rank := range userMap {
		caseStr = fmt.Sprintf("%s WHEN mid = %d THEN %d", caseStr, mid, rank)
		mids = append(mids, mid)
	}
	if res, err = d.db.Exec(ctx, fmt.Sprintf(_updateUserRankSQL, caseStr, xstr.JoinInts(mids))); err != nil {
		err = errors.Wrap(err, "UpdateUserRank:db.Exec error")
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpUserRankZero(c context.Context) (err error) {
	if _, err = d.db.Exec(c, _updateUserRankZeroSQL); err != nil {
		log.Errorc(c, "UpUserRankZero:d.db.Exec(%d,%d) error(%v)", _updateUserRankZeroSQL, err)
	}
	return
}

func (d *Dao) UpdateUserData(c context.Context, userInfo *quesmdl.AnswerUserInfo) (err error) {
	if _, err = d.db.Exec(c, _updateUserDataSQL, userInfo.UserScore, userInfo.AnswerTimes, userInfo.UserRank, userInfo.ID); err != nil {
		log.Errorc(c, "UpdateUserData:d.db.Exec(%d,%d) error(%v)", _updateUserDataSQL, err)
	}
	return
}

func (d *Dao) UpdateQuestionDetail(ctx context.Context, detail *question.Detail, state int) (err error) {
	if _, err = d.db.Exec(ctx, _updateDetailSQL, detail.Name, detail.RightAnswer, detail.WrongAnswer, state, detail.ID); err != nil {
		log.Errorc(ctx, "UpdateQuestionDetail:d.db.Exec(%d,%d) error(%v)", _updateDetailSQL, err)
	}
	return
}
