package http

import (
	"fmt"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/datasource-ng/admin/internal/model"

	"go-gateway/app/web-svr/datasource-ng/admin/api"
)

var noArg = &api.NoArgRequest{}

func ModelList(c *bm.Context) {
	req := &api.ModelListRequest{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.ModelList(c, req))
}

func ModelAll(c *bm.Context) {
	c.JSON(svc.ModelAll(c, noArg))
}

func ModelDetail(c *bm.Context) {
	req := &api.ModelDetailRequest{}
	if err := c.Bind(req); err != nil {
		return
	}
	detail, err := svc.ModelDetail(c, req)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(modelDetailMarshal(detail.Detail), nil)
}

func modelDetailMarshal(res *api.ModelSchema) *model.JsonSchema {
	rsp := &model.JsonSchema{}
	// 基本类型 string、bool、int
	if !model.IsReference(res.Type) {
		switch res.Type {
		case model.MdlFieldTypeInt:
			rsp.Type = model.ReflectJsonSchemaType[res.Type]
			rsp.Default = res.DefaultInt
			rsp.Description = res.Description
		case model.MdlFieldTypeString:
			rsp.Type = model.ReflectJsonSchemaType[res.Type]
			rsp.Default = res.DefaultString
			rsp.Description = res.Description
		case model.MdlFieldTypeBool:
			rsp.Type = model.ReflectJsonSchemaType[res.Type]
			rsp.Default = model.TranBool(int(res.DefaultInt))
			rsp.Description = res.Description
		}
		if res.Component != nil {
			rsp.Component = &api.ModelComponent{
				Type:     res.Component.Type,
				Metadata: res.Component.Metadata,
			}
		}
		return rsp
	}
	// Object类型
	if res.Type == model.MdlFieldTypeObject {
		rsp.Type = model.ReflectJsonSchemaType[res.Type]
		rsp.Properties = make(map[string]*model.JsonSchema)
		rsp.Description = res.Description
		for key, item := range res.Properties {
			schema := modelDetailMarshal(item)
			rsp.Properties[key] = schema
			if item.Required {
				rsp.Required = append(rsp.Required, key)
			}
		}
		if res.Component != nil {
			rsp.Component = &api.ModelComponent{
				Type:     res.Component.Type,
				Metadata: res.Component.Metadata,
			}
		}
		return rsp
	}
	// Array类型
	genType, _ := model.SplitGenericType(res.Type)
	switch genType {
	case model.MdlFieldTypeInt:
		fallthrough
	case model.MdlFieldTypeString:
		fallthrough
	case model.MdlFieldTypeBool:
		rsp.Type = model.JsonSchemaTypeArray
		rsp.Description = res.Description
		rsp.Items = &model.JsonSchemaItems{Type: model.ReflectJsonSchemaType[genType]}
		if res.Required {
			rsp.MinItems = 1
		}
	default:
		rsp.Type = model.JsonSchemaTypeArray
		rsp.Items = &model.JsonSchemaItems{
			Ref: fmt.Sprintf(model.JsonSchemaRef, "def"),
		}
		rsp.Definitions = make(map[string]*model.JsonSchema)
		rsp.Description = res.Description
		definitionsTmp := &model.JsonSchema{
			Type:        model.JsonSchemaTypeObject,
			Description: res.Description,
			Properties:  make(map[string]*model.JsonSchema),
		}
		for key, item := range res.Properties {
			schema := modelDetailMarshal(item)
			definitionsTmp.Properties[key] = schema
			if item.Required {
				definitionsTmp.Required = append(definitionsTmp.Required, key)
			}
		}
		rsp.Definitions["def"] = definitionsTmp
	}
	return rsp
}

func ModelItemList(c *bm.Context) {
	req := &api.ModelItemListRequest{}
	if err := c.Bind(req); err != nil {
		return
	}
	c.JSON(svc.ModelItemList(c, req))
}

func ModelItemDetail(c *bm.Context) {
	req := &api.ModelItemDetailRequest{}
	if err := c.Bind(req); err != nil {
		return
	}
	result, err := svc.ModelItemDetail(c, req)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(FlatItem(result))
}

func FlatItem(in *api.ItemReply) (interface{}, error) {
	rsp := make(map[string]interface{})
	switch item := in.Item.(type) {
	case *api.ItemReply_ValueRaw:
		switch v := item.ValueRaw.Value.(type) {
		case *api.FieldValue_ValueInt:
			return v.ValueInt, nil
		case *api.FieldValue_ValueIntList:
			return v.ValueIntList, nil
		case *api.FieldValue_ValueString:
			return v.ValueString, nil
		case *api.FieldValue_ValueStringList:
			return v.ValueStringList, nil
		case *api.FieldValue_ValueBool:
			return v.ValueBool, nil
		case *api.FieldValue_ValueBoolList:
			return v.ValueBoolList, nil
		case *api.FieldValue_ValueItem:
			return FlatItem(v.ValueItem)
		case *api.FieldValue_ValueItemList:
			out := make([]interface{}, 0, len(v.ValueItemList.List))
			for _, it := range v.ValueItemList.List {
				flat, err := FlatItem(it)
				if err != nil {
					return nil, err
				}
				out = append(out, flat)
			}
			return out, nil
		}
	case *api.ItemReply_ValueStruct:
		for key, item := range item.ValueStruct.Item {
			switch value := item.Value.(type) {
			case *api.FieldValue_ValueInt:
				rsp[key] = value.ValueInt
			case *api.FieldValue_ValueIntList:
				rsp[key] = value.ValueIntList.List
			case *api.FieldValue_ValueString:
				rsp[key] = value.ValueString
			case *api.FieldValue_ValueStringList:
				rsp[key] = value.ValueStringList.List
			case *api.FieldValue_ValueBool:
				rsp[key] = value.ValueBool
			case *api.FieldValue_ValueBoolList:
				rsp[key] = value.ValueBoolList.List
			case *api.FieldValue_ValueItem:
				objValue, err := FlatItem(value.ValueItem)
				if err != nil {
					return nil, err
				}
				rsp[key] = objValue
			case *api.FieldValue_ValueItemList:
				var objValues []interface{}
				for _, subItem := range value.ValueItemList.List {
					objValue, err := FlatItem(subItem)
					if err != nil {
						return nil, err
					}
					objValues = append(objValues, objValue)
				}
				rsp[key] = objValues
			}
		}
	}
	return rsp, nil
}

func ModelCreate(c *bm.Context) {
	req := &api.ModelCreateRequest{}
	if err := c.Bind(req); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, _ := username.(string)
	req.CreatedBy = userName
	c.JSON(svc.ModelCreate(c, req))
}

func ModelItemCreate(c *bm.Context) {
	req := &api.ModelItemCreateRequest{}
	if err := c.Bind(req); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, _ := username.(string)
	req.CreatedBy = userName
	if req.Expirable != 0 && req.ExpireAt == 0 {
		c.JSON(nil, ecode.Error(ecode.RequestErr, "未填写过期时间"))
		return
	}
	c.JSON(svc.ModelItemCreate(c, req))
}
