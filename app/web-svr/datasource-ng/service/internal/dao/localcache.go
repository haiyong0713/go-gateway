package dao

import (
	"context"
	"sync/atomic"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/datasource-ng/service/internal/model"
)

type modelCache struct {
	storeList []*model.Model
	storeMap  map[string]*model.Model
}

type modelFieldCache struct {
	storeMap map[string]map[string]*model.ModelField
}

type modelItemCache struct {
	storeList []*model.ModelItem
	storeMap  map[string]*model.ModelItem
}

type itemFieldValueCache struct {
	storeMap map[string]map[string]*model.ItemFieldValue
}

type localcache interface {
	GetAllModel() []*model.Model
	GetAllModelMap() map[string]*model.Model
	GetAllModelField() map[string]map[string]*model.ModelField
	GetAllModelItem() []*model.ModelItem
	GetAllModelItemMap() map[string]*model.ModelItem
	GetAllItemFieldValueMap() map[string]map[string]*model.ItemFieldValue

	SetAllModel([]*model.Model)
	SetAllModelField([]*model.ModelField)
	SetAllModelItem([]*model.ModelItem)
	SetAllItemFieldValue([]*model.ItemFieldValue)
}

type cacheImpl struct {
	allModel          atomic.Value
	allModelField     atomic.Value
	allModelItem      atomic.Value
	allItemFieldValue atomic.Value
}

var _ localcache = &cacheImpl{}

// SetAllModel is
func (c *cacheImpl) SetAllModel(in []*model.Model) {
	storeMap := make(map[string]*model.Model, len(in))
	for _, i := range in {
		storeMap[i.Name] = i
	}
	c.allModel.Store(&modelCache{
		storeList: in,
		storeMap:  storeMap,
	})
}

// SetAllModel is
func (c *cacheImpl) SetAllModelField(in []*model.ModelField) {
	storeMap := make(map[string]map[string]*model.ModelField, len(in))
	for _, i := range in {
		_, ok := storeMap[i.ModelName]
		if !ok {
			storeMap[i.ModelName] = make(map[string]*model.ModelField)
		}
		storeMap[i.ModelName][i.Name] = i
	}
	c.allModelField.Store(&modelFieldCache{
		storeMap: storeMap,
	})
}

// SetAllModelItem is
func (c *cacheImpl) SetAllModelItem(in []*model.ModelItem) {
	storeMap := make(map[string]*model.ModelItem, len(in))
	for _, i := range in {
		storeMap[i.ItemUuid] = i
	}
	c.allModelItem.Store(&modelItemCache{
		storeList: in,
		storeMap:  storeMap,
	})
}

// SetAllModelItem is
func (c *cacheImpl) SetAllItemFieldValue(in []*model.ItemFieldValue) {
	storeMap := make(map[string]map[string]*model.ItemFieldValue, len(in))
	for _, i := range in {
		_, ok := storeMap[i.ItemUuid]
		if !ok {
			storeMap[i.ItemUuid] = make(map[string]*model.ItemFieldValue)
		}
		storeMap[i.ItemUuid][i.FieldName] = i
	}
	c.allItemFieldValue.Store(&itemFieldValueCache{
		storeMap: storeMap,
	})
}

// GetAllModel is
func (c *cacheImpl) GetAllModel() []*model.Model {
	cache := c.allModel.Load().(*modelCache)
	return cache.storeList
}

// GetAllModelGetAllModelMap is
func (c *cacheImpl) GetAllModelMap() map[string]*model.Model {
	cache := c.allModel.Load().(*modelCache)
	return cache.storeMap
}

// GetAllModelField is
func (c *cacheImpl) GetAllModelField() map[string]map[string]*model.ModelField {
	cache := c.allModelField.Load().(*modelFieldCache)
	return cache.storeMap
}

// GetAllModelItem is
func (c *cacheImpl) GetAllModelItem() []*model.ModelItem {
	cache := c.allModelItem.Load().(*modelItemCache)
	return cache.storeList
}

// GetAllModelItemMap is
func (c *cacheImpl) GetAllModelItemMap() map[string]*model.ModelItem {
	cache := c.allModelItem.Load().(*modelItemCache)
	return cache.storeMap
}

