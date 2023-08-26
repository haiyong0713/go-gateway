package question

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/question"

	"github.com/pkg/errors"
)

const (
	_basesSQL         = "SELECT id,business_id,foreign_id,count,one_ts,retry_ts,stime,etime,name,`separator`,`distribute_type` FROM question_base WHERE stime <= ? AND etime >= ?"
	_detailSQL        = "SELECT id,base_id,`name`,right_answer,wrong_answer,attribute,pic FROM question_detail WHERE id = ? AND state = 1"
	_detailsSQL       = "SELECT id,base_id,`name`,right_answer,wrong_answer,attribute,pic FROM question_detail WHERE id IN (%s) AND state = 1"
	_lastQuesTsSQL    = "SELECT id,mid,base_id,detail_id,pool_id,answer,is_right,`index`,question_time,answer_time,ctime,mtime FROM %s WHERE mid = ? AND base_id = ? ORDER BY mtime DESC LIMIT 1"
	_userLogsSQL      = "SELECT /*master*/ id,mid,base_id,detail_id,pool_id,answer,is_right,`index`,question_time,answer_time,ctime,mtime FROM %s WHERE mid = ? AND base_id = ? AND pool_id = ?"
	_userLogAddSQL    = "INSERT INTO %s(mid,base_id,detail_id,pool_id,`index`,question_time) VALUES(?,?,?,?,?,?)"
	_userLogUpSQL     = "UPDATE %s SET is_right=?,answer_time=? WHERE id = ?"
	_userRecordAddSQL = "INSERT INTO question_user_records(mid,base_id,pool_id,pool_count,answer_count,right_count,state) VALUES(?,?,?,?,?,?,?)"
	_userRecordUpSQL  = "UPDATE question_user_records SET answer_count=?, right_count=?,state=? WHERE mid=? AND base_id=? and pool_id=?"
	_userRecordsSQL   = "SELECT id,mid,base_id,pool_id,pool_count,answer_count,right_count,state,ctime,mtime FROM question_user_records WHERE mid = ? AND base_id IN (%s) AND state IN (%s) ORDER BY ctime DESC LIMIT ?,?"
	_userRecordSQL    = "SELECT id,mid,base_id,pool_id,pool_count,answer_count,right_count,state,ctime,mtime FROM question_user_records WHERE mid = ? AND base_id = ? AND pool_id = ?"
	_selectTotalRank  = "SELECT count(*) from act_gaokao_dati_record "
	_selectUserRank   = "SELECT count(*) from act_gaokao_dati_record where  province = ? and course = ? and rank_socre >= ?"
	_insertUserCore   = "INSERT INTO act_gaokao_dati_record(mid,year,province,course,score,used_time,rank_socre)VALUES(?,?,?,?,?,?,?)"
)

func userLogTableName(baseID int64) string {
	name := "question_user_log"
	if baseID > 1 {
		name = fmt.Sprintf("question_user_log_%d", baseID)
	}
	return name
}

// RawBases get question base.
func (d *Dao) RawBases(c context.Context, ts xtime.Time) (data []*question.Base, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, _basesSQL, ts, ts)
	if err != nil {
		log.Error("RawBases:d.db.Query(%d) error(%v)", ts, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(question.Base)
		if err = rows.Scan(&n.ID, &n.BusinessID, &n.ForeignID, &n.Count, &n.OneTs, &n.RetryTs, &n.Stime, &n.Etime, &n.Name, &n.Separator, &n.DistributeType); err != nil {
			log.Error("RawBases:rows.Scan() error(%v)", err)
			return
		}
		data = append(data, n)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawBases:rows.Err() error(%v)", err)
	}
	return
}

// RawDetail .
func (d *Dao) RawDetail(c context.Context, id int64) (data *question.Detail, err error) {
	data = new(question.Detail)
	row := d.db.QueryRow(c, _detailSQL, id)
	if err = row.Scan(&data.ID, &data.BaseID, &data.Name, &data.RightAnswer, &data.WrongAnswer, &data.Attribute, &data.Pic); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "RawDetail:QueryRow")
		}
	}
	return
}

// RawDetails .
func (d *Dao) RawDetails(c context.Context, ids []int64) (data map[int64]*question.Detail, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, fmt.Sprintf(_detailsSQL, xstr.JoinInts(ids)))
	if err != nil {
		log.Error("RawDetails:d.db.Query(%v) error(%v)", ids, err)
		return
	}
	defer rows.Close()
	data = make(map[int64]*question.Detail, len(ids))
	for rows.Next() {
		n := new(question.Detail)
		if err = rows.Scan(&n.ID, &n.BaseID, &n.Name, &n.RightAnswer, &n.WrongAnswer, &n.Attribute, &n.Pic); err != nil {
			log.Error("RawDetails:rows.Scan() error(%v)", err)
			return
		}
		data[n.ID] = n
	}
	if err = rows.Err(); err != nil {
		log.Error("RawDetails:rows.Err() error(%v)", err)
	}
	return
}

// RawLastQuesTs get last question time.
func (d *Dao) RawLastQuesLog(c context.Context, mid, baseID int64) (data *question.UserAnswerLog, err error) {
	data = new(question.UserAnswerLog)
	row := d.db.QueryRow(c, fmt.Sprintf(_lastQuesTsSQL, userLogTableName(baseID)), mid, baseID)
	if err = row.Scan(&data.ID, &data.Mid, &data.BaseID, &data.DetailID, &data.PoolID, &data.Answer, &data.IsRight, &data.Index, &data.QuestionTime, &data.AnswerTime, &data.Ctime, &data.Mtime); err != nil {
		if err == sql.ErrNoRows {
			data = nil
			err = nil
		} else {
			err = errors.Wrap(err, "RawLastQuesTs:QueryRow")
		}
	}
	return
}

