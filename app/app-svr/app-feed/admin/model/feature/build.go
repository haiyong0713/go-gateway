package feature

const (
	OpLt = "lt"
	OpLe = "le"
	OpGt = "gt"
	OpGe = "ge"
	OpEq = "eq"
	OpNe = "ne"
)

var OpList = map[string]struct{}{
	OpLt: {},
	OpLe: {},
	OpGt: {},
	OpGe: {},
	OpEq: {},
	OpNe: {},
}
