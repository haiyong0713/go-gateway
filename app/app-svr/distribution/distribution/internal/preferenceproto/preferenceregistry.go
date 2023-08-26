package preferenceproto

import (
	"fmt"
	"strings"
	"sync/atomic"

	"go-gateway/app/app-svr/distribution/distribution/internal/distributionconst"

	"go-common/library/log"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

const (
	E_Preference_Field    = 76001
	E_StorageDriver_Field = 76002
	E_Disabled_Field      = 76003
	E_Feature_Field       = 76004
	E_Default_Field       = 75004
	E_ABTest_Field        = 75005
	E_Tus_Field           = 75006
	E_Multiple_Tus_Field  = 75007
	E_Refenum_Field       = 75001
)

var (
	initialized = int64(0)
)

var (
	defaultExtensionRegistry = dynamic.NewExtensionRegistryWithDefaults()

	DefaultDistributionExtensionDesc = &ExtensionDesc{}
	DistributionPrimitiveType        = &PrimitiveTypeDesc{}
	GlobalRegistry                   *PreferenceRegistry
)

type ExtensionDesc struct {
	FileOptions struct {
		Preference    *desc.FieldDescriptor
		StorageDriver *desc.FieldDescriptor
		Disabled      *desc.FieldDescriptor
		Feature       *desc.FieldDescriptor
	}
	FieldOptions struct {
		Default         *desc.FieldDescriptor
		ABTestFlagValue *desc.FieldDescriptor
		TusValue        *desc.FieldDescriptor
		TusValues       *desc.FieldDescriptor
		Refenum         *desc.FieldDescriptor
	}
}

type PrimitiveTypeDesc struct {
	DoubleValue *desc.MessageDescriptor
	FloatValue  *desc.MessageDescriptor
	Int64Value  *desc.MessageDescriptor
	UInt64Value *desc.MessageDescriptor
	Int32Value  *desc.MessageDescriptor
	UInt32Value *desc.MessageDescriptor
	BoolValue   *desc.MessageDescriptor
	StringValue *desc.MessageDescriptor
	BytesValue  *desc.MessageDescriptor
}

func (p *PrimitiveTypeDesc) NewStringValue() *dynamic.Message {
	dm := dynamic.NewMessage(p.StringValue)
	return dm
}

func (p *PrimitiveTypeDesc) NewInt64Value() *dynamic.Message {
	dm := dynamic.NewMessage(p.Int64Value)
	return dm
}

func (p *PrimitiveTypeDesc) NewInt32Value() *dynamic.Message {
	dm := dynamic.NewMessage(p.Int32Value)
	return dm
}

func (p *PrimitiveTypeDesc) GetTypeDescriptor(typeName string) (*desc.MessageDescriptor, bool) {
	pureTypeName := strings.TrimPrefix(typeName, "bilibili.app.distribution.v1.")
	typeDef := map[string]*desc.MessageDescriptor{
		DoubleValue: p.DoubleValue,
		FloatValue:  p.FloatValue,
		Int64Value:  p.Int64Value,
		UInt64Value: p.UInt64Value,
		Int32Value:  p.Int32Value,
		UInt32Value: p.UInt32Value,
		BoolValue:   p.BoolValue,
		StringValue: p.StringValue,
		BytesValue:  p.BytesValue,
	}
	dst, ok := typeDef[pureTypeName]
	if !ok {
		return nil, false
	}
	return dst, true
}

func (p *PrimitiveTypeDesc) Add(msgDescs ...*desc.MessageDescriptor) error {
	typeNames := map[string]**desc.MessageDescriptor{
		DoubleValueFQN: &p.DoubleValue,
		FloatValueFQN:  &p.FloatValue,
		Int64ValueFQN:  &p.Int64Value,
		UInt64ValueFQN: &p.UInt64Value,
		Int32ValueFQN:  &p.Int32Value,
		UInt32ValueFQN: &p.UInt32Value,
		BoolValueFQN:   &p.BoolValue,
		StringValueFQN: &p.StringValue,
		BytesValueFQN:  &p.BytesValue,
	}
	for _, msgDesc := range msgDescs {
		dstField, ok := typeNames[msgDesc.GetFullyQualifiedName()]
		if !ok {
			continue
		}
		*dstField = msgDesc
	}
	return nil
}

func (p *PrimitiveTypeDesc) Fulfill() bool {
	if p.DoubleValue == nil {
		return false
	}
	if p.FloatValue == nil {
		return false
	}
	if p.Int64Value == nil {
		return false
	}
	if p.UInt64Value == nil {
		return false
	}
	if p.Int32Value == nil {
		return false
	}
	if p.UInt32Value == nil {
		return false
	}
	if p.BoolValue == nil {
		return false
	}
	if p.StringValue == nil {
		return false
	}
	if p.BytesValue == nil {
		return false
	}
	return true
}

func (e *ExtensionDesc) Add(exts ...*desc.FieldDescriptor) error {
	for _, ext := range exts {
		if !ext.IsExtension() {
			return errors.Errorf("given field is not an extension: %s", ext.GetFullyQualifiedName())
		}
		switch ext.GetNumber() {
		case E_Preference_Field:
			e.FileOptions.Preference = ext
		case E_StorageDriver_Field:
			e.FileOptions.StorageDriver = ext
		case E_Disabled_Field:
			e.FileOptions.Disabled = ext
		case E_Feature_Field:
			e.FileOptions.Feature = ext
		case E_Default_Field:
			e.FieldOptions.Default = ext
		case E_ABTest_Field:
			e.FieldOptions.ABTestFlagValue = ext
		case E_Tus_Field:
			e.FieldOptions.TusValue = ext
		case E_Multiple_Tus_Field:
			e.FieldOptions.TusValues = ext
		case E_Refenum_Field:
			e.FieldOptions.Refenum = ext
		default:
			log.Warn("Unrecognized extension: %q", ext.GetFullyQualifiedName())
		}
	}
	return nil
}

func asStringSlice(in interface{}) ([]string, bool) {
	slice, ok := in.([]interface{})
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(slice))
	for _, v := range slice {
		str, ok := v.(string)
		if !ok {
			return nil, false
		}
		out = append(out, str)
	}
	return out, true
}

