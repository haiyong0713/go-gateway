package model

import (
	"fmt"
	"regexp"

	"go-common/library/time"

	"github.com/pkg/errors"
)

const (
	MdlFieldTypeObject = "object"
	MdlFieldTypeInt    = "int"
	MdlFieldTypeString = "string"
	MdlFieldTypeBool   = "bool"
	TypeGeneric        = "List<%s>"
	RuleTypePrimary    = "primary"
	RuleTypeJavascript = "javascript"
	RuleTypeLiteral    = "literal"

	RoleAdmin = "admin"
)

var (
	ValueTypeListPattern = regexp.MustCompile(`^List\<(.*)\>$`)
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

func SplitGenericType(valueType string) (string, error) {
	if ValueTypeListPattern.MatchString(valueType) {
		return ValueTypeListPattern.FindStringSubmatch(valueType)[1], nil
	}
	return "", errors.Errorf("Invalid generic value type: %s", valueType)
}

func FieldRuleKey(modelName, fieldName, datasourceAPI string) string {
	return fmt.Sprintf("%s.%s.%s", modelName, fieldName, datasourceAPI)
}

type ItemFieldRule struct {
	Id            int64     `json:"id,omitempty"`
	ModelName     string    `json:"model_name,omitempty"`
	FieldName     string    `json:"field_name,omitempty"`
	DatasourceApi string    `json:"datasource_api,omitempty"`
	ExternalRule  string    `json:"external_rule,omitempty"`
	RuleType      string    `json:"rule_type,omitempty"`
	ValueSource   string    `json:"value_source,omitempty"`
	Ctime         time.Time `json:"ctime,omitempty"`
	IsDeleted     int32     `json:"is_deleted,omitempty"`
}

type RoleContext struct {
	TreeID   int64  `form:"tree_id" validate:"required"`
	Cookie   string `form:"-"`
	Username string `form:"-"`
	Role     string `form:"-"`
}
