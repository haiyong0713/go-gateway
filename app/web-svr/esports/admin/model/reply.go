package model

// Reply .
type Reply struct {
	ID       int64
	Business int
}

// TableName .
func (t Reply) TableName() string {
	return "es_reply"
}