func (e *ExtensionDesc) FileOptionsPreference(in *dynamic.Message) ([]string, error) {
	fieldV, err := in.TryGetField(e.FileOptions.Preference)
	if err != nil {
		return nil, err
	}
	out, ok := asStringSlice(fieldV)
	if !ok {
		return nil, errors.Errorf("Unexpected field value type: %+v", fieldV)
	}
	return out, nil
}

func (e *ExtensionDesc) FieldOptionsDefaultValue(in *desc.FieldDescriptor) (*dynamic.Message, error) {
	dm, err := dynamic.AsDynamicMessageWithExtensionRegistry(in.GetFieldOptions(), defaultExtensionRegistry)
	if err != nil {
		return nil, err
	}
	fieldV, err := dm.TryGetField(e.FieldOptions.Default)
	if err != nil {
		return nil, err
	}
	dm, ok := fieldV.(*dynamic.Message)
	if !ok {
		return nil, errors.Errorf("Unexpected field value type: %T", fieldV)
	}
	return dm, nil
}

func (e *ExtensionDesc) FieldOptionsABTestFlagValue(in *desc.FieldDescriptor) (string, error) {
	dm, err := dynamic.AsDynamicMessageWithExtensionRegistry(in.GetFieldOptions(), defaultExtensionRegistry)
	if err != nil {
		return "", err
	}
	fieldV, err := dm.TryGetField(e.FieldOptions.ABTestFlagValue)
	if err != nil {
		return "", err
	}
	fieldVS, ok := fieldV.(string)
	if !ok {
		return "", errors.Errorf("Unexpected field value type: %T", fieldV)
	}
	return fieldVS, nil
}

func (e *ExtensionDesc) FieldOptionsTusValue(in *desc.FieldDescriptor) (string, error) {
	dm, err := dynamic.AsDynamicMessageWithExtensionRegistry(in.GetFieldOptions(), defaultExtensionRegistry)
	if err != nil {
		return "", err
	}
	fieldV, err := dm.TryGetField(e.FieldOptions.TusValue)
	if err != nil {
		return "", err
	}
	fieldVS, ok := fieldV.(string)
	if !ok {
		return "", errors.Errorf("Unexpected field value type: %T", fieldV)
	}
	return fieldVS, nil
}

