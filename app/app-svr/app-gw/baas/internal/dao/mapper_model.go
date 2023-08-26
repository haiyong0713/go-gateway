package dao

import (
	"context"
	"fmt"
	"strings"

	xsql "go-common/library/database/sql"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-gw/baas/api"
	"go-gateway/app/app-svr/app-gw/baas/internal/model"
)

const (
	_modelAll             = "SELECT id,name,description,tree_id,ctime FROM mapper_model where tree_id=? ORDER BY ctime DESC LIMIT ? OFFSET ?"
	_modelAllCount        = "SELECT count(1) FROM mapper_model where tree_id=?"
	_modelByName          = "SELECT id,name,description,tree_id,ctime FROM mapper_model WHERE name=? and tree_id=?"
	_modelFieldByName     = "SELECT id,model_name,field_name,field_type,ctime,json_alias FROM mapper_model_field WHERE model_name=? AND is_delete=0 ORDER BY id DESC"
	_modelField           = "SELECT id,model_name,field_name,field_type,ctime,json_alias FROM mapper_model_field WHERE is_delete=0 ORDER BY id DESC"
	_addModel             = "INSERT INTO mapper_model(name,description,tree_id) VALUES (?,?,?)"
	_addModelField        = "INSERT INTO mapper_model_field(model_name,field_name,field_type,json_alias) VALUES (?,?,?,?)"
	_updateModelField     = "UPDATE mapper_model_field SET field_name=?,field_type=?,json_alias=? WHERE id=?"
	_delModelField        = "UPDATE mapper_model_field SET is_delete = 1 WHERE id=?"
	_modelFieldRule       = "SELECT id,model_name,field_name,datasource_api,external_rule,rule_type,value_source,ctime FROM mapper_model_field_rule"
	_addModelFieldRule    = "INSERT INTO mapper_model_field_rule(model_name,field_name,datasource_api,external_rule,rule_type,value_source) VALUES %s"
	_updateModelFieldRule = "UPDATE mapper_model_field_rule SET datasource_api=?,external_rule=?,rule_type=?,value_source=? WHERE id=?"
)

func (d *dao) ModelAll(ctx context.Context, param *api.ModelListRequest) ([]*api.MapperModel, error) {
	rows, err := d.db.Query(ctx, _modelAll, param.TreeId, param.Ps, (param.Pn-1)*param.Ps)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*api.MapperModel
	for rows.Next() {
		item := &api.MapperModel{}
		if err := rows.Scan(&item.Id, &item.Name, &item.Description, &item.TreeId, &item.Ctime); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return result, nil
}

func (d *dao) ModelAllCount(ctx context.Context, treeID int64) (int64, error) {
	row := d.db.QueryRow(ctx, _modelAllCount, treeID)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (d *dao) ModelByName(ctx context.Context, modelName string, treeID int64) (*api.MapperModel, error) {
	row := d.db.QueryRow(ctx, _modelByName, modelName, treeID)
	out := &api.MapperModel{}
	if err := row.Scan(&out.Id, &out.Name, &out.Description, &out.TreeId, &out.Ctime); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *dao) ModelFieldByName(ctx context.Context, modelName string) ([]*api.MapperModelField, error) {
	rows, err := d.db.Query(ctx, _modelFieldByName, modelName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*api.MapperModelField
	for rows.Next() {
		field := &api.MapperModelField{}
		if err := rows.Scan(&field.Id, &field.ModelName, &field.FieldName, &field.FieldType, &field.Ctime, &field.JsonAlias); err != nil {
			return nil, err
		}
		out = append(out, field)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return out, nil
}

func (d *dao) ModelField(ctx context.Context) ([]*api.MapperModelField, error) {
	rows, err := d.db.Query(ctx, _modelField)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*api.MapperModelField
	for rows.Next() {
		field := &api.MapperModelField{}
		if err := rows.Scan(&field.Id, &field.ModelName, &field.FieldName, &field.FieldType, &field.Ctime, &field.JsonAlias); err != nil {
			return nil, err
		}
		out = append(out, field)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return out, nil
}

func (d *dao) Transact(ctx context.Context, txFunc func(tx *xsql.Tx) error) error {
	tx, err := d.TxBegin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			if err := tx.Rollback(); err != nil {
				log.Error("Failed to rollback: %+v", err)
			}
			log.Error("%+v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%+v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%+v)", err)
		}
	}()
	err = txFunc(tx)
	return err
}

func (d *dao) TxInsertModel(tx *xsql.Tx, modelName, desc string, treeID int64) error {
	_, err := tx.Exec(_addModel, modelName, desc, treeID)
	return err
}

func (d *dao) TxInsertModelField(tx *xsql.Tx, field *api.ModelField) error {
	_, err := tx.Exec(_addModelField, field.ModelName, field.FieldName, field.FieldType, field.JsonAlias)
	return err
}

func (d *dao) AddModelField(ctx context.Context, field *api.AddModelFieldRequest) error {
	_, err := d.db.Exec(ctx, _addModelField, field.ModelName, field.FieldName, field.FieldType, field.JsonAlias)
	return err
}

func (d *dao) UpdateModelField(ctx context.Context, field *api.UpdateModelFieldRequest) error {
	_, err := d.db.Exec(ctx, _updateModelField, field.FieldName, field.FieldType, field.JsonAlias, field.Id)
	return err
}

func (d *dao) DelModelField(ctx context.Context, id int64) error {
	_, err := d.db.Exec(ctx, _delModelField, id)
	return err
}

func (d *dao) ModelFieldRule(ctx context.Context) (map[string]*api.MapperModelFieldRule, error) {
	rows, err := d.db.Query(ctx, _modelFieldRule)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]*api.MapperModelFieldRule)
	for rows.Next() {
		item := &api.MapperModelFieldRule{}
		if err := rows.Scan(&item.Id, &item.ModelName, &item.FieldName, &item.DatasourceApi, &item.ExternalRule,
			&item.RuleType, &item.ValueSource, &item.Ctime); err != nil {
			return nil, err
		}
		result[model.FieldRuleKey(item.ModelName, item.FieldName, item.DatasourceApi)] = item
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return result, nil
}

func (d *dao) AddFieldRule(ctx context.Context, params []*model.ItemFieldRule) error {
	var (
		sqls = make([]string, 0, len(params))
		args = make([]interface{}, 0, len(params)*2)
	)
	for _, param := range params {
		sqls = append(sqls, "(?,?,?,?,?,?)")
		args = append(args, param.ModelName, param.FieldName, param.DatasourceApi, param.ExternalRule, param.RuleType,
			param.ValueSource)
	}
	_, err := d.db.Exec(ctx, fmt.Sprintf(_addModelFieldRule, strings.Join(sqls, ",")), args...)
	if err != nil {
		return err
	}
	return nil
}

func (d *dao) UpdateFieldRule(ctx context.Context, param *api.UpdateModelFieldRuleRequest) error {
	_, err := d.db.Exec(ctx, _updateModelFieldRule, param.DatasourceApi, param.ExternalRule, param.RuleType,
		param.ValueSource, param.Id)
	return err
}
