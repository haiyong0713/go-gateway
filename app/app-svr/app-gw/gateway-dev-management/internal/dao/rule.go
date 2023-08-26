package dao

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/gateway-dev-management/internal/model"
)

const (
	_insertCodeRule         = "INSERT INTO gwmng_code_rule (team, rule_type, method, code, rule_id) VALUES (?,?,?,?,?)"
	_selectRuleId           = "SELECT rule_id FROM gwmng_code_rule WHERE team = ? AND rule_type = ? AND method = ? AND code = ?"
	_deleteCodeRule         = "Delete FROM gwmng_code_rule WHERE rule_id = ?"
	_selectOwnerService     = "SELECT service FROM gwmng_service_owner WHERE primary_owner=? or secondary_owner=?"
	_selectPrimaryService   = "SELECT service FROM gwmng_service_owner WHERE primary_owner=?"
	_selectSecondaryService = "SELECT service FROM gwmng_service_owner WHERE secondary_owner=?"
)

func (d *dao) SelectRuleId(ctx context.Context, rule *model.CodeRule) (int64, error) {
	var value int64
	err := d.db.QueryRow(ctx, _selectRuleId, rule.Team, rule.Type, rule.Method, rule.Code).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return -1, err
	}
	return value, nil
}

func (d *dao) InsertCodeRule(ctx context.Context, rule *model.CodeRule) error {
	if _, err := d.db.Exec(ctx, _insertCodeRule, rule.Team, rule.Type, rule.Method, rule.Code, rule.RuleId); err != nil {
		log.Error("d.db.Exec error(%v)", err)
		return err
	}
	return nil
}

func (d *dao) DeleteCodeRule(ctx context.Context, ruleId int64) error {
	if _, err := d.db.Exec(ctx, _deleteCodeRule, ruleId); err != nil {
		log.Error("d.db.Exec error(%v)", err)
		return err
	}
	return nil
}

func (d *dao) GetUserService(ctx context.Context, username string) ([]string, error) {
	rows, err := d.db.Query(ctx, _selectOwnerService, username, username)
	if err != nil {
		return nil, err
	}
	var res []string
	defer rows.Close()
	for rows.Next() {
		var s string
		if err = rows.Scan(&s); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *dao) GetPrimaryService(ctx context.Context, username string) ([]string, error) {
	rows, err := d.db.Query(ctx, _selectPrimaryService, username)
	if err != nil {
		return nil, err
	}
	var res []string
	defer rows.Close()
	for rows.Next() {
		var s string
		if err = rows.Scan(&s); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *dao) GetSecondaryService(ctx context.Context, username string) ([]string, error) {
	rows, err := d.db.Query(ctx, _selectSecondaryService, username)
	if err != nil {
		return nil, err
	}
	var res []string
	defer rows.Close()
	for rows.Next() {
		var s string
		if err = rows.Scan(&s); err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
