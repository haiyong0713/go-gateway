package rank

type List struct {
	Aid    int64   `json:"aid"`
	Score  int64   `json:"score"`
	Others []*List `json:"others"`
}
