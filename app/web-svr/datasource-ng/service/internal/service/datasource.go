package service

import (
	"context"
	"sort"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/datasource-ng/service/api"
	"go-gateway/app/web-svr/datasource-ng/service/internal/model"

	"github.com/pkg/errors"
)

func castAsBool(valueInt int64) bool {
	if valueInt > 0 {
		return true
	}
	return false
}

func (s *Service) cvtFieldValue(ctx context.Context, in *model.ItemFieldValue, valueType string) (*api.FieldValue, error) {
	out := &api.FieldValue{}
	switch valueType {
	case model.ValueTypeString:
		out.Value = &api.FieldValue_ValueString{
			ValueString: in.ValueString,
		}
	case model.ValueTypeInt:
		out.Value = &api.FieldValue_ValueInt{
			ValueInt: in.ValueInt,
		}
	case model.ValueTypeBool:
		out.Value = &api.FieldValue_ValueBool{
			ValueBool: castAsBool(in.ValueInt),
		}
	default:
		refItem, err := s.dao.GetItem(ctx, in.ValueItemUuid)
		if err != nil {
			return nil, err
		}
		itemReply, err := s.resolveValueByType(ctx, refItem)
		if err != nil {
			return nil, err
		}
		out.Value = &api.FieldValue_ValueItem{
			ValueItem: itemReply,
		}
	}
	return out, nil
}

func splitGenericType(valueType string) (string, error) {
	if model.ValueTypeListPattern.MatchString(valueType) {
		return model.ValueTypeListPattern.FindStringSubmatch(valueType)[1], nil
	}
	return "", errors.Errorf("Invalid generic value type: %s", valueType)
}

func iterByKeyOrder(in map[string]*model.ItemFieldValue, reverse bool, iter func(string, *model.ItemFieldValue) error) {
	keys := make([]string, 0, len(in))
	for k := range in {
		keys = append(keys, k)
	}
	sort.Sort(sort.StringSlice(keys))
	if reverse {
		sort.Reverse(sort.StringSlice(keys))
	}
	for _, k := range keys {
		v := in[k]
		if err := iter(k, v); err != nil {
			log.Error("Stop to iterate: %+v", err)
			return
		}
	}
}

func (s *Service) cvtGenericFieldValue(ctx context.Context, in map[string]*model.ItemFieldValue, innerType string) (*api.FieldValue, error) {
	out := &api.FieldValue{}
	switch innerType {
	case model.ValueTypeString:
		apiFv := &api.FieldValue_ValueStringList{
			ValueStringList: &api.StringList{
				List: []string{},
			},
		}
		refList := &apiFv.ValueStringList.List
		out.Value = apiFv
		iterByKeyOrder(in, false, func(_ string, fv *model.ItemFieldValue) error {
			*refList = append(*refList, fv.ValueString)
			return nil
		})
	case model.ValueTypeInt:
		apiFv := &api.FieldValue_ValueIntList{
			ValueIntList: &api.IntList{
				List: []int64{},
			},
		}
		refList := &apiFv.ValueIntList.List
		out.Value = apiFv
		iterByKeyOrder(in, false, func(_ string, fv *model.ItemFieldValue) error {
			*refList = append(*refList, fv.ValueInt)
			return nil
		})
	case model.ValueTypeBool:
		apiFv := &api.FieldValue_ValueBoolList{
			ValueBoolList: &api.BoolList{
				List: []bool{},
			},
		}
		refList := &apiFv.ValueBoolList.List
		out.Value = apiFv
		iterByKeyOrder(in, false, func(_ string, fv *model.ItemFieldValue) error {
			*refList = append(*refList, castAsBool(fv.ValueInt))
			return nil
		})
	default:
		apiFv := &api.FieldValue_ValueItemList{
			ValueItemList: &api.ItemList{
				List: []*api.ItemReply{},
			},
		}
		refList := &apiFv.ValueItemList.List
		out.Value = apiFv
		iterByKeyOrder(in, false, func(_ string, fv *model.ItemFieldValue) error {
			refItem, err := s.dao.GetItem(ctx, fv.ValueItemUuid)
			if err != nil {
				log.Error("Failed to get item: %+v", err)
				return nil
			}
			itemReply, err := s.resolveValueByType(ctx, refItem)
			if err != nil {
				log.Error("Failed to resolve value by type: %+v", err)
				return nil
			}
			*refList = append(*refList, itemReply)
			return nil
		})
	}
	return out, nil
}

func getOne(in map[string]*model.ItemFieldValue) (string, *model.ItemFieldValue, bool) {
	for k, v := range in {
		return k, v, true
	}
	return "", nil, false
}

func (s *Service) resolveArbitrary(ctx context.Context, typeName string, fvalues map[string]*model.ItemFieldValue) (*api.ItemReply, error) {
	innerType, err := splitGenericType(typeName)
	if err != nil {
		return nil, err
	}

	apiFv, err := s.cvtGenericFieldValue(ctx, fvalues, innerType)
	if err != nil {
		return nil, err
	}
	out := &api.ItemReply{
		Item: &api.ItemReply_ValueRaw{
			ValueRaw: apiFv,
		},
	}
	return out, nil
}

