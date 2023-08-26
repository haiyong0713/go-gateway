package tool

import (
	"encoding/json"
	"strings"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"

	"go-common/library/ecode"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

// 通过proto的preference和filed中特定的option获取message descriptor
func FindMessageDescriptor(preference, targetOptionValue string, getOptionValue func(in *desc.FieldDescriptor) (string, error)) (*desc.MessageDescriptor, error) {
	var (
		preferenceMeta *preferenceproto.PreferenceMeta
	)
	for _, meta := range preferenceproto.ALLPreference() {
		if meta.Preference() == preference {
			preferenceMeta = meta
			break
		}
	}
	if preferenceMeta == nil {
		return nil, errors.Errorf("Failed to Find preference meta by preference(%s)", preference)
	}
	var expectedFiled *desc.FieldDescriptor
	for _, filed := range preferenceMeta.ProtoDesc.GetFields() {
		optionValue, err := getOptionValue(filed)
		if err != nil {
			return nil, err
		}
		if optionValue == targetOptionValue {
			expectedFiled = filed
			break
		}
	}
	if expectedFiled == nil {
		return nil, errors.Errorf("Failed to Find filed descriptor by filed value(%s)", targetOptionValue)
	}
	expectedMessageDescriptor := expectedFiled.GetMessageType()
	if expectedMessageDescriptor == nil {
		return nil, errors.Errorf("Failed to convert filed descriptor to message descriptor(%v)", expectedFiled)
	}
	return expectedMessageDescriptor, nil
}

func MessageDescriptorToJson(dm *desc.MessageDescriptor, rawData []byte) (json.RawMessage, error) {
	ctr := dynamic.NewMessage(dm)
	if err := func() error {
		if len(rawData) == 0 {
			return nil
		}
		if err := ctr.UnmarshalJSONPB(&jsonpb.Unmarshaler{AllowUnknownFields: true}, rawData); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal raw data to dynamic message")
	}
	descMarshaler := jsonpb.Marshaler{
		EmitDefaults: true,
		OrigName:     true,
		Indent:       " ",
	}
	descString, err := descMarshaler.MarshalToString(ctr)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal message data to json")
	}
	return json.RawMessage(descString), nil
}

func FieldBasicInfo(in *desc.MessageDescriptor) []*model.FieldBasicInfo {
	var fieldBasicInfos []*model.FieldBasicInfo
	for _, v := range in.GetFields() {
		if v.GetSourceInfo() == nil {
			continue
		}
		fieldBasicInfo := &model.FieldBasicInfo{
			Name: v.GetName(),
			Type: castDistributionTypeAsJsonType(v.GetMessageType().GetName()),
		}
		if v.GetSourceInfo().LeadingComments != nil {
			fieldBasicInfo.Chinese = RemoveCRLF(v.GetSourceInfo().GetLeadingComments())
		}
		if v.GetSourceInfo().TrailingComments != nil {
			fieldBasicInfo.Chinese = RemoveCRLF(v.GetSourceInfo().GetTrailingComments())
		}
		fieldBasicInfo.Enum = GetRefNum(in, v)
		fieldBasicInfos = append(fieldBasicInfos, fieldBasicInfo)
	}
	return fieldBasicInfos
}

func GetRefNum(dm *desc.MessageDescriptor, fdm *desc.FieldDescriptor) []int64 {
	refenum, err := preferenceproto.DefaultDistributionExtensionDesc.FieldOptionsRefenum(fdm)
	if err != nil && refenum == "" {
		return nil
	}
	var enum []int64
	for _, nestedEnum := range dm.GetNestedEnumTypes() {
		if nestedEnum.GetName() != refenum {
			continue
		}
		for _, value := range nestedEnum.GetValues() {
			enum = append(enum, int64(value.GetNumber()))
		}
	}
	return enum
}

func RemoveCRLF(in string) string {
	return strings.ReplaceAll(in, "\n", "")
}

func GetFiledOptionValuesFromPreferenceproto(msgFullName string, parseFiledOptionValue func(in *desc.FieldDescriptor) (string, error)) ([]string, error) {
	meta, ok := preferenceproto.TryGetPreference(msgFullName)
	if !ok {
		return nil, errors.Wrapf(ecode.NothingFound, "Failed to fetch proto meta from %s", msgFullName)
	}
	var filedOptionValues []string
	for _, v := range meta.ProtoDesc.GetFields() {
		value, err := parseFiledOptionValue(v)
		if err != nil {
			return nil, err
		}
		filedOptionValues = append(filedOptionValues, value)
	}
	return filedOptionValues, nil
}

func castDistributionTypeAsJsonType(in string) string {
	switch in {
	case preferenceproto.DoubleValue:
		return "number"
	case preferenceproto.Int32Value, preferenceproto.Int64Value:
		return "integer"
	case preferenceproto.BoolValue:
		return "boolean"
	case preferenceproto.StringValue:
		return "string"
	default:
		return ""
	}
}
