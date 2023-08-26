package util

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

func GetFieldAsDynamicMessage(dm *dynamic.Message, field *desc.FieldDescriptor) (*dynamic.Message, error) {
	fieldV, err := dm.TryGetField(field)
	if err != nil {
		return nil, err
	}
	dmFieldV, ok := fieldV.(*dynamic.Message)
	if !ok {
		return nil, errors.Errorf("Failed to cast field value: %T as dynamic message", fieldV)
	}
	return dmFieldV, nil
}

func GetFieldAsRepeatedDynamicMessage(dm *dynamic.Message, field *desc.FieldDescriptor) ([]*dynamic.Message, error) {
	fieldV, err := dm.TryGetField(field)
	if err != nil {
		return nil, err
	}

	interfaceSlice, ok := fieldV.([]interface{})
	if !ok {
		return nil, errors.Errorf("Failed to cast field value: %T as interface{} slice", fieldV)
	}

	out := make([]*dynamic.Message, 0, len(interfaceSlice))
	for _, dmi := range interfaceSlice {
		dm, ok := dmi.(*dynamic.Message)
		if !ok {
			return nil, errors.Errorf("Failed to cast field value: %T as dynamic message", fieldV)
		}
		out = append(out, dm)
	}
	return out, nil
}

func MessageJSONify(in *dynamic.Message) string {
	marshaler := &jsonpb.Marshaler{
		Indent: "  ",
	}
	out, _ := marshaler.MarshalToString(in)
	return out
}
