package ab

import (
	"strconv"
	"strings"
)

type varType int32

const (
	//nolint:deadcode,varcheck
	typeUnknown varType = iota
	typeString
	typeInt64
	typeFloat64
	typeBool
	typeVersion
)

const versionSeparator = "."

// Version is detailed struct for semantic versions. One can check whether one version is equal to or greater than another.
type Version struct {
	raw string
	// digits separated by dots
	digits []int
}

func newVersion(s string) (*Version, error) {
	parts := strings.Split(s, versionSeparator)
	digits := make([]int, len(parts))
	for index, part := range parts {
		ipart, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		digits[index] = ipart
	}
	return &Version{
		raw:    s,
		digits: digits,
	}, nil
}

func (v *Version) String() string {
	return v.raw
}

func (v *Version) ge(otherV *Version) bool {
	if otherV == nil {
		return true
	}
	for i := 0; i < len(otherV.digits); i++ {
		if i >= len(v.digits) {
			return false
		}
		if v.digits[i] > otherV.digits[i] {
			return true
		} else if v.digits[i] < otherV.digits[i] {
			return false
		}
	}
	return true
}

func (v *Version) eq(otherV *Version) bool {
	if v == otherV {
		return true
	}
	if otherV == nil {
		return false
	}
	if len(v.digits) != len(otherV.digits) {
		return false
	}
	for index, digit := range v.digits {
		if otherV.digits[index] != digit {
			return false
		}
	}
	return true
}

// KV holds values of various types.
type KV struct {
	Key     string
	Type    varType
	String  string
	Int64   int64
	Float64 float64
	Bool    bool
	Version *Version
}

// KVString construct KV with string value.
func KVString(key string, value string) KV {
	return KV{
		Key:    key,
		Type:   typeString,
		String: value,
	}
}

// KVInt construct KV with int64 value.
func KVInt(key string, value int64) KV {
	return KV{
		Key:   key,
		Type:  typeInt64,
		Int64: value,
	}
}

// KVFloat construct KV with float64 value.
func KVFloat(key string, value float64) KV {
	return KV{
		Key:     key,
		Type:    typeFloat64,
		Float64: value,
	}
}

// KVBool construct KV with bool value.
func KVBool(key string, value bool) KV {
	return KV{
		Key:  key,
		Type: typeBool,
		Bool: value,
	}
}

// KVVersion construct KV with version value.
func KVVersion(key string, value *Version) KV {
	return KV{
		Key:     key,
		Type:    typeVersion,
		Version: value,
	}
}

// Value extracts value from KV.
func (kv KV) Value() interface{} {
	switch kv.Type {
	case typeString:
		return kv.String
	case typeInt64:
		return kv.Int64
	case typeFloat64:
		return kv.Float64
	case typeBool:
		return kv.Bool
	case typeVersion:
		return kv.Version
	default:
	}
	return nil
}

// Clone returns a deep copy of KV.
func (kv KV) Clone() (nkv KV) {
	nkv = kv
	v := new(Version)
	*v = *kv.Version
	nkv.Version = v
	return
}

//nolint:unparam
func parseKV(v KV, val string) (value KV, err error) {
	switch v.Type {
	case typeString:
		value = KVString(v.Key, val)
	case typeInt64:
		ival, err := strconv.Atoi(val)
		if err == nil {
			value = KVInt(v.Key, int64(ival))
		}
	case typeFloat64:
		fval, err := strconv.ParseFloat(val, 64)
		if err == nil {
			value = KVFloat(v.Key, fval)
		}
	case typeBool:
		bval, err := strconv.ParseBool(val)
		if err == nil {
			value = KVBool(v.Key, bval)
		}
	case typeVersion:
		ver, err := newVersion(val)
		if err == nil {
			value = KVVersion(v.Key, ver)
		}
	default:
	}
	return
}

type envVar struct {
	index uint
	kv    KV
}
