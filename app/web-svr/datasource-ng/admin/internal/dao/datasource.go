package dao

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/datasource-ng/admin/api"
)

const (
	_modelAll           = "SELECT id,name,description,ctime,created_by,is_deleted FROM data_model ORDER BY ctime DESC LIMIT ? OFFSET ?"
	_modelAllCount      = "SELECT count(1) FROM data_model"
	_modelNameAll       = "SELECT name FROM data_model ORDER BY ctime DESC"
	_modelItemList      = "SELECT id,business,item_uuid,type_name,expirable,expire_at,ctime,mtime,created_by FROM data_model_item WHERE is_element=0 %s ORDER BY id DESC LIMIT ? OFFSET ?"
	_modelItemListCount = "SELECT count(1) FROM data_model_item WHERE is_element=0 %s"
	_modelFieldByName   = "SELECT id,model_name,name,component_uuid,description,value_type,ctime,created_by FROM data_model_field WHERE model_name=? ORDER BY id DESC"
	_components         = "SELECT id,component_uuid,type,metadata,default_string,default_int,required,ctime,mtime FROM data_model_component WHERE component_uuid IN(%s)"
	_itemValues         = "SELECT id,item_uuid,field_name,value_int,value_string,value_item_uuid,ctime,mtime FROM data_item_field_value WHERE item_uuid=?"
	_item               = "SELECT id,business,item_uuid,type_name,expirable,expire_at,ctime,mtime,created_by FROM data_model_item WHERE item_uuid=?"
	_modelByName        = "SELECT id,name,description,ctime,mtime,is_deleted,created_by FROM data_model WHERE name=?"
	_modelExists        = "SELECT count(1) FROM data_model WHERE name=?"
	_txAddModel         = "INSERT INTO data_model(name,description,created_by) VALUES(?,?,?)"
	_txInsertModelField = "INSERT INTO data_model_field(model_name,name,description,value_type,created_by,component_uuid) VALUES(?,?,?,?,?,?)"
	_txInsertComponent  = "INSERT INTO data_model_component(component_uuid,type,metadata,required,default_string,default_int,created_by) VALUES(?,?,?,?,?,?,?)"
	_txInsertModelItem  = "INSERT INTO data_model_item(business,item_uuid,type_name,expirable,expire_at,created_by) VALUES(?,?,?,?,?,?)"
	_txInsertItemValue  = "INSERT INTO data_item_field_value(item_uuid,field_name,value_int,value_string,value_item_uuid) VALUES(?,?,?,?,?)"
)

