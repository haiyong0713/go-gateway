package funny

import (
	"context"
	"errors"
	"fmt"
	"go-common/library/log"
)

const (
	_likesSQL = "SELECT wid FROM likes force index (`ix_like_0`) WHERE sid in (14751,14570,14569,14568,14567,14566,14565) order by `id` asc limit 50 offset ?"
	_funnySQL = "SELECT id FROM act_funny_mid WHERE mid = ?"
)

// get likes user data.
func (d *dao) GetUserBatchData(c context.Context, sid, limit, page int64) ([]int64, error) {
	var aids []int64
	rows, err := d.db.Query(c, _likesSQL, page*limit)
	if err != nil {
		log.Errorc(c, "Caculate GetUserBatchData query Err sql:%v limit:%v offset:%v err:%v", _likesSQL, limit, page, err)
		return nil, errors.New(fmt.Sprintf("Caculate GetUserBatchData rows.Scan Err sql:%v limit:%v offset:%v err:%v", _likesSQL, limit, page, err))
	}
	defer rows.Close()
	for rows.Next() {
		n := new(int64)
		if err := rows.Scan(&n); err != nil {
			return nil, errors.New(fmt.Sprintf("Caculate GetUserBatchData rows.Scan Err sql:%v limit:%v offset:%v err:%v", _likesSQL, limit, page, err))
		}
		aids = append(aids, *n)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New(fmt.Sprintf("Caculate GetUserBatchData rows.Err Err sql:%v limit:%v offset:%v err:%v", _likesSQL, limit, page, err))
	}

	return aids, nil
}

// if the record is a new user
func (d *dao) IsNewUser(c context.Context, Mid int64) (bool, error) {
	var data []int64
	rows, err := d.db.Query(c, _funnySQL, Mid)
	if err != nil {
		log.Errorc(c, "Caculate IsNewUser query Err sql:%v mid:%v err:%v", _funnySQL, Mid, err)
		return false, errors.New(fmt.Sprintf("Caculate IsNewUser query Err sql:%v mid:%v err:%v", _funnySQL, Mid, err))
	}

	defer rows.Close()
	for rows.Next() {
		n := new(int64)
		if err := rows.Scan(&n); err != nil {
			return false, errors.New(fmt.Sprintf("Judge IsNewUser rows.Scan Err sql:%v mid:%v err:%v", _funnySQL, Mid, err))
		}
		// 找到记录
		data = append(data, *n)
	}
	if err := rows.Err(); err != nil {
		return false, errors.New(fmt.Sprintf("Judge IsNewUser rows.Err Err sql:%v mid:%v err:%v", _funnySQL, Mid, err))
	}

	if len(data) == 0 {
		return true, nil
	}

	return false, nil
}
