package model

import (
	"regexp"
	"time"
)

// consts
const (
	ValueTypeString = "string"
	ValueTypeInt    = "int"
	ValueTypeBool   = "bool"
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
	return (typeName != ValueTypeString) &&
		(typeName != ValueTypeInt) &&
		(typeName != ValueTypeBool)
}

// Expired is
func (mi *ModelItem) Expired(at time.Time) bool {
	return mi.Expirable && (at.Unix() > mi.ExpireAt)
}
