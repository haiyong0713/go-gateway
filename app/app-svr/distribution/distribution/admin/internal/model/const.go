package model

const (
	ValidateKeyLen      = 3
	DefaultTusValue     = "default"
	DefaultTusValueName = "默认值"
	ActionSave          = "save"
)

type FieldBasicInfo struct {
	Name    string  `json:"name"`
	Chinese string  `json:"chinese"`
	Type    string  `json:"type"`
	Enum    []int64 `json:"enum"`
}
