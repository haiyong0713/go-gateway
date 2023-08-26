package dao

import (
	"context"
	"database/sql"
	"fmt"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	"strings"

	"go-gateway/app/app-svr/archive-extra/service/api"
	model "go-gateway/app/app-svr/archive-extra/service/model/extra"

	"github.com/pkg/errors"
)

const (
	// 查询
	_extraSQL     = "SELECT aid, biz_type, biz_value, is_deleted FROM archive_extra_biz WHERE aid=?"
	_extrasSQL    = "SELECT aid, biz_type, biz_value, is_deleted FROM archive_extra_biz WHERE aid IN (%s)"
	_extraByKeys  = "SELECT aid, biz_type, biz_value, is_deleted FROM archive_extra_biz WHERE aid=? AND biz_type IN (%s)"
	_queryExtraId = "SELECT id, aid FROM archive_extra_biz WHERE aid=? AND biz_type=? AND is_deleted=0"
	// 删除
	_delExtraSQL = "UPDATE archive_extra_biz SET is_deleted = 1 WHERE id=?"
	// 添加
	_inExtraSQL     = "INSERT INTO archive_extra_biz (aid, biz_type, biz_value) VALUES (?,?,?)"
	_updateExtraSQL = "UPDATE archive_extra_biz SET biz_value=? WHERE id=?"
	// log
	_inExtraLogSQL = "INSERT INTO archive_extra_log (aid, biz_type, act) VALUES (?,?,?)"
)

