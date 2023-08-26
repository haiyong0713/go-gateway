package like

type ArcType struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Rank        int64      `json:"rank"`
	Children    []*ArcType `json:"children"`
}
