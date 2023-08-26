package like

import (
	"context"
	"database/sql"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/question"
)

const (
	_questionDetailSQL      = "SELECT id,base_id,`name`,right_answer,wrong_answer,attribute,pic FROM question_detail WHERE base_id=? AND state=1"
	_AddQuestionSQL         = "INSERT IGNORE INTO  act_s10_answer(mid) VALUES(?)"
	_UpQuestionAllTimesSQL  = "UPDATE act_s10_answer SET all_times=all_times+1 WHERE mid=?"
	_UpQuestionRuleSQL      = "UPDATE act_s10_answer SET know_rule=1 WHERE mid=?"
	_UpQuestionUpPendantSQL = "UPDATE act_s10_answer SET have_pendant=1 WHERE mid=?"
)

// QuestionDetails .
func (d *Dao) QuestionDetails(c context.Context, baseID int64) (data map[int64]*question.Detail, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _questionDetailSQL, baseID)
	if err != nil {
		log.Error("QuestionDetails:d.db.Query(%v) error(%v)", baseID, err)
		return
	}
	defer rows.Close()
	data = make(map[int64]*question.Detail, 500)
	for rows.Next() {
		n := new(question.Detail)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.Name, &n.RightAnswer, &n.WrongAnswer, &n.Attribute, &n.Pic); err != nil {
			log.Error("QuestionDetails:rows.Scan() error(%v)", err)
			return
		}
		data[n.ID] = n
	}
	if err = rows.Err(); err != nil {
		log.Error("QuestionDetails:rows.Err() error(%v)", err)
	}
	return
}

// AddUserQuestion .
func (d *Dao) AddUserQuestion(ctx context.Context, mid int64) (int64, error) {
	res, err := d.db.Exec(ctx, _AddQuestionSQL, mid)
	if err != nil {
		log.Error("d.AddUserQuestion(%d) error(%+v)", mid, err)
		return 0, err
	}
	return res.LastInsertId()
}

// UpQuestionAllTimes .
func (d *Dao) UpQuestionAllTimes(ctx context.Context, mid int64) (ef int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(ctx, _UpQuestionAllTimesSQL, mid); err != nil {
		log.Error("d.UpQuestionAllTimes(%d) Exec error(%+v)", mid, err)
		return
	}
	ef, _ = res.RowsAffected()
	return
}

// UpKnowRule .
func (d *Dao) UpKnowRule(ctx context.Context, mid int64) (ef int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(ctx, _UpQuestionRuleSQL, mid); err != nil {
		log.Error("d.UpKnowRule(%d) Exec error(%+v)", mid, err)
		return
	}
	ef, _ = res.RowsAffected()
	return
}

// UpPendant .
func (d *Dao) UpPendant(ctx context.Context, mid int64) (ef int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(ctx, _UpQuestionUpPendantSQL, mid); err != nil {
		log.Error("d.UpPendant(%d) Exec error(%+v)", mid, err)
		return
	}
	ef, _ = res.RowsAffected()
	return
}