// GetAllItemFieldMap is
func (c *cacheImpl) GetAllItemFieldValueMap() map[string]map[string]*model.ItemFieldValue {
	cache := c.allItemFieldValue.Load().(*itemFieldValueCache)
	return cache.storeMap
}

// GetAllModel is
func (d *Dao) GetAllModel(ctx context.Context) ([]*model.Model, error) {
	rows, err := d.db.Query(ctx, `SELECT id,name,description FROM data_model ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*model.Model{}
	for rows.Next() {
		item := &model.Model{}
		if err := rows.Scan(&item.Id, &item.Name, &item.Description); err != nil {
			log.Warn("Failed to scan data model: %+v", err)
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

// GetAllModelField is
func (d *Dao) GetAllModelField(ctx context.Context) ([]*model.ModelField, error) {
	rows, err := d.db.Query(ctx, `SELECT id,model_name,name,description,value_type FROM data_model_field ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*model.ModelField{}
	for rows.Next() {
		item := &model.ModelField{}
		if err := rows.Scan(&item.Id, &item.ModelName, &item.Name, &item.Description, &item.ValueType); err != nil {
			log.Warn("Failed to scan data model field: %+v", err)
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

// GetAllModelItem is
func (d *Dao) GetAllModelItem(ctx context.Context) ([]*model.ModelItem, error) {
	rows, err := d.db.Query(ctx, `SELECT id,business,item_uuid,type_name,expirable,expire_at FROM data_model_item ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*model.ModelItem{}
	for rows.Next() {
		item := &model.ModelItem{}
		if err := rows.Scan(&item.Id, &item.Business, &item.ItemUuid, &item.TypeName, &item.Expirable, &item.ExpireAt); err != nil {
			log.Warn("Failed to scan data model item: %+v", err)
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

// GetAllItemFieldValue is
func (d *Dao) GetAllItemFieldValue(ctx context.Context) ([]*model.ItemFieldValue, error) {
	rows, err := d.db.Query(ctx, `SELECT id,item_uuid,field_name,value_int,value_string,value_item_uuid FROM data_item_field_value ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []*model.ItemFieldValue{}
	for rows.Next() {
		item := &model.ItemFieldValue{}
		if err := rows.Scan(&item.Id, &item.ItemUuid, &item.FieldName, &item.ValueInt, &item.ValueString, &item.ValueItemUuid); err != nil {
			log.Warn("Failed to scan item field value: %+v", err)
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func (d *Dao) cacheloadproc() {
	for {
		log.Info("Load datasource meta cache at: %+v", time.Now())

		func() {
			allModel, err := d.GetAllModel(context.Background())
			if err != nil {
				log.Warn("Failed to load all model: %+v", err)
				return
			}
			d.localcache.SetAllModel(allModel)
		}()

		func() {
			allModelField, err := d.GetAllModelField(context.Background())
			if err != nil {
				log.Warn("Failed to load all model field: %+v", err)
				return
			}
			d.localcache.SetAllModelField(allModelField)
		}()

		func() {
			allModelItem, err := d.GetAllModelItem(context.Background())
			if err != nil {
				log.Warn("Failed to load all model item: %+v", err)
				return
			}
			d.localcache.SetAllModelItem(allModelItem)
		}()

		func() {
			allItemFieldValue, err := d.GetAllItemFieldValue(context.Background())
			if err != nil {
				log.Warn("Failed to load all item field value: %+v", err)
				return
			}
			d.localcache.SetAllItemFieldValue(allItemFieldValue)
		}()

		time.Sleep(time.Second * 60)
	}
}

func (d *Dao) initCache() {
	cache := &cacheImpl{}

	allModel, err := d.GetAllModel(context.Background())
	if err != nil {
		panic(err)
	}
	cache.SetAllModel(allModel)

	allModelField, err := d.GetAllModelField(context.Background())
	if err != nil {
		panic(err)
	}
	cache.SetAllModelField(allModelField)

	allModelItem, err := d.GetAllModelItem(context.Background())
	if err != nil {
		panic(err)
	}
	cache.SetAllModelItem(allModelItem)

	allItemFieldValue, err := d.GetAllItemFieldValue(context.Background())
	if err != nil {
		panic(err)
	}
	cache.SetAllItemFieldValue(allItemFieldValue)

	d.localcache = cache
}