// RawExtra 指定AID的全部业务信息
func (d *Dao) RawExtra(c context.Context, aid int64) (res map[string]string, err error) {
	var rows *xsql.Rows
	res = make(map[string]string)
	if rows, err = d.db.Query(c, _extraSQL, aid); err != nil {
		log.Error("d.db.Query aid(%d) error(%v)", aid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		extra := &model.ArchiveExtra{}
		if err = rows.Scan(&extra.Aid, &extra.BizType, &extra.BizValue, &extra.IsDeleted); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		if extra.IsDeleted == 0 {
			res[extra.BizType] = extra.BizValue
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// RawExtras 批量查AIDS的全部业务信息
func (d *Dao) RawExtras(c context.Context, aids []int64) (res map[int64]*api.ArchiveExtraValueReply, err error) {
	var rows *xsql.Rows
	res = make(map[int64]*api.ArchiveExtraValueReply, len(aids))
	if rows, err = d.db.Query(c, fmt.Sprintf(_extrasSQL, xstr.JoinInts(aids))); err != nil {
		log.Error("d.db.Query aid(%d) error(%v)", aids, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		extra := &model.ArchiveExtra{}
		if err = rows.Scan(&extra.Aid, &extra.BizType, &extra.BizValue, &extra.IsDeleted); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		if extra.IsDeleted == 0 {
			info, ok := res[extra.Aid]
			if !ok {
				tmp := make(map[string]string)
				info = &api.ArchiveExtraValueReply{ExtraInfo: tmp}
				res[extra.Aid] = info
			}
			info.ExtraInfo[extra.BizType] = extra.BizValue
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// RawExtrasByKeys 批量查指定稿件的Keys的全部业务信息
func (d *Dao) RawExtrasByKeys(c context.Context, aid int64, keys []string) (res map[string]string, err error) {
	var (
		rows    *xsql.Rows
		sqlKeys []string
		args    []interface{}
	)
	res = make(map[string]string)

	args = append(args, aid)
	if len(keys) > 0 {
		for _, item := range keys {
			sqlKeys = append(sqlKeys, "?")
			args = append(args, item)
		}
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_extraByKeys, strings.Join(sqlKeys, ",")), args...); err != nil {
		log.Error("d.db.Query aid(%d) error(%v)", aid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		extra := &model.ArchiveExtra{}
		if err = rows.Scan(&extra.Aid, &extra.BizType, &extra.BizValue, &extra.IsDeleted); err != nil {
			log.Error("rows.Scan error(%v)", err)
			return
		}
		if extra.IsDeleted == 0 {
			res[extra.BizType] = extra.BizValue
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// QueryExtraId 查找有无数据存在
func (d *Dao) QueryExtraId(c context.Context, aid int64, bizType string) (int64, error) {
	row := d.db.QueryRow(c, _queryExtraId, aid, bizType)
	extra := &model.ArchiveExtra{}
	if err := row.Scan(&extra.Id, &extra.Aid); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
	}
	return extra.Id, nil
}

// TXInsertExtra 插入业务信息
func (d *Dao) TXInsertExtra(tx *xsql.Tx, aid int64, bizType, bizValue string) (eff int64, err error) {
	res, err := tx.Exec(_inExtraSQL, aid, bizType, bizValue)
	if err != nil {
		err = errors.Wrapf(err, "InsertExtra _insertExtra Exec err)")
		return
	}
	eff, _ = res.RowsAffected()
	if eff <= 0 {
		err = errors.Wrap(err, "InsertExtra _insertExtra RowsAffected 0")
		return
	}

	return
}

// TXUpdateExtra 更新业务信息
func (d *Dao) TXUpdateExtra(tx *xsql.Tx, id int64, bizValue string) (eff int64, err error) {
	res, err := tx.Exec(_updateExtraSQL, bizValue, id)
	if err != nil {
		err = errors.Wrapf(err, "UpExtra _upExtra Exec err")
		return
	}
	eff, _ = res.RowsAffected()
	if eff <= 0 {
		err = errors.Wrap(err, "UpExtra _upExtra RowsAffected 0")
		return
	}

	return
}

// TXUpExtraLog 更新日志信息表
func (d *Dao) TXUpExtraLog(tx *xsql.Tx, aid int64, bizType string, act int) (err error) {
	res, err := tx.Exec(_inExtraLogSQL, aid, bizType, act)
	if err != nil {
		err = errors.Wrapf(err, "UpExtraLog _upExtraLog Exec err, aid(%d) key(%s) act(%d)", aid, bizType, act)
		return
	}
	if eff, _ := res.RowsAffected(); eff <= 0 {
		err = errors.Wrap(err, "UpExtraLog _upExtraLog RowsAffected 0")
		return
	}
	return
}

// TXDelExtra 删除指定AID的业务信息
func (d *Dao) TXDelExtra(tx *xsql.Tx, id int64) (err error) {
	res, err := tx.Exec(_delExtraSQL, id)
	if err != nil {
		err = errors.Wrapf(err, "DelExtra _delExtra Exec err")
		return
	}
	if eff, _ := res.RowsAffected(); eff <= 0 {
		err = errors.Wrap(err, "DelExtra _delExtra RowsAffected 0")
		return
	}
	return
}

// TXExtraAndLog 事务修改extra两张表数据
func (d *Dao) TXExtraAndLog(ctx context.Context, id, aid int64, bizType, bizValue string, act int) (err error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		err = errors.Wrap(err, "d.db.Begin err")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("Failed to rollback: %+v", err1)
				return
			}
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("Failed to rollback: %+v", err1)
				return
			}
			return
		}
		if err = tx.Commit(); err != nil {
			err = errors.Wrap(err, "tx.Commit() err")
			return
		}
		return
	}()

	// 更新archive_extra_biz表业务字段, act=1新增 act=2修改 act=3删除
	switch act {
	case _insertAct:
		_, err = d.TXInsertExtra(tx, aid, bizType, bizValue)
		if err != nil {
			return
		}
	case _updateAct:
		_, err = d.TXUpdateExtra(tx, id, bizValue)
		if err != nil {
			return
		}
	case _deleteAct:
		err = d.TXDelExtra(tx, id)
		if err != nil {
			return
		}
	}

	// 更新archive_extra_log表操作记录字段
	if err = d.TXUpExtraLog(tx, aid, bizType, act); err != nil {
		return
	}

	return
}
