package like

type SteinList struct {
	AwardOne int64 `json:"award_one"`
	AwardTwo int64 `json:"award_two"`
}

type SteinWebData struct {
	List []*struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Data *struct {
			Aids string `json:"aids"`
			Name string `json:"name"`
		}
	}
}

type SteinMemData struct {
	Name string
	Aids []int64
}

type SteinData struct {
	Name string       `json:"name"`
	List []*SimpleArc `json:"list"`
}
