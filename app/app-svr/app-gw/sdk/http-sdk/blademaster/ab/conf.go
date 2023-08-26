package ab

// Group represents exp group in public protocol.
type Group struct {
	// group id
	ID int64
	// buckets allocated by exp conf server
	Buckets []uint16
	// flags defined in this exp group
	Flags map[string]string
	// env variable -> val1, val2...
	// requests which match any whitelist would freeze this exp group directly, ignoring normal diversion process
	Whitelist map[string][]string
}

// Meta data of both domain and exp.
type Meta struct {
	// domain id or exp id
	ID int64
	// buckets allocated by exp conf server
	Buckets []uint16
	// accept only requests matching this condition
	Condition string
	// conf version. each time conf changes, version updates.
	Version string
}

// Exp represents exp in public protocol. Exp contains groups.
type Exp struct {
	Meta
	// control group and treatment group 1,2,3...
	Groups []*Group
}

// Layer represents layer in public protocol. Both domain and exp can be contained in layers.
type Layer struct {
	// layer id
	ID int64
	// env variable as diversion
	Diversion string
	// whether this layer is a launch layer
	Launched bool
	// contained exps
	Exps []*Exp
	// contained sub-domains
	Domains []*Domain
}

// Domain represents domain in public protocol. Multiple layers can be contained in domains.
type Domain struct {
	Meta
	// contained layers
	Layers []*Layer
}
