package defaultvalue

import (
	"go-gateway/app/app-svr/distribution/distribution/internal/distributionconst"
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/util"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"

	"go-common/library/log"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

func Init() {
	msgDesc, ok := preferenceproto.TryGetMessage(distributionconst.MakeFullyQualifiedName("defaultValue"))
	if !ok {
		panic(errors.Errorf("unable to find defaultValue message descriptor"))
	}

	oneOfs := msgDesc.GetOneOfs()
	for _, oneOf := range oneOfs {
		if oneOf.GetName() == "value" {
			defaultValueOneOfDesc = oneOf
			break
		}
	}
	if defaultValueOneOfDesc == nil {
		panic(errors.Errorf("unable to find value one of descriptor"))
	}
}

var (
	defaultValueOneOfDesc *desc.OneOfDescriptor
)

func initializeOnRepeated(dst *dynamic.Message, field *desc.FieldDescriptor) {
	dmSlice, err := util.GetFieldAsRepeatedDynamicMessage(dst, field)
	if err != nil {
		log.Error("Failed to get repeated field: %q: %+v", field.GetFullyQualifiedName(), err)
		return
	}

	fieldType := field.GetMessageType().GetFullyQualifiedName()
	for _, dm := range dmSlice {
		if _, ok := preferenceproto.DistributionPrimitiveType.GetTypeDescriptor(fieldType); !ok {
			InitializeWithDefaultValue(dm)
			continue
		}
		setDefaultValueField(dm, field)
	}
}

func setDefaultValueField(dm *dynamic.Message, field *desc.FieldDescriptor) {
	defV, err := preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsDefaultValue(field)
	if err != nil {
		log.Error("Failed to get field default value: %q: %+v", field.GetFullyQualifiedName(), err)
		return
	}
	defaultValue, ok := extractDefaultValueDef(dm, defV)
	if ok {
		if err := dm.TrySetFieldByName("default_value", defaultValue); err != nil {
			log.Error("Failed to set default field: %+v", err)
			return
		}
		setValueFieldIfNotModified(dm, defaultValue)
	}
}

func setValueFieldIfNotModified(dm *dynamic.Message, defaultValue interface{}) {
	lastModifiedV, err := dm.TryGetFieldByName("last_modified")
	if err != nil {
		log.Error("Failed to get last_modified: %+v", errors.WithStack(err))
		return
	}
	lastModified, ok := lastModifiedV.(int64)
	if !ok {
		log.Error("Failed to cast last_modified as int64: %+v/%T", lastModifiedV, lastModifiedV)
		return
	}
	if lastModified > 0 {
		return
	}
	if err := dm.TrySetFieldByName("value", defaultValue); err != nil {
		log.Error("Failed to set value field: %+v", errors.WithStack(err))
		return
	}
}

func InitializeWithDefaultValue(dst *dynamic.Message) {
	for _, field := range dst.GetMessageDescriptor().GetFields() {
		if field.GetType() != dpb.FieldDescriptorProto_TYPE_MESSAGE {
			continue
		}
		if field.IsRepeated() {
			initializeOnRepeated(dst, field)
			continue
		}
		fieldV, err := util.GetFieldAsDynamicMessage(dst, field)
		if err != nil {
			log.Warn("Failed to get field value: %+v as dynamic message: %+v", field, err)
			continue
		}
		if fieldV == nil {
			fieldV = dynamic.NewMessage(field.GetMessageType())
		}

		func() {
			fieldType := field.GetMessageType().GetFullyQualifiedName()
			if _, ok := preferenceproto.DistributionPrimitiveType.GetTypeDescriptor(fieldType); !ok {
				InitializeWithDefaultValue(fieldV)
				return
			}
			setDefaultValueField(fieldV, field)
		}()
		dst.SetField(field, fieldV)
	}
}

func extractDefaultValueDef(dst *dynamic.Message, defaultValueDef *dynamic.Message) (interface{}, bool) {
	fieldDesc, value, err := defaultValueDef.TryGetOneOfField(defaultValueOneOfDesc)
	if err != nil {
		log.Error("Failed to extract default value: %+v", err)
		return nil, false
	}
	if fieldDesc == nil {
		return nil, false
	}
	// https://developers.google.com/protocol-buffers/docs/proto3#scalar
	switch dst.GetMessageDescriptor().GetFullyQualifiedName() {
	case preferenceproto.Int64ValueFQN:
		if fieldDesc.GetType() == dpb.FieldDescriptorProto_TYPE_INT64 {
			return value.(int64), true
		}
	case preferenceproto.Int32ValueFQN:
		if fieldDesc.GetType() == dpb.FieldDescriptorProto_TYPE_INT32 {
			return value.(int32), true
		}
	case preferenceproto.DoubleValueFQN:
		if fieldDesc.GetType() == dpb.FieldDescriptorProto_TYPE_DOUBLE {
			return value.(float64), true
		}
	case preferenceproto.FloatValueFQN:
		if fieldDesc.GetType() == dpb.FieldDescriptorProto_TYPE_FLOAT {
			return value.(float32), true
		}
	case preferenceproto.UInt64ValueFQN:
		if fieldDesc.GetType() == dpb.FieldDescriptorProto_TYPE_UINT64 {
			return value.(uint64), true
		}
	case preferenceproto.UInt32ValueFQN:
		if fieldDesc.GetType() == dpb.FieldDescriptorProto_TYPE_UINT32 {
			return value.(uint32), true
		}
	case preferenceproto.BoolValueFQN:
		if fieldDesc.GetType() == dpb.FieldDescriptorProto_TYPE_BOOL {
			return value.(bool), true
		}
	case preferenceproto.StringValueFQN:
		if fieldDesc.GetType() == dpb.FieldDescriptorProto_TYPE_STRING {
			return value.(string), true
		}
	case preferenceproto.BytesValueFQN:
		if fieldDesc.GetType() == dpb.FieldDescriptorProto_TYPE_BYTES {
			return value.([]byte), true
		}
	}
	return nil, false
}
