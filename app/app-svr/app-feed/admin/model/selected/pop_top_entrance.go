package selected

type PopTopEntrance struct {
	ID   int   `json:"id"`
	Rank int64 `json:"rank"`
}

func (e *PopTopEntrance) TableName() string {
	return "popular_top_entrance"
}
