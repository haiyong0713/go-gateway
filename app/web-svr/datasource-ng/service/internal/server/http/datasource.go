package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/datasource-ng/service/api"

	"github.com/pkg/errors"
)

func flat(in *api.ItemReply) (interface{}, error) {
	switch item := in.Item.(type) {
	case *api.ItemReply_ValueStruct:
		out := make(map[string]interface{})
		for fname, fvalue := range item.ValueStruct.Item {
			switch v := fvalue.Value.(type) {
			case *api.FieldValue_ValueString:
				out[fname] = v.ValueString
			case *api.FieldValue_ValueInt:
				out[fname] = v.ValueInt
			case *api.FieldValue_ValueBool:
				out[fname] = v.ValueBool
			case *api.FieldValue_ValueItem:
				flated, err := flat(v.ValueItem)
				if err != nil {
					return nil, err
				}
				out[fname] = flated
			}
		}
		return out, nil
	case *api.ItemReply_ValueRaw:
		switch v := item.ValueRaw.Value.(type) {
		case *api.FieldValue_ValueString:
			return v.ValueString, nil
		case *api.FieldValue_ValueInt:
			return v.ValueInt, nil
		case *api.FieldValue_ValueItem:
			return flat(v.ValueItem)
		case *api.FieldValue_ValueStringList:
			return v.ValueStringList.List, nil
		case *api.FieldValue_ValueIntList:
			return v.ValueIntList.List, nil
		case *api.FieldValue_ValueBoolList:
			return v.ValueBoolList.List, nil
		case *api.FieldValue_ValueItemList:
			out := make([]interface{}, 0, len(v.ValueItemList.List))
			for _, i := range v.ValueItemList.List {
				flated, err := flat(i)
				if err != nil {
					log.Error("Failed to flat item: %+v: %+v", i, err)
					continue
				}
				out = append(out, flated)
			}
			return out, nil
		}
	}

	return nil, errors.Errorf("unable to flat item reply: %+v", in)
}

func flatItem(ctx *bm.Context) {
	req := &api.ItemReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	item, err := svc.Item(ctx, req)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(flat(item))
}

func item(ctx *bm.Context) {
	req := &api.ItemReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.Item(ctx, req))
}
