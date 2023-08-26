package service

import (
	"context"
	"encoding/json"
	"fmt"
	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"strconv"

	"github.com/pkg/errors"

	"go-gateway/app/web-svr/datasource-ng/admin/api"
	"go-gateway/app/web-svr/datasource-ng/admin/internal/model"
)

func (s *Service) ModelList(c context.Context, req *api.ModelListRequest) (*api.ModelListReply, error) {
	result, err := s.dao.ModelAll(c, req.Pn, req.Ps)
	if err != nil {
		return nil, err
	}
	total, err := s.dao.ModelAllCount(c)
	if err != nil {
		return nil, err
	}
	rsp := &api.ModelListReply{
		List:  result,
		Pn:    req.Pn,
		Ps:    int32(len(result)),
		Total: total,
	}
	return rsp, nil
}

func (s *Service) ModelAll(c context.Context, request *api.NoArgRequest) (*api.ModelAllReply, error) {
	result, err := s.dao.ModelNameAll(c)
	if err != nil {
		return nil, err
	}
	rsp := &api.ModelAllReply{
		List: result,
	}
	return rsp, nil
}

func (s *Service) ModelItemList(c context.Context, request *api.ModelItemListRequest) (*api.ModelItemListReply, error) {
	result, err := s.dao.ModelItemList(c, request.ModelName, request.Pn, request.Ps)
	if err != nil {
		return nil, err
	}
	total, err := s.dao.ModelItemListCount(c, request.ModelName)
	if err != nil {
		return nil, err
	}
	rsp := &api.ModelItemListReply{
		List:  result,
		Pn:    request.Pn,
		Ps:    int32(len(result)),
		Total: total,
	}
	return rsp, nil
}

func (s *Service) ModelDetail(c context.Context, request *api.ModelDetailRequest) (*api.ModelDetailReply, error) {
	modelFieldsRes, componentRes, err := s.relatedModelInfoByName(c, request.ModelName)
	result, err := s.modelSchema(c, request.ModelName, false, modelFieldsRes, componentRes)
	if err != nil {
		return nil, err
	}
	rsp := &api.ModelDetailReply{
		Detail: result,
	}
	return rsp, nil
}

func (s *Service) relatedModelInfoByName(c context.Context, modelName string) (map[string][]*api.ModelField, map[string]*api.ModelComponent, error) {
	fields, err := s.dao.ModelFieldByName(c, modelName)
	if err != nil {
		return nil, nil, err
	}
	modelFields := make(map[string][]*api.ModelField)
	modelFields[modelName] = fields
	componentUUID := []string{}
	for _, field := range fields {
		componentUUID = append(componentUUID, field.ComponentUuid)
		valueType := field.ValueType
		if model.IsGeneric(field.ValueType) {
			valueType, err = model.SplitGenericType(valueType)
			if err != nil {
				return nil, nil, err
			}
		}
		if model.IsReference(valueType) {
			if _, ok := modelFields[valueType]; ok {
				continue
			}
			fieldsTmp, err := s.dao.ModelFieldByName(c, valueType)
			if err != nil {
				return nil, nil, err
			}
			fields = append(fields, fieldsTmp...)
			modelFields[valueType] = fieldsTmp
		}
	}
	components, err := s.dao.Components(c, componentUUID)
	if err != nil {
		return nil, nil, err
	}
	return modelFields, components, nil
}

