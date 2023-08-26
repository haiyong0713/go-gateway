package ab

type flag struct {
	name  string
	desc  string
	index uint
	val   KV
}

func (f *flag) init(name, desc string, val KV) {
	f.name = name
	f.desc = desc
	f.val = val
	Registry.RegisterFlag(f)
	//nolint:gosimple
	return
}

// Value returns value of flag based on T.
func (f *flag) Value(t *T) *KV {
	return t.value(f)
}

// BoolFlag is a flag of bool value
type BoolFlag struct {
	flag
}

// Bool constructs a BoolFlag.
func Bool(name, desc string, defaultValue bool) (b *BoolFlag) {
	b = &BoolFlag{}
	b.init(name, desc, KVBool(name, defaultValue))
	return
}

// BoolVar populates BoolFlag parameter.
func BoolVar(f *BoolFlag, name, desc string, defaultValue bool) {
	f.init(name, desc, KVBool(name, defaultValue))
}

// Value returns bool value of BoolFlag.
func (f *BoolFlag) Value(t *T) bool {
	val := f.flag.Value(t)
	if val == nil {
		return false
	}
	return val.Bool
}

// IntFlag is a flag of int64 value
type IntFlag struct {
	flag
}

// Int constructs a IntFlag.
func Int(name, desc string, defaultValue int64) (i *IntFlag) {
	i = &IntFlag{}
	i.init(name, desc, KVInt(name, defaultValue))
	return
}

// IntVar populates IntFlag parameter.
func IntVar(f *IntFlag, name, desc string, defaultValue int64) {
	f.init(name, desc, KVInt(name, defaultValue))
}

// Value returns int64 value of IntFlag.
func (f *IntFlag) Value(t *T) int64 {
	val := f.flag.Value(t)
	if val == nil {
		return 0
	}
	return val.Int64
}

// FloatFlag is a flag of float64 value
type FloatFlag struct {
	flag
}

// Float constructs a FloatFlag.
func Float(name, desc string, defaultValue float64) (f *FloatFlag) {
	f = &FloatFlag{}
	f.init(name, desc, KVFloat(name, defaultValue))
	return
}

// FloatVar populates FloatFlag parameter.
func FloatVar(f *FloatFlag, name, desc string, defaultValue float64) {
	f.init(name, desc, KVFloat(name, defaultValue))
}

// Value returns float64 value of FloatFlag.
func (f *FloatFlag) Value(t *T) float64 {
	val := f.flag.Value(t)
	if val == nil {
		return 0
	}
	return val.Float64
}

// StringFlag is a flag of string value
type StringFlag struct {
	flag
}

// String constructs a StringFlag.
func String(name, desc string, defaultValue string) (s *StringFlag) {
	s = &StringFlag{}
	s.init(name, desc, KVString(name, defaultValue))
	return
}

// StringVar populates StringFlag parameter.
func StringVar(f *StringFlag, name, desc string, defaultValue string) {
	f.init(name, desc, KVString(name, defaultValue))
}

// Value returns string value of StringFlag.
func (f *StringFlag) Value(t *T) string {
	val := f.flag.Value(t)
	if val == nil {
		return ""
	}
	return val.String
}
