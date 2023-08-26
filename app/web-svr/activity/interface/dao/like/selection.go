package like

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const _addSelectionQA = "INSERT INTO act_selection_qa (mid,question_order,question,product,role) VALUES %s"

func (d *Dao) AddSelectionQA(ctx context.Context, mid int64, list []*like.SelectionQA) (int64, error) {
	var (
		rowStrs []string
		args    []interface{}
	)
	for orderID, v := range list {
		if len(v.Answer) == 0 {
			rowStrs = append(rowStrs, "(?,?,?,?,?)")
			args = append(args, mid, orderID+1, v.Question, "", "")
		} else {
			for _, answer := range v.Answer {
				rowStrs = append(rowStrs, "(?,?,?,?,?)")
				args = append(args, mid, orderID+1, v.Question, answer.Product, answer.Role)
			}
		}
	}
	row, err := d.db.Exec(ctx, fmt.Sprintf(_addSelectionQA, strings.Join(rowStrs, ",")), args...)
	if err != nil {
		log.Errorc(ctx, "AddSelectionQA mid(%d) error(%+v)", mid, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _SelSelectionQA = "Select id,mid,question_order,question,product,role FROM  act_selection_qa WHERE mid=? ORDER BY question_order ASC,id ASC"

func (d *Dao) SelSelectionDBsByMid(ctx context.Context, mid int64) (list []*like.SelectionQADB, err error) {
	var rows *xsql.Rows
	rows, err = d.db.Query(ctx, _SelSelectionQA, mid)
	if err != nil {
		log.Errorc(ctx, "SelSelectionDBsByMid:d.db.Query(%v) error(%v)", mid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		n := new(like.SelectionQADB)
		if err = rows.Scan(&n.ID, &n.Mid, &n.QuestionOrder, &n.Question, &n.Product, &n.Role); err != nil {
			log.Error("SelSelectionDBsByMid:rows.Scan() mid(%d) error(%v)", mid, err)
			return
		}
		list = append(list, n)
	}
	if err = rows.Err(); err != nil {
		log.Error("SelSelectionDBsByMid:rows.Err() mid(%d) error(%v)", mid, err)
		return
	}
	return
}

const _selAllSelectionSQL = "Select id,mid,question_order,question,product,role FROM  act_selection_qa WHERE id > ? ORDER BY id ASC LIMIT 1000"

func (d *Dao) SelAllSelection(c context.Context, maxID int64) (res []*like.SelectionQADB, err error) {
	res = make([]*like.SelectionQADB, 0, 1000)
	rows, err := d.db.Query(c, _selAllSelectionSQL, maxID)
	if err != nil {
		log.Error("SelAllSelection.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		row := &like.SelectionQADB{}
		if err = rows.Scan(&row.ID, &row.Mid, &row.QuestionOrder, &row.Question, &row.Product, &row.Role); err != nil {
			log.Error("SelAllSelection row.Scan() error(%v)", err)
			return
		}
		res = append(res, row)
	}
	if err = rows.Err(); err != nil {
		log.Error("SelAllSelection rows.Err() error(%v)", err)
	}
	return
}

const _selProductRoleSQL = "Select id,category_id,category_type,role,product,tags,tags_type,vote_num,ctime,mtime FROM act_selection_productrole WHERE is_deleted=0 AND category_id=?"

func (d *Dao) SelProductRoleByCategory(ctx context.Context, categoryID int64) (res []*like.ProductRoleDB, err error) {
	res = make([]*like.ProductRoleDB, 0)
	rows, err := d.db.Query(ctx, _selProductRoleSQL, categoryID)
	if err != nil {
		log.Error("SelProductRoleByCategory.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		row := &like.ProductRoleDB{}
		if err = rows.Scan(&row.ID, &row.CategoryID, &row.CategoryType, &row.Role, &row.Product, &row.Tags, &row.TagsType, &row.VoteNum, &row.Ctime, &row.Mtime); err != nil {
			log.Error("SelProductRoleByCategory row.Scan() error(%v)", err)
			return
		}
		res = append(res, row)
	}
	if err = rows.Err(); err != nil {
		log.Error("SelProductRoleByCategory rows.Err() error(%v)", err)
	}
	return
}

const _upPRVoteSQL = "UPDATE act_selection_productrole SET vote_num=vote_num+1 WHERE id=?"

func (d *Dao) txUpPRVote(ctx context.Context, tx *xsql.Tx, id int64) (res int64, err error) {
	var sqlRes sql.Result
	if sqlRes, err = tx.Exec(_upPRVoteSQL, id); err != nil {
		log.Error("upPRVote:tx.Exec(%s) id(%d) error(%v)", _upPRVoteSQL, id, err)
		return
	}
	return sqlRes.LastInsertId()
}

const _insertUserLog = "INSERT INTO act_selection_vote_log (`mid`,`category_id`,`productrole_id`,`vote_date`) VALUES(?,?,?,?)"

func (dao *Dao) txAddUserLog(c context.Context, tx *xsql.Tx, mid, categoryID, productroleID int64, nowTime time.Time) (res int64, err error) {
	var sqlRes sql.Result
	if sqlRes, err = tx.Exec(_insertUserLog, mid, categoryID, productroleID, nowTime.Format("2006-01-02")); err != nil {
		log.Error("txAddUserLog:tx.Exec(%s) error(%v)", _insertUserLog, err)
		return
	}
	return sqlRes.LastInsertId()
}

// ItemAndContent .
func (d *Dao) VoteTransact(c context.Context, mid, categoryID, productroleID int64, nowTime time.Time) (res int64, err error) {
	var (
		tx *xsql.Tx
	)
	if tx, err = d.db.Begin(c); err != nil {
		log.Error("d.db.Begin() error(%v)", err)
		return
	}
	defer func() {
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if res, err = d.txUpPRVote(c, tx, productroleID); err != nil {
		return
	}
	if res, err = d.txAddUserLog(c, tx, mid, categoryID, productroleID, nowTime); err != nil {
		return
	}
	return
}

const _selPrNoVoteSQL = "Select id,category_id,category_type,role,product FROM act_selection_productrole WHERE is_deleted=0 AND category_id=? ORDER BY CONVERT(%s USING gbk) ASC"

func (d *Dao) SelPrNotVoteByCategory(ctx context.Context, categoryID int64, field string) (res []*like.ProductRoleDB, err error) {
	res = make([]*like.ProductRoleDB, 0)
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selPrNoVoteSQL, field), categoryID)
	if err != nil {
		log.Error("SelPrNotVoteByCategory.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		row := &like.ProductRoleDB{}
		if err = rows.Scan(&row.ID, &row.CategoryID, &row.CategoryType, &row.Role, &row.Product); err != nil {
			log.Error("SelPrNotVoteByCategory row.Scan() error(%v)", err)
			return
		}
		res = append(res, row)
	}
	if err = rows.Err(); err != nil {
		log.Error("SelPrNotVoteByCategory rows.Err() error(%v)", err)
	}
	return
}
