package dao

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"
)

const (
	_insertScript     = "INSERT INTO gwmng_script (userid, script_type, parameter, app) VALUES (?,?,?,?)"
	_selectUserScript = "SELECT * FROM gwmng_script WHERE userid = ?"
	_selectScript     = "SELECT * FROM gwmng_script WHERE id = ?"
	_deleteScript     = "Delete FROM gwmng_script WHERE id = ?"
)

func (d *dao) InsertScript(ctx context.Context, script *model.Script) error {
	if _, err := d.db.Exec(ctx, _insertScript, script.UserName, script.Type, script.Parameter, script.APP); err != nil {
		log.Error("d.db.Exec error(%v)", err)
		return err
	}
	return nil
}

func (d *dao) GetUserScript(ctx context.Context, userid string) ([]*model.Script, error) {
	rows, err := d.db.Query(ctx, _selectUserScript, userid)
	if err != nil {
		return nil, err
	}
	var res []*model.Script
	defer rows.Close()
	for rows.Next() {
		s := &model.Script{}
		if err = rows.Scan(&s.ID, &s.UserName, &s.Type, &s.Parameter, &s.CTime, &s.MTime, &s.APP); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *dao) GetScript(ctx context.Context, id string) (*model.Script, error) {
	script := &model.Script{}
	row := d.db.QueryRow(ctx, _selectScript, id)
	if err := row.Scan(&script.ID, &script.UserName, &script.Type, &script.Parameter, &script.CTime, &script.MTime, &script.APP); err != nil {
		return nil, err
	}
	return script, nil
}

func (d *dao) DeleteScript(ctx context.Context, id string) error {
	if _, err := d.db.Exec(ctx, _deleteScript, id); err != nil {
		log.Error("d.db.Exec error(%v)", err)
		return err
	}
	return nil
}