// RawUserLogs get user logs.
func (d *Dao) RawUserLogs(c context.Context, mid, baseID, poolID int64) (list []*question.UserAnswerLog, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, fmt.Sprintf(_userLogsSQL, userLogTableName(baseID)), mid, baseID, poolID)
	if err != nil {
		log.Error("RawDetails:d.db.Query(%d,%d,%d) error(%v)", mid, baseID, poolID, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(question.UserAnswerLog)
		if err = rows.Scan(&n.ID, &n.Mid, &n.BaseID, &n.DetailID, &n.PoolID, &n.Answer, &n.IsRight, &n.Index, &n.QuestionTime, &n.AnswerTime, &n.Ctime, &n.Mtime); err != nil {
			log.Error("RawDetails:rows.Scan() error(%v)", err)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawDetails:rows.Err() error(%v)", err)
	}
	return
}

// AddUserLog add user log.
func (d *Dao) AddUserLog(c context.Context, mid, baseID, detailID, poolID, index int64, questionTime time.Time) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, fmt.Sprintf(_userLogAddSQL, userLogTableName(baseID)), mid, baseID, detailID, poolID, index, questionTime); err != nil {
		log.Error("AddUserLog d.db.Exec(%d,%d,%d,%d,%d,%v) error(%v)", mid, baseID, detailID, poolID, index, questionTime, err)
		return
	}
	return res.LastInsertId()
}

// UpUserLog .
func (d *Dao) UpUserLog(c context.Context, isRight int, answerTime time.Time, id, baseID int64) (err error) {
	if _, err = d.db.Exec(c, fmt.Sprintf(_userLogUpSQL, userLogTableName(baseID)), isRight, answerTime, id); err != nil {
		log.Error("UpUserLog d.db.Exec(%d,%v,%d) baseID(%d) error(%v)", isRight, answerTime, baseID, id, err)
	}
	return
}

// AddUserRecords add user log.
func (d *Dao) AddUserRecords(c context.Context, mid, baseID, poolID, poolCount, answerCount, rightCount, state int64) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _userRecordAddSQL, mid, baseID, poolID, poolCount, answerCount, rightCount, state); err != nil {
		log.Errorc(c, "AddUserRecords d.db.Exec(%d,%d,%d,%d) error(%v)", mid, baseID, poolID, poolCount, err)
		return
	}
	return res.LastInsertId()
}

// UpUserRecords up user log.
func (d *Dao) UpUserRecords(c context.Context, mid, baseID, poolID, answerCount, rightCount, state int64) (err error) {
	if _, err = d.db.Exec(c, _userRecordUpSQL, answerCount, rightCount, state, mid, baseID, poolID); err != nil {
		log.Errorc(c, "UpUserRecords d.db.Exec(%d,%d,%d,%d,%d,%d) error(%v)", answerCount, rightCount, state, mid, baseID, poolID, err)
	}
	return
}

func (d *Dao) SelectUserRank(c context.Context, province string, course string, rankScore int64) (rank int64, err error) {
	row := d.db.QueryRow(c, _selectUserRank, province, course, rankScore)
	if err = row.Scan(&rank); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "SelectUserRank QueryRow")
		}
	}
	return
}

func (d *Dao) SelectTotalCount(c context.Context) (rank int64, err error) {
	row := d.db.QueryRow(c, _selectTotalRank)
	if err = row.Scan(&rank); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "SelectUserRank QueryRow")
		}
	}
	return
}

func (d *Dao) AddUserScore(c context.Context, mid int64, year int, province string, course string, score int, usedTime int, rankScore int64) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _insertUserCore, mid, year, province, course, score, usedTime, rankScore); err != nil {
		log.Error("AddUserScore d.db.Exec(%v,%v,%v,%v,%v,%v,%v) error(%+v)", mid, year, province, course, score, usedTime, rankScore, err)
		return
	}
	return res.LastInsertId()
}

// RawUserRecords get user logs.
func (d *Dao) RawUserRecords(c context.Context, mid int64, baseIDs, state []int64, offset, limit int64) (list []*question.UserAnswerRecord, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(c, fmt.Sprintf(_userRecordsSQL, xstr.JoinInts(baseIDs), xstr.JoinInts(state)), mid, offset, limit)
	if err != nil {
		log.Errorc(c, "RawUserRecords:d.db.Query(%s,%s,%d) error(%v)", xstr.JoinInts(baseIDs), xstr.JoinInts(state), mid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(question.UserAnswerRecord)
		if err = rows.Scan(&n.ID, &n.Mid, &n.BaseID, &n.PoolID, &n.PoolCount, &n.AnswerCount, &n.RightCount, &n.State, &n.Ctime, &n.Mtime); err != nil {
			log.Errorc(c, "RawUserRecords:rows.Scan() error(%v)", err)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(c, "RawUserRecords:rows.Err() error(%v)", err)
	}
	return
}

// RawUserRecord get user logs.
func (d *Dao) RawUserRecord(c context.Context, mid, baseID, poolID int64) (record *question.UserAnswerRecord, err error) {
	row := d.db.QueryRow(c, _userRecordSQL, mid, baseID, poolID)
	record = new(question.UserAnswerRecord)
	if err = row.Scan(&record.ID, &record.Mid, &record.BaseID, &record.PoolID, &record.PoolCount, &record.AnswerCount, &record.RightCount, &record.State, &record.Ctime, &record.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			record = nil
		} else {
			log.Errorc(c, "RawUserRecord:rows.Scan() error(%v)", err)
		}
	}
	return
}
