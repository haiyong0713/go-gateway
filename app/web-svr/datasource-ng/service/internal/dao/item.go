package dao

import (
	"context"

	"go-common/library/ecode"
	"go-gateway/app/web-svr/datasource-ng/service/internal/model"

	"github.com/pkg/errors"
)

// GetItem is
func (d *Dao) GetItem(ctx context.Context, itemUUID string) (*model.ModelItem, error) {
	allItem := d.localcache.GetAllModelItemMap()
	item, ok := allItem[itemUUID]
	if !ok {
		return nil, errors.Wrap(ecode.NothingFound, itemUUID)
	}
	return item, nil
}

// GetModel is
func (d *Dao) GetModel(ctx context.Context, modelName string) (*model.Model, error) {
	allModel := d.localcache.GetAllModelMap()
	model, ok := allModel[modelName]
	if !ok {
		return nil, errors.Wrap(ecode.NothingFound, modelName)
	}
	return model, nil
}

// GetModelField is
func (d *Dao) GetModelField(ctx context.Context, modelName string) (map[string]*model.ModelField, error) {
	allModelField := d.localcache.GetAllModelField()
	fields, ok := allModelField[modelName]
	if !ok {
		return nil, nil
	}
	return fields, nil
}

// GetModelFieldValue is
func (d *Dao) GetModelFieldValue(ctx context.Context, itemUUID string) (map[string]*model.ItemFieldValue, error) {
	allItemFieldValue := d.localcache.GetAllItemFieldValueMap()
	fvalues, ok := allItemFieldValue[itemUUID]
	if !ok {
		return nil, nil
	}
	return fvalues, nil
}