func (s *Service) modelSchema(c context.Context, modelName string, isList bool, modelFieldsRes map[string][]*api.ModelField, componentRes map[string]*api.ModelComponent) (*api.ModelSchema, error) {
	// Step 1. 查模型字段数据
	modelFields := modelFieldsRes[modelName]
	rsp := &api.ModelSchema{Type: model.MdlFieldTypeObject}
	if isList {
		rsp.Type = fmt.Sprintf(model.TypeGeneric, model.MdlFieldTypeObject)
	}
	properties := make(map[string]*api.ModelSchema)
	var cptUUIDs []string
	for _, item := range modelFields {
		if item.ComponentUuid != "" {
			cptUUIDs = append(cptUUIDs, item.ComponentUuid)
		}
	}
	// Step 2. 根据类型构建回包。基本类型直接构建，引用类型递归
	for _, field := range modelFields {
		generic := false
		valueType := field.ValueType
		if model.IsGeneric(valueType) {
			var err error
			generic = true
			if valueType, err = model.SplitGenericType(valueType); err != nil {
				return nil, err
			}
		}
		switch valueType {
		case model.MdlFieldTypeInt:
			if generic {
				valueType = fmt.Sprintf(model.TypeGeneric, valueType)
			}
			in := &api.ModelSchema{
				Type:        field.ValueType,
				Description: field.Description,
			}
			component := model.DefaultSchemaComponent()
			mdlComp, ok := componentRes[field.ComponentUuid]
			if ok {
				component.Type = mdlComp.Type
				component.Metadata = mdlComp.Metadata
				in.Required = mdlComp.Required
				in.DefaultString = mdlComp.DefaultString
				in.DefaultInt = mdlComp.DefaultInt
			}
			in.Component = component
			properties[field.Name] = in
		case model.MdlFieldTypeBool:
			if generic {
				valueType = fmt.Sprintf(model.TypeGeneric, valueType)
			}
			bo := &api.ModelSchema{
				Type:        field.ValueType,
				Description: field.Description,
			}
			component := model.DefaultSchemaComponent()
			mdlComp, ok := componentRes[field.ComponentUuid]
			if ok {
				component.Type = mdlComp.Type
				component.Metadata = mdlComp.Metadata
				bo.Required = mdlComp.Required
				bo.DefaultString = mdlComp.DefaultString
				bo.DefaultInt = mdlComp.DefaultInt
			}
			bo.Component = component
			properties[field.Name] = bo
		case model.MdlFieldTypeString:
			if generic {
				valueType = fmt.Sprintf(model.TypeGeneric, valueType)
			}
			st := &api.ModelSchema{
				Type:        field.ValueType,
				Description: field.Description,
			}
			component := model.DefaultSchemaComponent()
			mdlComp, ok := componentRes[field.ComponentUuid]
			if ok {
				component.Type = mdlComp.Type
				component.Metadata = mdlComp.Metadata
				st.Required = mdlComp.Required
				st.DefaultString = mdlComp.DefaultString
				st.DefaultInt = mdlComp.DefaultInt
			}
			st.Component = component
			properties[field.Name] = st
		default:
			schema, err := s.modelSchema(c, valueType, generic, modelFieldsRes, componentRes)
			if err != nil {
				return nil, err
			}
			properties[field.Name] = schema
		}
	}
	rsp.Properties = properties
	return rsp, nil
}

func (s *Service) ModelItemDetail(c context.Context, request *api.ModelItemDetailRequest) (*api.ItemReply, error) {
	return s.itemDetailByUUID(c, request.ItemUuid)
}

