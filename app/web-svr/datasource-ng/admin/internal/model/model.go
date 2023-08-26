package model

import (
	"go-gateway/app/web-svr/datasource-ng/admin/api"
	"regexp"

	"github.com/pkg/errors"

	"github.com/google/uuid"
)

const (
	MdlFieldTypeObject   = "object"
	MdlFieldTypeInt      = "int"
	MdlFieldTypeString   = "string"
	MdlFieldTypeBool     = "bool"
	TypeGeneric          = "List<%s>"
	CptTypeText          = "input"
	JsonSchemaTypeInt    = "integer"
	JsonSchemaTypeString = "string"
	JsonSchemaTypeBool   = "boolean"
	JsonSchemaTypeObject = "object"
	JsonSchemaTypeArray  = "array"
	JsonSchemaRef        = "#/definitions/%s"
)

var (
	ValueTypeListPattern  = regexp.MustCompile(`^List\<(.*)\>$`)
	ReflectJsonSchemaType = map[string]string{
		MdlFieldTypeString: JsonSchemaTypeString,
		MdlFieldTypeInt:    JsonSchemaTypeInt,
		MdlFieldTypeBool:   JsonSchemaTypeBool,
		MdlFieldTypeObject: JsonSchemaTypeObject,
	}
)

// IsGeneric is
func IsGeneric(typeName string) bool {
	return ValueTypeListPattern.MatchString(typeName)
}

// IsReference is
func IsReference(typeName string) bool {
	return (typeName != MdlFieldTypeInt) &&
		(typeName != MdlFieldTypeString) &&
		(typeName != MdlFieldTypeBool)
}

func TranBool(value int) bool {
	if value == 0 {
		return false
	}
	return true
}

func BoolToInt64(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func UUID4() string {
	return uuid.New().String()
}

func DefaultSchemaComponent() *api.SchemaComponent {
	return &api.SchemaComponent{
		Type: "input",
	}
}

func NewModelComponent() *api.ModelComponent {
	return &api.ModelComponent{
		ComponentUuid: UUID4(),
		Type:          "input",
	}
}

type JsonSchema struct {
	Type        string                 `json:"type,omitempty"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]*JsonSchema `json:"properties,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Component   *api.ModelComponent    `json:"component,omitempty"`
	Items       *JsonSchemaItems       `json:"items,omitempty"`
	MinItems    int                    `json:"minItems,omitempty"`
	Definitions map[string]*JsonSchema `json:"definitions,omitempty"`
}

type JsonSchemaItems struct {
	Type string `json:"type,omitempty"`
	Ref  string `json:"$ref,omitempty"`
}

func SplitGenericType(valueType string) (string, error) {
	if ValueTypeListPattern.MatchString(valueType) {
		return ValueTypeListPattern.FindStringSubmatch(valueType)[1], nil
	}
	return "", errors.Errorf("Invalid generic value type: %s", valueType)
}

type GetItemValueArgsParams struct {
	ModelName      string
	Business       string
	Expirable      int32
	ExpireAt       int64
	Values         map[string]interface{}
	ModelFieldsRes map[string][]*api.ModelField
	ComponentRes   map[string]*api.ModelComponent
}
