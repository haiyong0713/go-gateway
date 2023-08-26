package column

type Item struct {
	Id         int64  `json:"id"`
	Wid        int64  `json:"wid"`
	Likes      int64  `json:"likes"`
	TotalLikes int64  `json:"total_likes"`
	Title      string `json:"title"`
}

type Items []*Item

func (s Items) Len() int {
	return len(s)
}

func (s Items) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Items) Less(i, j int) bool {
	return s[i].TotalLikes > s[j].TotalLikes
}
