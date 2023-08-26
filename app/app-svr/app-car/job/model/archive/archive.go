package archive

type Message struct {
	Action string  `json:"action"`
	Table  string  `json:"table"`
	Old    *ArcMsg `json:"old"`
	New    *ArcMsg `json:"new"`
}

// ArcMsg is
type ArcMsg struct {
	Aid       int64  `json:"aid"`
	State     int64  `json:"state"`
	Attribute int64  `json:"attribute"`
	Action    string `json:"action"`
}

// AttrVal get attr val by bit.
func (a *ArcMsg) AttrVal(bit uint) int32 {
	return int32((a.Attribute >> bit) & int64(1))
}