func (d *Dao) ModelAll(c context.Context, pn, ps int32) ([]*api.Model, error) {
	rows, err := d.db.Query(c, _modelAll, ps, (pn-1)*ps)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*api.Model
	for rows.Next() {
		item := &api.Model{}
		if err := rows.Scan(&item.Id, &item.Name, &item.Description, &item.Ctime, &item.CreatedBy, &item.IsDeleted); err != nil {
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

func (d *Dao) ModelAllCount(c context.Context) (int64, error) {
	row := d.db.QueryRow(c, _modelAllCount)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (d *Dao) ModelNameAll(c context.Context) ([]string, error) {
	rows, err := d.db.Query(c, _modelNameAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []string
	for rows.Next() {
		var str string
		if err := rows.Scan(&str); err != nil {
			return nil, err
		}
		res = append(res, str)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModelItemList(c context.Context, modelName string, pn, ps int32) ([]*api.ModelItem, error) {
	var (
		where string
		par   []interface{}
	)
	if modelName != "" {
		where = "AND type_name=?"
		par = append(par, modelName)
	}
	par = append(par, ps)
	par = append(par, (pn-1)*ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_modelItemList, where), par...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*api.ModelItem
	for rows.Next() {
		item := &api.ModelItem{}
		if err := rows.Scan(&item.Id, &item.Business, &item.ItemUuid, &item.TypeName, &item.Expirable, &item.ExpireAt, &item.Ctime,
			&item.Mtime, &item.CreatedBy); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModelItemListCount(c context.Context, modelName string) (int64, error) {
	var (
		where string
		par   []interface{}
	)
	if modelName != "" {
		where = "AND type_name=?"
		par = append(par, modelName)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_modelItemListCount, where), par...)
	var total int64
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (d *Dao) ModelFieldByName(c context.Context, modelName string) ([]*api.ModelField, error) {
	rows, err := d.db.Query(c, _modelFieldByName, modelName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*api.ModelField
	for rows.Next() {
		field := &api.ModelField{}
		if err := rows.Scan(&field.Id, &field.ModelName, &field.Name, &field.ComponentUuid, &field.Description, &field.ValueType,
			&field.Ctime, &field.CreatedBy); err != nil {
			return nil, err
		}
		res = append(res, field)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) Components(c context.Context, uuids []string) (map[string]*api.ModelComponent, error) {
	sqls := []string{}
	args := []interface{}{}
	for _, uuid := range uuids {
		sqls = append(sqls, "?")
		args = append(args, uuid)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_components, strings.Join(sqls, ",")), args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return make(map[string]*api.ModelComponent), nil
		}
		return nil, err
	}
	defer rows.Close()
	var res []*api.ModelComponent
	for rows.Next() {
		com := &api.ModelComponent{}
		if err := rows.Scan(&com.Id, &com.ComponentUuid, &com.Type, &com.Metadata, &com.DefaultString, &com.DefaultInt, &com.Required,
			&com.Ctime, &com.Mtime); err != nil {
			return nil, err
		}
		res = append(res, com)
	}
	rsp := make(map[string]*api.ModelComponent)
	for _, item := range res {
		rsp[item.ComponentUuid] = item
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return rsp, nil
}

func (d *Dao) ItemValues(c context.Context, uuid string) ([]*api.ItemFieldValue, error) {
	rows, err := d.db.Query(c, _itemValues, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []*api.ItemFieldValue{}
	for rows.Next() {
		item := &api.ItemFieldValue{}
		if err := rows.Scan(&item.Id, &item.ItemUuid, &item.FieldName, &item.ValueInt, &item.ValueString, &item.ValueItemUuid,
			&item.Ctime, &item.Mtime); err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return res, err
}

func (d *Dao) Item(c context.Context, uuid string) (*api.ModelItem, error) {
	row := d.db.QueryRow(c, _item, uuid)
	item := &api.ModelItem{}
	err := row.Scan(&item.Id, &item.Business, &item.ItemUuid, &item.TypeName, &item.Expirable, &item.ExpireAt, &item.Ctime, &item.Mtime, &item.CreatedBy)
	if err != nil {
		return nil, err
	}
	return item, nil

}

func (d *Dao) ModelByName(c context.Context, modelName string) (*api.Model, error) {
	row := d.db.QueryRow(c, _modelByName, modelName)
	res := &api.Model{}
	if err := row.Scan(&res.Id, &res.Name, &res.Description, &res.Ctime, &res.Mtime, &res.IsDeleted, &res.CreatedBy); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ModelExists(c context.Context, modelName string) (bool, error) {
	row := d.db.QueryRow(c, _modelExists, modelName)
	var total int64
	if err := row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if total != 0 {
		return true, nil
	}
	return false, nil

}

func (d *Dao) Transact(c context.Context, txFunc func(tx *xsql.Tx) error) error {
	tx, err := d.TxBegin(c)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
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

func (d *Dao) TxInsertModel(tx *xsql.Tx, modelName, desc, createdBy string) error {
	_, err := tx.Exec(_txAddModel, modelName, desc, createdBy)
	if err != nil {
		log.Error("dao.TxAddModel() DB INSERT failed. error(%+v)", err)
	}
	return err
}

func (d *Dao) TxInsertModelField(tx *xsql.Tx, field *api.ModelField, createdBy string) error {
	_, err := tx.Exec(_txInsertModelField, field.ModelName, field.Name, field.Description, field.ValueType, createdBy, field.ComponentUuid)
	if err != nil {
		log.Error("dao.TxInsertModelFields() DB INSERT failed. error(%+v)", err)
	}
	return err
}

func (d *Dao) TxInsertComponent(tx *xsql.Tx, component *api.ModelComponent, createdBy string) error {
	_, err := tx.Exec(_txInsertComponent, component.ComponentUuid, component.Type, component.Metadata, component.Required, component.DefaultString, component.DefaultInt, createdBy)
	if err != nil {
		log.Error("dao.TxInsertComponents() DB INSERT failed. error(%+v)", err)
	}
	return err
}

func (d *Dao) TxInsertModelItem(tx *xsql.Tx, arg *api.ModelItem, createdBy string) error {
	_, err := tx.Exec(_txInsertModelItem, arg.Business, arg.ItemUuid, arg.TypeName, arg.Expirable, arg.ExpireAt, createdBy)
	if err != nil {
		log.Error("dao.TxInsertModelItem() DB INSERT failed. error(%+v)", err)
	}
	return err
}

func (d *Dao) TxInsertItemValue(tx *xsql.Tx, value *api.ItemFieldValue) error {
	_, err := tx.Exec(fmt.Sprintf(_txInsertItemValue), value.ItemUuid, value.FieldName, value.ValueInt, value.ValueString, value.ValueItemUuid)
	if err != nil {
		log.Error("dao.TxInsertItemValues() DB INSERT failed. error(%+v)", err)
	}
	return err
}
