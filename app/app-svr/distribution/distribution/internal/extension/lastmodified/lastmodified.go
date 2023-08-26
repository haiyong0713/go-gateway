package lastmodified

import (
	"reflect"
	"time"

	"go-gateway/app/app-svr/distribution/distribution/internal/extension/util"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"

	"go-common/library/log"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

func Init() {}

func setLastModifiedOnRepeated(dst *dynamic.Message, field *desc.FieldDescriptor, t time.Time) {
	dmSlice, err := util.GetFieldAsRepeatedDynamicMessage(dst, field)
	if err != nil {
		log.Error("Failed to get repeated field: %q: %+v", field.GetFullyQualifiedName(), err)
		return
	}

	fieldType := field.GetMessageType().GetFullyQualifiedName()
	for _, dm := range dmSlice {
		if _, ok := preferenceproto.DistributionPrimitiveType.GetTypeDescriptor(fieldType); !ok {
			SetPreferenceValueLastModified(dm, t)
			continue
		}
		setLastModified(dm, t)
	}
}

func setLastModified(dst *dynamic.Message, t time.Time) {
	if err := dst.TrySetFieldByName("last_modified", t.Unix()); err != nil {
		log.Error("Failed to set `last_modified` field: %+v", err)
		return
	}
}

func compareAndSetLastModifiedOnRepeated(dst *dynamic.Message, origin *dynamic.Message, field *desc.FieldDescriptor, t time.Time) {
	dmSlice, err := util.GetFieldAsRepeatedDynamicMessage(dst, field)
	if err != nil {
		log.Error("Failed to get repeated field: %q: %+v", field.GetFullyQualifiedName(), err)
		return
	}

	originDMSlice, err := util.GetFieldAsRepeatedDynamicMessage(origin, field)
	if err != nil {
		log.Error("Failed to get origin repeated field: %q: %+v", field.GetFullyQualifiedName(), err)
		return
	}

	fieldType := field.GetMessageType().GetFullyQualifiedName()
	for i, dm := range dmSlice {
		originDM := func() *dynamic.Message {
			if (i + 1) > len(originDMSlice) {
				return nil
			}
			return originDMSlice[i]
		}()
		if _, ok := preferenceproto.DistributionPrimitiveType.GetTypeDescriptor(fieldType); !ok {
			CompareAndSetPreferenceValueLastModified(dm, originDM, t)
			continue
		}
		compareAndSetLastModified(dm, originDM, t)
	}
}

func compareAndSetLastModified(dst *dynamic.Message, origin *dynamic.Message, t time.Time) {
	if origin == nil {
		setLastModified(dst, t)
		return
	}

	equal, err := compareUpdateValueField(dst, origin)
	if err != nil {
		log.Error("Failed to compare update field: %+v", err)
		equal = false // treat as false
	}
	switch equal {
	case true: // ensure field `last_modified` as origin
		originLastModified, err := getLastModified(origin)
		if err != nil {
			log.Error("Failed to get origin `last_modified` field: %+v: %+v", origin, err)
			return
		}
		if err := dst.TrySetFieldByName("last_modified", originLastModified); err != nil {
			log.Error("Failed to set `last_modified` field: %+v", err)
			return
		}
	case false: // set field `last_modified` as current timestamp
		if err := dst.TrySetFieldByName("last_modified", t.Unix()); err != nil {
			log.Error("Failed to set `last_modified` field: %+v", err)
			return
		}
	}
}

func SetPreferenceValueLastModified(dst *dynamic.Message, t time.Time) {
	for _, field := range dst.GetMessageDescriptor().GetFields() {
		if field.GetType() != dpb.FieldDescriptorProto_TYPE_MESSAGE {
			continue
		}
		if field.IsRepeated() {
			setLastModifiedOnRepeated(dst, field, t)
			continue
		}
		dmFieldV, err := util.GetFieldAsDynamicMessage(dst, field)
		if err != nil {
			log.Error("Failed to get dst field: %q: %+v", field.GetFullyQualifiedName(), err)
			continue
		}
		if dmFieldV == nil {
			// skip nil field
			continue
		}

		fieldType := field.GetMessageType().GetFullyQualifiedName()
		if _, ok := preferenceproto.DistributionPrimitiveType.GetTypeDescriptor(fieldType); !ok {
			SetPreferenceValueLastModified(dmFieldV, t)
			continue
		}
		setLastModified(dmFieldV, t)
	}
}

func CompareAndSetPreferenceValueLastModified(dst *dynamic.Message, origin *dynamic.Message, t time.Time) {
	if origin == nil {
		SetPreferenceValueLastModified(dst, t)
		return
	}

	if dst.GetMessageDescriptor().GetFullyQualifiedName() != origin.GetMessageDescriptor().GetFullyQualifiedName() {
		log.Error("Mismatched message type: %q and %q", dst.GetMessageDescriptor().GetFullyQualifiedName(), origin.GetMessageDescriptor().GetFullyQualifiedName())
		return
	}

	for _, field := range dst.GetMessageDescriptor().GetFields() {
		if field.GetType() != dpb.FieldDescriptorProto_TYPE_MESSAGE {
			continue
		}
		if field.IsRepeated() {
			compareAndSetLastModifiedOnRepeated(dst, origin, field, t)
			continue
		}

		dmFieldV, err := util.GetFieldAsDynamicMessage(dst, field)
		if err != nil {
			log.Error("Failed to get dst field: %q: %+v", field.GetFullyQualifiedName(), err)
			continue
		}
		if dmFieldV == nil {
			// skip nil field
			continue
		}
		originDmFieldV, err := util.GetFieldAsDynamicMessage(origin, field)
		if err != nil {
			log.Error("Failed to get origin field: %q: %+v", field.GetFullyQualifiedName(), err)
			continue
		}
		if originDmFieldV == nil {
			SetPreferenceValueLastModified(dmFieldV, t)
			continue
		}

		fieldType := field.GetMessageType().GetFullyQualifiedName()
		if _, ok := preferenceproto.DistributionPrimitiveType.GetTypeDescriptor(fieldType); !ok {
			CompareAndSetPreferenceValueLastModified(dmFieldV, originDmFieldV, t)
			continue
		}
		compareAndSetLastModified(dmFieldV, originDmFieldV, t)
	}
}

func getLastModified(dm *dynamic.Message) (int64, error) {
	lastModifiedV, err := dm.TryGetFieldByName("last_modified")
	if err != nil {
		return 0, err
	}
	lastModified, ok := lastModifiedV.(int64)
	if !ok {
		return 0, errors.Errorf("Invalid last_modified value: %+v", lastModifiedV)
	}
	return lastModified, nil
}

func compareUpdateValueField(dst *dynamic.Message, origin *dynamic.Message) (bool, error) {
	if dst.GetMessageDescriptor().GetFullyQualifiedName() != origin.GetMessageDescriptor().GetFullyQualifiedName() {
		return false, errors.Errorf("Mismatched message type to compare: %+v/%T, %+v/%T", dst, dst, origin, origin)
	}
	dstValue, err := dst.TryGetFieldByName("value")
	if err != nil {
		return false, err
	}

	originValue, err := judgeOriginValue(origin)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(dstValue, originValue), nil
}

func judgeOriginValue(in *dynamic.Message) (interface{}, error) {
	lastModifiedV, err := in.TryGetFieldByName("last_modified")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	lastModified, ok := lastModifiedV.(int64)
	if !ok {
		return nil, errors.Errorf("Failed to cast last_modified as int64: %+v/%T", lastModifiedV, lastModifiedV)
	}
	valueV, err := in.TryGetFieldByName("value")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defaultValueV, err := in.TryGetFieldByName("default_value")
	if err != nil {
		return nil, err
	}

	if lastModified > 0 {
		return valueV, nil
	}
	return defaultValueV, nil
}