func (e *ExtensionDesc) FieldOptionsRefenum(in *desc.FieldDescriptor) (string, error) {
	dm, err := dynamic.AsDynamicMessageWithExtensionRegistry(in.GetFieldOptions(), defaultExtensionRegistry)
	if err != nil {
		return "", err
	}
	fieldV, err := dm.TryGetField(e.FieldOptions.Refenum)
	if err != nil {
		return "", err
	}
	fieldVS, ok := fieldV.(string)
	if !ok {
		return "", errors.Errorf("Unexpected field value type: %T", fieldV)
	}
	return fieldVS, nil
}

func (e *ExtensionDesc) FieldOptionsTusValues(in *desc.FieldDescriptor) ([]string, error) {
	dm, err := dynamic.AsDynamicMessageWithExtensionRegistry(in.GetFieldOptions(), defaultExtensionRegistry)
	if err != nil {
		return nil, err
	}
	fieldV, err := dm.TryGetField(e.FieldOptions.TusValues)
	if err != nil {
		return nil, err
	}
	out, ok := asStringSlice(fieldV)
	if !ok {
		return nil, errors.Errorf("Unexpected field value type: %+v", out)
	}
	return out, nil
}

func (e *ExtensionDesc) FileOptionsStorageDriver(in *dynamic.Message) (string, error) {
	fieldV, err := in.TryGetField(e.FileOptions.StorageDriver)
	if err != nil {
		return "", err
	}
	out, ok := fieldV.(string)
	if !ok {
		return "", errors.Errorf("Unexpected field value type: %+v", out)
	}
	if out == "" {
		out = distributionconst.DefaultStorageDriver
	}
	return out, nil
}

func (e *ExtensionDesc) FileOptionsDisabled(in *dynamic.Message) (bool, error) {
	fieldV, err := in.TryGetField(e.FileOptions.Disabled)
	if err != nil {
		return false, err
	}
	out, ok := fieldV.(bool)
	if !ok {
		return false, errors.Errorf("Unexpected field value type: %+v", out)
	}
	return out, nil
}

func (e *ExtensionDesc) FileOptionsFeature(in *dynamic.Message) ([]string, error) {
	fieldV, err := in.TryGetField(e.FileOptions.Feature)
	if err != nil {
		return nil, err
	}
	out, ok := asStringSlice(fieldV)
	if !ok {
		return nil, errors.Errorf("Unexpected field value type: %+v", out)
	}
	return out, nil
}

func (e *ExtensionDesc) Fulfill() bool {
	//nolint:gosimple
	if e.FileOptions.Preference == nil {
		return false
	}
	return true
}

type PreferenceRegistry struct {
	preferenceStore map[string]*PreferenceMeta
	messageStore    map[string]*desc.MessageDescriptor
}

func (pr *PreferenceRegistry) TryGetMessage(msgFullName string) (*desc.MessageDescriptor, bool) {
	msgDesc, ok := pr.messageStore[msgFullName]
	return msgDesc, ok
}

func (pr *PreferenceRegistry) ALLPreference() map[string]*PreferenceMeta {
	out := map[string]*PreferenceMeta{}
	for k, v := range pr.preferenceStore {
		dup := *v
		out[k] = &dup
	}
	return out
}

func (pr *PreferenceRegistry) TryGetPreference(msgFullName string) (*PreferenceMeta, bool) {
	msgDesc, ok := pr.preferenceStore[msgFullName]
	return msgDesc, ok
}

func ensureInitialized() {
	if atomic.LoadInt64(&initialized) <= 0 {
		panic(errors.Errorf("Global registry is not initialized"))
	}
}

func TryGetMessage(msgFullName string) (*desc.MessageDescriptor, bool) {
	ensureInitialized()
	return GlobalRegistry.TryGetMessage(msgFullName)
}

func ALLPreference() map[string]*PreferenceMeta {
	ensureInitialized()
	return GlobalRegistry.ALLPreference()
}

func TryGetPreference(msgFullName string) (*PreferenceMeta, bool) {
	ensureInitialized()
	return GlobalRegistry.TryGetPreference(msgFullName)
}

