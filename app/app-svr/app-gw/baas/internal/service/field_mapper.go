package service

import (
	"context"
	"encoding/json"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/baas/api"
	"go-gateway/app/app-svr/app-gw/baas/internal/model"
	"go-gateway/app/app-svr/app-gw/baas/utils/sets"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

func (s *Service) ModelList(ctx context.Context, req *api.ModelListRequest) (*api.ModelListReply, error) {
	result, err := s.dao.ModelAll(ctx, req)
	if err != nil {
		return nil, err
	}
	total, err := s.dao.ModelAllCount(ctx, req.TreeId)
	if err != nil {
		return nil, err
	}
	out := &api.ModelListReply{
		List:  constructMapperModels(result),
		Pn:    req.Pn,
		Ps:    req.Ps,
		Total: total,
	}
	return out, nil
}

func constructMapperModels(in []*api.MapperModel) []*api.MapperModelItem {
	out := make([]*api.MapperModelItem, 0)
	for _, val := range in {
		out = append(out, api.ConstructMapperModelItem(val))
	}
	return out
}

func (s *Service) ModelItemList(ctx context.Context, req *api.ModelItemListRequest) (*api.ModelItemListReply, error) {
	export, err := s.exportList(ctx, req.TreeId, req.ExportApi)
	if err != nil {
		return nil, err
	}
	if len(export) == 0 {
		return nil, ecode.Error(ecode.NothingFound, fmt.Sprintf("找不到导出api对应的数据: %s", req.ExportApi))
	}
	//mdl, err := s.dao.ModelByName(ctx, req.ModelName, req.TreeId)
	//if err != nil {
	//	return nil, err
	//}
	//if mdl.Name != export[0].ModelName {
	//	return nil, ecode.Error(ecode.NothingFound,
	//		fmt.Sprintf("导出模型与所查模型不一致: %s, %s", mdl.Name, export[0].ModelName))
	//}
	importMap, err := s.dao.ImportByExportIds(ctx, []int64{export[0].Id})
	if err != nil {
		return nil, err
	}
	datasourceAPISet := sets.String{}
	for _, v := range importMap[export[0].Id] {
		datasourceAPISet.Insert(v.DatasourceApi)
	}
	out := &api.ModelItemListReply{
		ModelName:     req.ModelName,
		ExportApi:     req.ExportApi,
		DatasourceApi: datasourceAPISet.List(),
	}
	fields, err := s.dao.ModelFieldByName(ctx, req.ModelName)
	if err != nil {
		return nil, err
	}
	rules, err := s.dao.ModelFieldRule(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*api.FieldRuleMetadata, 0, len(fields))
	for _, field := range fields {
		item := setModelItem(api.ConstructModelField(field), rules, datasourceAPISet)
		list = append(list, item)
	}
	out.List = list
	return out, nil
}

func setModelItem(field *api.ModelField, rules map[string]*api.MapperModelFieldRule, datasourceAPISet sets.String) *api.FieldRuleMetadata {
	out := &api.FieldRuleMetadata{
		Id:        field.Id,
		ModelName: field.ModelName,
		FieldName: field.FieldName,
		FieldType: field.FieldType,
		Ctime:     field.Ctime,
	}
	for _, datasourceAPI := range datasourceAPISet.List() {
		rule, ok := rules[model.FieldRuleKey(field.ModelName, field.FieldName, datasourceAPI)]
		if !ok {
			continue
		}
		out.RuleId = rule.Id
		out.DatasourceApi = rule.DatasourceApi
		out.ExternalRule = rule.ExternalRule.String
		out.ValueSource = rule.ValueSource
		out.RuleType = rule.RuleType
	}
	return out
}

func (s *Service) ModelDetail(ctx context.Context, request *api.ModelDetailRequest) (*api.ModelDetailReply, error) {
	result, err := s.modelSchema(ctx, request.ModelName, false)
	if err != nil {
		return nil, err
	}
	rsp := &api.ModelDetailReply{
		Detail: result,
	}
	return rsp, nil
}

func (s *Service) modelSchema(ctx context.Context, modelName string, isList bool) (*api.ModelSchema, error) {
	// Step 1. 查模型下所有字段
	modelFields, err := s.dao.ModelFieldByName(ctx, modelName)
	if err != nil {
		return nil, err
	}
	if len(modelFields) == 0 {
		return nil, ecode.Error(ecode.NothingFound, "找不到对应的模型")
	}
	rsp := &api.ModelSchema{Type: model.MdlFieldTypeObject}
	if isList {
		rsp.Type = fmt.Sprintf(model.TypeGeneric, model.MdlFieldTypeObject)
	}
	properties := make(map[string]*api.ModelSchema)
	// Step 2. 根据类型构建回包。基本类型直接构建，引用类型递归
	for _, field := range modelFields {
		generic := false
		valueType := field.FieldType
		if model.IsGeneric(valueType) {
			var err error
			generic = true
			if valueType, err = model.SplitGenericType(valueType); err != nil {
				return nil, err
			}
		}
		switch valueType {
		case model.MdlFieldTypeInt:
			in := &api.ModelSchema{
				Type: field.FieldType,
			}
			properties[field.FieldName] = in
		case model.MdlFieldTypeBool:
			bo := &api.ModelSchema{
				Type: field.FieldType,
			}
			properties[field.FieldName] = bo
		case model.MdlFieldTypeString:
			st := &api.ModelSchema{
				Type: field.FieldType,
			}
			properties[field.FieldName] = st
		default:
			schema, err := s.modelSchema(ctx, valueType, generic)
			if err != nil {
				return nil, err
			}
			properties[field.FieldName] = schema
		}
	}
	rsp.Properties = properties
	return rsp, nil
}

func (s *Service) AddModel(ctx context.Context, req *api.AddModelRequest) (*empty.Empty, error) {
	ok, err := s.modelExists(ctx, req.ModelName, req.TreeId)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("model: %s已存在", req.ModelName))
	}
	fieldArgs := make([]*api.ModelField, 0)
	if req.ModelFields != "" {
		modelFields := &api.ModelSchema{}
		if err := json.Unmarshal([]byte(req.ModelFields), &modelFields); err != nil {
			log.Error("Failed to unmarshal model fields: %+v", errors.WithStack(err))
			return nil, err
		}
		fieldArgs, err = s.procModelFieldArgs(ctx, req.ModelName, modelFields, req.TreeId)
		if err != nil {
			return nil, err
		}
	}
	err = s.dao.Transact(ctx, func(tx *xsql.Tx) error {
		if err := s.dao.TxInsertModel(tx, req.ModelName, req.Description, req.TreeId); err != nil {
			return err
		}
		for _, field := range fieldArgs {
			if err := s.dao.TxInsertModelField(tx, field); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error("Failed to transact: %+v", err)
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Service) modelExists(ctx context.Context, modelName string, treeID int64) (bool, error) {
	_, err := s.dao.ModelByName(ctx, modelName, treeID)
	if err == nil {
		return true, nil
	}
	if err != xsql.ErrNoRows {
		return false, ecode.Error(ecode.RequestErr, "model查询失败，请稍后再试")
	}
	return false, nil
}

func (s *Service) procModelFieldArgs(c context.Context, modelName string, fields *api.ModelSchema, treeID int64) ([]*api.ModelField, error) {
	FieldArgs := []*api.ModelField{}
	for name, field := range fields.Properties {
		valueType := field.Type
		var err error
		if model.IsGeneric(valueType) {
			valueType, err = model.SplitGenericType(valueType)
			if err != nil {
				return nil, err
			}
		}
		switch valueType {
		case model.MdlFieldTypeString:
			fallthrough
		case model.MdlFieldTypeInt:
			fallthrough
		case model.MdlFieldTypeBool:
			fieldArg := &api.ModelField{
				ModelName: modelName,
				FieldName: name,
				FieldType: field.Type,
			}
			FieldArgs = append(FieldArgs, fieldArg)
		default:
			exist, err := s.modelExists(c, valueType, treeID)
			if err != nil {
				return nil, err
			}
			if !exist {
				return nil, errors.New(fmt.Sprintf("模型(%s)不存在", valueType))
			}
			fieldArg := &api.ModelField{
				ModelName: modelName,
				FieldName: name,
				FieldType: field.Type,
			}
			FieldArgs = append(FieldArgs, fieldArg)
		}
	}
	return FieldArgs, nil
}

func (s *Service) ModelFieldList(ctx context.Context, req *api.ModelDetailRequest) (*api.ModelFieldReply, error) {
	mdl, err := s.dao.ModelByName(ctx, req.ModelName, req.TreeId)
	if err != nil {
		log.Error("Failed to modelByName: %+v", err)
		return nil, ecode.Error(ecode.NothingFound, "找不到对应的模型")
	}
	fields, err := s.dao.ModelFieldByName(ctx, mdl.Name)
	if err != nil {
		return nil, err
	}
	out := &api.ModelFieldReply{
		List: constructModelFields(fields),
	}
	return out, nil
}

func constructModelFields(in []*api.MapperModelField) []*api.ModelField {
	out := make([]*api.ModelField, 0)
	for _, val := range in {
		out = append(out, api.ConstructModelField(val))
	}
	return out
}

func (s *Service) AddModelField(ctx context.Context, req *api.AddModelFieldRequest) (*empty.Empty, error) {
	if err := s.dao.AddModelField(ctx, req); err != nil {
		log.Error("Failed to update model field: %+v", err)
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Service) UpdateModelField(ctx context.Context, req *api.UpdateModelFieldRequest) (*empty.Empty, error) {
	if err := s.dao.UpdateModelField(ctx, req); err != nil {
		log.Error("Failed to update model field: %+v", err)
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Service) DeleteModelField(ctx context.Context, req *api.DeleteModelFieldRequest) (*empty.Empty, error) {
	if err := s.dao.DelModelField(ctx, req.Id); err != nil {
		log.Error("Failed to delete model field: %+v", err)
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Service) AddModelFieldRule(ctx context.Context, req *api.AddModelFieldRuleRequest) (*empty.Empty, error) {
	params := make([]*model.ItemFieldRule, 0)
	if err := json.Unmarshal([]byte(req.FieldRuleList), &params); err != nil {
		return nil, err
	}
	if err := s.dao.AddFieldRule(ctx, params); err != nil {
		log.Error("Failed to add model field rule: %+v", err)
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Service) UpdateModelFieldRule(ctx context.Context, req *api.UpdateModelFieldRuleRequest) (*empty.Empty, error) {
	if err := s.dao.UpdateFieldRule(ctx, req); err != nil {
		log.Error("Failed to update model field rule: %+v", err)
		return nil, err
	}
	return &empty.Empty{}, nil
}
