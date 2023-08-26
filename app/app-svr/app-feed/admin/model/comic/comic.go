package comic

// Comic .
type Comic struct {
	Title string
}

// Comic .
type ComicInfo struct {
	Title         string
	VerticalCover string `json:"vertical_cover"`
}

// Comics .
type Comics struct {
	Code int
	Msg  string
	Data *Comic
}

// Comics .
type ComicInfos struct {
	Code int
	Data []*ComicInfo
}