func InitPreferenceRegistry(in *ProtoStore) error {
	firstInit := atomic.CompareAndSwapInt64(&initialized, 0, 1)
	if !firstInit {
		return errors.Errorf("Preference registry is already initialized")
	}

	in.ALLProtos.Iter(func(bp *BAPIProto) bool {
		exts := bp.FileDescriptor.GetExtensions()
		if len(exts) <= 0 {
			return true
		}
		if err := DefaultDistributionExtensionDesc.Add(exts...); err != nil {
			log.Warn("Failed to register distribution extension: %+v: %+v", exts, err)
			return true
		}
		if err := defaultExtensionRegistry.AddExtension(exts...); err != nil {
			log.Warn("Failed to register extension: %+v: %+v", exts, err)
			return true
		}
		return true
	})
	if !DefaultDistributionExtensionDesc.Fulfill() {
		return errors.Errorf("unsatisfied distribution proto extension desc: %+v", DefaultDistributionExtensionDesc)
	}

	primitiveTypeProto, ok := in.ALLProtos.Find(func(bp *BAPIProto) bool {
		return bp.FileDescriptor.GetFullyQualifiedName() == "bilibili/app/distribution/distribution.proto"
	})
	if !ok {
		return errors.Errorf("Failed to find primitive type proto")
	}
	if err := DistributionPrimitiveType.Add(primitiveTypeProto.FileDescriptor.GetMessageTypes()...); err != nil {
		return err
	}
	if !DistributionPrimitiveType.Fulfill() {
		return errors.Errorf("unsatisfied distribution primitive type: %+v", DistributionPrimitiveType)
	}

	registry := &PreferenceRegistry{
		preferenceStore: map[string]*PreferenceMeta{},
		messageStore:    map[string]*desc.MessageDescriptor{},
	}
	in.ALLProtos.Iter(func(bp *BAPIProto) bool {
		fileDesc := bp.FileDescriptor
		dm, err := dynamic.AsDynamicMessageWithExtensionRegistry(fileDesc.GetFileOptions(), defaultExtensionRegistry)
		if err != nil {
			log.Warn("Failed to cast file option as dynamic message: %+v", err)
			return true
		}
		preferenceMessageNames, err := DefaultDistributionExtensionDesc.FileOptionsPreference(dm)
		if err != nil {
			log.Error("Failed to get file preference type: %+v", err)
			return true
		}
		storageDriver, err := DefaultDistributionExtensionDesc.FileOptionsStorageDriver(dm)
		if err != nil {
			log.Error("Failed to get file preference storage driver: %+v", err)
			return true
		}
		disabled, err := DefaultDistributionExtensionDesc.FileOptionsDisabled(dm)
		if err != nil {
			log.Error("Failed to get file preference is disabled: %+v", err)
			return true
		}
		feature, err := DefaultDistributionExtensionDesc.FileOptionsFeature(dm)
		if err != nil {
			log.Error("Failed to get file preference feature: %+v", err)
			return true
		}
		for _, p := range preferenceMessageNames {
			msgFullName := fmt.Sprintf("%s.%s", bp.FileDescriptor.GetPackage(), p)
			msgDesc := bp.FileDescriptor.FindMessage(msgFullName)
			if msgDesc == nil {
				log.Error("Failed to find file preference type descriptor: %q", msgFullName)
				continue
			}
			if _, ok := registry.preferenceStore[msgDesc.GetFullyQualifiedName()]; ok {
				panic(errors.Errorf("Duplicate prederence: %q", msgDesc.GetFullyQualifiedName()))
			}
			preferenceMeta := &PreferenceMeta{
				ProtoDesc:     msgDesc,
				storageDriver: storageDriver,
				disabled:      disabled,
				feature:       feature,
				preference:    p,
			}
			registry.preferenceStore[msgDesc.GetFullyQualifiedName()] = preferenceMeta
		}
		return true
	})
	in.ALLProtos.Iter(func(bp *BAPIProto) bool {
		fileDesc := bp.FileDescriptor
		for _, msgDesc := range fileDesc.GetMessageTypes() {
			if _, ok := registry.messageStore[msgDesc.GetFullyQualifiedName()]; ok {
				panic(errors.Errorf("Duplicate message: %q", msgDesc.GetFullyQualifiedName()))
			}
			registry.messageStore[msgDesc.GetFullyQualifiedName()] = msgDesc
		}
		return true
	})

	GlobalRegistry = registry
	return nil
}
