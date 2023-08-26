package ab

type itemType int32

const (
	typeNil itemType = iota
	typeDomain
	typeExp
	typeGroup
)

var emptyItem = &item{
	Type: typeNil,
}

type item struct {
	Type   itemType
	Domain *domain
	Exp    *exp
	Group  *group
}

// Cond returns condition of current item. Nil if none.
func (i *item) Cond() (c Condition) {
	switch i.Type {
	case typeDomain:
		c = i.Domain.cond
	case typeExp:
		c = i.Exp.cond
	default:
	}
	return
}
