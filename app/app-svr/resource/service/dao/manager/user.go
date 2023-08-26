package manager

import (
	"context"

	"go-common/library/database/sql"
	xecode "go-common/library/ecode"

	model "go-gateway/app/app-svr/app-feed/admin/model/frontpage"
)

const (
	getActiveUsersSQL = "SELECT id, username, nickname, email, phone, department_id, state, wx_id, ctime, mtime FROM user WHERE state = 0"
)

func (d *Dao) GetActiveUsers(ctx context.Context) (res []*model.User, err error) {
	res = make([]*model.User, 0)

	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, getActiveUsersSQL); err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			err = nil
		}
		return nil, err
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	for rows.Next() {
		row := new(model.User)
		if err = rows.Scan(
			&row.ID, &row.Username, &row.Nickname, &row.Email, &row.Phone, &row.DepartmentID, &row.State, &row.WXID, &row.Ctime, &row.Mtime); err != nil {
			return nil, err
		}
		res = append(res, row)
	}

	if err = rows.Err(); err != nil {
		res = nil
		return
	}

	return
}