func createEmptyValue(typeName string) (*api.FieldValue, error) {
	out := &api.FieldValue{}

	if model.IsGeneric(typeName) {
		innerType, err := splitGenericType(typeName)
		if err != nil {
			return nil, err
		}
		switch innerType {
		case model.ValueTypeString:
			out.Value = &api.FieldValue_ValueStringList{
				ValueStringList: &api.StringList{
					List: []string{},
				},
			}
		case model.ValueTypeInt:
			out.Value = &api.FieldValue_ValueIntList{
				ValueIntList: &api.IntList{
					List: []int64{},
				},
			}
		case model.ValueTypeBool:
			out.Value = &api.FieldValue_ValueBoolList{
				ValueBoolList: &api.BoolList{
					List: []bool{},
				},
			}
		default:
			out.Value = &api.FieldValue_ValueItemList{
				ValueItemList: &api.ItemList{
					List: []*api.ItemReply{},
				},
			}
		}
	}

	switch typeName {
	case model.ValueTypeString:
		out.Value = &api.FieldValue_ValueString{
			ValueString: "",
		}
	case model.ValueTypeInt:
		out.Value = &api.FieldValue_ValueInt{
			ValueInt: 0,
		}
	case model.ValueTypeBool:
		out.Value = &api.FieldValue_ValueBool{
			ValueBool: false,
		}
	default:
		out.Value = &api.FieldValue_ValueItem{
			ValueItem: &api.ItemReply{},
		}
	}

	return out, nil
}

func (s *Service) resolveReference(ctx context.Context, typeName string, fvalues map[string]*model.ItemFieldValue) (*api.ItemReply, error) {
	if model.IsGeneric(typeName) {
		return s.resolveArbitrary(ctx, typeName, fvalues)
	}

	customType, err := s.dao.GetModel(ctx, typeName)
	if err != nil {
		return nil, err
	}
	fields, err := s.dao.GetModelField(ctx, customType.Name)
	if err != nil {
		return nil, err
	}

	structed := &api.ItemReply_ValueStruct{
		ValueStruct: &api.StructedItem{
			Item: make(map[string]*api.FieldValue),
		},
	}
	refItem := structed.ValueStruct.Item
	out := &api.ItemReply{
		Item: structed,
	}
	for fname, mf := range fields {
		fv, ok := fvalues[fname]
		if !ok {
			empty, err := createEmptyValue(mf.ValueType)
			if err != nil {
				log.Error("Failed to create empty value: %+v", err)
				continue
			}
			refItem[fname] = empty
			continue
		}

		if model.IsReference(mf.ValueType) {
			rFvalues, err := s.dao.GetModelFieldValue(ctx, fv.ValueItemUuid)
			if err != nil {
				log.Error("Failed to get model field value: %+v", err)
				continue
			}
			refReply, err := s.resolveReference(ctx, mf.ValueType, rFvalues)
			if err != nil {
				log.Error("Failed to resolve reference: %+v", err)
				continue
			}
			refItem[fname] = &api.FieldValue{
				Value: &api.FieldValue_ValueItem{
					ValueItem: refReply,
				},
			}
			continue
		}

		apiFv, err := s.cvtFieldValue(ctx, fv, mf.ValueType)
		if err != nil {
			log.Error("Failed to convert field value: %+v", err)
			continue
		}
		refItem[fname] = apiFv
	}

	return out, nil
}

func (s *Service) resolveValueByType(ctx context.Context, item *model.ModelItem) (*api.ItemReply, error) {
	fvalues, err := s.dao.GetModelFieldValue(ctx, item.ItemUuid)
	if err != nil {
		return nil, err
	}

	out := &api.ItemReply{}
	// builtin types
	switch item.TypeName {
	case model.ValueTypeString:
		valueString := ""
		_, v, ok := getOne(fvalues)
		if ok {
			valueString = v.ValueString
		}
		out.Item = &api.ItemReply_ValueRaw{
			ValueRaw: &api.FieldValue{
				Value: &api.FieldValue_ValueString{
					ValueString: valueString,
				},
			},
		}
		return out, nil
	case model.ValueTypeInt:
		valueInt := int64(0)
		_, v, ok := getOne(fvalues)
		if ok {
			valueInt = v.ValueInt
		}
		out.Item = &api.ItemReply_ValueRaw{
			ValueRaw: &api.FieldValue{
				Value: &api.FieldValue_ValueInt{
					ValueInt: valueInt,
				},
			},
		}
		return out, nil
	case model.ValueTypeBool:
		valueInt := int64(0)
		_, v, ok := getOne(fvalues)
		if ok {
			valueInt = v.ValueInt
		}
		out.Item = &api.ItemReply_ValueRaw{
			ValueRaw: &api.FieldValue{
				Value: &api.FieldValue_ValueBool{
					ValueBool: castAsBool(valueInt),
				},
			},
		}
		return out, nil
	}

	// reference
	return s.resolveReference(ctx, item.TypeName, fvalues)
}

// Item is
func (s *Service) Item(ctx context.Context, req *api.ItemReq) (*api.ItemReply, error) {
	item, err := s.dao.GetItem(ctx, req.ItemUuid)
	if err != nil {
		return nil, err
	}
	if item.Expired(time.Now()) {
		return nil, errors.Wrapf(ecode.RequestErr, "item expired: %+v", item)
	}
	if req.TypeName != "" {
		if req.TypeName != item.TypeName {
			return nil, errors.WithStack(ecode.RequestErr)
		}
	}
	return s.resolveValueByType(ctx, item)
}