func (s *Service) itemDetailByUUID(c context.Context, uuid string) (*api.ItemReply, error) {
	item, err := s.dao.Item(c, uuid)
	if err != nil {
		return nil, err
	}
	modelFields, err := s.dao.ModelFieldByName(c, item.TypeName)
	if err != nil {
		return nil, err
	}
	values, err := s.dao.ItemValues(c, uuid)
	if err != nil {
		return nil, err
	}
	valuesMap := make(map[string]*api.ItemFieldValue)
	for _, item := range values {
		valuesMap[item.FieldName] = item
	}
	res, err := s.resolveItemByType(c, modelFields, valuesMap)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (s *Service) resolveItemByType(c context.Context, modelFields []*api.ModelField, values map[string]*api.ItemFieldValue) (*api.ItemReply, error) {
	rsp := &api.ItemReply{}
	res := &api.StructedItem{
		Item: make(map[string]*api.FieldValue),
	}
	for _, item := range modelFields {
		value, ok := values[item.Name]
		if !ok {
			empty, err := createEmptyValue(item.ValueType)
			if err != nil {
				return nil, err
			}
			res.Item[item.Name] = empty
			continue
		}
		// 字段类型是数组
		if model.IsGeneric(item.ValueType) {
			gen, err := s.resolveGeneral(c, item.ValueType, value)
			if err != nil {
				return nil, err
			}
			res.Item[item.Name] = gen
			continue
		}
		// 字段类型是基础类型或Object
		switch item.ValueType {
		case model.MdlFieldTypeInt:
			intTmp := &api.FieldValue{
				Value: &api.FieldValue_ValueInt{
					ValueInt: value.ValueInt,
				},
			}
			res.Item[item.Name] = intTmp
		case model.MdlFieldTypeString:
			strTmp := &api.FieldValue{
				Value: &api.FieldValue_ValueString{
					ValueString: value.ValueString,
				},
			}
			res.Item[item.Name] = strTmp
		case model.MdlFieldTypeBool:
			boolTmp := &api.FieldValue{
				Value: &api.FieldValue_ValueBool{
					ValueBool: model.TranBool(int(value.ValueInt)),
				},
			}
			res.Item[item.Name] = boolTmp
		default:
			modelFields, err := s.dao.ModelFieldByName(c, item.ValueType)
			if err != nil {
				return nil, err
			}
			values, err := s.dao.ItemValues(c, value.ValueItemUuid)
			if err != nil {
				return nil, err
			}
			valuesMap := make(map[string]*api.ItemFieldValue)
			for _, item := range values {
				valuesMap[item.FieldName] = item
			}
			objTmp, err := s.resolveItemByType(c, modelFields, valuesMap)
			if err != nil {
				return nil, err
			}
			res.Item[item.Name] = &api.FieldValue{
				Value: &api.FieldValue_ValueItem{
					ValueItem: objTmp,
				},
			}
		}
	}
	rsp.Item = &api.ItemReply_ValueStruct{
		ValueStruct: res,
	}
	return rsp, nil
}

func (s *Service) resolveGeneral(c context.Context, valueType string, value *api.ItemFieldValue) (*api.FieldValue, error) {
	subValues, err := s.dao.ItemValues(c, value.ValueItemUuid)
	if err != nil {
		log.Errorc(c, "resolveGeneral() s.dao.ItemValues() failed. error(%v)", err)
		return nil, err
	}
	innerType, err := model.SplitGenericType(valueType)
	if err != nil {
		return nil, err
	}
	rsp := &api.FieldValue{}
	switch innerType {
	case model.MdlFieldTypeInt:
		res := &api.IntList{}
		for _, item := range subValues {
			res.List = append(res.List, item.ValueInt)
		}
		rsp.Value = &api.FieldValue_ValueIntList{
			ValueIntList: res,
		}
	case model.MdlFieldTypeString:
		res := &api.StringList{}
		for _, item := range subValues {
			res.List = append(res.List, item.ValueString)
		}
		rsp.Value = &api.FieldValue_ValueStringList{
			ValueStringList: res,
		}
	case model.MdlFieldTypeBool:
		res := &api.BoolList{}
		for _, item := range subValues {
			res.List = append(res.List, model.TranBool(int(item.ValueInt)))
		}
		rsp.Value = &api.FieldValue_ValueBoolList{
			ValueBoolList: res,
		}
	default:
		res := &api.ItemList{}
		for _, item := range subValues {
			elementRes, err := s.itemDetailByUUID(c, item.ValueItemUuid)
			if err != nil {
				return nil, err
			}
			res.List = append(res.List, elementRes)
		}
		rsp.Value = &api.FieldValue_ValueItemList{
			ValueItemList: res,
		}
	}
	return rsp, nil
}

func createEmptyValue(valueType string) (*api.FieldValue, error) {
	rsp := &api.FieldValue{}
	if model.IsGeneric(valueType) {
		innerType, err := model.SplitGenericType(valueType)
		if err != nil {
			return nil, err
		}
		switch innerType {
		case model.MdlFieldTypeInt:
			rsp.Value = &api.FieldValue_ValueIntList{
				ValueIntList: &api.IntList{
					List: []int64{},
				},
			}
		case model.MdlFieldTypeString:
			rsp.Value = &api.FieldValue_ValueStringList{
				ValueStringList: &api.StringList{
					List: []string{},
				},
			}
		case model.MdlFieldTypeBool:
			rsp.Value = &api.FieldValue_ValueBoolList{
				ValueBoolList: &api.BoolList{
					List: []bool{},
				},
			}
		default:
			rsp.Value = &api.FieldValue_ValueItemList{
				ValueItemList: &api.ItemList{
					List: []*api.ItemReply{},
				},
			}
		}
		return rsp, nil
	}
	switch valueType {
	case model.MdlFieldTypeInt:
		rsp.Value = &api.FieldValue_ValueInt{
			ValueInt: 0,
		}
	case model.MdlFieldTypeString:
		rsp.Value = &api.FieldValue_ValueString{
			ValueString: "",
		}
	case model.MdlFieldTypeBool:
		rsp.Value = &api.FieldValue_ValueBool{
			ValueBool: false,
		}
	default:
		rsp.Value = &api.FieldValue_ValueItem{
			ValueItem: &api.ItemReply{},
		}
	}
	return rsp, nil
}

func (s *Service) ModelCreate(c context.Context, request *api.ModelCreateRequest) (*api.EmptyReply, error) {
	modelFields := &api.ModelSchema{}
	err := json.Unmarshal([]byte(request.ModelFields), &modelFields)
	if err != nil {
		log.Error("ModelCreate() json.Unmarshal() failed. error(%v)", err)
		return nil, err
	}
	exist, err := s.dao.ModelExists(c, request.ModelName)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, ecode.Error(ecode.RequestErr, "model_name已存在")
	}
	fieldArgs, componentArgs, err := s.procModelFieldArgs(c, request.ModelName, modelFields)
	if err != nil {
		return nil, err
	}
	err = s.dao.Transact(c, func(tx *xsql.Tx) error {
		err := s.dao.TxInsertModel(tx, request.ModelName, request.Description, request.CreatedBy)
		if err != nil {
			return err
		}
		for _, field := range fieldArgs {
			err := s.dao.TxInsertModelField(tx, field, request.CreatedBy)
			if err != nil {
				return err
			}
		}
		for _, component := range componentArgs {
			err := s.dao.TxInsertComponent(tx, component, request.CreatedBy)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return &api.EmptyReply{}, err
}

func (s *Service) procModelFieldArgs(c context.Context, modelName string, fields *api.ModelSchema) ([]*api.ModelField, []*api.ModelComponent, error) {
	FieldArgs := []*api.ModelField{}
	componentArgs := []*api.ModelComponent{}
	for name, field := range fields.Properties {
		valueType := field.Type
		var err error
		if model.IsGeneric(valueType) {
			valueType, err = model.SplitGenericType(valueType)
			if err != nil {
				return nil, nil, err
			}
		}
		switch valueType {
		case model.MdlFieldTypeString:
			fallthrough
		case model.MdlFieldTypeInt:
			fallthrough
		case model.MdlFieldTypeBool:
			fieldArg := &api.ModelField{
				ModelName:   modelName,
				Name:        name,
				Description: field.Description,
				ValueType:   field.Type,
			}
			component := model.NewModelComponent()
			component.DefaultInt = field.DefaultInt
			component.DefaultString = field.DefaultString
			component.Required = field.Required
			if field.Component != nil {
				component.Type = field.Component.Type
				component.Metadata = field.Component.Metadata
			}
			fieldArg.ComponentUuid = component.ComponentUuid
			componentArgs = append(componentArgs, component)
			FieldArgs = append(FieldArgs, fieldArg)
		default:
			exist, err := s.dao.ModelExists(c, valueType)
			if err != nil {
				return nil, nil, err
			}
			if !exist {
				return nil, nil, errors.New(fmt.Sprintf("模型(%s)不存在", valueType))
			}
			fieldArg := &api.ModelField{
				ModelName:   modelName,
				Name:        name,
				Description: field.Description,
				ValueType:   field.Type,
			}
			component := model.NewModelComponent()
			component.DefaultInt = field.DefaultInt
			component.DefaultString = field.DefaultString
			component.Required = field.Required
			if field.Component != nil {
				component.Type = field.Component.Type
				component.Metadata = field.Component.Metadata
			}
			fieldArg.ComponentUuid = component.ComponentUuid
			componentArgs = append(componentArgs, component)
			FieldArgs = append(FieldArgs, fieldArg)
		}
	}
	return FieldArgs, componentArgs, nil
}

func (s *Service) ModelItemCreate(c context.Context, request *api.ModelItemCreateRequest) (*api.EmptyReply, error) {
	values := make(map[string]interface{})
	err := json.Unmarshal([]byte(request.Value), &values)
	if err != nil {
		return nil, err
	}
	modelFieldsRes, componentRes, err := s.relatedModelInfoByName(c, request.ModelName)
	if err != nil {
		return nil, err
	}
	itemValueArgsParams := &model.GetItemValueArgsParams{
		ModelName:      request.ModelName,
		Business:       request.Business,
		Expirable:      request.Expirable,
		ExpireAt:       request.ExpireAt,
		Values:         values,
		ModelFieldsRes: modelFieldsRes,
		ComponentRes:   componentRes,
	}
	modelItemArgs, fieldValueArgs, err := s.getItemValueArgs(c, itemValueArgsParams)
	if err != nil {
		return nil, err
	}
	err = s.dao.Transact(c, func(tx *xsql.Tx) error {
		for _, modelItem := range modelItemArgs {
			if err := s.dao.TxInsertModelItem(tx, modelItem, request.CreatedBy); err != nil {
				return err
			}
		}
		for _, fieldValue := range fieldValueArgs {
			if err := s.dao.TxInsertItemValue(tx, fieldValue); err != nil {
				return err
			}
		}
		return nil
	})
	return &api.EmptyReply{}, err
}

func (s *Service) getItemValueArgs(c context.Context, params *model.GetItemValueArgsParams) ([]*api.ModelItem, []*api.ItemFieldValue, error) {
	modelItem := getModelItem(params.ModelName, params.Business, params.Expirable, params.ExpireAt)
	modelItemArgs, fieldValueArgs, err := s.getObjectArgs(c, params.ModelName, modelItem.ItemUuid, params.Values, params)
	if err != nil {
		return nil, nil, err
	}
	modelItemArgRes := []*api.ModelItem{modelItem}
	if modelItemArgs != nil {
		modelItemArgRes = append(modelItemArgRes, modelItemArgs...)
	}
	return modelItemArgRes, fieldValueArgs, nil
}

func (s *Service) getObjectArgs(c context.Context, modelName, uuid string, values map[string]interface{}, params *model.GetItemValueArgsParams) ([]*api.ModelItem, []*api.ItemFieldValue, error) {
	fields := params.ModelFieldsRes[modelName]
	fieldsValueRes := []*api.ItemFieldValue{}
	modelItemsRes := []*api.ModelItem{}
	for _, field := range fields {
		// 是数组
		if model.IsGeneric(field.ValueType) {
			va, ok := values[field.Name]
			if !ok {
				component, ok := params.ComponentRes[field.ComponentUuid]
				if ok && component.Required {
					return nil, nil, errors.New(fmt.Sprintf("%v字段为必填字段", field.Name))
				}
				continue
			}
			mdlargs, fiargs, err := s.procGenItemValue(c, field.ValueType, field.Name, uuid, va, params)
			if err != nil {
				return nil, nil, err
			}
			fieldsValueRes = append(fieldsValueRes, fiargs...)
			if mdlargs != nil {
				modelItemsRes = append(modelItemsRes, mdlargs...)
			}
			continue
		}
		// 非数组
		va, ok := values[field.Name]
		if !ok {
			component, ok := params.ComponentRes[field.ComponentUuid]
			if ok && component.Required {
				return nil, nil, errors.New(fmt.Sprintf("%v字段为必填字段", field.Name))
			}
			continue
		}
		modelItemArgs, fieldValueArgs, err := s.procItemValue(c, field.ValueType, field.Name, uuid, va, params)
		if err != nil {
			return nil, nil, err
		}
		fieldsValueRes = append(fieldsValueRes, fieldValueArgs...)
		if modelItemArgs != nil {
			modelItemsRes = append(modelItemsRes, modelItemArgs...)
		}
	}
	return modelItemsRes, fieldsValueRes, nil
}

func (s *Service) procGenItemValue(c context.Context, valueType, name, uuid string, valueIn interface{}, params *model.GetItemValueArgsParams) ([]*api.ModelItem, []*api.ItemFieldValue, error) {
	in, err := model.SplitGenericType(valueType)
	if err != nil {
		return nil, nil, err
	}
	modelItemRes := []*api.ModelItem{}
	fieldValueRes := []*api.ItemFieldValue{}
	item := getReferenceArg(name, uuid)
	fieldValueRes = append(fieldValueRes, item)
	values, ok := valueIn.([]interface{})
	if !ok {
		return nil, nil, errors.New(fmt.Sprintf("字段(%s)不是[]interface{}类型", name))
	}
	switch in {
	case model.MdlFieldTypeString:
		for key, value := range values {
			va, ok := value.(string)
			if !ok {
				return nil, nil, errors.New(fmt.Sprintf("字段(%s)存在非string类型元素", name))
			}
			arg := getStringArg(strconv.Itoa(key), item.ValueItemUuid, va)
			fieldValueRes = append(fieldValueRes, arg)
		}
	case model.MdlFieldTypeInt:
		for key, value := range values {
			va, ok := value.(float64)
			if !ok {
				return nil, nil, errors.New(fmt.Sprintf("字段(%s)存在非int类型元素", name))
			}
			arg := getIntArg(strconv.Itoa(key), item.ValueItemUuid, int64(va))
			fieldValueRes = append(fieldValueRes, arg)
		}
	case model.MdlFieldTypeBool:
		for key, value := range values {
			va, ok := value.(bool)
			if !ok {
				return nil, nil, errors.New(fmt.Sprintf("字段(%s)存在非bool类型元素", name))
			}
			arg := getBoolArg(strconv.Itoa(key), item.ValueItemUuid, va)
			fieldValueRes = append(fieldValueRes, arg)
		}
	default:
		for key, va := range values {
			v, ok := va.(map[string]interface{})
			if !ok {
				return nil, nil, errors.New(fmt.Sprintf("字段（%s)不是map[string]interface{}类型", name))
			}
			indexArg := getArrayIndexArg(key, item.ValueItemUuid)
			modelItemTmp := getModelItem(in, params.Business, params.Expirable, params.ExpireAt)
			modelItemTmp.ItemUuid = indexArg.ValueItemUuid
			modelItemResult, args, err := s.getObjectArgs(c, in, modelItemTmp.ItemUuid, v, params)
			if err != nil {
				return nil, nil, err
			}
			fieldValueRes = append(fieldValueRes, indexArg)
			fieldValueRes = append(fieldValueRes, args...)
			modelItemRes = append(modelItemRes, modelItemTmp)
			if modelItemResult != nil {
				modelItemRes = append(modelItemRes, modelItemResult...)
			}
		}
	}
	return modelItemRes, fieldValueRes, nil
}

func (s *Service) procItemValue(c context.Context, valueType, name, uuid string, value interface{}, params *model.GetItemValueArgsParams) ([]*api.ModelItem, []*api.ItemFieldValue, error) {
	fieldValueRes := []*api.ItemFieldValue{}
	modelItemRes := []*api.ModelItem{}
	switch valueType {
	case model.MdlFieldTypeString:
		v, ok := value.(string)
		if !ok {
			return nil, nil, errors.New(fmt.Sprintf("字段(%s)不是string类型", name))
		}
		arg := getStringArg(name, uuid, v)
		fieldValueRes = append(fieldValueRes, arg)
	case model.MdlFieldTypeInt:
		v, ok := value.(float64)
		if !ok {
			return nil, nil, errors.New(fmt.Sprintf("字段(%s)不是int64类型", name))
		}
		intValue := int64(v)
		arg := getIntArg(name, uuid, intValue)
		fieldValueRes = append(fieldValueRes, arg)
	case model.MdlFieldTypeBool:
		v, ok := value.(bool)
		if !ok {
			return nil, nil, errors.New(fmt.Sprintf("字段(%s)不是bool类型", name))
		}
		arg := getBoolArg(name, uuid, v)
		fieldValueRes = append(fieldValueRes, arg)
	default:
		v, ok := value.(map[string]interface{})
		if !ok {
			return nil, nil, errors.New(fmt.Sprintf("字段(%s)不是map[string]interface{}类型", name))
		}
		item := getReferenceArg(name, uuid)
		modelItemArgs, fieldValueArgs, err := s.getObjectArgs(c, valueType, item.ValueItemUuid, v, params)
		if err != nil {
			return nil, nil, err
		}
		fieldValueRes = append(fieldValueRes, item)
		fieldValueRes = append(fieldValueRes, fieldValueArgs...)
		if modelItemArgs != nil {
			modelItemRes = append(modelItemRes, modelItemArgs...)
		}
	}
	return modelItemRes, fieldValueRes, nil
}

func getReferenceArg(name, uuid string) *api.ItemFieldValue {
	res := &api.ItemFieldValue{
		ItemUuid:      uuid,
		FieldName:     name,
		ValueItemUuid: model.UUID4(),
	}
	return res
}

func getStringArg(name, uuid, value string) *api.ItemFieldValue {
	res := &api.ItemFieldValue{
		ItemUuid:    uuid,
		FieldName:   name,
		ValueString: value,
	}
	return res
}

func getIntArg(name, uuid string, value int64) *api.ItemFieldValue {
	res := &api.ItemFieldValue{
		ItemUuid:  uuid,
		FieldName: name,
		ValueInt:  value,
	}
	return res
}

func getBoolArg(name, uuid string, value bool) *api.ItemFieldValue {
	res := &api.ItemFieldValue{
		ItemUuid:  uuid,
		FieldName: name,
		ValueInt:  model.BoolToInt64(value),
	}
	return res
}

func getArrayIndexArg(index int, uuid string) *api.ItemFieldValue {
	res := &api.ItemFieldValue{
		ItemUuid:      uuid,
		FieldName:     strconv.Itoa(index),
		ValueItemUuid: model.UUID4(),
	}
	return res
}

func getModelItem(modelName, business string, expirable int32, expireAt int64) *api.ModelItem {
	arg := &api.ModelItem{
		Business:  business,
		ItemUuid:  model.UUID4(),
		TypeName:  modelName,
		Expirable: expirable,
		ExpireAt:  expireAt,
	}
	return arg
}
